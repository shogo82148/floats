package floats

import "math"

// Asinh returns the inverse hyperbolic sine of x.
//
// Special cases are:
//
//	Asinh(±0) = ±0
//	Asinh(±Inf) = ±Inf
//	Asinh(NaN) = NaN
func (a Float64) Asinh() Float64 {
	return NewFloat64(math.Asinh(a.BuiltIn()))
}
