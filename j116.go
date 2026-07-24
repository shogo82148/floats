package floats

import "math"

// J1 returns the order-one Bessel function of the first kind.
//
// Special cases are:
//
//	J1(±Inf) = 0
//	J1(NaN) = NaN
func (a Float16) J1() Float16 {
	return NewFloat16(math.J1(a.Float64().BuiltIn()))
}
