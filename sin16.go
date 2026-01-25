package floats

import "math"

// Sin returns the sine of the radian argument a.
//
// Special cases are:
//
//	±0.Sin() = ±0
//	±Inf.Sin() = NaN
//	NaN.Sin() = NaN
func (a Float16) Sin() Float16 {
	return NewFloat16(math.Sin(a.Float64().BuiltIn()))
}

// Cos returns the cosine of the radian argument a.
//
// Special cases are:
//
//	±Inf.Cos() = NaN
//	NaN.Cos() = NaN
func (a Float16) Cos() Float16 {
	return NewFloat16(math.Cos(a.Float64().BuiltIn()))
}

// Sincos returns Sin(a), Cos(a).
//
// Special cases are:
//
//	±0.Sincos() = ±0, 1
//	±Inf.Sincos() = NaN, NaN
//	NaN.Sincos() = NaN, NaN
func (a Float16) Sincos() (sin, cos Float16) {
	s, c := math.Sincos(a.Float64().BuiltIn())
	return NewFloat16(s), NewFloat16(c)
}
