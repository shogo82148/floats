package floats

import (
	"math/bits"

	"github.com/shogo82148/ints"
)

const (
	mask256  = 0x7_ffff     // mask for exponent
	shift256 = 256 - 19 - 1 // shift for exponent
	bias256  = 262143       // bias for exponent
)

var (
	// Infinity value for Float256
	uvinf256 = ints.Uint256{
		0x7fff_f000_0000_0000, 0x0000_0000_0000_0000,
		0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
	}
	// Negative Infinity value for Float256
	uvneginf256 = ints.Uint256{
		0xffff_f000_0000_0000, 0x0000_0000_0000_0000,
		0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
	}
	uvnan256 = ints.Uint256{
		0x7fff_f800_0000_0000, 0x0000_0000_0000_0000,
		0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
	}
	uvone256 = ints.Uint256{
		0x3fff_f000_0000_0000, 0x0000_0000_0000_0000,
		0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
	}
	// mask for sign bit
	signMask256 = ints.Uint256{
		0x8000_0000_0000_0000, 0x0000_0000_0000_0000,
		0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
	}
	fracMask256 = ints.Uint256{
		0x0000_0fff_ffff_ffff, 0xffff_ffff_ffff_ffff,
		0xffff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff,
	}
)

// Float256 is a 256-bit floating-point number.
type Float256 ints.Uint256

// NewFloat256 converts f to Float256.
func NewFloat256(f float64) Float256 {
	return Float64(f).Float256()
}

// NewFloat256FromBits converts the IEEE 754 binary representation b to Float256.
func NewFloat256FromBits(b ints.Uint256) Float256 {
	return Float256(b)
}

// NewFloat256NaN returns a NaN Float256 value.
func NewFloat256NaN() Float256 {
	return Float256(uvnan256)
}

// NewFloat256Inf positive infinity if sign >= 0, negative infinity if sign < 0.
func NewFloat256Inf(sign int) Float256 {
	if sign >= 0 {
		return Float256(uvinf256)
	}
	return Float256(uvneginf256)
}

// Bits returns the IEEE 754 binary representation of a.
func (a Float256) Bits() ints.Uint256 {
	return ints.Uint256(a)
}

// IsNaN reports whether a is an IEEE 754 “not-a-number” value.
func (a Float256) IsNaN() bool {
	return a[0]&(mask256<<(shift256-192)) == (mask256<<(shift256-192)) &&
		!ints.Uint256(a).And(fracMask256).IsZero()
}

// IsInf reports whether a is an infinity, according to sign.
// If sign > 0, IsInf reports whether a is positive infinity.
// If sign < 0, IsInf reports whether a is negative infinity.
// If sign == 0, IsInf reports whether a is either infinity.
func (a Float256) IsInf(sign int) bool {
	b := ints.Uint256(a)
	return sign >= 0 && b == uvinf256 || sign <= 0 && b == uvneginf256
}

// Signbit reports whether a is negative or negative zero.
func (a Float256) Signbit() bool {
	return a[0]&signMask256[0] != 0
}

// Copysign returns a value with the magnitude of a
// and the sign of sign.
func (a Float256) Copysign(sign Float256) Float256 {
	return Float256{(a[0] &^ signMask256[0]) | (sign[0] & signMask256[0]), a[1], a[2], a[3]}
}

// Int64 returns the integer value of a, rounding towards zero.
// If a cannot be represented in an int64, the result is undefined.
func (a Float256) Int64() int64 {
	sign, exp, frac := a.normalize()
	frac = frac.Rsh(uint(shift256 - exp))
	ret := int64(frac.Uint64())
	if sign != 0 {
		ret = -ret
	}
	return ret
}

// Uint64 returns the unsigned integer value of a, rounding towards zero.
// If a cannot be represented in a uint64, the result is undefined.
func (a Float256) Uint64() uint64 {
	_, exp, frac := a.normalize()
	frac = frac.Rsh(uint(shift256 - exp))
	return uint64(frac.Uint64())
}

// Int128 returns the signed 128-bit integer value of a, rounding towards zero.
// If a cannot be represented in a int128, the result is undefined.
func (a Float256) Int128() ints.Int128 {
	sign, exp, frac := a.normalize()
	frac = frac.Rsh(uint(shift256 - exp))
	ret := ints.Int128(frac.Uint128())
	if sign != 0 {
		ret = ret.Neg()
	}
	return ret
}

// Uint128 returns the unsigned 128-bit integer value of a, rounding towards zero.
// If a cannot be represented in a uint128, the result is undefined.
func (a Float256) Uint128() ints.Uint128 {
	_, exp, frac := a.normalize()
	frac = frac.Rsh(uint(shift256 - exp))
	return ints.Uint128(frac.Uint128())
}

// Int256 returns the signed 256-bit integer value of a, rounding towards zero.
// If a cannot be represented in a int256, the result is undefined.
func (a Float256) Int256() ints.Int256 {
	sign, exp, frac := a.normalize()
	if exp <= shift256 {
		frac = frac.Rsh(uint(shift256 - exp))
	} else {
		frac = frac.Lsh(uint(exp - shift256))
	}
	ret := ints.Int256(frac.Uint256())
	if sign != 0 {
		ret = ret.Neg()
	}
	return ret
}

// Uint256 returns the unsigned 256-bit integer value of a, rounding towards zero.
// If a cannot be represented in a uint256, the result is undefined.
func (a Float256) Uint256() ints.Uint256 {
	_, exp, frac := a.normalize()
	if exp <= shift256 {
		frac = frac.Rsh(uint(shift256 - exp))
	} else {
		frac = frac.Lsh(uint(exp - shift256))
	}
	return frac
}

// IsZero reports whether a is zero (+0 or -0).
func (a Float256) IsZero() bool {
	return (a[0]&^signMask256[0])|a[1]|a[2]|a[3] == 0
}

// Neg returns the negation of a.
func (a Float256) Neg() Float256 {
	return Float256{a[0] ^ signMask256[0], a[1], a[2], a[3]}
}

// Abs returns the absolute value of a.
//
// Special cases:
//
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func (a Float256) Abs() Float256 {
	return Float256{a[0] &^ signMask256[0], a[1], a[2], a[3]}
}

// Mul returns the product of a and b.
func (a Float256) Mul(b Float256) Float256 {
	if a.IsNaN() || b.IsNaN() {
		// a * NaN = NaN
		// NaN * b = NaN
		return Float256(uvnan256)
	}
	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()
	sign := signA ^ signB

	// handle special cases
	if expA == mask256-bias256 {
		// NaN check is done above; a is ±inf
		if b.IsZero() {
			// ±inf * 0 = NaN
			return Float256(uvnan256)
		} else {
			// ±inf * +finite = ±inf
			// ±inf * -finite = ∓inf
			return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
		}
	}
	if expB == mask256-bias256 {
		// NaN check is done above; b is ±inf
		if a.IsZero() {
			// 0 * ±inf = NaN
			return Float256(uvnan256)
		} else {
			// +finite * ±inf = ±inf
			// -finite * ±inf = ∓inf
			return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
		}
	}
	if a.IsZero() || b.IsZero() {
		// +0 * ±finite = ±0
		// -0 * ±finite = ∓0
		return Float256{sign, 0, 0, 0}
	}

	// p = fracA * fracB is in [2^472, 2^474).
	p := fracA.Mul512(fracB)

	// Shift p right by 220 bits, keeping the shifted-out bits as a sticky bit.
	// frac is in [2^252, 2^254), and frac * 2^(base-253) is the exact product.
	const extra = 256 - 1 - (shift256 + 1) - 1
	frac := ints.Uint256{
		p[0]<<36 | p[1]>>28,
		p[1]<<36 | p[2]>>28,
		p[2]<<36 | p[3]>>28,
		p[3]<<36 | p[4]>>28,
	}
	if p[4]<<36|p[5]|p[6]|p[7] != 0 {
		frac[3] |= 1
	}
	base := expA + expB + 1

	// the exponent of the result
	exp := max(base+frac.BitLen()-(shift256+1+extra),
		// the result is subnormal
		1-bias256)
	if exp >= mask256-bias256 {
		// overflow
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}

	// round the fraction
	shift := exp - base + extra
	if shift >= 256 {
		// underflow: frac < 2^254, so it is rounded to zero.
		return Float256{sign, 0, 0, 0}
	}
	frac = roundToNearestEven256(frac, uint(shift))

	// The hidden bit of frac is added to the exponent.
	// It also handles carry-out caused by rounding, and subnormal results.
	// If the result overflows, it becomes infinity.
	frac[0] += uint64(exp-1+bias256) << (shift256 - 192)
	return Float256{sign | frac[0], frac[1], frac[2], frac[3]}
}

// Quo returns the quotient of a and b.
func (a Float256) Quo(b Float256) Float256 {
	if a.IsNaN() || b.IsNaN() {
		// a / NaN = NaN
		// NaN / b = NaN
		return Float256(uvnan256)
	}

	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()
	sign := signA ^ signB

	if b.IsZero() {
		if a.IsZero() {
			// 0 / 0 = NaN
			return Float256(uvnan256)
		}
		// ±finite / 0 = ±inf
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}
	if a.IsZero() {
		// 0 / finite = 0
		return Float256{sign, 0, 0, 0}
	}
	if expA == mask256-bias256 {
		// NaN check is done above; a is ±inf
		if expB == mask256-bias256 {
			// ±inf / ±inf = NaN
			return Float256(uvnan256)
		}
		// ±inf / finite = ±inf
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}
	if expB == mask256-bias256 {
		// NaN check is done above; b is ±inf
		// NaN and Inf checks are done above; a is finite.
		// ±finite / ±inf = 0
		return Float256{sign, 0, 0, 0}
	}

	exp := expA - expB + bias256
	if fracA.Cmp(fracB) < 0 {
		exp--
		fracA = lsh256small(fracA, 1)
	}
	if exp >= mask256 {
		// overflow
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}

	const shift = shift256 + 3 // 1 for the implicit bit, 1 for the rounding bit, 1 for the guard bit
	// normalize the divisor so that its most significant bit is set.
	const norm = 256 - (shift256 + 1)
	// fracA << (shift + norm) == (fracA << 2) << 256
	frac, inexact := quo512by256(lsh256small(fracA, shift+norm-256), lsh256small(fracB, norm))
	if inexact {
		frac[3] |= 1
	}

	if exp <= 0 {
		// the result is subnormal
		shift := -exp + 3 + 1
		frac = roundToNearestEven256(frac, uint(shift))
		return Float256{
			sign | frac[0],
			frac[1],
			frac[2],
			frac[3],
		}
	}

	// round-to-nearest-even (guard+round+sticky are in the low 3 bits)
	frac = roundToNearestEven256(frac, 3)
	// detect carry-out caused by rounding
	if frac[0]>>(shift256+1-192) != 0 {
		frac = ints.Uint256{
			frac[0] >> 1,
			frac[1]>>1 | frac[0]<<63,
			frac[2]>>1 | frac[1]<<63,
			frac[3]>>1 | frac[2]<<63,
		}
		exp++
		if exp >= mask256 {
			// overflow
			return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
		}
	}
	return Float256{
		sign | uint64(exp)<<(shift256-192) | frac[0]&fracMask256[0],
		frac[1],
		frac[2],
		frac[3],
	}
}

// Add returns the sum of a and b.
func (a Float256) Add(b Float256) Float256 {
	if a.IsNaN() || b.IsNaN() {
		// a + NaN = NaN
		// NaN + b = NaN
		return Float256(uvnan256)
	}
	if a.IsZero() {
		if b.IsZero() {
			//  0 +  0 =  0
			//  0 + -0 =  0
			// -0 +  0 =  0
			// -0 + -0 = -0
			return Float256{a[0] & b[0], 0, 0, 0}
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
	if expA == mask256-bias256 {
		// NaN check is done above; a is ±inf
		if expB == mask256-bias256 {
			if signA == signB {
				// ±inf + ±inf = ±inf
				return Float256{signA | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
			}
			// ±inf + ∓inf = NaN
			return Float256(uvnan256)
		}
		// b is finite, the result is ±inf
		return a
	}
	if expB == mask256-bias256 {
		// NaN check is done above; b is ±inf
		// NaN and Inf checks are done above; a is finite.
		return b
	}

	// make |a| >= |b|
	if expA < expB || (expA == expB && fracA.Cmp(fracB) < 0) {
		signA, signB = signB, signA
		expA, expB = expB, expA
		fracA, fracB = fracB, fracA
	}

	// Place the fractions at the top of 256-bit integers, leaving one bit for carry-out.
	// The extra low bits work as guard, round, and sticky bits.
	const extra = 256 - 1 - (shift256 + 1) - 1
	fracA = fracA.Lsh(extra)
	fracB = shrcompress256(fracB.Lsh(extra), uint(expA-expB))

	// add the fractions
	var frac ints.Uint256
	if signA == signB {
		frac = fracA.Add(fracB)
	} else {
		frac = fracA.Sub(fracB)
		if frac.IsZero() {
			// x - x = +0
			return Float256{}
		}
	}

	// the exponent of the result
	exp := max(expA+frac.BitLen()-(shift256+1+extra),
		// the result is subnormal
		1-bias256)
	if exp >= mask256-bias256 {
		// overflow
		return Float256{signA | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}

	// round the fraction
	if shift := exp - expA + extra; shift > 0 {
		frac = roundToNearestEven256(frac, uint(shift))
	} else {
		frac = frac.Lsh(uint(-shift))
	}

	// The hidden bit of frac is added to the exponent.
	// It also handles carry-out caused by rounding, and subnormal results.
	// If the result overflows, it becomes infinity.
	frac[0] += uint64(exp-1+bias256) << (shift256 - 192)
	return Float256{signA | frac[0], frac[1], frac[2], frac[3]}
}

// Sub returns the difference of a and b.
func (a Float256) Sub(b Float256) Float256 {
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
func (a Float256) Sqrt() Float256 {
	switch {
	case a.IsZero() || a.IsNaN() || a.IsInf(1):
		return a
	case a[0]&signMask256[0] != 0:
		return Float256(uvnan256)
	}

	_, exp, frac := a.normalize()
	if exp%2 != 0 {
		// odd exp. double x to make it even
		frac = frac.Lsh(1)
	}
	// exponent of square root
	exp >>= 1

	// q = floor(sqrt(frac << (shift256 + 2))) has shift256 + 2 bits,
	// including one guard bit for rounding.
	// Scale frac so that its top word is normalized for sqrtRem512.
	const extra = (512 - (shift256 + 2) - (shift256 + 2)) / 2
	n := ints.Uint512{0, 0, 0, 0, frac[0], frac[1], frac[2], frac[3]}.Lsh(shift256 + 2 + 2*extra)
	root, inexact := sqrtRem512(n)
	q := root.Rsh(extra)

	// final rounding
	if root[3]&(1<<extra-1) != 0 || inexact {
		q = q.Add(q.And(ints.Uint256{0, 0, 0, 1}))
	}
	q = q.Rsh(1)
	q = q.Add(ints.Uint256{uint64(exp-1+bias256) << (shift256 - 192), 0, 0, 0})
	return Float256(q)
}

// Eq returns a == b.
// NaNs are not equal to anything, including NaN.
// -0.0 and 0.0 are equal.
func (a Float256) Eq(b Float256) bool {
	if a.IsNaN() || b.IsNaN() {
		return false
	}
	if a == b {
		// a and b have the same bit pattern.
		return true
	}

	// check -0 == 0
	return (a[0]|b[0])&^signMask256[0]|a[1]|b[1]|a[2]|b[2]|a[3]|b[3] == 0
}

// Ne returns a != b.
// NaNs are not equal to anything, including NaN.
// -0.0 and 0.0 are equal.
func (a Float256) Ne(b Float256) bool {
	return !a.Eq(b)
}

// Lt returns a < b.
//
// Special cases are:
//
//	Lt(NaN, x) == false
//	Lt(x, NaN) == false
func (a Float256) Lt(b Float256) bool {
	if a.IsNaN() || b.IsNaN() {
		return false
	}
	return a.comparable().Cmp(b.comparable()) < 0
}

// Gt returns a > b.
//
// Special cases are:
//
//	Gt(x, NaN) == false
//	Gt(NaN, x) == false
func (a Float256) Gt(b Float256) bool {
	return b.Lt(a)
}

// Le returns a <= b.
//
// Special cases are:
//
//	Le(x, NaN) == false
//	Le(NaN, x) == false
func (a Float256) Le(b Float256) bool {
	if a.IsNaN() || b.IsNaN() {
		return false
	}
	return a.comparable().Cmp(b.comparable()) <= 0
}

// Ge returns a >= b.
//
// Special cases are:
//
//	Ge(x, NaN) == false
//	Ge(NaN, x) == false
func (a Float256) Ge(b Float256) bool {
	return b.Le(a)
}

// normalize returns the sign, exponent, and normalized fraction of a.
func (a Float256) normalize() (sign uint64, exp int, frac ints.Uint256) {
	sign = a[0] & signMask256[0]
	exp = int((a[0]>>(shift256-192))&mask256) - bias256
	frac = ints.Uint256{a[0] & fracMask256[0], a[1], a[2], a[3]}
	if exp == -bias256 {
		// a is subnormal
		// normalize
		l := frac.BitLen()
		frac = lsh256(frac, uint(shift256-l)+1)
		exp = l - (bias256 + shift256)
		return
	}

	// a is normal
	frac[0] = frac[0] | (1 << (shift256 - 192))
	return
}

// split returns the sign, exponent, and fraction of a.
func (a Float256) split() (sign uint64, exp int, frac ints.Uint256) {
	b := ints.Uint256(a)
	sign = b[0] & signMask256[0]
	exp = int((b[0]>>(shift256-192))&mask256) - bias256
	frac = b.And(fracMask256)
	if exp == -bias256 {
		// a is subnormal
		exp++
	} else {
		// a is normal
		frac[0] = frac[0] | (1 << (shift256 - 192))
	}
	return
}

// comparable returns a comparable value for a.
func (a Float256) comparable() ints.Int256 {
	i := ints.Int256(a)
	sign := uint64(int64(i[0]) >> 63)
	i = i.Xor(ints.Int256{
		sign & 0x7fff_ffff_ffff_ffff, sign, sign, sign,
	})
	i = i.Add(ints.Int256{0, 0, 0, sign & 1})
	return i
}

// FMA256 returns x * y + z, computed with only one rounding.
// (That is, FMA256 returns the fused multiply-add of x, y, and z.)
func FMA256(x, y, z Float256) Float256 {
	if x.IsZero() || y.IsZero() || x[0]&(mask256<<(shift256-192)) == mask256<<(shift256-192) || y[0]&(mask256<<(shift256-192)) == mask256<<(shift256-192) {
		return x.Mul(y).Add(z)
	}
	if z.IsZero() {
		return x.Mul(y)
	}
	// Handle non-finite z separately. Evaluating x*y+z where
	// x and y are finite, but z is infinite, should always result in z.
	if z[0]&(mask256<<(shift256-192)) == mask256<<(shift256-192) {
		return z
	}

	// Split x, y, z into sign, exponent, mantissa.
	signX, expX, fracX := x.normalize()
	signY, expY, fracY := y.normalize()
	signZ, expZ, fracZ0 := z.normalize()

	// Compute product p = x*y as sign, exponent, mantissa.
	expP := expX + expY + 1
	fracP := lsh256(fracX, 18).Mul512(lsh256(fracY, 19))
	signP := signX ^ signY // product sign

	// Normalize the product without branches; the result is random in general.
	n := (^fracP[0] >> 62) & 1
	fracP = ints.Uint512{
		fracP[0]<<n | fracP[1]>>(64-n),
		fracP[1]<<n | fracP[2]>>(64-n),
		fracP[2]<<n | fracP[3]>>(64-n),
		fracP[3]<<n | fracP[4]>>(64-n),
		fracP[4]<<n | fracP[5]>>(64-n),
		fracP[5]<<n | fracP[6]>>(64-n),
		fracP[6]<<n | fracP[7]>>(64-n),
		fracP[7] << n,
	}
	expP -= int(n)

	// fracZ = fracZ0 << (18 + 256)
	zz := lsh256(fracZ0, 18)
	fracZ := ints.Uint512{zz[0], zz[1], zz[2], zz[3], 0, 0, 0, 0}

	// Swap addition operands so |p| >= |z|
	// The low 256 bits of fracZ are zero, so fracP >= fracZ if the high 256 bits are equal.
	if expP < expZ || expP == expZ && cmpHigh256(fracP, fracZ) < 0 {
		signP, signZ = signZ, signP
		expP, expZ = expZ, expP
		fracP, fracZ = fracZ, fracP
	}

	// Special case: if p == -z the result is always +0 since neither operand is zero.
	if expP == expZ && cmpHigh256(fracP, fracZ) == 0 && fracP[4]|fracP[5]|fracP[6]|fracP[7] == 0 && signP != signZ {
		return Float256{0, 0, 0, 0}
	}

	// Align mantissa
	if d := uint(expP - expZ); d != 0 {
		fracZ = shrcompress512(fracZ, d)
	}

	// Compute resulting significands, normalizing if necessary.
	var frac ints.Uint256
	if signP == signZ {
		// Adding fracP + fracZ
		fracP = fracP.Add(fracZ)
		carry := fracP[0] >> 63
		expP += int(carry)
		// frac = shrcompress512(fracP, 256+carry)
		frac = ints.Uint256{
			fracP[0] >> carry,
			fracP[1]>>carry | fracP[0]<<(64-carry),
			fracP[2]>>carry | fracP[1]<<(64-carry),
			fracP[3]>>carry | fracP[2]<<(64-carry),
		}
		frac[3] |= nonzero64(fracP[3]&carry | fracP[4] | fracP[5] | fracP[6] | fracP[7])
	} else {
		// Subtracting fracP - fracZ
		fracP = fracP.Sub(fracZ)
		nz := fracP.LeadingZeros() - 1
		expP -= nz
		// frac = shrcompress512(fracP << nz, 256)
		fracP = lsh512(fracP, uint(nz))
		frac = ints.Uint256{fracP[0], fracP[1], fracP[2], fracP[3]}
		frac[3] |= nonzero64(fracP[4] | fracP[5] | fracP[6] | fracP[7])
	}

	// check for underflow
	expP += bias256
	if expP <= 0 {
		n := uint(1 - expP)
		frac = roundToNearestEven256(frac, n+18)
		return Float256{signP | frac[0], frac[1], frac[2], frac[3]}
	}

	// Round and break ties to even
	// frac >> 18 rounded to nearest even.
	// Adding half - 1 + (the least significant bit of the result) never overflows
	// because the most significant bit of frac is bit 254.
	var c uint64
	frac[3], c = bits.Add64(frac[3], 1<<17-1+(frac[3]>>18)&1, 0)
	frac[2], c = bits.Add64(frac[2], 0, c)
	frac[1], c = bits.Add64(frac[1], 0, c)
	frac[0] += c
	frac = ints.Uint256{
		frac[0] >> 18,
		frac[1]>>18 | frac[0]<<46,
		frac[2]>>18 | frac[1]<<46,
		frac[3]>>18 | frac[2]<<46,
	}
	if frac[0]&(1<<(shift256+1-192)) != 0 {
		expP++
		frac = frac.Rsh(1)
	}
	if expP >= mask256 {
		// Overflow
		return Float256{signP | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}
	return Float256{
		signP | uint64(expP)<<(shift256-192) | frac[0]&fracMask256[0],
		frac[1],
		frac[2],
		frac[3],
	}
}

// cmpHigh256 compares the high 256 bits of x and y.
func cmpHigh256(x, y ints.Uint512) int {
	for i := 0; i < 4; i++ {
		if x[i] != y[i] {
			if x[i] < y[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

// Nextafter returns the next representable float256 value after a towards b.
//
// Special cases are:
//
//	a.Nextafter(a)   = a
//	NaN.Nextafter(b) = NaN
//	a.Nextafter(NaN) = NaN
func (a Float256) Nextafter(b Float256) (r Float256) {
	switch {
	case a.IsNaN() || b.IsNaN(): // special case
		r = NewFloat256NaN()
	case a.Eq(b): // special case
		r = a
	case a.IsZero():
		r = Float256{0, 0, 0, 1}.Copysign(b)
	case b.Gt(a) == a.Gt(Float256{}):
		r = Float256(a.Bits().Add(ints.Uint256{0, 0, 0, 1}))
	default:
		r = Float256(a.Bits().Sub(ints.Uint256{0, 0, 0, 1}))
	}
	return
}

// Modf returns integer and fractional floating-point numbers
// that sum to f. Both values have the same sign as f.
//
// Special cases are:
//
//	Modf(±Inf) = ±Inf, NaN
//	Modf(NaN) = NaN, NaN
func (a Float256) Modf() (int Float256, frac Float256) {
	if a.Lt(Float256(uvone256)) { // a < 1
		switch {
		case a.Lt(Float256{}): // a < 0
			int, frac = a.Neg().Modf()
			return int.Neg(), frac.Neg()
		case a.IsZero(): // a == 0
			return a, a
		default: // 0 < a < 1
			return Float256{}, a
		}
	}

	x := a.Bits()
	e := uint((a[0]>>(shift256-192))&mask256) - bias256

	// Keep the top 20+e bits, the integer part; clear the rest.
	if e < shift256 {
		one := ints.Uint256{0, 0, 0, 1}
		x = x.AndNot(one.Lsh(shift256 - e).Sub(one))
	}
	int = Float256(x)
	frac = a.Sub(int)
	return
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
func (a Float256) Frexp() (frac Float256, exp int) {
	// special cases
	if a.IsZero() || a.IsNaN() || a.IsInf(0) {
		return a, 0
	}

	sign, e, bits := a.normalize()
	e++
	bits[0] = sign | (-1+bias256)<<(shift256-192) | (bits[0] & fracMask256[0])
	return Float256(bits), e
}

// Ldexp is the inverse of [Frexp].
// It returns a × 2**exp.
//
// Special cases are:
//
//	±0.Ldexp(exp) = ±0
//	±Inf.Ldexp(exp) = ±Inf
//	NaN.Ldexp(exp) = NaN
func (a Float256) Ldexp(exp int) Float256 {
	// special cases
	if a.IsZero() || a.IsNaN() || a.IsInf(0) {
		return a
	}

	sign, e, bits := a.normalize()
	e += exp
	if e >= mask256-bias256 {
		// overflow
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}
	if e <= -bias256-shift256 {
		// underflow
		return Float256{sign, 0, 0, 0}
	}
	if e <= -bias256 {
		// the result is subnormal
		shift := -e - bias256 + 1
		frac := roundToNearestEven256(bits, uint(shift))
		return Float256{sign | frac[0], frac[1], frac[2], frac[3]}
	}
	bits[0] = sign | uint64(e+bias256)<<(shift256-192) | (bits[0] & fracMask256[0])
	return Float256(bits)
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
func (a Float256) Mod(b Float256) Float256 {
	// special cases
	if b.IsZero() || a.IsInf(0) || a.IsNaN() || b.IsNaN() {
		return NewFloat256NaN()
	}

	b = b.Abs()
	bfr, bexp := b.Frexp()
	r := a
	if a.Lt(Float256{}) {
		r = a.Neg()
	}
	for r.Ge(b) {
		rfr, rexp := r.Frexp()
		if rfr.Lt(bfr) {
			rexp--
		}
		r = r.Sub(b.Ldexp(rexp - bexp))
	}
	if a.Lt(Float256{}) {
		r = r.Neg()
	}
	return r
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
func (a Float256) Remainder(b Float256) Float256 {
	// special cases
	if b.IsZero() || a.IsInf(0) || a.IsNaN() || b.IsNaN() {
		return NewFloat256NaN()
	}
	if b.IsInf(0) {
		return a
	}

	sign := false
	if a.Lt(Float256{}) {
		a = a.Neg()
		sign = true
	}
	if b.Lt(Float256{}) {
		b = b.Neg()
	}
	if a.Eq(b) {
		if sign {
			return Float256(signMask256)
		}
		return Float256{}
	}
	if b.Le(Float256{0x7fff_dfff_ffff_ffff, 0xffff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}) { // = MAX_FLOAT256 / 2
		// now a < 2b
		a = a.Mod(b.Add(b))
	}
	if b.Lt(Float256{0x0000_2000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000}) { // smallest positive normal number * 2
		// To avoid loss of precision, we will bypass the calculation b * 0.5.
		if a.Add(a).Gt(b) {
			a = a.Sub(b)
			if a.Add(a).Ge(b) {
				a = a.Sub(b)
			}
		}
	} else {
		bHalf := b.Mul(Float256{0x3fff_e000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000}) // b * 0.5
		if a.Gt(bHalf) {
			a = a.Sub(b)
			if a.Ge(bHalf) {
				a = a.Sub(b)
			}
		}
	}

	if sign {
		a = a.Neg()
	}
	return a
}
