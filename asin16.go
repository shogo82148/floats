package floats

import "math"

// Asin returns the arcsine, in radians, of a.
//
// Special cases are:
//
//	±0.Asin() = ±0
//	x.Asin() = NaN if x < -1 or x > 1
func (a Float16) Asin() Float16 {
	ix := a &^ signMask16
	if ix < 0x2800 { // |a| < 2**-5
		// asin(a) = a + a**3/6 + ... rounds to a.
		return a
	}
	if ix > uvone16 { // |a| > 1, Inf, or NaN
		return NewFloat16NaN()
	}

	// a is a normal Float16 value, so the conversion is exact.
	x := normal16ToFloat64(ix)
	var r float64
	if x <= 0.5 {
		r = x * asinSeries16(x*x)
	} else {
		// asin(x) = pi/2 - 2 asin(sqrt((1-x)/2))
		t := (1 - x) * 0.5 // exact
		r = math.Pi/2 - 2*math.Sqrt(t)*asinSeries16(t)
	}
	if a&signMask16 != 0 {
		r = -r
	}
	return NewFloat16(r)
}

// asinSeries16 returns asin(sqrt(z))/sqrt(z) for 0 <= z <= 1/4.
// It is a degree-7 minimax polynomial; the relative error is about 2e-11,
// far below the Float16 precision.
func asinSeries16(z float64) float64 {
	return 0.9999999999854742 + z*(0.16666667399182022+z*(0.07499939642496817+z*(0.04466168828589267+z*(0.030096185611744593+z*(0.02468817769822816+z*(0.00729188997896001+z*(0.03487368125810243)))))))
}

// Acos returns the arccosine, in radians, of a.
//
// Special case is:
//
//	x.Acos() = NaN if x < -1 or x > 1
func (a Float16) Acos() Float16 {
	ix := a &^ signMask16
	if ix > uvone16 { // |a| > 1, Inf, or NaN
		return NewFloat16NaN()
	}

	var x float64
	if ix < 0x0400 {
		// a is zero or subnormal.
		x = float64(ix) * 0x1p-24
	} else {
		x = normal16ToFloat64(ix)
	}
	var r float64
	if x <= 0.5 {
		// acos(x) = pi/2 - asin(x)
		s := x * asinSeries16(x*x)
		if a&signMask16 != 0 {
			s = -s
		}
		return NewFloat16(math.Pi/2 - s)
	}

	// acos(x) = 2 asin(sqrt((1-x)/2))
	t := (1 - x) * 0.5 // exact
	r = 2 * math.Sqrt(t) * asinSeries16(t)
	if a&signMask16 != 0 {
		// acos(-x) = pi - acos(x)
		r = math.Pi - r
	}
	return NewFloat16(r)
}

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float16) Atan() Float16 {
	ix := a &^ signMask16
	if ix < 0x2400 { // |a| < 2**-6
		// atan(a) = a - a**3/3 + ... rounds to a.
		return a
	}
	if ix >= uvinf16 {
		if ix > uvinf16 {
			// atan(NaN) = NaN
			return NewFloat16NaN()
		}
		// atan(±Inf) = ±Pi/2
		return a&signMask16 | NewFloat16(math.Pi/2)
	}

	// a is a normal Float16 value, so the conversion is exact.
	x := normal16ToFloat64(ix)
	var r float64
	if x <= 1 {
		r = atanKernel16(x)
	} else {
		// atan(x) = pi/2 - atan(1/x)
		r = math.Pi/2 - atanKernel16(1/x)
	}
	if a&signMask16 != 0 {
		r = -r
	}
	return NewFloat16(r)
}

// atanKernel16 returns atan(x) for 0 <= x <= 1.
func atanKernel16(x float64) float64 {
	// atan(x) = 2 atan(u), where u = x/(1+sqrt(1+x**2)) <= tan(pi/8).
	u := x / (1 + math.Sqrt(1+x*x))
	return 2 * u * atanSeries16(u*u)
}

// atanSeries16 returns atan(sqrt(z))/sqrt(z) for 0 <= z <= tan(pi/8)**2.
// It is a degree-6 minimax polynomial; the relative error is about 2e-11,
// far below the Float16 precision.
func atanSeries16(z float64) float64 {
	return 0.9999999999791293 + z*(-0.3333333212787915+z*(0.19999885898675399+z*(-0.1428163925942391+z*(0.11041054162292493+z*(-0.08459109187133435+z*(0.04712997339340858))))))
}

// Atan2 returns the arc tangent of a/b, using
// the signs of the two to determine the quadrant
// of the return value.
//
// Special cases are (in order):
//
//	y.Atan2(NaN) = NaN
//	NaN.Atan2(x) = NaN
//	+0.Atan2(x>=0) = +0
//	-0.Atan2(x>=0) = -0
//	+0.Atan2(x<=-0) = +Pi
//	-0.Atan2(x<=-0) = -Pi
//	y>0.Atan2(0) = +Pi/2
//	y<0.Atan2(0) = -Pi/2
//	+Inf.Atan2(+Inf) = +Pi/4
//	-Inf.Atan2(+Inf) = -Pi/4
//	+Inf.Atan2(-Inf) = 3Pi/4
//	-Inf.Atan2(-Inf) = -3Pi/4
//	y.Atan2(+Inf) = 0
//	(y>0).Atan2(-Inf) = +Pi
//	(y<0).Atan2(-Inf) = -Pi
//	+Inf.Atan2(x) = +Pi/2
//	-Inf.Atan2(x) = -Pi/2
func (a Float16) Atan2(b Float16) Float16 {
	iy := a &^ signMask16
	ix := b &^ signMask16
	if iy < 0x0400 || iy >= uvinf16 || ix < 0x0400 || ix >= uvinf16 {
		// zeros, subnormals, infinities and NaNs.
		return NewFloat16(math.Atan2(a.Float64().BuiltIn(), b.Float64().BuiltIn()))
	}

	// Both are normal Float16 values, so the conversions are exact.
	y := normal16ToFloat64(iy)
	x := normal16ToFloat64(ix)
	var r float64
	if y <= x {
		r = atanKernel16(y / x)
	} else {
		// atan(y/x) = pi/2 - atan(x/y)
		r = math.Pi/2 - atanKernel16(x/y)
	}
	if b&signMask16 != 0 {
		// atan2(y, -x) = pi - atan2(y, x)
		r = math.Pi - r
	}
	// The kernel has a relative error of about 2e-11. If r is that close to
	// a rounding boundary of Float16 (or the result is subnormal),
	// the rounding may be wrong, so fall back to the accurate path.
	const window = 1 << 19 // in units of the last place of float64
	frac := math.Float64bits(r) & (1<<42 - 1)
	if r < 0x1p-14 || frac-(1<<41)+window < 2*window {
		return NewFloat16(math.Atan2(a.Float64().BuiltIn(), b.Float64().BuiltIn()))
	}
	if a&signMask16 != 0 {
		r = -r
	}
	return NewFloat16(r)
}
