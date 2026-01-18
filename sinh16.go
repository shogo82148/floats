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
