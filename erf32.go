package floats

import "math"

// Erf returns the error function of a.
//
// Special cases are:
//
//	+Inf.Erf() = 1
//	-Inf.Erf() = -1
//	NaN.Erf() = NaN
func (a Float32) Erf() Float32 {
	b := a.Bits()
	ix := b &^ signMask32
	switch {
	case ix > 0x7f80_0000:
		return a // NaN
	case ix >= 0x407a_d445:
		// erf(x) is rounded to 1 for |x| >= 3.9192, including ±Inf.
		return NewFloat32FromBits(b&signMask32 | 0x3f80_0000)
	}

	// erf is an odd function, so the polynomials for |x| are used and the sign of a is restored.
	x := math.Float64frombits(math.Float64bits(float64(a)) &^ (1 << 63))
	var y float64
	if x < 0.125 {
		// erf(x) = x P(x**2). It also handles ±0 and the subnormal numbers.
		c := &erf16SmallCoeffs
		u := x * x
		y = x * (c[0] + u*(c[1]+u*(c[2]+u*(c[3]+u*(c[4]+u*c[5])))))
	} else {
		y = erf32Poly(x)
	}
	// The relative error of y is less than 2**-38, that is, 2**14 ulps of float64.
	if z, ok := float32Round(y, 1<<16); ok {
		return NewFloat32FromBits(z.Bits() | b&signMask32)
	}
	// The result may not be correctly rounded because it is close to the midpoint of two adjacent Float32 values.
	// math.Erf rounded to Float32 is correctly rounded for all Float32 values, which was checked by calculating
	// the arguments whose results are close to the midpoints with mpmath.
	return NewFloat32(math.Erf(float64(a)))
}

// erf32Poly returns erf(x) for x in [1/8, 3.92) with the relative error less than 2**-39,
// by the polynomial of the segment that includes x.
func erf32Poly(x float64) float64 {
	b := math.Float64bits(x)
	c := &erf32Coeffs[b>>49-8160]
	t := x - math.Float64frombits(b&^(1<<49-1)|1<<48) // the center of the segment
	t2 := t * t
	t4 := t2 * t2
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	return ((c[0] + t*c[1]) + t2*(c[2]+t*c[3])) + t4*((c[4]+t*c[5])+t2*(c[6]+t*c[7]))
}

// Erfc returns the complementary error function of x.
//
// Special cases are:
//
//	+Inf.Erfc() = 0
//	-Inf.Erfc() = 2
//	NaN.Erfc() = NaN
func (a Float32) Erfc() Float32 {
	return NewFloat32(math.Erfc(a.Float64().BuiltIn()))
}

// Erfinv returns the inverse error function of a.
//
// Special cases are:
//
//	1.Erfinv() = +Inf
//	-1.Erfinv() = -Inf
//	x.Erfinv() = NaN if x < -1 or x > 1
//	NaN.Erfinv() = NaN
func (a Float32) Erfinv() Float32 {
	return NewFloat32(math.Erfinv(a.Float64().BuiltIn()))
}

// Erfcinv returns the inverse of [Erfc](a).
//
// Special cases are:
//
//	0.Erfcinv() = +Inf
//	2.Erfcinv() = -Inf
//	x.Erfcinv() = NaN if x < 0 or x > 2
//	NaN.Erfcinv() = NaN
func (a Float32) Erfcinv() Float32 {
	return NewFloat32(math.Erfcinv(a.Float64().BuiltIn()))
}
