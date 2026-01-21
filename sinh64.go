package floats

import "math"

// Sinh returns the hyperbolic sine of x.
//
// Special cases are:
//
//	±0.Sinh() = ±0
//	±Inf.Sinh() = ±Inf
//	NaN.Sinh() = NaN
func (a Float64) Sinh() Float64 {
	return NewFloat64(math.Sinh(a.BuiltIn()))
}

// Cosh returns the hyperbolic cosine of x.
//
// Special cases are:
//
//	±0.Cosh() = 1
//	±Inf.Cosh() = +Inf
//	NaN.Cosh() = NaN
func (a Float64) Cosh() Float64 {
	return NewFloat64(math.Cosh(a.BuiltIn()))
}
