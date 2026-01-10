package floats

import "math"

// Floor returns the greatest integer value less than or equal to x.
//
// Special cases are:
//
//	Floor(±0) = ±0
//	Floor(±Inf) = ±Inf
//	Floor(NaN) = NaN
func (a Float32) Floor() Float32 {
	return NewFloat32(math.Floor(a.Float64().BuiltIn()))
}
