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

// Trunc returns the integer value of a.
//
// Special cases are:
//
//	±0.Trunc() = ±0
//	±Inf.Trunc() = ±Inf
//	NaN.Trunc() = NaN
func (a Float64) Trunc() Float64 {
	return NewFloat64(math.Trunc(a.BuiltIn()))
}

// Round returns the nearest integer, rounding half away from zero.
//
// Special cases are:
//
//	±0.Round() = ±0
//	±Inf.Round() = ±Inf
//	NaN.Round() = NaN
func (a Float64) Round() Float64 {
	return NewFloat64(math.Round(a.BuiltIn()))
}

// RoundToEven returns the nearest integer, rounding ties to even.
//
// Special cases are:
//
//	±0.RoundToEven() = ±0
//	±Inf.RoundToEven() = ±Inf
//	NaN.RoundToEven() = NaN
func (a Float64) RoundToEven() Float64 {
	return NewFloat64(math.RoundToEven(a.BuiltIn()))
}
