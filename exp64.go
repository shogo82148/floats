package floats

import "math"

// Exp returns e**x, the base-e exponential of a.
//
// Special cases are:
//
//	+Inf.Exp() = +Inf
//	NaN.Exp() = NaN
//
// Very large values overflow to 0 or +Inf.
// Very small values underflow to 1.
func (a Float64) Exp() Float64 {
	return NewFloat64(math.Exp(a.BuiltIn()))
}
