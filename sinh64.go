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
