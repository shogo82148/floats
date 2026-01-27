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

// Erfc returns the complementary error function of x.
//
// Special cases are:
//
//	+Inf.Erfc() = 0
//	-Inf.Erfc() = 2
//	NaN.Erfc() = NaN
func (a Float16) Erfc() Float16 {
	return NewFloat16(math.Erfc(a.Float64().BuiltIn()))
}
