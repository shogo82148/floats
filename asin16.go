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
	return NewFloat16(math.Acos(a.Float64().BuiltIn()))
}

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float16) Atan() Float16 {
	return NewFloat16(math.Atan(a.Float64().BuiltIn()))
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
	return NewFloat16(math.Atan2(a.Float64().BuiltIn(), b.Float64().BuiltIn()))
}
