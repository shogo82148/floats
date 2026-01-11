package floats

import "math"

// Hypot64 returns [Sqrt](p*p + q*q), taking care to avoid
// unnecessary overflow and underflow.
//
// Special cases are:
//
//	Hypot64(±Inf, q) = +Inf
//	Hypot64(p, ±Inf) = +Inf
//	Hypot64(NaN, q) = NaN
//	Hypot64(p, NaN) = NaN
func Hypot64(p, q Float64) Float64 {
	return NewFloat64(math.Hypot(p.BuiltIn(), q.BuiltIn()))
}
