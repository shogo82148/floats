package floats

import "math"

// Floor returns the greatest integer value less than or equal to x.
//
// Special cases are:
//
//	Floor(±0) = ±0
//	Floor(±Inf) = ±Inf
//	Floor(NaN) = NaN
func (a Float64) Floor() Float64 {
	return NewFloat64(math.Floor(a.BuiltIn()))
}
