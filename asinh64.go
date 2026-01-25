package floats

import "math"

// Asinh returns the inverse hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Asinh() = ±0
//	±Inf.Asinh() = ±Inf
//	NaN.Asinh() = NaN
func (a Float64) Asinh() Float64 {
	return NewFloat64(math.Asinh(a.BuiltIn()))
}
