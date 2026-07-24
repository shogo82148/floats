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
func (a Float16) Y1() Float16 {
	return NewFloat16(math.Y1(a.Float64().BuiltIn()))
}
