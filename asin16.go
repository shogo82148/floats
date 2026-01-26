package floats

import "math"

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float16) Atan() Float16 {
	return NewFloat16(math.Atan(a.Float64().BuiltIn()))
}
