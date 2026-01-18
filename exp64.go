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

// Exp2 returns 2**x, the base-2 exponential of x.
//
// Special cases are the same as [Exp].
func (a Float64) Exp2() Float64 {
	return NewFloat64(math.Exp2(a.BuiltIn()))
}
