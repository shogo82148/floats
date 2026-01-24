package floats

import (
	"math"
	"math/bits"

	"github.com/shogo82148/ints"
)

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

// NewFloat64NaN returns a NaN Float64 value.
func NewFloat64NaN() Float64 {
	return Float64(math.NaN())
}

// NewFloat64Inf positive infinity if sign >= 0, negative infinity if sign < 0.
func NewFloat64Inf(sign int) Float64 {
	return Float64(math.Inf(sign))
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

// Copysign returns a value with the magnitude of a
// and the sign of sign.
func (a Float64) Copysign(sign Float64) Float64 {
	return NewFloat64FromBits((a.Bits() &^ signMask64) | (sign.Bits() & signMask64))
}

// Int64 returns the integer value of a, rounding towards zero.
// If a cannot be represented in an int64, the result is undefined.
func (a Float64) Int64() int64 {
	return int64(a)
}

// Uint64 returns the unsigned integer value of a, rounding towards zero.
// If a cannot be represented in a uint64, the result is undefined.
func (a Float64) Uint64() uint64 {
	return uint64(a)
}

// Int128 returns the signed 128-bit integer value of a, rounding towards zero.
// If a cannot be represented in a int128, the result is undefined.
func (a Float64) Int128() ints.Int128 {
	sign, exp, frac := a.normalize()
	frac128 := ints.Int128{0, frac}
	if exp <= shift64 {
		frac128 = frac128.Rsh(uint(shift64 - exp))
	} else {
		frac128 = frac128.Lsh(uint(exp - shift64))
	}
	if sign != 0 {
		frac128 = frac128.Neg()
	}
	return frac128
}

// Uint128 returns the unsigned 128-bit integer value of a, rounding towards zero.
// If a cannot be represented in a uint128, the result is undefined.
func (a Float64) Uint128() ints.Uint128 {
	_, exp, frac := a.normalize()
	frac128 := ints.Uint128{0, frac}
	if exp <= shift64 {
		frac128 = frac128.Rsh(uint(shift64 - exp))
	} else {
		frac128 = frac128.Lsh(uint(exp - shift64))
	}
	return frac128
}

// Int256 returns the signed 256-bit integer value of a, rounding towards zero.
// If a cannot be represented in a int256, the result is undefined.
func (a Float64) Int256() ints.Int256 {
	sign, exp, frac := a.normalize()
	frac256 := ints.Int256{0, 0, 0, frac}
	if exp <= shift64 {
		frac256 = frac256.Rsh(uint(shift64 - exp))
	} else {
		frac256 = frac256.Lsh(uint(exp - shift64))
	}
	if sign != 0 {
		frac256 = frac256.Neg()
	}
	return frac256
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

func (a Float64) normalize() (sign uint64, exp int, frac uint64) {
	b := math.Float64bits(float64(a))
	sign = b & signMask64
	exp = int((b>>shift64)&mask64) - bias64
	frac = b & fracMask64

	if exp == -bias64 {
		// a is subnormal
		// normalize
		l := bits.Len64(frac)
		frac <<= uint(shift64 + 1 - l)
		exp = l - (bias64 + shift64)
		return
	}

	// a is normal
	frac |= 1 << shift64
	return
}

// FMA64 returns x * y + z, computed with only one rounding.
// (That is, FMA64 returns the fused multiply-add of x, y, and z.)
func FMA64(x, y, z Float64) Float64 {
	return Float64(math.FMA(float64(x), float64(y), float64(z)))
}

// Nextafter returns the next representable float64 value after a towards b.
//
// Special cases are:
//
//	a.Nextafter(a)   = a
//	NaN.Nextafter(b) = NaN
//	a.Nextafter(NaN) = NaN
func (a Float64) Nextafter(b Float64) (r Float64) {
	return Float64(math.Nextafter(a.BuiltIn(), b.BuiltIn()))
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

// Frexp breaks a into a normalized fraction
// and an integral power of two.
// It returns frac and exp satisfying f == frac × 2**exp,
// with the absolute value of frac in the interval [½, 1).
//
// Special cases are:
//
//	±0.Frexp() = ±0, 0
//	±Inf.Frexp() = ±Inf, 0
//	NaN.Frexp() = NaN, 0
func (a Float64) Frexp() (frac Float64, exp int) {
	f, e := math.Frexp(a.BuiltIn())
	return NewFloat64(f), e
}

// Ldexp is the inverse of [Frexp].
// It returns a × 2**exp.
//
// Special cases are:
//
//	±0.Ldexp(exp) = ±0
//	±Inf.Ldexp(exp) = ±Inf
//	NaN.Ldexp(exp) = NaN
func (a Float64) Ldexp(exp int) Float64 {
	f := math.Ldexp(a.BuiltIn(), exp)
	return NewFloat64(f)
}

// Mod returns the floating-point remainder of a/b.
// The magnitude of the result is less than b and its
// sign agrees with that of a.
//
// Special cases are:
//
//	±Inf.Mod(b) = NaN
//	NaN.Mod(b) = NaN
//	a.Mod(0) = NaN
//	a.Mod(±Inf) = a
//	a.Mod(NaN) = NaN
func (a Float64) Mod(b Float64) Float64 {
	return NewFloat64(math.Mod(a.BuiltIn(), b.BuiltIn()))
}

// Remainder returns the IEEE 754 floating-point remainder of a/b.
//
// Special cases are:
//
//	±Inf.Remainder(b) = NaN
//	NaN.Remainder(b) = NaN
//	a.Remainder(0) = NaN
//	a.Remainder(±Inf) = a
//	a.Remainder(NaN) = NaN
func (a Float64) Remainder(b Float64) Float64 {
	return NewFloat64(math.Remainder(a.BuiltIn(), b.BuiltIn()))
}
