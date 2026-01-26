package floats

import "math"

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float64) Atan() Float64 {
	return NewFloat64(math.Atan(a.BuiltIn()))
}
