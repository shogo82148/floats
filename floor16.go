package floats

import "math"

// Floor returns the greatest integer value less than or equal to a.
//
// Special cases are:
//
//	±0.Floor() = ±0
//	±Inf.Floor() = ±Inf
//	NaN.Floor() = NaN
func (a Float16) Floor() Float16 {
	return NewFloat16(math.Floor(a.Float64().BuiltIn()))
}

// Ceil returns the least integer value greater than or equal to a.
//
// Special cases are:
//
//	±0.Ceil() = ±0
//	±Inf.Ceil() = ±Inf
//	NaN.Ceil() = NaN
func (a Float16) Ceil() Float16 {
	return NewFloat16(math.Ceil(a.Float64().BuiltIn()))
}

// Trunc returns the integer value of a.
//
// Special cases are:
//
//	±0.Trunc() = ±0
//	±Inf.Trunc() = ±Inf
//	NaN.Trunc() = NaN
func (a Float16) Trunc() Float16 {
	return NewFloat16(math.Trunc(a.Float64().BuiltIn()))
}
