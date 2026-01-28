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

// Erfc returns the complementary error function of x.
//
// Special cases are:
//
//	+Inf.Erfc() = 0
//	-Inf.Erfc() = 2
//	NaN.Erfc() = NaN
func (a Float64) Erfc() Float64 {
	return NewFloat64(math.Erfc(a.BuiltIn()))
}

// Erfinv returns the inverse error function of a.
//
// Special cases are:
//
//	1.Erfinv() = +Inf
//	-1.Erfinv() = -Inf
//	x.Erfinv() = NaN if x < -1 or x > 1
//	NaN.Erfinv() = NaN
func (a Float64) Erfinv() Float64 {
	return NewFloat64(math.Erfinv(a.BuiltIn()))
}

// Erfcinv returns the inverse of [Erfc](a).
//
// Special cases are:
//
//	0.Erfcinv() = +Inf
//	2.Erfcinv() = -Inf
//	x.Erfcinv() = NaN if x < 0 or x > 2
//	NaN.Erfcinv() = NaN
func (a Float64) Erfcinv() Float64 {
	return NewFloat64(math.Erfcinv(a.BuiltIn()))
}
