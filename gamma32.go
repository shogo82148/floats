package floats

import "math"

// Gamma returns the Gamma function of a.
//
// Special cases are:
//
//	+Inf.Gamma() = +Inf
//	+0.Gamma() = +Inf
//	-0.Gamma() = -Inf
//	x.Gamma() = NaN for integer x < 0
//	-Inf.Gamma() = NaN
//	NaN.Gamma() = NaN
func (a Float32) Gamma() Float32 {
	b := a.Bits()
	ix := b &^ signMask32

	switch {
	case ix < 0x0020_0000:
		// |Gamma(x)| ~ 1/|x| overflows for |x| < 2**-128, including ±0.
		return NewFloat32FromBits(b&signMask32 | 0x7f80_0000)
	case ix < 0x2d80_0000:
		// |x| < 2**-36. Gamma(x) = 1/x - 0.577... + ..., and the error of 1/x is less than 2**-36.
		if y, ok := gamma32Round(1 / float64(a)); ok {
			return y
		}
	case b < 0x420c_6666:
		// 2**-36 <= x < 35.1
		if y, ok := gamma32Fast(float64(a)); ok {
			return y
		}
	case b < 0x7f80_0000:
		// Gamma(x) overflows for x >= 35.1.
		return NewFloat32Inf(1)
	case b >= signMask32 && ix < 0x422c_0000:
		// -43 < x <= -2**-36
		if y, ok := gamma32Fast(float64(a)); ok {
			return y
		}
	case b >= signMask32 && ix < 0x7f80_0000:
		// x <= -43. All Float32 values with |x| >= 2**23 are integers.
		// Gamma(x) underflows for the others, and its sign is the sign of sin(pi x) = (-1)**n sin(pi r) for x = n + r.
		x := float64(a)
		n := math.RoundToEven(x)
		r := x - n
		if r == 0 {
			return NewFloat32NaN() // negative integer
		}
		sign := uint32(int64(n)&1) << 31
		if r < 0 {
			sign ^= signMask32
		}
		return NewFloat32FromBits(sign)
	}
	// The others are ±Inf, NaN, and the arguments whose results may not be correctly rounded.
	return gamma32Slow(a)
}

// gamma32Slow returns Gamma(a) with math.Gamma.
// math.Gamma rounded to Float32 is correctly rounded except for the arguments whose results are
// very close to the midpoint of two adjacent Float32 values. gamma32Hard has the correct results of the ones
// that are not correctly rounded.
func gamma32Slow(a Float32) Float32 {
	b := a.Bits()
	for _, h := range gamma32Hard {
		if h[0] == b {
			return NewFloat32FromBits(h[1])
		}
	}
	return NewFloat32(math.Gamma(float64(a)))
}

// gamma32Fast calculates Gamma(x) in float64, whose relative error is less than 2**-34,
// for 2**-36 <= |x| < 43, and returns it rounded to Float32.
// ok is false if the result may not be correctly rounded, because it is close to the midpoint
// of two adjacent Float32 values, or if x is a negative integer.
func gamma32Fast(x float64) (y Float32, ok bool) {
	var g float64
	switch {
	case x >= 1:
		g = gamma32Poly(x)
	case x > 0:
		// Gamma(x) = Gamma(1+x)/x for 0 < x < 1
		g = gamma32Poly(1+x) / x
	default:
		// Gamma(x) = pi / (sin(pi x) Gamma(1-x)), where sin(pi x) = (-1)**n sin(pi r) for x = n + r, |r| <= 1/2.
		// r and 1-x are exact.
		n := math.RoundToEven(x)
		r := x - n
		if r == 0 {
			return 0, false // negative integer
		}
		g = math.Pi / (math.Pi * r * gamma16Sinc(r) * gamma32Poly(1-x))
		// flip the sign if n is odd, without a branch.
		g = math.Float64frombits(math.Float64bits(g) ^ uint64(int64(n)&1)<<63)
	}
	return gamma32Round(g)
}

// gamma32Round rounds g, whose relative error is less than 2**-34, to Float32.
// ok is false if the result may not be correctly rounded, because g is close to the midpoint
// of two adjacent Float32 values.
func gamma32Round(g float64) (y Float32, ok bool) {
	// The error of g is less than 2**18 ulps of float64.
	return float32Round(g, 1<<21)
}

// float32Round rounds g to Float32. ok is false if the result may not be correctly rounded, because g is
// closer than the window ulps of float64 to the midpoint of two adjacent Float32 values.
// The window must be larger than the error of g in ulps of float64.
func float32Round(g float64, window int64) (y Float32, ok bool) {
	// round to Float32 if g is not close to the midpoint of two adjacent Float32 values.
	// The midpoint is 2**28 in the lower 29 bits.
	bits := math.Float64bits(g)
	sign := uint32(bits>>32) & signMask32
	abs := bits &^ (1 << 63)
	if abs < 0x3810_0000_0000_0000 {
		// 2**-126 > |g|: the result is subnormal. It is m × 2**-149 for the integer m rounded from |g| × 2**149.
		m := math.Float64frombits(abs) * 0x1p149
		k := float64(uint32(m))
		if d := m - k - 0.5; -0x1p-10 < d && d < 0x1p-10 {
			return 0, false
		}
		if m-k > 0.5 {
			k++
		}
		return NewFloat32FromBits(sign | uint32(k)), true
	}
	if abs >= 0x47f0_0000_0000_0000 { // 2**128 <= |g|
		return NewFloat32FromBits(sign | 0x7f80_0000), true
	}
	low := abs & (1<<29 - 1)
	if d := int64(low) - 1<<28; -window < d && d < window {
		return 0, false
	}
	// the rounding may carry into the exponent, and it becomes ±Inf if the result overflows.
	r := abs>>29 + low>>28 - (1023-127)<<23
	return NewFloat32FromBits(sign | uint32(r)), true
}

// gamma32Poly returns Gamma(x) for x in [1, 44) with the relative error less than 2**-35,
// by the polynomial of the segment of the width 1/4 that includes x.
func gamma32Poly(x float64) float64 {
	s := int(x * 4)
	c := &gamma32Coeffs[s-4]
	t := x - (float64(s)+0.5)*(1.0/4)
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	t2 := t * t
	t4 := t2 * t2
	return ((c[0] + t*c[1]) + t2*(c[2]+t*c[3])) + t4*((c[4]+t*c[5])+t2*(c[6]+t*c[7])+t4*c[8])
}

// gamma32Hard are the arguments whose results are not correctly rounded by math.Gamma rounded to Float32,
// and the correct results. They were found by checking all Float32 values, and the exact results
// were calculated with mpmath.
var gamma32Hard = [...][2]uint32{
	{0x27de86a9, 0x57134133},
	{0x27e05475, 0x57121211},
	{0x41e886d1, 0x709989b5},
}
