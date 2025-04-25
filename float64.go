package floats

const (
	mask64     = 0x7ff       // mask for exponent
	shift64    = 64 - 11 - 1 // shift for exponent
	bias64     = 1023        // bias for exponent
	signMask64 = 1 << 63     // mask for sign bit
	fracMask64 = 1<<shift64 - 1
)

// Float64 is a 64-bit floating-point number.
type Float64 float64

// IsNaN reports whether a is an IEEE 754 “not-a-number” value.
func (a Float64) IsNaN() bool {
	return a != a
}
