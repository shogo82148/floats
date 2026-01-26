package floats

import "math"

// Erf returns the error function of a.
//
// Special cases are:
//
//	+Inf.Erf() = 1
//	-Inf.Erf() = -1
//	NaN.Erf() = NaN
func (a Float16) Erf() Float16 {
	return NewFloat16(math.Erf(a.Float64().BuiltIn()))
}
