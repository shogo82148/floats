package floats

import (
	"math/big"
	"math/rand/v2"
	"testing"

	"github.com/shogo82148/ints"
)

// toBig converts x to an exact big.Float.
func (x ynExt128) toBig() *big.Float {
	m := new(big.Int)
	for _, w := range x.m {
		m.Lsh(m, 64).Or(m, new(big.Int).SetUint64(w))
	}
	f := new(big.Float).SetPrec(128).SetInt(m)
	f.SetMantExp(f, x.e-127)
	if x.neg {
		f.Neg(f)
	}
	return f
}

// ynExt128FromBig converts f, which must be representable in 128 bits, to ynExt128.
func ynExt128FromBig(f *big.Float) ynExt128 {
	if f.Sign() == 0 {
		return ynExt128{}
	}
	mant := new(big.Float).SetPrec(128)
	exp := f.MantExp(mant) // f = mant * 2**exp, 0.5 <= |mant| < 1
	neg := mant.Sign() < 0
	mant.Abs(mant)
	mant.SetMantExp(mant, 128)
	mi, _ := mant.Int(nil)
	var m ints.Uint128
	for i := range 2 {
		m[1-i] = new(big.Int).And(new(big.Int).Rsh(mi, uint(64*i)), new(big.Int).SetUint64(^uint64(0))).Uint64()
	}
	return ynExt128{m: m, e: exp - 1, neg: neg}
}

// randYnExt128 returns a random value with the exponent e, including the special significands.
func randYnExt128(r *rand.Rand, e int) ynExt128 {
	m := ints.Uint128{r.Uint64() | 1<<63, r.Uint64()}
	switch r.IntN(8) {
	case 0:
		m = ints.Uint128{1 << 63, 0} // a power of two
	case 1:
		m[1] = 0 // trailing zeros
	}
	return ynExt128{m: m, e: e, neg: r.IntN(2) == 0}
}

// TestYnExt128 compares Add, Mul and Inv with the 128-bit math/big.Float, which is correctly rounded.
func TestYnExt128(t *testing.T) {
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
			eb = ea - 122 - r.IntN(15) // around the limit of the alignment
		default:
			eb = ea + r.IntN(300) - 150
		}
		a, b := randYnExt128(r, ea), randYnExt128(r, eb)
		if r.IntN(4) == 0 && ea == eb {
			// the nearly equal significands
			b.m = a.m
			b.m[1] ^= uint64(r.IntN(8))
			if b.m[1] == 0 && b.m[0] == 1<<63 {
				b.m[1] = 1
			}
		}
		ab, bb := a.toBig(), b.toBig()

		sum := new(big.Float).SetPrec(128).Add(ab, bb)
		if got, want := ynExt128Add(a, b), ynExt128FromBig(sum); got != want {
			t.Fatalf("Add(%v, %v) = %v; want %v", a, b, got, want)
		}
		prod := new(big.Float).SetPrec(128).Mul(ab, bb)
		if got, want := ynExt128Mul(a, b), ynExt128FromBig(prod); got != want {
			t.Fatalf("Mul(%v, %v) = %v; want %v", a, b, got, want)
		}
		quo := new(big.Float).SetPrec(128).Quo(big.NewFloat(1), ab)
		if got, want := ynExt128Inv(a), ynExt128FromBig(quo); got != want {
			t.Fatalf("Inv(%v) = %v; want %v", a, got, want)
		}
	}
}

// TestYnExt128Float128 checks the conversion between ynExt128 and Float128.
func TestYnExt128Float128(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(3, 4))
	for range 20000 {
		a := NewFloat128FromBits(ints.Uint128{r.Uint64(), r.Uint64()})
		if a.IsNaN() || a.IsInf(0) {
			continue
		}
		// the conversion is exact for the Float128 values.
		if got := ynExt128FromFloat128(a).Float128(); !eq128(got, a) {
			t.Fatalf("Float128(%v) = %v", a, got)
		}
	}
	// rounding to nearest even
	for range 20000 {
		x := randYnExt128(r, r.IntN(2000)-1000)
		if r.IntN(2) == 0 {
			// ties
			x.m[1] = x.m[1]&^(1<<15-1) | 1<<14
		}
		want := ynExt128FromBig(new(big.Float).SetPrec(113).SetMode(big.ToNearestEven).Set(x.toBig()))
		if got := ynExt128FromFloat128(x.Float128()); got != want {
			t.Fatalf("Float128(%v) = %v; want %v", x, got, want)
		}
	}
	// overflow and underflow
	if got := (ynExt128{m: ints.Uint128{1 << 63}, e: 20000}).Float128(); !got.IsInf(1) {
		t.Errorf("overflow = %v", got)
	}
	if got := (ynExt128{m: ints.Uint128{1 << 63}, e: -20000, neg: true}).Float128(); !got.IsZero() || !got.Signbit() {
		t.Errorf("underflow = %v", got)
	}
}

// TestYnExt128Special checks the zeros, the carry-out of the rounding, and the cancellation.
func TestYnExt128Special(t *testing.T) {
	t.Parallel()
	one := ynExt128FromUint(1)
	var zero ynExt128
	if got := ynExt128FromFloat128(Float128{}); got != zero {
		t.Errorf("FromFloat128(0) = %v", got)
	}
	if got := zero.Float128(); !eq128(got, Float128{}) {
		t.Errorf("zero.Float128() = %v", got)
	}
	if got := ynExt128Add(one, zero); got != one {
		t.Errorf("1 + 0 = %v", got)
	}
	if got := ynExt128Add(zero, one); got != one {
		t.Errorf("0 + 1 = %v", got)
	}
	if got := ynExt128Mul(one, zero); got != zero {
		t.Errorf("1 * 0 = %v", got)
	}
	if got := ynExt128Add(one, ynExt128{m: one.m, e: one.e, neg: true}); got != zero {
		t.Errorf("1 - 1 = %v", got)
	}
	// far smaller values do not change the result.
	if got := ynExt128Add(one, ynExt128{m: one.m, e: -1000}); got != one {
		t.Errorf("1 + 2**-1000 = %v", got)
	}

	// the significand of all ones is rounded up to 2**128, and the exponent is carried.
	ones := ynExt128{m: ints.Uint128{^uint64(0), ^uint64(0)}, e: 3}
	if got, want := ones.Float128(), exact128(16); !eq128(got, want) {
		t.Errorf("Float128(%v) = %v; want %v", ones, got, want)
	}
}
