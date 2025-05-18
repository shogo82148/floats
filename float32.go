package floats

import (
	"math"
	"math/bits"
)

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

// Sqrt returns the square root of a.
//
// Special cases are:
//
//	Sqrt(+Inf) = +Inf
//	Sqrt(±0) = ±0
//	Sqrt(x < 0) = NaN
//	Sqrt(NaN) = NaN
func (a Float32) Sqrt() Float32 {
	// This operation involves two rounds of rounding, so it is technically incorrect.
	// However, it always returns the correct result in practice.
	return Float32(math.Sqrt(float64(a)))
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

// Gt returns a > b.
//
// Special cases are:
//
//	Gt(x, NaN) == false
//	Gt(NaN, x) == false
func (a Float32) Gt(b Float32) bool {
	return a > b
}

// Le returns a <= b.
//
// Special cases are:
//
//	Le(x, NaN) == false
//	Le(NaN, x) == false
func (a Float32) Le(b Float32) bool {
	return a <= b
}

// Ge returns a >= b.
//
// Special cases are:
//
//	Ge(x, NaN) == false
//	Ge(NaN, x) == false
func (a Float32) Ge(b Float32) bool {
	return a >= b
}

func (a Float32) split() (sign uint32, exp int, frac uint32) {
	b := math.Float32bits(float32(a))
	sign = b & signMask32
	exp = int((b>>shift32)&mask32) - bias32
	frac = b & fracMask32

	if exp == -bias32 {
		// a is subnormal
		// normalize
		l := bits.Len32(frac)
		frac <<= uint(shift32 + 1 - l)
		exp = l - (bias32 + shift32)
		return
	}

	// a is normal
	frac |= 1 << shift32
	return
}

// FMA32 returns x * y + z, computed with only one rounding.
// (That is, FMA32 returns the fused multiply-add of x, y, and z.)
func FMA32(x, y, z Float32) Float32 {
	// Split x, y, z into sign, exponent, mantissa.
	signX, expX, fracX := x.split()
	signY, expY, fracY := y.split()
	signZ, expZ, fracZ0 := z.split()

	// Inf or NaN involved. At most one rounding will occur.
	if x == 0 || y == 0 || expX == mask32-bias32 || expY == mask32-bias32 {
		return x*y + z
	}
	if z == 0 {
		return x * y
	}
	// Handle non-finite z separately. Evaluating x*y+z where
	// x and y are finite, but z is infinite, should always result in z.
	if expZ == mask32-bias32 {
		return z
	}

	// Compute product p = x*y as sign, exponent, mantissa.
	expP := expX + expY + 1
	fracP := uint64(fracX<<7) * uint64(fracY<<8)
	signP := signX ^ signY // product sign

	// Normalize product.
	is62zero := uint((^fracP >> 62) & 1)
	fracP <<= is62zero
	expP -= int(is62zero)

	fracZ := uint64(fracZ0) << (7 + 32)

	// Swap addition operands so |p| >= |z|
	if expP < expZ || expP == expZ && fracP < fracZ {
		signP, signZ = signZ, signP
		expP, expZ = expZ, expP
		fracP, fracZ = fracZ, fracP
	}

	// Special case: if p == -z the result is always +0 since neither operand is zero.
	if signP != signZ && expP == expZ && fracP == fracZ {
		return 0
	}

	// Align mantissa
	fracZ = shrcompress64(fracZ, uint(expP-expZ))

	// Compute resulting significands, normalizing if necessary.
	var frac uint32
	if signP == signZ {
		// Adding fracP + fracZ
		fracP += fracZ
		expP += int(fracP >> 63)
		frac = uint32(shrcompress64(fracP, uint(32+fracP>>63)))
	} else {
		// Subtracting fracP - fracZ
		fracP -= fracZ
		nz := bits.LeadingZeros64(fracP) - 1
		expP -= nz
		frac = uint32(shrcompress64(fracP<<uint(nz), 32))
	}

	// check for underflow
	expP += bias32
	if expP <= 0 {
		n := uint(1 - expP)
		frac = roundToNearestEven32(frac, n+7)
		return Float32(math.Float32frombits(signP | frac))
	}

	// Round and break ties to even
	frac = roundToNearestEven32(frac, 7)
	if frac&(1<<(shift32+1)) != 0 {
		expP++
		frac >>= 1
	}
	if expP >= mask32 {
		// overflow
		return Float32(math.Float32frombits(signP | uvinf32))
	}
	return Float32(math.Float32frombits(signP | uint32(expP<<shift32) | frac&fracMask32))
}
