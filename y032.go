package floats

import "math"

// Y0 returns the order-zero Bessel function of the second kind.
//
// Special cases are:
//
//	Y0(+Inf) = 0
//	Y0(0) = -Inf
//	Y0(x < 0) = NaN
//	Y0(NaN) = NaN
func (a Float32) Y0() Float32 {
	return NewFloat32(math.Y0(a.Float64().BuiltIn()))
}
