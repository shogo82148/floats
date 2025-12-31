package floats

import (
	"github.com/shogo82148/ints"
)

const (
	mask128  = 0x7fff       // mask for exponent
	shift128 = 128 - 15 - 1 // shift for exponent
	bias128  = 16383        // bias for exponent
)

var (
	uvinf128    = ints.Uint128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000} // Infinity value for Float128
	uvneginf128 = ints.Uint128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000} // Negative Infinity value for Float128
	uvnan128    = ints.Uint128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000} // NaN value for Float128
	// uvone128    = ints.Uint128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000} // One value for Float128

	signMask128 = ints.Uint128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000} // mask for sign bit
	fracMask128 = ints.Uint128{0x0000_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}
)

// Float128 is a 128-bit floating-point number.
type Float128 ints.Uint128

// NewFloat128 converts f to Float128.
func NewFloat128(f float64) Float128 {
	return Float64(f).Float128()
}

// IsNaN reports whether a is an IEEE 754 “not-a-number” value.
func (a Float128) IsNaN() bool {
	return a[0]&(mask128<<(shift128-64)) == (mask128<<(shift128-64)) &&
		!ints.Uint128(a).And(fracMask128).IsZero()
}

// IsInf reports whether a is an infinity, according to sign.
// If sign > 0, IsInf reports whether a is positive infinity.
// If sign < 0, IsInf reports whether a is negative infinity.
// If sign == 0, IsInf reports whether a is either infinity.
func (a Float128) IsInf(sign int) bool {
	b := ints.Uint128(a)
	return sign >= 0 && b == uvinf128 || sign <= 0 && b == uvneginf128
}

// Signbit reports whether x is negative or negative zero.
func (a Float128) Signbit() bool {
	return a[0]&signMask128[0] != 0
}

// Int64 returns the integer value of a, rounding towards zero.
// If a cannot be represented in an int64, the result is undefined.
func (a Float128) Int64() int64 {
	sign, exp, frac := a.normalize()
	frac = frac.Rsh(uint(shift128 - exp))
	ret := int64(frac.Uint64())
	if sign != 0 {
		ret = -ret
	}
	return ret
}

// IsZero reports whether a is zero (+0 or -0).
func (a Float128) IsZero() bool {
	return (a[0]&^signMask128[0])|a[1] == 0
}

// Neg returns the negation of a.
func (a Float128) Neg() Float128 {
	return Float128{a[0] ^ signMask128[0], a[1]}
}

// Mul returns the product of a and b.
func (a Float128) Mul(b Float128) Float128 {
	if a.IsNaN() || b.IsNaN() {
		// a * NaN = NaN
		// NaN * b = NaN
		return Float128(uvnan128)
	}

	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()
	sign := signA ^ signB

	// handle special cases
	if expA == mask128-bias128 {
		// NaN check is done above; a is ±inf
		if b.IsZero() {
			// ±inf * 0 = NaN
			return Float128(uvnan128)
		} else {
			// ±inf * +finite = ±inf
			// ±inf * -finite = ∓inf
			return Float128{sign | uvinf128[0], uvinf128[1]}
		}
	}
	if expB == mask128-bias128 {
		// NaN check is done above; b is ±inf
		if a.IsZero() {
			// 0 * ±inf = NaN
			return Float128(uvnan128)
		} else {
			// +finite * ±inf = ±inf
			// -finite * ±inf = ∓inf
			return Float128{sign | uvinf128[0], uvinf128[1]}
		}
	}
	if a.IsZero() || b.IsZero() {
		// 0 * finite = 0
		return Float128{sign, 0}
	}

	exp := expA + expB
	frac := fracA.Mul256(fracB)
	shift := frac.BitLen() - (shift128 + 1)
	exp += shift - shift128

	if exp < -(bias128 + shift128) {
		// underflow
		return Float128{sign, 0}
	} else if exp <= -bias128 {
		// the result is subnormal
		// normalize
		shift := shift128 - (expA + expB + bias128) + 1
		frac = roundToNearestEven256(frac, uint(shift))
		return Float128{sign | frac[2], frac[3]}
	}

	exp = expA + expB + bias128
	frac = roundToNearestEven256(frac, uint(shift))
	shift = frac.BitLen() + shift - (shift128 + 1)
	exp += shift - shift128
	if exp >= mask128 {
		// overflow
		return Float128{sign | uvinf128[0], uvinf128[1]}
	}
	return Float128{
		sign | uint64(exp)<<(shift128-64) | frac[2]&fracMask128[0],
		frac[3],
	}
}

// Quo returns the quotient of a and b.
func (a Float128) Quo(b Float128) Float128 {
	if a.IsNaN() || b.IsNaN() {
		// a / NaN = NaN
		// NaN / b = NaN
		return Float128(uvnan128)
	}

	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()
	sign := signA ^ signB

	if b.IsZero() {
		if a.IsZero() {
			// 0 / 0 = NaN
			return Float128(uvnan128)
		}
		// ±finite / 0 = ±inf
		return Float128{sign | uvinf128[0], uvinf128[1]}
	}
	if a.IsZero() {
		// 0 / finite = 0
		return Float128{sign, 0}
	}
	if expA == mask128-bias128 {
		// NaN check is done above; a is ±inf
		if expB == mask128-bias128 {
			// ±inf / ±inf = NaN
			return Float128(uvnan128)
		} else {
			// ±inf / finite = ±inf
			return Float128{sign | uvinf128[0], uvinf128[1]}
		}
	}
	if expB == mask128-bias128 {
		// NaN check is done above; b is ±inf
		// NaN and Inf checks are done above; a is finite.
		// ±finite / ±inf = 0
		return Float128{sign, 0}
	}

	exp := expA - expB + bias128
	if fracA.Cmp(fracB) < 0 {
		exp--
		fracA = fracA.Lsh(1)
	}
	if exp >= mask128 {
		// overflow
		return Float128{sign | uvinf128[0], uvinf128[1]}
	}

	shift := shift128 + 3 // 1 for the implicit bit, 1 for the rounding bit, 1 for the guard bit
	fracA256 := fracA.Uint256().Lsh(uint(shift))
	fracB256 := fracB.Uint256()
	frac256, mod := fracA256.DivMod(fracB256)
	frac256[3] |= nonzero64(mod[0]) | nonzero64(mod[1]) | nonzero64(mod[2]) | nonzero64(mod[3])
	frac := frac256.Uint128()

	if exp <= 0 {
		// the result is subnormal
		shift := -exp + 3 + 1
		frac = roundToNearestEven128(frac, uint(shift))
		return Float128{sign | frac[0], frac[1]}
	}

	// round-to-nearest-even (guard+round+sticky are in the low 3 bits)
	frac = roundToNearestEven128(frac, uint(3))
	// detect carry-out caused by rounding
	if frac[0]&(1<<(shift128-64+1)) != 0 {
		frac = frac.Rsh(1)
		exp++
		if exp >= mask128 { // overflow -> ±Inf
			return Float128{sign | uvinf128[0], uvinf128[1]}
		}
	}
	return Float128{sign | uint64(exp)<<(shift128-64) | frac[0]&fracMask128[0], frac[1] & fracMask128[1]}
}

// Add returns the sum of a and b.
func (a Float128) Add(b Float128) Float128 {
	if a.IsNaN() || b.IsNaN() {
		// a + NaN = NaN
		// NaN + b = NaN
		return Float128(uvnan128)
	}
	if a.IsZero() {
		if b.IsZero() {
			//  0 +  0 =  0
			//  0 + -0 =  0
			// -0 +  0 =  0
			// -0 + -0 = -0
			return Float128{a[0] & b[0], 0}
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
	if expA == mask128-bias128 {
		// NaN check is done above; a is ±inf
		if expB == mask128-bias128 {
			if signA == signB {
				// ±inf + ±inf = ±inf
				return Float128{signA | uvinf128[0], uvinf128[1]}
			}
			// ±inf + ∓inf = NaN
			return Float128(uvnan128)
		}
		// b is finite, the result is ±inf
		return a
	}
	if expB == mask128-bias128 {
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
	const offset = 128
	fracA256 := ints.Int256{fracA[0], fracA[1], 0, 0}
	fracB256 := ints.Int256{fracB[0], fracB[1], 0, 0}
	fracB256 = fracB256.Rsh(uint(expA - expB))
	if signA != 0 {
		fracA256 = fracA256.Neg()
	}
	if signB != 0 {
		fracB256 = fracB256.Neg()
	}
	frac256 := fracA256.Add(fracB256)
	sign := uint64(0)
	if frac256.Sign() < 0 {
		sign = signMask128[0]
		frac256 = frac256.Neg()
	}

	shift := ints.Uint256(frac256).BitLen() - (shift128 + 1)
	exp := expA + shift - offset

	if frac256.IsZero() || exp < -(bias128+shift128) {
		// underflow
		return Float128{sign, 0}
	}
	if exp <= -bias128 {
		// the result is subnormal
		shift := offset - (expA + bias128) + 1
		frac256 = ints.Int256(roundToNearestEven256(ints.Uint256(frac256), uint(shift)))
		return Float128{sign | frac256[2]&fracMask128[0], frac256[3]}
	}
	if exp >= mask128-bias128 {
		// overflow
		return Float128{sign | uvinf128[0], uvinf128[1]}
	}

	frac256 = ints.Int256(roundToNearestEven256(ints.Uint256(frac256), uint(shift)))
	// detect carry-out caused by rounding
	if ints.Uint256(frac256).BitLen() > shift128+1 {
		frac256 = frac256.Rsh(1)
		exp++
		if exp >= mask128 {
			// overflow
			return Float128{sign | uvinf128[0], uvinf128[1]}
		}
	}
	return Float128{sign | uint64(exp+bias128)<<(shift128-64) | frac256[2]&fracMask128[0], frac256[3]}
}

// Sub returns the difference of a and b.
func (a Float128) Sub(b Float128) Float128 {
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
func (a Float128) Sqrt() Float128 {
	switch {
	case a.IsZero() || a.IsNaN() || a.IsInf(1):
		return a
	case a[0]&signMask128[0] != 0:
		return Float128(uvnan128)
	}

	_, exp, frac := a.normalize()
	if exp%2 != 0 {
		// odd exp, double x to make it even
		frac = frac.Lsh(1)
	}
	// exponent of square root
	exp >>= 1

	// generate sqrt(frac) bit by bit
	frac = frac.Lsh(1)
	var q, s ints.Uint128
	r := ints.Uint128{1 << (shift128 + 1 - 64), 0}
	for !r.IsZero() {
		t := s.Add(r)
		if t.Cmp(frac) <= 0 {
			s = t.Add(r)
			frac = frac.Sub(t)
			q = q.Add(r)
		}
		frac = frac.Lsh(1)
		r = r.Rsh(1)
	}

	// final rounding
	if !frac.IsZero() {
		q = q.Add(q.And(ints.Uint128{0, 1}))
	}
	q = q.Rsh(1)
	q = q.Add(ints.Uint128{uint64(exp-1+bias128) << (shift128 - 64), 0})
	return Float128(q)
}

// Eq returns a == b.
// NaNs are not equal to anything, including NaN.
// -0.0 and 0.0 are equal.
func (a Float128) Eq(b Float128) bool {
	if a.IsNaN() || b.IsNaN() {
		return false
	}
	if a == b {
		// a and b have the same bit pattern.
		return true
	}

	// check -0 == 0
	return (a[0]|b[0])&^signMask128[0] == 0 && (a[1]|b[1]) == 0
}

// Ne returns a != b.
// NaNs are not equal to anything, including NaN.
// -0.0 and 0.0 are equal.
func (a Float128) Ne(b Float128) bool {
	return !a.Eq(b)
}

// Lt returns a < b.
//
// Special cases are:
//
//	Lt(NaN, x) == false
//	Lt(x, NaN) == false
func (a Float128) Lt(b Float128) bool {
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
func (a Float128) Gt(b Float128) bool {
	return b.Lt(a)
}

// Le returns a <= b.
//
// Special cases are:
//
//	Le(x, NaN) == false
//	Le(NaN, x) == false
func (a Float128) Le(b Float128) bool {
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
func (a Float128) Ge(b Float128) bool {
	return b.Le(a)
}

// normalize returns the sign, exponent, and normalized fraction of a.
func (a Float128) normalize() (sign uint64, exp int, frac ints.Uint128) {
	b := ints.Uint128(a)
	sign = b[0] & signMask128[0]
	exp = int((b[0]>>(shift128-64))&mask128) - bias128
	frac = b.And(fracMask128)
	if exp == -bias128 {
		// a is subnormal
		// normalize
		l := frac.BitLen()
		frac = frac.Lsh(uint(shift128-l) + 1)
		exp = l - (bias128 + shift128)
		return
	}

	// a is normal
	frac[0] = frac[0] | (1 << (shift128 - 64))
	return
}

// split returns the sign, exponent, and fraction of a.
func (a Float128) split() (sign uint64, exp int, frac ints.Uint128) {
	b := ints.Uint128(a)
	sign = b[0] & signMask128[0]
	exp = int((b[0]>>(shift128-64))&mask128) - bias128
	frac = b.And(fracMask128)
	if exp == -bias128 {
		// a is subnormal
		exp++
	} else {
		// a is normal
		frac[0] = frac[0] | (1 << (shift128 - 64))
	}
	return
}

func (a Float128) comparable() ints.Int128 {
	i := ints.Int128(a)
	i = i.Xor(ints.Int128{
		uint64(int64(i[0]) >> 63 & 0x7fff_ffff_ffff_ffff),
		uint64(int64(i[0]) >> 63),
	})
	i = i.Add(ints.Int128{0, i[0] >> 63}) // normalize -0 to 0
	return i
}

// FMA128 returns x * y + z, computed with only one rounding.
// (That is, FMA128 returns the fused multiply-add of x, y, and z.)
func FMA128(x, y, z Float128) Float128 {
	if x.IsZero() || y.IsZero() || x[0]&(mask128<<(shift128-64)) == (mask128<<(shift128-64)) || y[0]&(mask128<<(shift128-64)) == (mask128<<(shift128-64)) {
		return x.Mul(y).Add(z)
	}
	if z.IsZero() {
		return x.Mul(y)
	}
	// Handle non-finite z separately. Evaluating x*y+z where
	// x and y are finite, but z is infinite, should always result in z.
	if z[0]&(mask128<<(shift128-64)) == (mask128 << (shift128 - 64)) {
		return z
	}

	// Split x, y, z into sign, exponent, mantissa.
	signX, expX, fracX := x.normalize()
	signY, expY, fracY := y.normalize()
	signZ, expZ, fracZ0 := z.normalize()

	// Compute product p = x*y as sign, exponent, mantissa.
	expP := expX + expY + 1
	fracP := fracX.Lsh(14).Mul256(fracY.Lsh(15))
	signP := signX ^ signY // product sign

	// Normalize the product
	is254zero := uint((^fracP[0] >> 62) & 1)
	fracP = fracP.Lsh(is254zero)
	expP -= int(is254zero)

	fracZ := fracZ0.Uint256().Lsh(14 + 128)

	// Swap addition operands so |p| >= |z|
	if expP < expZ || expP == expZ && fracP.Cmp(fracZ) < 0 {
		signP, signZ = signZ, signP
		expP, expZ = expZ, expP
		fracP, fracZ = fracZ, fracP
	}

	// Special case: if p == -z the result is always +0 since neither operand is zero.
	if signP != signZ && expP == expZ && fracP.Cmp(fracZ) == 0 {
		return Float128{0, 0}
	}

	// Align mantissa
	fracZ = shrcompress256(fracZ, uint(expP-expZ))

	// Compute resulting significands, normalizing if necessary.
	var frac ints.Uint128
	if signP == signZ {
		// Adding fracP + fracZ
		fracP = fracP.Add(fracZ)
		expP += int(fracP[0] >> 63)
		frac = shrcompress256(fracP, uint(128+fracP[0]>>63)).Uint128()
	} else {
		// Subtracting fracP - fracZ
		fracP = fracP.Sub(fracZ)
		nz := fracP.LeadingZeros() - 1
		expP -= nz
		frac = shrcompress256(fracP.Lsh(uint(nz)), 128).Uint128()
	}

	// check for underflow
	expP += bias128
	if expP <= 0 {
		n := uint(1 - expP)
		frac = roundToNearestEven128(frac, n+14)
		return Float128{signP | frac[0], frac[1]}
	}

	// Round and break ties to even
	frac = roundToNearestEven128(frac, 14)
	if frac[0]&(1<<(shift128+1-64)) != 0 {
		expP++
		frac = frac.Rsh(1)
	}
	if expP >= mask128 {
		// overflow
		return Float128{signP | uvinf128[0], uvinf128[1]}
	}
	return Float128{
		signP | uint64(expP)<<(shift128-64) | frac[0]&fracMask128[0],
		frac[1] & fracMask128[1],
	}
}
