package floats

import "math"

// Hypot returns [Sqrt](p*p + q*q), taking care to avoid
// unnecessary overflow and underflow.
//
// Special cases are:
//
//	Hypot(±Inf, q) = +Inf
//	Hypot(p, ±Inf) = +Inf
//	Hypot(NaN, q) = NaN
//	Hypot(p, NaN) = NaN
func Hypot64(p, q Float64) Float64 {
	return NewFloat64(math.Hypot(p.BuiltIn(), q.BuiltIn()))
}
