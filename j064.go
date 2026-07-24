package floats

import "math"

// J0 returns the order-zero Bessel function of the first kind.
//
// Special cases are:
//
//	J0(±Inf) = 0
//	J0(0) = 1
//	J0(NaN) = NaN
func (a Float64) J0() Float64 {
	return NewFloat64(math.J0(a.BuiltIn()))
}
