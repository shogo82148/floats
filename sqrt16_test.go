package floats

import (
	"math/rand/v2"
	"runtime"
	"testing"
)

// sqrtReference16 is the reference implementation of Float16.Sqrt
// that uses integer arithmetic.
func sqrtReference16(a Float16) Float16 {
	// special cases
	switch {
	case a.IsZero() || a.IsNaN() || a.IsInf(1):
		return a
	case a&signMask16 != 0:
		return uvnan16
	}

	_, exp, frac := a.normalize()
	if exp%2 != 0 {
		// odd exp, double x to make it even
		frac <<= 1
	}
	// exponent of square root
	exp >>= 1

	// generate sqrt(frac) bit by bit
	frac <<= 1
	var q, s uint16 // q = sqrt(frac)
	r := uint16(1 << (shift16 + 1))
	for r != 0 {
		t := s + r
		if t <= frac {
			s = t + r
			frac -= t
			q += r
		}
		frac <<= 1
		r >>= 1
	}

	// final rounding
	if frac != 0 {
		q += q & 1
	}
	return Float16((exp-1+bias16)<<shift16) + Float16(q>>1)
}

func TestFloat16_SqrtReference(t *testing.T) {
	t.Parallel()
	for a := range 1 << 16 {
		if got, want := Float16(a).Sqrt(), sqrtReference16(Float16(a)); got != want {
			t.Fatalf("Float16(%#x).Sqrt() = %#x, want %#x", a, got, want)
		}
	}
}

func BenchmarkFloat16_SqrtCases(b *testing.B) {
	// the exponents are in a typical range and in a wide range
	r := rand.New(rand.NewPCG(1, 2))
	for _, c := range []struct {
		name   string
		lo, hi int
	}{{"typical", -5, 5}, {"wide", -14, 14}} {
		var xs [1024]Float16
		for i := range xs {
			xs[i] = Float16(r.Uint32()&0x03ff | uint32(15+c.lo+r.IntN(c.hi-c.lo+1))<<10)
		}
		b.Run(c.name, func(b *testing.B) {
			for i := 0; b.Loop(); i++ {
				runtime.KeepAlive(xs[i%1024].Sqrt())
			}
		})
	}
}
