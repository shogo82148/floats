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

// Erfc returns the complementary error function of a.
//
// Special cases are:
//
//	+Inf.Erfc() = 0
//	-Inf.Erfc() = 2
//	NaN.Erfc() = NaN
func (a Float32) Erfc() Float32 {
	b := a.Bits()
	ix := b &^ signMask32
	switch {
	case ix > 0x7f80_0000:
		return a // NaN
	case b&signMask32 == 0 && ix >= 0x4120_ddfc:
		// erfc(x) is rounded to 0 for x >= 10.0541, including +Inf.
		return 0
	case b&signMask32 != 0 && ix >= 0x4075_47cb:
		// erfc(x) is rounded to 2 for x <= -3.8325, including -Inf.
		return NewFloat32FromBits(0x4000_0000)
	}

	// y = erfc(|x|). erfc(-x) = 2 - erfc(x).
	x := math.Float64frombits(math.Float64bits(float64(a)) &^ (1 << 63))
	y := erfc32Poly(x)
	if b&signMask32 != 0 {
		y = 2 - y
	}
	// The relative error of y is less than 2**-40, that is, 2**13 ulps of float64.
	if z, ok := float32Round(y, 1<<16); ok {
		return z
	}
	// The result may not be correctly rounded because it is close to the midpoint of two adjacent Float32 values.
	// math.Erfc rounded to Float32 is correctly rounded for all Float32 values except one, which was checked by
	// calculating the arguments whose results are close to the midpoints with mpmath.
	if b == 0xb76c9f62 {
		// erfc(x) = 1 + 133.4999999998... × 2**-23, but math.Erfc rounds it up.
		return NewFloat32FromBits(0x3f80_0085)
	}
	return NewFloat32(math.Erfc(float64(a)))
}

// erfc32Poly returns erfc(x) for x in [0, 10.06) with the relative error less than 2**-41,
// by the polynomial of the segment of the width 1/32 that includes x.
func erfc32Poly(x float64) float64 {
	s := int(x * 32)
	c := &erfc32Coeffs[s]
	t := x - (float64(s)+0.5)*(1.0/32)
	t2 := t * t
	t4 := t2 * t2
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	return ((c[0] + t*c[1]) + t2*(c[2]+t*c[3])) + t4*((c[4]+t*c[5])+t2*(c[6]+t*c[7])+t4*c[8])
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
	b := a.Bits()
	ix := b &^ signMask32
	switch {
	case ix > 0x3f80_0000:
		return NewFloat32(math.NaN()) // |x| > 1, ±Inf, and NaN
	case ix == 0x3f80_0000:
		return NewFloat32FromBits(0x7f80_0000 | b&signMask32)
	}

	// erfinv is an odd function, so the polynomials for |x| are used and the sign of a is restored.
	x := math.Float64frombits(math.Float64bits(float64(a)) &^ (1 << 63))
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
		y = erfinv32Poly(1 - x)
	}
	// The relative error of y is less than 2**-42, that is, 2**10 ulps of float64.
	if z, ok := float32Round(y, 1<<16); ok {
		return NewFloat32FromBits(z.Bits() | b&signMask32)
	}
	// The result may not be correctly rounded because it is close to the midpoint of two adjacent Float32 values.
	// math.Erfinv rounded to Float32 is correctly rounded for all Float32 values except one, which was checked by
	// calculating the arguments whose results are close to the midpoints with mpmath.
	if b&^signMask32 == 0x3bba_61fd {
		// erfinv(x) = 0x1.a52dc7.8000000a... × 2**-8, but math.Erfinv rounds it down.
		return NewFloat32FromBits(0x3ba5_2dc8 | b&signMask32)
	}
	return NewFloat32(math.Erfinv(float64(a)))
}

// erfinv32Poly returns erfinv(1-t) for t in [2**-24, 1/2] with the relative error less than 2**-43,
// by the polynomial of the segment that includes t.
func erfinv32Poly(t float64) float64 {
	b := math.Float64bits(t)
	c := &erfinv32HiCoeffs[b>>49-7992]
	s := t - math.Float64frombits(b&^(1<<49-1)|1<<48) // the center of the segment
	s2 := s * s
	s4 := s2 * s2
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	return ((c[0] + s*c[1]) + s2*(c[2]+s*c[3])) + s4*((c[4]+s*c[5])+s2*(c[6]+s*c[7]))
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
	b := a.Bits()
	switch {
	case b&^signMask32 == 0:
		return NewFloat32FromBits(0x7f80_0000) // +Inf for ±0
	case b&signMask32 != 0 || b > 0x4000_0000:
		return NewFloat32(math.NaN()) // x < 0, x > 2, and NaN
	case b == 0x4000_0000:
		return NewFloat32FromBits(0xff80_0000) // -Inf for 2
	}

	// erfcinv(x) = -erfcinv(2-x), where 2 - x is exact.
	x := float64(a)
	neg := b > 0x3f80_0000
	if neg {
		x = 2 - x
	}
	sign := uint32(0)
	if neg {
		sign = signMask32
	}

	if x >= 0.5 {
		// erfcinv(x) = erfinv(1-x), where 1 - x is exact.
		return NewFloat32FromBits(Float32(1-x).Erfinv().Bits() | sign)
	}

	// 0 < x < 1/2. erfcinv(x) is calculated by the polynomials of t = x.
	var y float64
	if x >= 0x1p-24 {
		y = erfinv32Poly(x)
	} else {
		y = erfcinv32Poly(x)
	}
	// The relative error of y is less than 2**-43, that is, 2**9 ulps of float64.
	if z, ok := float32Round(y, 1<<11); ok {
		return NewFloat32FromBits(z.Bits() | sign)
	}
	// The result may not be correctly rounded because it is close to the midpoint of two adjacent Float32 values.
	// 1 - x is exact in Float256, and its Erfinv is correctly rounded.
	return NewFloat32FromBits(Float256(uvone256).Sub(NewFloat256(x)).Erfinv().Float32().Bits() | sign)
}

// erfcinv32Poly returns erfcinv(t) for t in [2**-149, 2**-24) with the relative error less than 2**-47,
// by the polynomial of the segment that includes t.
func erfcinv32Poly(t float64) float64 {
	b := math.Float64bits(t)
	c := &erfcinv32Coeffs[b>>49-6992]
	// z = 16 (t - the center of the segment)/2**e is in [-1, 1]. 16/2**e is a power of two, so that z is exact.
	z := (t - math.Float64frombits(b&^(1<<49-1)|1<<48)) * math.Float64frombits(uint64(2050-b>>52)<<52)
	z2 := z * z
	z4 := z2 * z2
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	return ((c[0] + z*c[1]) + z2*(c[2]+z*c[3])) + z4*((c[4]+z*c[5])+z2*(c[6]+z*c[7]))
}
