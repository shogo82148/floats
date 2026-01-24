package floats

import "math"

// Log1p returns the natural logarithm of 1 plus its argument a.
// It is more accurate than [Log](1 + a) when a is near zero.
//
// Special cases are:
//
//	+Inf.Log1p() = +Inf
//	±0.Log1p() = ±0
//	-1.Log1p() = -Inf
//	(a < -1).Log1p() = NaN
//	NaN.Log1p() = NaN
func (a Float64) Log1p() Float64 {
	return NewFloat64(math.Log1p(a.BuiltIn()))
}
