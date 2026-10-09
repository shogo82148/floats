package floats

import (
	"math/rand/v2"
	"runtime"
	"testing"
)

// quoReference16 is the reference implementation of Float16.Quo
// that uses integer arithmetic.
func quoReference16(a, b Float16) Float16 {
	if a.IsNaN() || b.IsNaN() {
		// a / NaN = NaN
		// NaN / b = NaN
		return uvnan16
	}

	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()
	sign := signA ^ signB

	if b.IsZero() {
		if a.IsZero() {
			// 0 / 0 = NaN
			return uvnan16
		}
		// ±finite / 0 = ±inf
		return Float16(sign | uvinf16)
	}
	if a.IsZero() {
		// 0 / ±finite = 0
		return Float16(sign)
	}
	if expA == mask16-bias16 {
		// NaN check is done above; a is ±inf
		if expB == mask16-bias16 {
			// ±inf / ±inf = NaN
			return uvnan16
		} else {
			// ±inf / finite = ±inf
			return Float16(sign | uvinf16)
		}
	}
	if expB == mask16-bias16 {
		// NaN check is done above; b is ±inf
		// NaN and Inf checks are done above; a is finite.
		// ±finite / ±inf = 0
		return Float16(sign)
	}

	exp := expA - expB + bias16
	if fracA < fracB {
		exp--
		fracA <<= 1
	}
	if exp >= mask16 {
		// overflow
		return Float16(sign | uvinf16)
	}

	shift := shift16 + 3 // 1 for the implicit bit, 1 for the rounding bit, 1 for the guard bit
	fracA32 := uint32(fracA) << shift
	frac := uint16(fracA32 / uint32(fracB))
	mod := uint16(fracA32 % uint32(fracB))
	frac |= nonzero16(mod)
	if exp <= 0 {
		// the result is subnormal
		shift := -exp + 3 + 1
		frac += (1<<(shift-1) - 1) + ((frac >> shift) & 1) // round to nearest even
		frac >>= shift
		return Float16(sign | uint16(frac))
	}

	frac += 0b11 + ((frac >> 3) & 1) // round to nearest even
	frac >>= 3
	return Float16(sign | uint16(exp)<<shift16 | frac&fracMask16)
}

// randomQuoOperands16 returns the operands for testing Float16.Quo.
func randomQuoOperands16(r *rand.Rand) (a, b Float16) {
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

func TestFloat16_QuoReference(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	for range 3_000_000 {
		a, b := randomQuoOperands16(r)
		if got, want := a.Quo(b), quoReference16(a, b); got != want {
			t.Fatalf("Float16(%#x).Quo(%#x) = %#x, want %#x", a, b, got, want)
		}
	}
}

func BenchmarkFloat16_QuoCases(b *testing.B) {
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
				runtime.KeepAlive(xs[i%1024].Quo(ys[i%1024]))
			}
		})
	}
}
