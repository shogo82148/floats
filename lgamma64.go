package floats

import "math"

// Lgamma returns the natural logarithm and sign (-1 or +1) of Gamma(a).
//
// Special cases are:
//
//	Lgamma(+Inf) = +Inf
//	Lgamma(0) = +Inf
//	Lgamma(-integer) = +Inf
//	Lgamma(-Inf) = -Inf
//	Lgamma(NaN) = NaN
func (a Float64) Lgamma() (Float64, int) {
	lgamma, sign := math.Lgamma(a.BuiltIn())
	return NewFloat64(lgamma), sign
}
