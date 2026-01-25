package floats

import (
	"math"
)

// Sin returns the sine of the radian argument a.
//
// Special cases are:
//
//	±0.Sin() = ±0
//	±Inf.Sin() = NaN
//	NaN.Sin() = NaN
func (a Float32) Sin() Float32 {
	return NewFloat32(math.Sin(a.Float64().BuiltIn()))
}

// Cos returns the cosine of the radian argument a.
//
// Special cases are:
//
//	±Inf.Cos() = NaN
//	NaN.Cos() = NaN
func (a Float32) Cos() Float32 {
	return NewFloat32(math.Cos(a.Float64().BuiltIn()))
}

// Sincos returns Sin(a), Cos(a).
//
// Special cases are:
//
//	±0.Sincos() = ±0, 1
//	±Inf.Sincos() = NaN, NaN
//	NaN.Sincos() = NaN, NaN
func (a Float32) Sincos() (sin, cos Float32) {
	s, c := math.Sincos(a.Float64().BuiltIn())
	return NewFloat32(s), NewFloat32(c)
}

// Tan returns the tangent of the radian argument a.
//
// Special cases are:
//
//	±0.Tan() = ±0
//	±Inf.Tan() = NaN
//	NaN.Tan() = NaN
func (a Float32) Tan() Float32 {
	return NewFloat32(math.Tan(a.Float64().BuiltIn()))
}
