package floats

import "github.com/shogo82148/ints"

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

// Int64 returns the integer value of a, rounding towards zero.
// If a cannot be represented in an int64, the result is undefined.
func (a Float256) Int64() int64 {
	sign, exp, frac := a.split()
	frac = frac.Rsh(uint(shift256 - exp))
	ret := int64(frac.Uint64())
	if sign != 0 {
		ret = -ret
	}
	return ret
}

func (a Float256) isZero() bool {
	return (a[0]&^signMask256[0])|a[1]|a[2]|a[3] == 0
}

// Neg returns the negation of a.
func (a Float256) Neg() Float256 {
	return Float256{a[0] ^ signMask256[0], a[1], a[2], a[3]}
}

// Mul returns the product of a and b.
func (a Float256) Mul(b Float256) Float256 {
	if a.IsNaN() || b.IsNaN() {
		// a * NaN = NaN
		// NaN * b = NaN
		return Float256(uvnan256)
	}
	signA, expA, fracA := a.split()
	signB, expB, fracB := b.split()
	sign := signA ^ signB

	// handle special cases
	if expA == mask256-bias256 {
		// NaN check is done above; a is ±inf
		if b.isZero() {
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
		if a.isZero() {
			// 0 * ±inf = NaN
			return Float256(uvnan256)
		} else {
			// +finite * ±inf = ±inf
			// -finite * ±inf = ∓inf
			return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
		}
	}
	if a.isZero() || b.isZero() {
		// +0 * ±finite = ±0
		// -0 * ±finite = ∓0
		return Float256{sign, 0, 0, 0}
	}

	// normal case
	exp := expA + expB
	frac := fracA.Mul512(fracB)
	shift := frac.BitLen() - (shift256 + 1)
	exp += shift - shift256

	if exp < -(bias256 + shift256) {
		// underflow
		return Float256{sign, 0, 0, 0}
	} else if exp <= -bias256 {
		// the result is subnormal
		// normalize
		shift := shift256 - (expA + expB + bias256) + 1
		frac = roundToNearestEven512(frac, uint(shift))
		frac = frac.Rsh(uint(shift))
		return Float256{sign | frac[4], frac[5], frac[6], frac[7]}
	}

	exp = expA + expB + bias256
	frac = roundToNearestEven512(frac, uint(shift))
	shift = frac.BitLen() - (shift256 + 1)
	exp += shift - shift256
	if exp >= mask256 {
		// overflow
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}

	frac = frac.Rsh(uint(shift))
	return Float256{
		sign | uint64(exp)<<(shift256-192) | frac[4]&fracMask256[0],
		frac[5],
		frac[6],
		frac[7],
	}
}

// Quo returns the quotient of a and b.
func (a Float256) Quo(b Float256) Float256 {
	if a.IsNaN() {
		return a
	}
	if b.IsNaN() {
		return b
	}

	signA, expA, fracA := a.split()
	signB, expB, fracB := b.split()
	sign := signA ^ signB

	if b.isZero() {
		if a.isZero() {
			// 0 / 0 = NaN
			return Float256(uvnan256)
		}
		// ±finite / 0 = ±inf
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}
	if a.isZero() {
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
		fracA = fracA.Lsh(1)
	}
	if exp >= mask256 {
		// overflow
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}

	shift := shift256 + 3 // 1 for the implicit bit, 1 for the rounding bit, 1 for the guard bit
	fracA512 := fracA.Uint512().Lsh(uint(shift))
	fracB512 := fracB.Uint512()
	frac512, mod := fracA512.DivMod(fracB512)
	frac512[7] |= squash512(mod)
	frac := frac512.Uint256()

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
	if frac.BitLen() > shift256+1 {
		frac = frac.Rsh(1)
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
	if a.IsNaN() {
		return a
	}
	if b.IsNaN() {
		return b
	}
	if a.isZero() {
		if b.isZero() {
			//  0 +  0 =  0
			//  0 + -0 =  0
			// -0 +  0 =  0
			// -0 + -0 = -0
			return Float256{a[0] & b[0], 0, 0, 0}
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

	if expA < expB {
		// swap a and b
		signA, signB = signB, signA
		expA, expB = expB, expA
		fracA, fracB = fracB, fracA
	}

	// add the fractions
	const offset = 256
	fracA512 := ints.Uint512{fracA[0], fracA[1], fracA[2], fracA[3], 0, 0, 0, 0}
	fracB512 := ints.Uint512{fracB[0], fracB[1], fracB[2], fracB[3], 0, 0, 0, 0}
	fracB512 = fracB512.Rsh(uint(expA - expB))
	if signA != 0 {
		fracA512 = fracA512.Neg()
	}
	if signB != 0 {
		fracB512 = fracB512.Neg()
	}
	frac512 := fracA512.Add(fracB512)
	sign := uint64(0)
	if ints.Int512(frac512).Sign() < 0 {
		sign = signMask256[0]
		frac512 = frac512.Neg()
	}

	shift := frac512.BitLen() - (shift256 + 1)
	exp := expA + shift - offset

	if frac512.IsZero() || exp < -(bias256+shift256) {
		// underflow
		return Float256{sign, 0, 0, 0}
	}
	if exp <= -bias256 {
		// the result is subnormal
		shift := offset - (expA + bias256) + 1
		frac512 = roundToNearestEven512(frac512, uint(shift))
		frac512 = frac512.Rsh(uint(shift))
		return Float256{sign | frac512[4], frac512[5], frac512[6], frac512[7]}
	}
	if exp >= mask256-bias256 {
		// overflow
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}

	frac512 = roundToNearestEven512(frac512, uint(shift))
	// detect carry-out caused by rounding
	if frac512.BitLen() > shift256+shift+1 {
		frac512 = frac512.Rsh(1)
		exp++
		if exp >= mask256 {
			// overflow
			return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
		}
	}
	frac512 = frac512.Rsh(uint(shift))
	return Float256{
		sign | uint64(exp+bias256)<<(shift256-192) | frac512[4]&fracMask256[0],
		frac512[5],
		frac512[6],
		frac512[7],
	}
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
	case a.isZero() || a.IsNaN() || a.IsInf(1):
		return a
	case a[0]&signMask256[0] != 0:
		return Float256(uvnan256)
	}

	_, exp, frac := a.split()
	if exp%2 != 0 {
		// odd exp. double x to make it even
		frac = frac.Lsh(1)
	}
	// exponent of square root
	exp >>= 1

	// generate sqrt(frac) bit by bit
	frac = frac.Lsh(1)
	var q, s ints.Uint256 // q = sqrt(frac)
	r := ints.Uint256{1 << (shift256 + 1 - 192), 0, 0, 0}
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
		q = q.Add(q.And(ints.Uint256{0, 1}))
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

func (a Float256) split() (sign uint64, exp int, frac ints.Uint256) {
	b := ints.Uint256(a)
	sign = b[0] & signMask256[0]
	exp = int((b[0]>>(shift256-192))&mask256) - bias256
	frac = b.And(fracMask256)
	if exp == -bias256 {
		// a is subnormal
		// normalize
		l := frac.BitLen()
		frac = frac.Lsh(uint(shift256-l) + 1)
		exp = l - (bias256 + shift256)
		return
	}

	// a is normal
	frac[0] = frac[0] | (1 << (shift256 - 192))
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
	if x.isZero() || y.isZero() || x[0]&(mask256<<(shift256-192)) == mask256<<(shift256-192) || y[0]&(mask256<<(shift256-192)) == mask256<<(shift256-192) {
		return x.Mul(y).Add(z)
	}
	if z.isZero() {
		return x.Mul(y)
	}
	// Handle non-finite z separately. Evaluating x*y+z where
	// x and y are finite, but z is infinite, should always result in z.
	if z[0]&(mask256<<(shift256-192)) == mask256<<(shift256-192) {
		return z
	}

	// Split x, y, z into sign, exponent, mantissa.
	signX, expX, fracX := x.split()
	signY, expY, fracY := y.split()
	signZ, expZ, fracZ0 := z.split()

	// Compute product p = x*y as sign, exponent, mantissa.
	expP := expX + expY + 1
	fracP := fracX.Lsh(18).Mul512(fracY.Lsh(19))
	signP := signX ^ signY // product sign

	// Normalize the product
	is510zero := uint((^fracP[0] >> 62) & 1)
	fracP = fracP.Lsh(is510zero)
	expP -= int(is510zero)

	fracZ := fracZ0.Uint512().Lsh(18 + 256)

	// Swap addition operands so |p| >= |z|
	if expP < expZ || expP == expZ && fracP.Cmp(fracZ) < 0 {
		signP, signZ = signZ, signP
		expP, expZ = expZ, expP
		fracP, fracZ = fracZ, fracP
	}

	// Special case: if p == -z the result is always +0 since neither operand is zero.
	if signP != signZ && expP == expZ && fracP.Cmp(fracZ) == 0 {
		return Float256{0, 0, 0, 0}
	}

	// Align mantissa
	fracZ = shrcompress512(fracZ, uint(expP-expZ))

	// Compute resulting significands, normalizing if necessary.
	var frac ints.Uint256
	if signP == signZ {
		// Adding fracP + fracZ
		fracP = fracP.Add(fracZ)
		expP += int(fracP[0] >> 63)
		frac = shrcompress512(fracP, uint(256+fracP[0]>>63)).Uint256()
	} else {
		// Subtracting fracP - fracZ
		fracP = fracP.Sub(fracZ)
		nz := fracP.LeadingZeros() - 1
		expP -= nz
		frac = shrcompress512(fracP.Lsh(uint(nz)), 256).Uint256()
	}

	// check for underflow
	expP += bias256
	if expP <= 0 {
		n := uint(1 - expP)
		frac = roundToNearestEven256(frac, n+18)
		return Float256{signP | frac[0], frac[1], frac[2], frac[3]}
	}

	// Round and break ties to even
	frac = roundToNearestEven256(frac, 18)
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
