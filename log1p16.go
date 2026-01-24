package floats

import "math"

// Log1p returns the natural logarithm of 1 plus its argument x.
// It is more accurate than [Log](1 + x) when x is near zero.
//
// Special cases are:
//
//	+Inf.Log1p() = +Inf
//	±0.Log1p() = ±0
//	-1.Log1p() = -Inf
//	(x < -1).Log1p() = NaN
//	NaN.Log1p() = NaN
func (a Float16) Log1p() Float16 {
	return NewFloat16(math.Log1p(a.Float64().BuiltIn()))
}
