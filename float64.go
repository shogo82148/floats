package floats

import "math"

const (
	uvnan64    = 0x7ff8000000000000 // NaN value for Float64
	uvinf64    = 0x7ff0000000000000 // Infinity value for Float64
	uvneginf64 = 0xfff0000000000000 // Negative Infinity value for Float64
	uvone64    = 0x3ff0000000000000 // One value for Float64
	mask64     = 0x7ff              // mask for exponent
	shift64    = 64 - 11 - 1        // shift for exponent
	bias64     = 1023               // bias for exponent
	signMask64 = 1 << 63            // mask for sign bit
	fracMask64 = 1<<shift64 - 1
)

// Float64 is a 64-bit floating-point number.
type Float64 float64

// IsNaN reports whether a is an IEEE 754 “not-a-number” value.
func (a Float64) IsNaN() bool {
	return a != a
}

// IsInf reports whether a is an infinity, according to sign.
// If sign > 0, IsInf reports whether a is positive infinity.
// If sign < 0, IsInf reports whether a is negative infinity.
// If sign == 0, IsInf reports whether a is either infinity.
func (a Float64) IsInf(sign int) bool {
	return math.IsInf(float64(a), sign)
}

// Int64 returns the integer value of a, rounding towards zero.
// If a cannot be represented in an int64, the result is undefined.
func (a Float64) Int64() int64 {
	return int64(a)
}
