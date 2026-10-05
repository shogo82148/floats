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
func (a Float16) Gamma() Float16 {
	ix := a &^ signMask16
	switch {
	case ix >= uvinf16:
		if a == uvinf16 {
			return a // +Inf
		}
		return NewFloat16NaN() // -Inf and NaN
	case ix < 0x0101:
		// |Gamma(x)| ~ 1/|x| overflows for |x| < 2**-16 (the smallest subnormal numbers are 2**-24), including ±0.
		return a&signMask16 | uvinf16
	}

	var x float64
	if ix < 0x0400 {
		x = float64(ix) * 0x1p-24 // subnormal
	} else {
		x = normal16ToFloat64(ix)
	}

	if a&signMask16 == 0 {
		if ix > 0x489c {
			return uvinf16 // Gamma(x) overflows for x > 9.2
		}
		// Gamma(x) = Gamma(1+x)/x for 0 < x < 1. They are selected without branches,
		// since the branch is unpredictable if x is random.
		xb := math.Float64bits(x)
		mask := uint64(int64(xb-math.Float64bits(1)) >> 63) // all ones if x < 1
		z := x + math.Float64frombits(mask&math.Float64bits(1))
		p := math.Float64frombits(mask&xb | ^mask&math.Float64bits(1))
		// the result is a normal number
		b := math.Float64bits(gamma16Poly(z) / p)
		b += 1<<(shift64-shift16-1) - 1 + (b>>(shift64-shift16))&1 // round to nearest even
		return Float16(b>>(shift64-shift16) - (bias64-bias16)<<shift16)
	}

	// Gamma(x) = pi / (sin(pi x) Gamma(1-x)), where sin(pi x) = (-1)**n sin(pi r) for x = n + r, |r| <= 1/2.
	// r and 1-x are calculated exactly.
	x = -x
	n := math.RoundToEven(x)
	r := x - n
	if ix > 0x4a07 || r == 0 {
		// Gamma(x) underflows for x < -12.1, and it is NaN for the negative integers.
		return NewFloat16(math.Gamma(x))
	}
	y := math.Pi / (math.Pi * r * gamma16Sinc(r) * gamma16Poly(1-x))
	// flip the sign if n is odd, without a branch.
	y = math.Float64frombits(math.Float64bits(y) ^ uint64(int64(n)&1)<<63)
	return NewFloat16(y)
}

// gamma16Poly returns Gamma(x) for x in [1, 13.06) with the relative error less than 2**-35,
// by the polynomial of the segment of the width 1/16 that includes x.
func gamma16Poly(x float64) float64 {
	s := int(x * 16)
	c := &gamma16Coeffs[s-16]
	t := x - (float64(s)+0.5)*(1.0/16)
	return c[0] + t*(c[1]+t*(c[2]+t*(c[3]+t*(c[4]+t*c[5]))))
}

// gamma16Sinc returns sin(pi r)/(pi r) for |r| <= 1/2 with the relative error less than 2**-36.
func gamma16Sinc(r float64) float64 {
	u := r * r
	c := &gamma16SincCoeffs
	return c[0] + u*(c[1]+u*(c[2]+u*(c[3]+u*(c[4]+u*(c[5]+u*c[6])))))
}
