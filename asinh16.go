package floats

import "math"

// Asinh returns the inverse hyperbolic sine of x.
//
// Special cases are:
//
//	Asinh(±0) = ±0
//	Asinh(±Inf) = ±Inf
//	Asinh(NaN) = NaN
func (a Float16) Asinh() Float16 {
	return NewFloat16(math.Asinh(a.Float64().BuiltIn()))
}
