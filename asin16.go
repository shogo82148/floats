package floats

import "math"

// Asin returns the arcsine, in radians, of a.
//
// Special cases are:
//
//	±0.Asin() = ±0
//	x.Asin() = NaN if x < -1 or x > 1
func (a Float16) Asin() Float16 {
	return NewFloat16(math.Asin(a.Float64().BuiltIn()))
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
