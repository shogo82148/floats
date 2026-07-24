package floats

import "math"

// Jn returns the order-n Bessel function of the first kind.
//
// Special cases are:
//
//	Jn(n, ±Inf) = 0
//	Jn(n, NaN) = NaN
func (a Float64) Jn(n int) Float64 {
	return NewFloat64(math.Jn(n, a.BuiltIn()))
}
