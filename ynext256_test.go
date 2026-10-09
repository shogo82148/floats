package floats

import (
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/shogo82148/ints"
)

// toBig converts x to an exact big.Float.
func (x ynExt256) toBig() *big.Float {
	m := new(big.Int)
	for _, w := range x.m {
		m.Lsh(m, 64).Or(m, new(big.Int).SetUint64(w))
	}
	f := new(big.Float).SetPrec(256).SetInt(m)
	f.SetMantExp(f, x.e-255)
	if x.neg {
		f.Neg(f)
	}
	return f
}

// ynExt256FromBig converts f, which must be representable in 256 bits, to ynExt256.
func ynExt256FromBig(f *big.Float) ynExt256 {
	if f.Sign() == 0 {
		return ynExt256{}
	}
	mant := new(big.Float).SetPrec(256)
	exp := f.MantExp(mant) // f = mant * 2**exp, 0.5 <= |mant| < 1
	neg := mant.Sign() < 0
	mant.Abs(mant)
	mant.SetMantExp(mant, 256)
	mi, _ := mant.Int(nil)
	var m ints.Uint256
	for i := range 4 {
		m[3-i] = new(big.Int).And(new(big.Int).Rsh(mi, uint(64*i)), new(big.Int).SetUint64(^uint64(0))).Uint64()
	}
	return ynExt256{m: m, e: exp - 1, neg: neg}
}

// randYnExt256 returns a random value with the exponent e, including the special significands.
func randYnExt256(r *rand.Rand, e int) ynExt256 {
	m := ints.Uint256{r.Uint64() | 1<<63, r.Uint64(), r.Uint64(), r.Uint64()}
	switch r.IntN(8) {
	case 0:
		m = ints.Uint256{1 << 63, 0, 0, 0} // a power of two
	case 1:
		m[3] = 0 // trailing zeros
	}
	return ynExt256{m: m, e: e, neg: r.IntN(2) == 0}
}

// TestYnExt256 compares Add, Mul and Inv with the 256-bit math/big.Float, which is correctly rounded.
func TestYnExt256(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	for range 200000 {
		ea := r.IntN(2000) - 1000
		var eb int
		switch r.IntN(4) {
		case 0:
			eb = ea // the cancellation
		case 1:
			eb = ea - r.IntN(5) + 2
		case 2:
			eb = ea - 250 - r.IntN(15) // around the limit of the alignment
		default:
			eb = ea + r.IntN(600) - 300
		}
		a, b := randYnExt256(r, ea), randYnExt256(r, eb)
		if r.IntN(4) == 0 && ea == eb {
			// the nearly equal significands
			b.m = a.m
			b.m[3] ^= uint64(r.IntN(8))
			if b.m[3]|b.m[2]|b.m[1] == 0 && b.m[0] == 1<<63 {
				b.m[3] = 1
			}
		}
		ab, bb := a.toBig(), b.toBig()

		sum := new(big.Float).SetPrec(256).Add(ab, bb)
		if got, want := ynExt256Add(a, b), ynExt256FromBig(sum); got != want {
			t.Fatalf("Add(%v, %v) = %v; want %v", a, b, got, want)
		}
		prod := new(big.Float).SetPrec(256).Mul(ab, bb)
		if got, want := ynExt256Mul(a, b), ynExt256FromBig(prod); got != want {
			t.Fatalf("Mul(%v, %v) = %v; want %v", a, b, got, want)
		}
		quo := new(big.Float).SetPrec(256).Quo(big.NewFloat(1), ab)
		if got, want := ynExt256Inv(a), ynExt256FromBig(quo); got != want {
			t.Fatalf("Inv(%v) = %v; want %v", a, got, want)
		}
	}
}

// TestYnExt256Float256 checks the conversion between ynExt256 and Float256.
func TestYnExt256Float256(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(3, 4))
	for range 20000 {
		a := NewFloat256FromBits(ints.Uint256{r.Uint64(), r.Uint64(), r.Uint64(), r.Uint64()})
		if a.IsNaN() || a.IsInf(0) {
			continue
		}
		// the conversion is exact for the Float256 values, and rounds to nearest even for the others.
		if got := ynExt256FromFloat256(a).Float256(); !eq256(got, a) {
			t.Fatalf("Float256(%v) = %v", a, got)
		}
	}
	// rounding to nearest even
	for range 20000 {
		x := randYnExt256(r, r.IntN(2000)-1000)
		if r.IntN(2) == 0 {
			// ties
			x.m[3] = x.m[3]&^(1<<19-1) | 1<<18
		}
		want := ynExt256FromBig(new(big.Float).SetPrec(237).SetMode(big.ToNearestEven).Set(x.toBig()))
		if got := ynExt256FromFloat256(x.Float256()); got != want {
			t.Fatalf("Float256(%v) = %v; want %v", x, got, want)
		}
	}
	// overflow and underflow
	if got := (ynExt256{m: ints.Uint256{1 << 63}, e: 300000}).Float256(); !got.IsInf(1) {
		t.Errorf("overflow = %v", got)
	}
	if got := (ynExt256{m: ints.Uint256{1 << 63}, e: -300000, neg: true}).Float256(); !got.IsZero() || !got.Signbit() {
		t.Errorf("underflow = %v", got)
	}
}

// TestYnExt256Special checks the zeros, the carry-out of the rounding, and the cancellation.
func TestYnExt256Special(t *testing.T) {
	t.Parallel()
	one := ynExt256FromUint(1)
	var zero ynExt256
	if got := ynExt256FromFloat256(Float256{}); got != zero {
		t.Errorf("FromFloat256(0) = %v", got)
	}
	if got := zero.Float256(); !eq256(got, Float256{}) {
		t.Errorf("zero.Float256() = %v", got)
	}
	if got := ynExt256Add(one, zero); got != one {
		t.Errorf("1 + 0 = %v", got)
	}
	if got := ynExt256Add(zero, one); got != one {
		t.Errorf("0 + 1 = %v", got)
	}
	if got := ynExt256Mul(one, zero); got != zero {
		t.Errorf("1 * 0 = %v", got)
	}
	if got := ynExt256Add(one, ynExt256{m: one.m, e: one.e, neg: true}); got != zero {
		t.Errorf("1 - 1 = %v", got)
	}
	// far smaller values do not change the result.
	if got := ynExt256Add(one, ynExt256{m: one.m, e: -1000}); got != one {
		t.Errorf("1 + 2**-1000 = %v", got)
	}

	// the significand of all ones is rounded up to 2**256, and the exponent is carried.
	ones := ynExt256{m: ints.Uint256{^uint64(0), ^uint64(0), ^uint64(0), ^uint64(0)}, e: 3}
	if got, want := ones.Float256(), exact256(16); !eq256(got, want) {
		t.Errorf("Float256(%v) = %v; want %v", ones, got, want)
	}
}
