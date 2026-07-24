package floats

import "math"

// Y1 returns the order-one Bessel function of the second kind.
//
// Special cases are:
//
//	Y1(+Inf) = 0
//	Y1(0) = -Inf
//	Y1(x < 0) = NaN
//	Y1(NaN) = NaN
func (a Float64) Y1() Float64 {
	return NewFloat64(math.Y1(a.BuiltIn()))
}
