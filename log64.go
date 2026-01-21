package floats

import "math"

// Log returns the natural logarithm of a.
//
// Special cases are:
//
//	+Inf.Log() = +Inf
//	0.Log() = -Inf
//	(x < 0).Log() = NaN
//	NaN.Log() = NaN
func (a Float64) Log() Float64 {
	return NewFloat64(math.Log(a.BuiltIn()))
}
