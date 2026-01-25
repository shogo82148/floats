package floats

import "math"

// Sin returns the sine of the radian argument a.
//
// Special cases are:
//
//	±0.Sin() = ±0
//	±Inf.Sin() = NaN
//	NaN.Sin() = NaN
func (a Float64) Sin() Float64 {
	return NewFloat64(math.Sin(a.BuiltIn()))
}

// Cos returns the cosine of the radian argument a.
//
// Special cases are:
//
//	±Inf.Cos() = NaN
//	NaN.Cos() = NaN
func (a Float64) Cos() Float64 {
	return NewFloat64(math.Cos(a.BuiltIn()))
}
