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
func (a Float16) Lgamma() (Float16, int) {
	lgamma, sign := math.Lgamma(a.Float64().BuiltIn())
	return NewFloat16(lgamma), sign
}
