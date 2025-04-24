package floats

const (
	mask32     = 0xff       // mask for exponent
	shift32    = 32 - 8 - 1 // shift for exponent
	bias32     = 127        // bias for exponent
	signMask32 = 1 << 31    // mask for sign bit
	fracMask32 = 1<<shift32 - 1
)

// Float32 is a 32-bit floating-point number.
type Float32 float32

// IsNaN reports whether a is an IEEE 754 “not-a-number” value.
func (a Float32) IsNaN() bool {
	return a != a
}
