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

// NewFloat16 converts f to Float16.
func NewFloat16(f float64) Float16 {
	return Float64(f).Float16()
}

// Bits returns the IEEE 754 binary representation of a.
func (a Float16) Bits() uint16 {
	return uint16(a)
}

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

// Signbit reports whether x is negative or negative zero.
func (a Float16) Signbit() bool {
	return a&signMask16 != 0
}

// Int64 returns the integer value of a, rounding towards zero.
// If a cannot be represented in an int64, the result is undefined.
func (a Float16) Int64() int64 {
	return int64(a.Float64())
}

// IsZero reports whether a is zero (+0 or -0).
func (a Float16) IsZero() bool {
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
	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()

	// special cases
	if expA == mask16-bias16 {
		// NaN check is done above; a is ±inf
		if b.IsZero() {
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
		if a.IsZero() {
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
	if a.IsNaN() || b.IsNaN() {
		// a / NaN = NaN
		// NaN / b = NaN
		return uvnan16
	}

	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()
	sign := signA ^ signB

	if b.IsZero() {
		if a.IsZero() {
			// 0 / 0 = NaN
			return uvnan16
		}
		// ±finite / 0 = ±inf
		return Float16(sign | uvinf16)
	}
	if a.IsZero() {
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
	frac |= nonzero16(mod)
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
	if a.IsNaN() || b.IsNaN() {
		// a + NaN = NaN
		// NaN + b = NaN
		return uvnan16
	}
	if a.IsZero() {
		if b.IsZero() {
			//  0 +  0 =  0
			//  0 + -0 =  0
			// -0 +  0 =  0
			// -0 + -0 = -0
			return a & b
		}
		// ±0 + b = b
		return b
	}
	if b.IsZero() {
		// a + ±0 = a
		return a
	}

	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()

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

// Sqrt returns the square root of a.
//
// Special cases are:
//
//	Sqrt(+Inf) = +Inf
//	Sqrt(±0) = ±0
//	Sqrt(x < 0) = NaN
//	Sqrt(NaN) = NaN
func (a Float16) Sqrt() Float16 {
	// special cases
	switch {
	case a.IsZero() || a.IsNaN() || a.IsInf(1):
		return a
	case a&signMask16 != 0:
		return uvnan16
	}

	_, exp, frac := a.normalize()
	if exp%2 != 0 {
		// odd exp, double x to make it even
		frac <<= 1
	}
	// exponent of square root
	exp >>= 1

	// generate sqrt(frac) bit by bit
	frac <<= 1
	var q, s uint16 // q = sqrt(frac)
	r := uint16(1 << (shift16 + 1))
	for r != 0 {
		t := s + r
		if t <= frac {
			s = t + r
			frac -= t
			q += r
		}
		frac <<= 1
		r >>= 1
	}

	// final rounding
	if frac != 0 {
		q += q & 1
	}
	return Float16((exp-1+bias16)<<shift16) + Float16(q>>1)
}

// Eq returns a == b.
// NaNs are not equal to anything, including NaN.
// -0.0 and 0.0 are equal.
func (a Float16) Eq(b Float16) bool {
	if a.IsNaN() || b.IsNaN() {
		return false
	}
	if a == b {
		// a and b have the same bit pattern.
		return true
	}

	// check -0 == 0
	return (a|b)&^signMask16 == 0
}

// Ne returns a != b.
// NaNs are not equal to anything, including NaN.
// -0.0 and 0.0 are equal.
func (a Float16) Ne(b Float16) bool {
	return !a.Eq(b)
}

// Lt returns a < b.
//
// Special cases are:
//
//	Lt(NaN, x) == false
//	Lt(x, NaN) == false
func (a Float16) Lt(b Float16) bool {
	if a.IsNaN() || b.IsNaN() {
		return false
	}
	return a.comparable() < b.comparable()
}

// Gt returns a > b.
//
// Special cases are:
//
//	Gt(x, NaN) == false
//	Gt(NaN, x) == false
func (a Float16) Gt(b Float16) bool {
	return b.Lt(a)
}

// Le returns a <= b.
//
// Special cases are:
//
//	Le(x, NaN) == false
//	Le(NaN, x) == false
func (a Float16) Le(b Float16) bool {
	if a.IsNaN() || b.IsNaN() {
		return false
	}
	return a.comparable() <= b.comparable()
}

// Ge returns a >= b.
//
// Special cases are:
//
//	Ge(x, NaN) == false
//	Ge(NaN, x) == false
func (a Float16) Ge(b Float16) bool {
	return b.Le(a)
}

// normalize returns the sign, exponent, and normalized fraction of a.
func (a Float16) normalize() (sign uint16, exp int, frac uint16) {
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

func (a Float16) split() (sign uint16, exp int, frac uint16) {
	sign = uint16(a & signMask16)
	exp = int((a>>shift16)&mask16) - bias16
	frac = uint16(a & fracMask16)

	if exp == -bias16 {
		// a is subnormal
		exp++
	} else {
		// a is normal
		frac |= 1 << shift16
	}
	return
}

// comparable converts a to a comparable form.
func (a Float16) comparable() int16 {
	i := int16(a)
	i ^= (i >> 15) & 0x7fff
	i += int16(a >> 15) // normalize -0 to 0
	return i
}

// FMA16 returns x * y + z, computed with only one rounding.
// (That is, FMA16 returns the fused multiply-add of x, y, and z.)
func FMA16(x, y, z Float16) Float16 {
	// Inf or NaN involved. At most one rounding will occur.
	if x.IsZero() || y.IsZero() || x&uvinf16 == uvinf16 || y&uvinf16 == uvinf16 {
		return x.Mul(y).Add(z)
	}
	if z.IsZero() {
		return x.Mul(y)
	}
	// Handle non-finite z separately. Evaluating x*y+z where
	// x and y are finite, but z is infinite, should always result in z.
	if z&uvinf16 == uvinf16 {
		return z
	}

	// Split x, y, z into sign, exponent, mantissa.
	signX, expX, fracX := x.normalize()
	signY, expY, fracY := y.normalize()
	signZ, expZ, fracZ0 := z.normalize()

	// Compute product p = x*y as sign, exponent, mantissa.
	expP := expX + expY + 1
	fracP := uint32(fracX<<4) * uint32(fracY<<5)
	signP := signX ^ signY // product sign

	// Normalize product.
	is30zero := uint((^fracP >> 30) & 1)
	fracP <<= is30zero
	expP -= int(is30zero)

	fracZ := uint32(fracZ0) << (4 + 16)

	// Swap addition operands so |p| >= |z|
	if expP < expZ || expP == expZ && fracP < fracZ {
		signP, signZ = signZ, signP
		expP, expZ = expZ, expP
		fracP, fracZ = fracZ, fracP
	}

	// Special case: if p == -z the result is always +0 since neither operand is zero.
	if signP != signZ && expP == expZ && fracP == fracZ {
		return Float16(0)
	}

	// Align mantissa
	fracZ = shrcompress32(fracZ, uint(expP-expZ))

	// Compute resulting significands, normalizing if necessary.
	var frac uint16
	if signP == signZ {
		// Adding fracP + fracZ
		fracP += fracZ
		expP += int(fracP >> 31)
		frac = uint16(shrcompress32(fracP, uint(16+fracP>>31)))
	} else {
		// Subtracting fracP - fracZ
		fracP -= fracZ
		nz := bits.LeadingZeros32(fracP) - 1
		expP -= nz
		frac = uint16(shrcompress32(fracP<<uint(nz), 16))
	}

	// check for underflow
	expP += bias16
	if expP <= 0 {
		n := uint(1 - expP)
		frac = roundToNearestEven16(frac, n+4)
		return Float16(signP | uint16(frac))
	}

	// Round and break ties to even
	frac = roundToNearestEven16(frac, 4)
	if frac&(1<<(shift16+1)) != 0 {
		expP++
		frac >>= 1
	}
	if expP >= mask16 {
		// overflow
		return Float16(signP | uvinf16)
	}
	return Float16(signP | uint16(expP<<shift16) | frac&fracMask16)
}
