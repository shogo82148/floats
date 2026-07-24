package floats

import "math"

// Jn returns the order-n Bessel function of the first kind.
//
// Special cases are:
//
//	Jn(n, ±Inf) = 0
//	Jn(n, NaN) = NaN
func (a Float16) Jn(n int) Float16 {
	return NewFloat16(math.Jn(n, a.Float64().BuiltIn()))
}
