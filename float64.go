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

// NewFloat64 converts f to Float64.
func NewFloat64(f float64) Float64 {
	return Float64(f)
}

// NewFloat64FromBits converts the IEEE 754 binary representation b to Float64.
func NewFloat64FromBits(b uint64) Float64 {
	return Float64(math.Float64frombits(b))
}

// Bits returns the IEEE 754 binary representation of a.
func (a Float64) Bits() uint64 {
	return math.Float64bits(float64(a))
}

// BuiltIn returns the built-in float64 value of a.
func (a Float64) BuiltIn() float64 {
	return float64(a)
}

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

// Signbit reports whether x is negative or negative zero.
func (a Float64) Signbit() bool {
	return math.Signbit(float64(a))
}

// Int64 returns the integer value of a, rounding towards zero.
// If a cannot be represented in an int64, the result is undefined.
func (a Float64) Int64() int64 {
	return int64(a)
}

// IsZero reports whether a is zero (+0 or -0).
func (a Float64) IsZero() bool {
	return a == 0
}

// Neg returns the negation of a.
func (a Float64) Neg() Float64 {
	return -a
}

// Abs returns the absolute value of a.
//
// Special cases:
//
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func (a Float64) Abs() Float64 {
	return NewFloat64FromBits(a.Bits() &^ signMask64)
}

// Mul returns the product of a and b.
func (a Float64) Mul(b Float64) Float64 {
	return a * b
}

// Quo returns the quotient of a and b.
func (a Float64) Quo(b Float64) Float64 {
	return a / b
}

// Add returns the sum of a and b.
func (a Float64) Add(b Float64) Float64 {
	return a + b
}

// Sub returns the difference of a and b.
func (a Float64) Sub(b Float64) Float64 {
	return a - b
}

// Sqrt returns the square root of a.
//
// Special cases are:
//
//	Sqrt(+Inf) = +Inf
//	Sqrt(±0) = ±0
//	Sqrt(x < 0) = NaN
//	Sqrt(NaN) = NaN
func (a Float64) Sqrt() Float64 {
	return Float64(math.Sqrt(float64(a)))
}

// Eq returns a == b.
// NaNs are not equal to anything, including NaN.
// -0.0 and 0.0 are equal.
func (a Float64) Eq(b Float64) bool {
	return a == b
}

// Ne returns a != b.
// NaNs are not equal to anything, including NaN.
// -0.0 and 0.0 are equal.
func (a Float64) Ne(b Float64) bool {
	return a != b
}

// Lt returns a < b.
//
// Special cases are:
//
//	Lt(NaN, x) == false
//	Lt(x, NaN) == false
func (a Float64) Lt(b Float64) bool {
	return a < b
}

// Gt returns a > b.
//
// Special cases are:
//
//	Gt(x, NaN) == false
//	Gt(NaN, x) == false
func (a Float64) Gt(b Float64) bool {
	return a > b
}

// Le returns a <= b.
//
// Special cases are:
//
//	Le(x, NaN) == false
//	Le(NaN, x) == false
func (a Float64) Le(b Float64) bool {
	return a <= b
}

// Ge returns a >= b.
//
// Special cases are:
//
//	Ge(x, NaN) == false
//	Ge(NaN, x) == false
func (a Float64) Ge(b Float64) bool {
	return a >= b
}

// FMA64 returns x * y + z, computed with only one rounding.
// (That is, FMA64 returns the fused multiply-add of x, y, and z.)
func FMA64(x, y, z Float64) Float64 {
	return Float64(math.FMA(float64(x), float64(y), float64(z)))
}

// Modf returns integer and fractional floating-point numbers
// that sum to f. Both values have the same sign as f.
//
// Special cases are:
//
//	Modf(±Inf) = ±Inf, NaN
//	Modf(NaN) = NaN, NaN
func (a Float64) Modf() (int Float64, frac Float64) {
	fint, ffrac := math.Modf(a.BuiltIn())
	return NewFloat64(fint), NewFloat64(ffrac)
}
