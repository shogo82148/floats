package floats

import (
	"math/bits"
	"math/rand/v2"
	"runtime"
	"testing"
)

// addReference16 is the reference implementation of Float16.Add
// that uses integer arithmetic.
func addReference16(a, b Float16) Float16 {
	if a.IsNaN() || b.IsNaN() {
		// a + NaN = NaN
		// NaN + b = NaN
		return uvnan16
	}
	if a.IsZero() {
		if b.IsZero() {
			//  0 +  0 =  0
			//  0 + -0 =  0
			// -0 +  0 =  0
			// -0 + -0 = -0
			return a & b
		}
		// ±0 + b = b
		return b
	}
	if b.IsZero() {
		// a + ±0 = a
		return a
	}

	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()

	// handle special cases
	if expA == mask16-bias16 {
		// NaN check is done above; a is ±inf
		if expB == mask16-bias16 {
			// NaN check is done above; b is ±inf
			if signA == signB {
				// ±inf + ±inf = ±inf
				return Float16(signA | uvinf16)
			}
			// ±inf + ∓inf = NaN
			return uvnan16
		}
		// b is finite, the result is ±inf
		return a
	}
	if expB == mask16-bias16 {
		// NaN check is done above; b is ±inf
		// NaN and Inf checks are done above; a is finite.
		return b
	}

	if expA < expB {
		// swap a and b
		signA, signB = signB, signA
		expA, expB = expB, expA
		fracA, fracB = fracB, fracA
	}

	// add the fractions
	const offset = 16
	fracA32 := int32(fracA) << offset
	fracB32 := int32(fracB) << offset
	fracB32 >>= uint(expA - expB)
	if signA != 0 {
		fracA32 = -fracA32
	}
	if signB != 0 {
		fracB32 = -fracB32
	}
	frac32 := fracA32 + fracB32
	sign := uint16(0)
	if frac32 < 0 {
		sign = signMask16
		frac32 = -frac32
	}

	shift := bits.Len32(uint32(frac32)) - shift16 - 1
	exp := expA + shift - offset

	// normalize
	if frac32 == 0 || exp < -(bias16+shift16) {
		// underflow
		return Float16(sign)
	}
	if exp <= -bias16 {
		// the result is subnormal
		shift := offset - (expA + bias16) + 1
		frac32 += (1<<uint(shift-1) - 1) + ((frac32 >> uint(shift)) & 1) // round to nearest even
		frac := uint16(frac32 >> shift)
		return Float16(sign | uint16(frac))
	}
	if exp >= mask16-bias16 {
		// overflow
		return Float16(sign | (mask16 << shift16))
	}

	frac32 += (1<<uint(shift-1) - 1) + ((frac32 >> uint(shift)) & 1) // round to nearest even
	if bits.Len32(uint32(frac32)) > shift16+shift+1 {
		frac32 >>= 1
		exp++
		if exp >= mask16 {
			// overflow
			return Float16(sign | (mask16 << shift16))
		}
	}
	frac := uint16(frac32 >> shift)
	return Float16(sign | uint16(exp+bias16)<<shift16 | frac&fracMask16)
}

// randomAddOperands16 returns the operands for testing Float16.Add.
func randomAddOperands16(r *rand.Rand) (a, b Float16) {
	a, b = Float16(r.Uint32()), Float16(r.Uint32())
	switch r.IntN(4) {
	case 0:
		// the exponents are close, to exercise the cancellation
		b = b&^0x7c00 | a&0x7c00
		b += Float16(r.IntN(5)-2) << 10
	case 1:
		// the results are subnormal or near the overflow
		a = a&^0x7c00 | Float16(r.IntN(31))<<10
		b = b&^0x7c00 | Float16(r.IntN(31))<<10
	case 2:
		// short fractions, to exercise the exact results and the ties
		n := uint(r.IntN(11))
		a, b = a>>n<<n, b>>n<<n
	}
	return
}

func TestFloat16_AddReference(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	for range 3_000_000 {
		a, b := randomAddOperands16(r)
		if got, want := a.Add(b), addReference16(a, b); got != want {
			t.Fatalf("Float16(%#x).Add(%#x) = %#x, want %#x", a, b, got, want)
		}
		if got, want := a.Sub(b), addReference16(a, b.Neg()); got != want {
			t.Fatalf("Float16(%#x).Sub(%#x) = %#x, want %#x", a, b, got, want)
		}
	}
}

func BenchmarkFloat16_AddCases(b *testing.B) {
	// the exponents are in a typical range and in a wide range
	// (about a half of the results of the product are subnormal).
	r := rand.New(rand.NewPCG(1, 2))
	for _, c := range []struct {
		name   string
		lo, hi int
	}{{"typical", -5, 5}, {"wide", -14, 14}} {
		var xs, ys [1024]Float16
		for i := range xs {
			xs[i] = Float16(r.Uint32()&0x83ff | uint32(15+c.lo+r.IntN(c.hi-c.lo+1))<<10)
			ys[i] = Float16(r.Uint32()&0x83ff | uint32(15+c.lo+r.IntN(c.hi-c.lo+1))<<10)
		}
		b.Run(c.name, func(b *testing.B) {
			for i := 0; b.Loop(); i++ {
				runtime.KeepAlive(xs[i%1024].Add(ys[i%1024]))
			}
		})
	}
}
