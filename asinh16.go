package floats

import "math"

// Asinh returns the inverse hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Asinh() = ±0
//	±Inf.Asinh() = ±Inf
//	NaN.Asinh() = NaN
func (a Float16) Asinh() Float16 {
	return NewFloat16(math.Asinh(a.Float64().BuiltIn()))
}

// Acosh returns the inverse hyperbolic cosine of a.
//
// Special cases are:
//
//	+Inf.Acosh() = +Inf
//	x.Acosh() = NaN if x < 1
//	NaN.Acosh() = NaN
func (a Float16) Acosh() Float16 {
	return NewFloat16(math.Acosh(a.Float64().BuiltIn()))
}
