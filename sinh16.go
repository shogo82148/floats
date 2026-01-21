package floats

import "math"

// Sinh returns the hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Sinh() = ±0
//	±Inf.Sinh() = ±Inf
//	NaN.Sinh() = NaN
func (a Float16) Sinh() Float16 {
	return NewFloat16(math.Sinh(a.Float64().BuiltIn()))
}

// Cosh returns the hyperbolic cosine of a.
//
// Special cases are:
//
//	±0.Cosh() = 1
//	±Inf.Cosh() = +Inf
//	NaN.Cosh() = NaN
func (a Float16) Cosh() Float16 {
	return NewFloat16(math.Cosh(a.Float64().BuiltIn()))
}

// Tanh returns the hyperbolic tangent of a.
//
// Special cases are:
//
//	±0.Tanh() = ±0
//	±Inf.Tanh() = ±1
//	NaN.Tanh() = NaN
func (a Float16) Tanh() Float16 {
	return NewFloat16(math.Tanh(a.Float64().BuiltIn()))
}
