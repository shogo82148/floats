package floats

import (
	"math/bits"
	"math/rand/v2"
	"runtime"
	"testing"
)

// mulReference16 is the reference implementation of Float16.Mul
// that uses integer arithmetic.
func mulReference16(a, b Float16) Float16 {
	if a.IsNaN() || b.IsNaN() {
		// a * NaN = NaN
		// NaN * b = NaN
		return uvnan16
	}
	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()

	// special cases
	if expA == mask16-bias16 {
		// NaN check is done above; a is ±inf
		if b.IsZero() {
			// ±inf * 0 = NaN
			return uvnan16
		} else {
			// ±inf * +finite = ±inf
			// ±inf * -finite = ∓inf
			return a ^ Float16(signB)
		}
	}
	if expB == mask16-bias16 {
		// NaN check is done above; b is ±inf
		if a.IsZero() {
			// 0 * ±inf = NaN
			return uvnan16
		} else {
			// +finite * ±inf = ±inf
			// -finite * ±inf = ∓inf
			return b ^ Float16(signA)
		}
	}

	sign := signA ^ signB
	exp := expA + expB
	frac := uint32(fracA) * uint32(fracB)
	shift := bits.Len32(frac) - (shift16 + 1)
	exp += shift - shift16

	if exp < -(bias16 + shift16) {
		// underflow
		return Float16(sign)
	} else if exp <= -bias16 {
		// the result is subnormal
		shift := shift16 - (expA + expB + bias16) + 1
		frac += (1<<(shift-1) - 1) + ((frac >> shift) & 1) // round to nearest even
		frac >>= shift
		return Float16(sign | uint16(frac))
	}

	exp = expA + expB + bias16
	frac += (1<<(shift-1) - 1) + ((frac >> shift) & 1) // round to nearest even
	shift = bits.Len32(frac) - (shift16 + 1)
	exp += shift - shift16
	if exp >= mask16 {
		// overflow
		return Float16(sign | (mask16 << shift16))
	}
	frac >>= shift
	frac &= fracMask16
	return Float16(sign | uint16(exp<<shift16) | uint16(frac))
}

// randomMulOperands16 returns the operands for testing Float16.Mul.
func randomMulOperands16(r *rand.Rand) (a, b Float16) {
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

func TestFloat16_MulReference(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	for range 3_000_000 {
		a, b := randomMulOperands16(r)
		if got, want := a.Mul(b), mulReference16(a, b); got != want {
			t.Fatalf("Float16(%#x).Mul(%#x) = %#x, want %#x", a, b, got, want)
		}
	}
}

func BenchmarkFloat16_MulCases(b *testing.B) {
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
				runtime.KeepAlive(xs[i%1024].Mul(ys[i%1024]))
			}
		})
	}
}
