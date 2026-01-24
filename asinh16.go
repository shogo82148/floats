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
