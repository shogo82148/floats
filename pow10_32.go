package floats

import "math"

// NewFloat32Pow10 returns 10**n, the base-10 exponential of n.
//
// Special cases are:
//
//	NewFloat32Pow10(n) =    0 for n < -45
//	NewFloat32Pow10(n) = +Inf for n > 38
func NewFloat32Pow10(n int) Float32 {
	return NewFloat32(math.Pow10(n))
}
