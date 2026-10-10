package floats

// Float8E4M3 is an 8-bit floating-point number in the E4M3 format
// (the "E4M3FN" variant of the OCP 8-bit Floating Point Specification).
// It has 1 sign bit, 4 exponent bits (bias 7) and 3 fraction bits.
//
// Unlike the IEEE 754 formats, it has no infinities;
// the only NaN encodings are 0x7f and 0xff, and the largest finite value is 448.
// The conversions to Float8E4M3 round to nearest even,
// and a value whose magnitude is too large to be represented, including an infinity,
// is converted to NaN.
type Float8E4M3 uint8

// Float8E5M2 is an 8-bit floating-point number in the E5M2 format
// (the "E5M2" format of the OCP 8-bit Floating Point Specification).
// It has 1 sign bit, 5 exponent bits (bias 15) and 2 fraction bits.
//
// It follows the conventions of IEEE 754, and it is the upper 8 bits of Float16;
// the largest finite value is 57344.
// The conversions to Float8E5M2 round to nearest even,
// and a value whose magnitude is too large to be represented is converted to infinity.
type Float8E5M2 uint8

// NewFloat8E4M3 converts f to Float8E4M3.
func NewFloat8E4M3(f float64) Float8E4M3 {
	return Float64(f).Float8E4M3()
}

// NewFloat8E5M2 converts f to Float8E5M2.
func NewFloat8E5M2(f float64) Float8E5M2 {
	return Float64(f).Float8E5M2()
}
