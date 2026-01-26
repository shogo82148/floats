package floats

import "math"

// Asin returns the arcsine, in radians, of a.
//
// Special cases are:
//
//	±0.Asin() = ±0
//	x.Asin() = NaN if x < -1 or x > 1
func (a Float16) Asin() Float16 {
	return NewFloat16(math.Asin(a.Float64().BuiltIn()))
}

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float16) Atan() Float16 {
	return NewFloat16(math.Atan(a.Float64().BuiltIn()))
}
