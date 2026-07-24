package floats

import "math"

// J0 returns the order-zero Bessel function of the first kind.
//
// Special cases are:
//
//	J0(±Inf) = 0
//	J0(0) = 1
//	J0(NaN) = NaN
func (a Float32) J0() Float32 {
	return NewFloat32(math.J0(a.Float64().BuiltIn()))
}
