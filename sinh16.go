package floats

import "math"

// Sinh returns the hyperbolic sine of x.
//
// Special cases are:
//
//	±0.Sinh() = ±0
//	±Inf.Sinh() = ±Inf
//	NaN.Sinh() = NaN
func (a Float16) Sinh() Float16 {
	return NewFloat16(math.Sinh(a.Float64().BuiltIn()))
}

// Cosh returns the hyperbolic cosine of x.
//
// Special cases are:
//
//	±0.Cosh() = 1
//	±Inf.Cosh() = +Inf
//	NaN.Cosh() = NaN
func (a Float16) Cosh() Float16 {
	return NewFloat16(math.Cosh(a.Float64().BuiltIn()))
}
