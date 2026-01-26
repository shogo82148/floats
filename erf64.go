package floats

import "math"

// Erf returns the error function of a.
//
// Special cases are:
//
//	+Inf.Erf() = 1
//	-Inf.Erf() = -1
//	NaN.Erf() = NaN
func (a Float64) Erf() Float64 {
	return NewFloat64(math.Erf(a.BuiltIn()))
}
