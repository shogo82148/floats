package floats

import "math"

// Cbrt returns the cube root of a.
//
// Special cases are:
//
//	±0.Cbrt() = ±0
//	±Inf.Cbrt() = ±Inf
//	NaN.Cbrt() = NaN
func (a Float16) Cbrt() Float16 {
	return NewFloat16(math.Cbrt(a.Float64().BuiltIn()))
}
