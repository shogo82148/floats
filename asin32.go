package floats

import "math"

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float32) Atan() Float32 {
	return NewFloat32(math.Atan(a.Float64().BuiltIn()))
}
