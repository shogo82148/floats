package floats

const (
	uvnan16    = 0x7e00     // NaN value for Float16
	uvinf16    = 0x7c00     // Infinity value for Float16
	uvneginf16 = 0xfc00     // Negative Infinity value for Float16
	uvone16    = 0x3c00     // One value for Float16
	mask16     = 0x1f       // mask for exponent
	shift16    = 16 - 5 - 1 // shift for exponent
	bias16     = 15         // bias for exponent
	signMask16 = 1 << 15    // mask for sign bit
	fracMask16 = 1<<shift16 - 1
)

// Float16 is a 16-bit floating-point number.
type Float16 uint16

// IsNaN reports whether a is an IEEE 754 “not-a-number” value.
func (a Float16) IsNaN() bool {
	return a&(mask16<<shift16) == (mask16<<shift16) && a&fracMask16 != 0
}

// Int64 returns the integer value of a, rounding towards zero.
// If a cannot be represented in an int64, the result is undefined.
func (a Float16) Int64() int64 {
	return int64(a.Float64())
}
