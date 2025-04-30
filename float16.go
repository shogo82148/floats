package floats

import (
	"math/bits"
)

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

// IsInf reports whether a is an infinity, according to sign.
// If sign > 0, IsInf reports whether a is positive infinity.
// If sign < 0, IsInf reports whether a is negative infinity.
// If sign == 0, IsInf reports whether a is either infinity.
func (a Float16) IsInf(sign int) bool {
	return sign >= 0 && a == uvinf16 || sign <= 0 && a == uvneginf16
}

// Int64 returns the integer value of a, rounding towards zero.
// If a cannot be represented in an int64, the result is undefined.
func (a Float16) Int64() int64 {
	return int64(a.Float64())
}

func (a Float16) isZero() bool {
	return a&^signMask16 == 0
}

// Neg returns the negation of a.
func (a Float16) Neg() Float16 {
	return a ^ signMask16
}

// Mul returns the product of a and b.
func (a Float16) Mul(b Float16) Float16 {
	if a.IsNaN() || b.IsNaN() {
		// a * NaN = NaN
		// NaN * b = NaN
		return uvnan16
	}
	signA, expA, fracA := a.split()
	signB, expB, fracB := b.split()

	// special cases
	if expA == mask16-bias16 {
		// NaN check is done above; a is ±inf
		if b.isZero() {
			// ±inf * 0 = NaN
			return uvnan16
		} else {
			// ±inf * +finite = ±inf
			// ±inf * -finite = ∓inf
			return a ^ Float16(signB)
		}
	}
	if expB == mask16-bias16 {
		// NaN check is done above; b is ±inf
		if a.isZero() {
			// 0 * ±inf = NaN
			return uvnan16
		} else {
			// +finite * ±inf = ±inf
			// -finite * ±inf = ∓inf
			return b ^ Float16(signA)
		}
	}

	sign := signA ^ signB
	exp := expA + expB
	frac := uint32(fracA) * uint32(fracB)
	shift := bits.Len32(frac) - (shift16 + 1)
	exp += shift - shift16

	if exp < -(bias16 + shift16) {
		// underflow
		return Float16(sign)
	} else if exp <= -bias16 {
		// the result is subnormal
		shift := shift16 - (expA + expB + bias16) + 1
		frac += (1<<(shift-1) - 1) + ((frac >> shift) & 1) // round to nearest even
		frac >>= shift
		return Float16(sign | uint16(frac))
	}

	exp = expA + expB + bias16
	frac += (1<<(shift-1) - 1) + ((frac >> shift) & 1) // round to nearest even
	shift = bits.Len32(frac) - (shift16 + 1)
	exp += shift - shift16
	if exp >= mask16 {
		// overflow
		return Float16(sign | (mask16 << shift16))
	}
	frac >>= shift
	frac &= fracMask16
	return Float16(sign | uint16(exp<<shift16) | uint16(frac))
}

// Quo returns the quotient of a and b.
func (a Float16) Quo(b Float16) Float16 {
	if a.IsNaN() {
		// NaN / anything = NaN
		return a
	}
	if b.IsNaN() {
		// anything / NaN = NaN
		return b
	}

	signA, expA, fracA := a.split()
	signB, expB, fracB := b.split()
	sign := signA ^ signB

	if b.isZero() {
		if a.isZero() {
			// 0 / 0 = NaN
			return uvnan16
		}
		// ±finite / 0 = ±inf
		return Float16(sign | uvinf16)
	}
	if a.isZero() {
		// 0 / ±finite = 0
		return Float16(sign)
	}
	if expA == mask16-bias16 {
		// NaN check is done above; a is ±inf
		if expB == mask16-bias16 {
			// ±inf / ±inf = NaN
			return uvnan16
		} else {
			// ±inf / finite = ±inf
			return Float16(sign | uvinf16)
		}
	}
	if expB == mask16-bias16 {
		// NaN check is done above; b is ±inf
		// NaN and Inf checks are done above; a is finite.
		// ±finite / ±inf = 0
		return Float16(sign)
	}

	exp := expA - expB + bias16
	if fracA < fracB {
		exp--
		fracA <<= 1
	}
	if exp >= mask16 {
		// overflow
		return Float16(sign | uvinf16)
	}

	shift := shift16 + 3 // 1 for the implicit bit, 1 for the rounding bit, 1 for the guard bit
	fracA32 := uint32(fracA) << shift
	frac := uint16(fracA32 / uint32(fracB))
	mod := uint16(fracA32 % uint32(fracB))
	frac |= squash16(mod)
	if exp <= 0 {
		// the result is subnormal
		shift := -exp + 3 + 1
		frac += (1<<(shift-1) - 1) + ((frac >> shift) & 1) // round to nearest even
		frac >>= shift
		return Float16(sign | uint16(frac))
	}

	frac += 0b11 + ((frac >> 3) & 1) // round to nearest even
	frac >>= 3
	return Float16(sign | uint16(exp)<<shift16 | frac&fracMask16)
}

// Add returns the sum of a and b.
func (a Float16) Add(b Float16) Float16 {
	if a.IsNaN() {
		// NaN + anything = NaN
		return a
	}
	if b.IsNaN() {
		// anything + NaN = NaN
		return b
	}
	if a.isZero() {
		if b.isZero() {
			//  0 +  0 =  0
			//  0 + -0 =  0
			// -0 +  0 =  0
			// -0 + -0 = -0
			return a & b
		}
		// ±0 + b = b
		return b
	}
	if b.isZero() {
		// a + ±0 = a
		return a
	}

	signA, expA, fracA := a.split()
	signB, expB, fracB := b.split()

	// handle special cases
	if expA == mask16-bias16 {
		// NaN check is done above; a is ±inf
		if expB == mask16-bias16 {
			// NaN check is done above; b is ±inf
			if signA == signB {
				// ±inf + ±inf = ±inf
				return Float16(signA | uvinf16)
			}
			// ±inf + ∓inf = NaN
			return uvnan16
		}
		// b is finite, the result is ±inf
		return a
	}
	if expB == mask16-bias16 {
		// NaN check is done above; b is ±inf
		// NaN and Inf checks are done above; a is finite.
		return b
	}

	if expA < expB {
		// swap a and b
		signA, signB = signB, signA
		expA, expB = expB, expA
		fracA, fracB = fracB, fracA
	}

	// add the fractions
	const offset = 16
	fracA32 := int32(fracA) << offset
	fracB32 := int32(fracB) << offset
	fracB32 >>= uint(expA - expB)
	if signA != 0 {
		fracA32 = -fracA32
	}
	if signB != 0 {
		fracB32 = -fracB32
	}
	frac32 := fracA32 + fracB32
	sign := uint16(0)
	if frac32 < 0 {
		sign = signMask16
		frac32 = -frac32
	}

	shift := bits.Len32(uint32(frac32)) - shift16 - 1
	exp := expA + shift - offset

	// normalize
	if frac32 == 0 || exp < -(bias16+shift16) {
		// underflow
		return Float16(sign)
	}
	if exp <= -bias16 {
		// the result is subnormal
		shift := offset - (expA + bias16) + 1
		frac32 += (1<<uint(shift-1) - 1) + ((frac32 >> uint(shift)) & 1) // round to nearest even
		frac := uint16(frac32 >> shift)
		return Float16(sign | uint16(frac))
	}
	if exp >= mask16-bias16 {
		// overflow
		return Float16(sign | (mask16 << shift16))
	}

	frac32 += (1<<uint(shift-1) - 1) + ((frac32 >> uint(shift)) & 1) // round to nearest even
	if bits.Len32(uint32(frac32)) > shift16+shift+1 {
		frac32 >>= 1
		exp++
		if exp >= mask16 {
			// overflow
			return Float16(sign | (mask16 << shift16))
		}
	}
	frac := uint16(frac32 >> shift)
	return Float16(sign | uint16(exp+bias16)<<shift16 | frac&fracMask16)
}

// Sub returns the difference of a and b.
func (a Float16) Sub(b Float16) Float16 {
	return a.Add(b.Neg())
}

func (a Float16) split() (sign uint16, exp int, frac uint16) {
	sign = uint16(a & signMask16)
	exp = int((a>>shift16)&mask16) - bias16
	frac = uint16(a & fracMask16)

	if exp == -bias16 {
		// a is subnormal
		// normalize
		l := bits.Len16(frac)
		frac <<= uint(shift16 + 1 - l)
		exp = l - (bias16 + shift16)
		return
	}

	// a is normal
	frac |= 1 << shift16
	return
}
