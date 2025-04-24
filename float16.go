package floats

const (
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
