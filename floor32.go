package floats

import "math"

// Floor returns the greatest integer value less than or equal to a.
//
// Special cases are:
//
//	±0.Floor() = ±0
//	±Inf.Floor() = ±Inf
//	NaN.Floor() = NaN
func (a Float32) Floor() Float32 {
	return NewFloat32(math.Floor(a.Float64().BuiltIn()))
}

// Ceil returns the least integer value greater than or equal to a.
//
// Special cases are:
//
//	±0.Ceil() = ±0
//	±Inf.Ceil() = ±Inf
//	NaN.Ceil() = NaN
func (a Float32) Ceil() Float32 {
	return NewFloat32(math.Ceil(a.Float64().BuiltIn()))
}

// Trunc returns the integer value of x.
//
// Special cases are:
//
//	±0.Trunc() = ±0
//	±Inf.Trunc() = ±Inf
//	NaN.Trunc() = NaN
func (a Float32) Trunc() Float32 {
	return NewFloat32(math.Trunc(a.Float64().BuiltIn()))
}
