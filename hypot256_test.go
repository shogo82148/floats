package floats

import (
	"math"
	"math/big"
	"math/rand"
	"testing"

	"github.com/shogo82148/ints"
)

func TestHypot256(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		y    Float256
		want string
	}{
		{exact256(3), exact256(4), "5"},
		{exact256(5), exact256(12), "13"},
		{exact256(1), exact256(1), "1.414213562373095048801688724209698078569671875376948073176679737990733"},
	}

	for _, tt := range tests {
		got := Hypot256(tt.x, tt.y)
		if !close256(got, tt.want) {
			t.Errorf("Hypot256(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		y    Float256
		want Float256
	}{
		// special cases
		{exact256(0), exact256(0), exact256(0)},
		{exact256(math.Inf(1)), exact256(1), exact256(math.Inf(1))},
		{exact256(math.Inf(-1)), exact256(1), exact256(math.Inf(1))},
		{exact256(1), exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(1), exact256(math.Inf(-1)), exact256(math.Inf(1))},
		{exact256(math.NaN()), exact256(1), exact256(math.NaN())},
		{exact256(1), exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := Hypot256(tt.x, tt.y)
		if !eq256(got, tt.want) {
			t.Errorf("Hypot256(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

// bigFromFloat256 converts a finite Float256 to an exact big.Float.
func bigFromFloat256(a Float256) *big.Float {
	m := new(big.Int)
	for i := range 4 {
		w := a[i]
		if i == 0 {
			w &= 0x0000_0fff_ffff_ffff
		}
		m.Lsh(m, 64).Or(m, new(big.Int).SetUint64(w))
	}
	e := int(a[0]>>44) & mask256
	if e == 0 {
		e = 1
	} else {
		m.SetBit(m, shift256, 1)
	}
	f := new(big.Float).SetPrec(300).SetInt(m)
	return f.SetMantExp(f, e-bias256-shift256)
}

func TestHypot256_Edge(t *testing.T) {
	t.Parallel()
	tiny := NewFloat256FromBits(ints.Uint256{0, 0, 0, 1})
	max := NewFloat256FromBits(ints.Uint256{0x7fff_efff_ffff_ffff, ^uint64(0), ^uint64(0), ^uint64(0)})
	if got := Hypot256(tiny, tiny); !eq256(got, tiny) {
		t.Errorf("Hypot256(tiny, tiny) = %v; want %v", got.Bits(), tiny.Bits())
	}
	if got := Hypot256(max, max); !got.IsInf(1) {
		t.Errorf("Hypot256(max, max) = %v; want +Inf", got)
	}
	if got := Hypot256(max, exact256(1)); !eq256(got, max) {
		t.Errorf("Hypot256(max, 1) = %v; want max", got)
	}
	if got := Hypot256(exact256(-3), exact256(0)); !eq256(got, exact256(3)) {
		t.Errorf("Hypot256(-3, 0) = %v; want 3", got)
	}
}

func TestHypot256_Reference(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(1))
	two := big.NewFloat(2)
	const prec = 2000
	maxF := new(big.Float).SetPrec(prec).Set(bigFromFloat256(NewFloat256FromBits(ints.Uint256{0x7fff_efff_ffff_ffff, ^uint64(0), ^uint64(0), ^uint64(0)})))
	randFloat := func() Float256 {
		return NewFloat256FromBits(ints.Uint256{r.Uint64(), r.Uint64(), r.Uint64(), r.Uint64()})
	}
	setExp := func(v Float256, e uint64) Float256 {
		b := v.Bits()
		b[0] = b[0]&^(0x7_ffff<<44) | e<<44
		return NewFloat256FromBits(b)
	}
	for range 5000 {
		p, q := randFloat(), randFloat()
		if r.Intn(2) == 0 {
			// make q's exponent near p's
			q = setExp(q, (p.Bits()[0]>>44)&0x7_ffff)
		}
		// exercise subnormal and near-overflow exponents
		for _, v := range []*Float256{&p, &q} {
			switch r.Intn(8) {
			case 0:
				*v = setExp(*v, uint64(r.Intn(3)))
			case 1:
				*v = setExp(*v, uint64(0x7_fffe-r.Intn(3)))
			}
		}
		if p.IsNaN() || q.IsNaN() || p.IsInf(0) || q.IsInf(0) {
			continue
		}
		x, y := bigFromFloat256(p), bigFromFloat256(q)
		sum := new(big.Float).SetPrec(prec).Add(
			new(big.Float).SetPrec(prec).Mul(x, x),
			new(big.Float).SetPrec(prec).Mul(y, y),
		)
		got := Hypot256(p, q)
		if got.IsInf(0) {
			// must overflow
			if sum.Cmp(new(big.Float).SetPrec(prec).Mul(maxF, maxF)) <= 0 {
				t.Errorf("Hypot256(%v, %v) = Inf; want finite", p.Bits(), q.Bits())
			}
			continue
		}
		// got is correctly rounded iff |sqrt(sum) - got| <= half an ulp.
		g := bigFromFloat256(got)
		ulp := new(big.Float).SetPrec(prec).Sub(bigFromFloat256(got.Nextafter(NewFloat256Inf(1))), g)
		h := ulp.Quo(ulp, two)
		d := new(big.Float).SetPrec(prec).Sub(new(big.Float).SetPrec(prec).Sqrt(sum), g)
		if d.Abs(d).Cmp(h) > 0 {
			t.Errorf("Hypot256(%v, %v) = %v: not correctly rounded", p.Bits(), q.Bits(), got)
		}
	}
}

func BenchmarkHypot256(b *testing.B) {
	p := exact256(3).Quo(exact256(7))
	q := exact256(5).Quo(exact256(11))
	for b.Loop() {
		Hypot256(p, q)
	}
}
