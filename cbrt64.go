package floats

import "math"

// Cbrt returns the cube root of a.
//
// Special cases are:
//
//	±0.Cbrt() = ±0
//	±Inf.Cbrt() = ±Inf
//	NaN.Cbrt() = NaN
func (a Float64) Cbrt() Float64 {
	return Float64(math.Cbrt(float64(a)))
}
