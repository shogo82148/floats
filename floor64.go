package floats

import "math"

// Floor returns the greatest integer value less than or equal to a.
//
// Special cases are:
//
//	±0.Floor() = ±0
//	±Inf.Floor() = ±Inf
//	NaN.Floor() = NaN
func (a Float64) Floor() Float64 {
	return NewFloat64(math.Floor(a.BuiltIn()))
}

// Ceil returns the least integer value greater than or equal to a.
//
// Special cases are:
//
//	±0.Ceil() = ±0
//	±Inf.Ceil() = ±Inf
//	NaN.Ceil() = NaN
func (a Float64) Ceil() Float64 {
	return NewFloat64(math.Ceil(a.BuiltIn()))
}

// Trunc returns the integer value of x.
//
// Special cases are:
//
//	±0.Trunc() = ±0
//	±Inf.Trunc() = ±Inf
//	NaN.Trunc() = NaN
func (a Float64) Trunc() Float64 {
	return NewFloat64(math.Trunc(a.BuiltIn()))
}
