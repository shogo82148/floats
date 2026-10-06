package floats

import "math"

// Erf returns the error function of a.
//
// Special cases are:
//
//	+Inf.Erf() = 1
//	-Inf.Erf() = -1
//	NaN.Erf() = NaN
func (a Float16) Erf() Float16 {
	ix := a &^ signMask16
	switch {
	case ix > uvinf16:
		return a // NaN
	case ix >= 0x4131:
		// erf(x) is rounded to 1 for |x| >= 2.5957, including ±Inf.
		return a&signMask16 | 0x3c00
	}

	var x float64
	if ix < 0x0400 {
		x = float64(ix) * 0x1p-24 // subnormal
	} else {
		x = normal16ToFloat64(ix)
	}

	// erf is an odd function, so the polynomials for |x| are used and the sign of a is restored.
	var y float64
	if x < 0.125 {
		// erf(x) = x P(x**2). It also handles ±0 and the subnormal numbers.
		c := &erf16SmallCoeffs
		u := x * x
		y = x * (c[0] + u*(c[1]+u*(c[2]+u*(c[3]+u*(c[4]+u*c[5])))))
	} else {
		y = erf16Poly(x)
	}
	// y is not negative, and the relative error is less than 2**-34, which is much smaller than the distance to
	// the midpoint of two adjacent Float16 values (at least 2**-15 ulp, checked with mpmath).
	return NewFloat16(y) | a&signMask16
}

// erf16Poly returns erf(x) for x in [1/8, 2.6) with the relative error less than 2**-34,
// by the polynomial of the segment that includes x.
func erf16Poly(x float64) float64 {
	b := math.Float64bits(x)
	c := &erf16Coeffs[b>>49-8160]
	t := x - math.Float64frombits(b&^(1<<49-1)|1<<48) // the center of the segment
	t2 := t * t
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	return ((c[0] + t*c[1]) + t2*(c[2]+t*c[3])) + t2*t2*((c[4]+t*c[5])+t2*c[6])
}

// Erfc returns the complementary error function of a.
//
// Special cases are:
//
//	+Inf.Erfc() = 0
//	-Inf.Erfc() = 2
//	NaN.Erfc() = NaN
func (a Float16) Erfc() Float16 {
	ix := a &^ signMask16
	switch {
	case ix > uvinf16:
		return a // NaN
	case a&signMask16 == 0 && ix >= 0x43d7:
		// erfc(x) is rounded to 0 for x >= 3.9199, including +Inf.
		return 0
	case a&signMask16 != 0 && ix >= 0x40ef:
		// erfc(x) is rounded to 2 for x <= -2.4668, including -Inf.
		return 0x4000
	}

	var x float64
	if ix < 0x0400 {
		x = float64(ix) * 0x1p-24 // subnormal
	} else {
		x = normal16ToFloat64(ix)
	}

	// y = erfc(|x|). erfc(-x) = 2 - erfc(x).
	var y float64
	if x < 0.125 {
		c := &erf16SmallCoeffs
		u := x * x
		y = 1 - x*(c[0]+u*(c[1]+u*(c[2]+u*(c[3]+u*(c[4]+u*c[5])))))
	} else {
		y = erfc16Poly(x)
	}
	if a&signMask16 != 0 {
		y = 2 - y
	}
	// The relative error of y is less than 2**-36, which is much smaller than the distance to
	// the midpoint of two adjacent Float16 values (at least 2**-30 of y, checked with mpmath).
	return NewFloat16(y)
}

// erfc16Poly returns erfc(x) for x in [1/8, 3.92) with the relative error less than 2**-36,
// by the polynomial of the segment that includes x.
func erfc16Poly(x float64) float64 {
	b := math.Float64bits(x)
	c := &erfc16Coeffs[b>>49-8160]
	t := x - math.Float64frombits(b&^(1<<49-1)|1<<48) // the center of the segment
	t2 := t * t
	t4 := t2 * t2
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	p01 := c[0] + t*c[1]
	p23 := c[2] + t*c[3]
	p45 := c[4] + t*c[5]
	p67 := c[6] + t*c[7]
	p89 := c[8] + t*c[9]
	return (p01 + t2*p23) + t4*((p45+t2*p67)+t4*(p89+t2*c[10]))
}

// Erfinv returns the inverse error function of a.
//
// Special cases are:
//
//	1.Erfinv() = +Inf
//	-1.Erfinv() = -Inf
//	x.Erfinv() = NaN if x < -1 or x > 1
//	NaN.Erfinv() = NaN
func (a Float16) Erfinv() Float16 {
	ix := a &^ signMask16
	switch {
	case ix > 0x3c00:
		return NewFloat16(math.NaN()) // |x| > 1, ±Inf, and NaN
	case ix == 0x3c00:
		return uvinf16 | a&signMask16
	}

	// erfinv is an odd function, so the polynomials for |x| are used and the sign of a is restored.
	var x float64
	if ix < 0x0400 {
		x = float64(ix) * 0x1p-24 // subnormal
	} else {
		x = normal16ToFloat64(ix)
	}
	var y float64
	switch {
	case x < 0.125:
		// erfinv(x) = x P(x**2). It also handles ±0 and the subnormal numbers.
		c := &erfinv16SmallCoeffs
		u := x * x
		y = x * (c[0] + u*(c[1]+u*(c[2]+u*(c[3]+u*(c[4]+u*c[5])))))
	case x < 0.5:
		y = erfinv16Poly(&erfinv16MidCoeffs[math.Float64bits(x)>>49-8160], x)
	default:
		// erfinv(x) is singular at x = 1, so that the polynomials of t = 1 - x, which is exact, are used.
		t := 1 - x
		y = erfinv16Poly(&erfinv16HiCoeffs[math.Float64bits(t)>>49-8096], t)
	}
	// The relative error of y is less than 2**-37, which is much smaller than the distance to
	// the midpoint of two adjacent Float16 values (at least 2**-26 of y, checked with mpmath).
	return NewFloat16(y) | a&signMask16
}

// erfinv16Poly returns the value of the polynomial c at x, whose segment is determined by the float64 representation of x.
func erfinv16Poly(c *[7]float64, x float64) float64 {
	b := math.Float64bits(x)
	t := x - math.Float64frombits(b&^(1<<49-1)|1<<48) // the center of the segment
	t2 := t * t
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	return ((c[0] + t*c[1]) + t2*(c[2]+t*c[3])) + t2*t2*((c[4]+t*c[5])+t2*c[6])
}

// Erfcinv returns the inverse of [Erfc](a).
//
// Special cases are:
//
//	0.Erfcinv() = +Inf
//	2.Erfcinv() = -Inf
//	x.Erfcinv() = NaN if x < 0 or x > 2
//	NaN.Erfcinv() = NaN
func (a Float16) Erfcinv() Float16 {
	return NewFloat16(math.Erfcinv(a.Float64().BuiltIn()))
}
