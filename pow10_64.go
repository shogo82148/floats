package floats

import "math"

// NewFloat64Pow10 returns 10**n, the base-10 exponential of n.
//
// Special cases are:
//
//	NewFloat64Pow10(n) =    0 for n < -323
//	NewFloat64Pow10(n) = +Inf for n > 308
func NewFloat64Pow10(n int) Float64 {
	return NewFloat64(math.Pow10(n))
}
