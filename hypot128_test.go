package floats

import (
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/shogo82148/ints"
)

func TestHypot128(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float128
		y    Float128
		want string
	}{
		{exact128(3), exact128(4), "5"},
		{exact128(5), exact128(12), "13"},
		{exact128(1), exact128(1), "1.4142135623730950488016887242096980786"}, // sqrt(2)
	}

	for _, tt := range tests {
		got := Hypot128(tt.x, tt.y)
		if !close128(got, tt.want) {
			t.Errorf("Hypot128(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		y    Float128
		want Float128
	}{
		// special cases
		{exact128(0), exact128(0), exact128(0)},
		{exact128(math.Inf(1)), exact128(1), exact128(math.Inf(1))},
		{exact128(math.Inf(-1)), exact128(1), exact128(math.Inf(1))},
		{exact128(1), exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(1), exact128(math.Inf(-1)), exact128(math.Inf(1))},
		{exact128(math.NaN()), exact128(1), exact128(math.NaN())},
		{exact128(1), exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := Hypot128(tt.x, tt.y)
		if !eq128(got, tt.want) {
			t.Errorf("Hypot128(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

// bigFromFloat128 converts a finite Float128 to an exact big.Float.
func bigFromFloat128(a Float128) *big.Float {
	bits := a.Bits()
	e := int(bits[0]>>48) & mask128
	m := new(big.Int).SetUint64(bits[0] & 0x0000_ffff_ffff_ffff)
	m.Lsh(m, 64).Or(m, new(big.Int).SetUint64(bits[1]))
	if e == 0 {
		e = 1
	} else {
		m.SetBit(m, 112, 1)
	}
	f := new(big.Float).SetPrec(200).SetInt(m)
	return f.SetMantExp(f, e-bias128-112)
}

func TestHypot128_Edge(t *testing.T) {
	t.Parallel()
	tiny := NewFloat128FromBits(ints.Uint128{0, 1})
	max := NewFloat128FromBits(ints.Uint128{0x7ffe_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff})
	if got := Hypot128(tiny, tiny); !eq128(got, tiny) {
		t.Errorf("Hypot128(tiny, tiny) = %v; want %v", got.Bits(), tiny.Bits())
	}
	if got := Hypot128(max, max); !got.IsInf(1) {
		t.Errorf("Hypot128(max, max) = %v; want +Inf", got)
	}
	if got := Hypot128(max, exact128(1)); !eq128(got, max) {
		t.Errorf("Hypot128(max, 1) = %v; want max", got)
	}
	if got := Hypot128(exact128(-3), exact128(0)); !eq128(got, exact128(3)) {
		t.Errorf("Hypot128(-3, 0) = %v; want 3", got)
	}
}

func TestHypot128_Reference(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(1))
	two := big.NewFloat(2)
	for range 20000 {
		// pick exponents close to each other most of the time
		p := NewFloat128FromBits(ints.Uint128{r.Uint64(), r.Uint64()})
		q := NewFloat128FromBits(ints.Uint128{r.Uint64(), r.Uint64()})
		if r.Intn(2) == 0 {
			// make q's exponent near p's
			qb := q.Bits()
			pb := p.Bits()
			qb[0] = qb[0]&^(0x7fff<<48) | (pb[0] & (0x7fff << 48))
			q = NewFloat128FromBits(qb)
		}
		// exercise subnormal and near-overflow exponents
		for _, v := range []*Float128{&p, &q} {
			switch r.Intn(8) {
			case 0:
				b := v.Bits()
				b[0] &^= 0x7fff << 48
				b[0] |= uint64(r.Intn(3)) << 48
				*v = NewFloat128FromBits(b)
			case 1:
				b := v.Bits()
				b[0] = b[0]&^(0x7fff<<48) | uint64(0x7ffe-r.Intn(3))<<48
				*v = NewFloat128FromBits(b)
			}
		}
		if p.IsNaN() || q.IsNaN() || p.IsInf(0) || q.IsInf(0) {
			continue
		}
		x, y := bigFromFloat128(p), bigFromFloat128(q)
		sum := new(big.Float).SetPrec(1000).Add(
			new(big.Float).SetPrec(1000).Mul(x, x),
			new(big.Float).SetPrec(1000).Mul(y, y),
		)
		got := Hypot128(p, q)
		if got.IsInf(0) {
			// must overflow: sum >= (max + ulp/2)^2
			max := bigFromFloat128(NewFloat128FromBits(ints.Uint128{0x7ffe_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}))
			if sum.Cmp(new(big.Float).SetPrec(1000).Mul(max, max)) <= 0 {
				t.Errorf("Hypot128(%v, %v) = Inf; want finite", p.Bits(), q.Bits())
			}
			continue
		}
		// got is correctly rounded iff (got-h)^2 <= sum <= (got+h)^2,
		// where h is half an ulp (the boundary is ignored; ties are measure zero).
		g := bigFromFloat128(got)
		ulp := new(big.Float).SetPrec(1000).Sub(bigFromFloat128(got.Nextafter(NewFloat128Inf(1))), g)
		h := ulp.Quo(ulp, two)
		lo := new(big.Float).SetPrec(1000).Sub(g, h)
		hi := new(big.Float).SetPrec(1000).Add(g, h)
		lo.Mul(lo, lo)
		hi.Mul(hi, hi)
		if g.Sign() == 0 || sum.Cmp(lo) < 0 || sum.Cmp(hi) > 0 {
			// at power-of-two boundaries the lower half-ulp is smaller; accept if the
			// neighbour below is not closer
			d := new(big.Float).SetPrec(1000).Sub(new(big.Float).SetPrec(1000).Sqrt(sum), g)
			d.Abs(d)
			if d.Cmp(h) > 0 {
				t.Errorf("Hypot128(%v, %v) = %v: not correctly rounded", p.Bits(), q.Bits(), got)
			}
		}
	}
}

func BenchmarkHypot128(b *testing.B) {
	p := exact128(3).Quo(exact128(7))
	q := exact128(5).Quo(exact128(11))
	for b.Loop() {
		Hypot128(p, q)
	}
}
