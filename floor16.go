package floats

import "math"

// Floor returns the greatest integer value less than or equal to x.
//
// Special cases are:
//
//	Floor(±0) = ±0
//	Floor(±Inf) = ±Inf
//	Floor(NaN) = NaN
func (a Float16) Floor() Float16 {
	return NewFloat16(math.Floor(a.Float64().BuiltIn()))
}
