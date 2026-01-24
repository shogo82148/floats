package floats

import "math"

// NewFloat16Pow10 returns 10**n, the base-10 exponential of n.
//
// Special cases are:
//
//	NewFloat16Pow10(n) =    0 for n < -7
//	NewFloat16Pow10(n) = +Inf for n > 4
func NewFloat16Pow10(n int) Float16 {
	return NewFloat16(math.Pow10(n))
}
