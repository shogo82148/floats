package floats

import "math"

const (
	uvnan32    = 0x7fc00000 // NaN value for Float32
	uvinf32    = 0x7f800000 // Infinity value for Float32
	uvneginf32 = 0xff800000 // Negative Infinity value for Float32
	uvone32    = 0x3f800000 // One value for Float32
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

// IsInf reports whether a is an infinity, according to sign.
// If sign > 0, IsInf reports whether a is positive infinity.
// If sign < 0, IsInf reports whether a is negative infinity.
// If sign == 0, IsInf reports whether a is either infinity.
func (a Float32) IsInf(sign int) bool {
	return math.IsInf(float64(a), sign)
}

// Int64 returns the integer value of a, rounding towards zero.
// If a cannot be represented in an int64, the result is undefined.
func (a Float32) Int64() int64 {
	return int64(a)
}

// Neg returns the negation of a.
func (a Float32) Neg() Float32 {
	return -a
}

// Mul returns the product of a and b.
func (a Float32) Mul(b Float32) Float32 {
	return a * b
}

// Quo returns the quotient of a and b.
func (a Float32) Quo(b Float32) Float32 {
	return a / b
}

// Add returns the sum of a and b.
func (a Float32) Add(b Float32) Float32 {
	return a + b
}

// Sub returns the difference of a and b.
func (a Float32) Sub(b Float32) Float32 {
	return a - b
}

// Eq returns a == b.
// NaNs are not equal to anything, including NaN.
// -0.0 and 0.0 are equal.
func (a Float32) Eq(b Float32) bool {
	return a == b
}

// Ne returns a != b.
// NaNs are not equal to anything, including NaN.
// -0.0 and 0.0 are equal.
func (a Float32) Ne(b Float32) bool {
	return a != b
}

// Lt returns a < b.
//
// Special cases are:
//
//	Lt(NaN, x) == false
//	Lt(x, NaN) == false
func (a Float32) Lt(b Float32) bool {
	return a < b
}
