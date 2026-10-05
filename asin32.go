package floats

import "math"

// Asin returns the arcsine, in radians, of a.
//
// Special cases are:
//
//	±0.Asin() = ±0
//	x.Asin() = NaN if x < -1 or x > 1
func (a Float32) Asin() Float32 {
	ix := a.Bits() &^ signMask32
	if ix < 0x39800000 { // |a| < 2**-12
		// asin(a) = a + a**3/6 + ... rounds to a.
		return a
	}
	if ix > 0x3f800000 { // |a| > 1, Inf, or NaN
		return NewFloat32NaN()
	}

	x := float64(math.Float32frombits(ix))
	var r float64
	if x <= 0.5 {
		r = x * asinSeries32(x*x)
	} else {
		// asin(x) = pi/2 - 2 asin(sqrt((1-x)/2))
		t := (1 - x) * 0.5 // exact
		r = math.Pi/2 - 2*math.Sqrt(t)*asinSeries32(t)
	}
	if a < 0 {
		r = -r
	}
	return Float32(r)
}

// asinSeries32 returns asin(sqrt(z))/sqrt(z) for 0 <= z <= 1/4.
// It is a degree-12 minimax polynomial; the relative error is about 3e-16.
func asinSeries32(z float64) float64 {
	return 1.0 + z*(0.16666666666669364+z*(0.07499999999658685+z*(0.04464285730828498+z*(0.03038194106964922+z*(0.022372165977951806+z*(0.017353810830872268+z*(0.013941617419028588+z*(0.011808053853317318+z*(0.008059973313918046+z*(0.015399414933173018+z*(-0.009823178487300003+z*(0.027566224119240496))))))))))))
}

// Acos returns the arccosine, in radians, of a.
//
// Special case is:
//
//	x.Acos() = NaN if x < -1 or x > 1
func (a Float32) Acos() Float32 {
	return NewFloat32(math.Acos(a.Float64().BuiltIn()))
}

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float32) Atan() Float32 {
	return NewFloat32(math.Atan(a.Float64().BuiltIn()))
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
func (a Float32) Atan2(b Float32) Float32 {
	return NewFloat32(math.Atan2(a.Float64().BuiltIn(), b.Float64().BuiltIn()))
}
