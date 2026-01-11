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
func (a Float16) Exp() Float16 {
	return NewFloat16(math.Exp(a.Float64().BuiltIn()))
}
