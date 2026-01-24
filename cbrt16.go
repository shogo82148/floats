package floats

import "math"

// Cbrt returns the cube root of x.
//
// Special cases are:
//
//	Cbrt(±0) = ±0
//	Cbrt(±Inf) = ±Inf
//	Cbrt(NaN) = NaN
func (a Float16) Cbrt() Float16 {
	return NewFloat16(math.Cbrt(a.Float64().BuiltIn()))
}
