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

// Int64 returns the integer value of a, rounding towards zero.
// If a cannot be represented in an int64, the result is undefined.
func (a Float128) Int64() int64 {
	sign, exp, frac := a.split()
	frac = frac.Rsh(uint(shift128 - exp))
	ret := int64(frac.Uint64())
	if sign != 0 {
		ret = -ret
	}
	return ret
}

func (a Float128) isZero() bool {
	return (a[0]&^signMask128[0])|a[1] == 0
}

// Mul returns the product of a and b.
func (a Float128) Mul(b Float128) Float128 {
	if a.IsNaN() || b.IsNaN() {
		// a * NaN = NaN
		// NaN * b = NaN
		return Float128(uvnan128)
	}

	signA, expA, fracA := a.split()
	signB, expB, fracB := b.split()
	sign := signA ^ signB

	// handle special cases
	if expA == mask128-bias128 {
		// NaN check is done above; a is ±inf
		if b.isZero() {
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
		if a.isZero() {
			// 0 * ±inf = NaN
			return Float128(uvnan128)
		} else {
			// +finite * ±inf = ±inf
			// -finite * ±inf = ∓inf
			return Float128{sign | uvinf128[0], uvinf128[1]}
		}
	}
	if a.isZero() || b.isZero() {
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
		one := ints.Uint256{0, 0, 0, 1}
		frac = frac.Add(one.Lsh(uint(shift - 1)).Sub(one)).Add(frac.Rsh(uint(shift)).And(one)) // round to nearest even
		frac = frac.Rsh(uint(shift))
		return Float128{sign | frac[2], frac[3]}
	}

	exp = expA + expB + bias128
	one := ints.Uint256{0, 0, 0, 1}
	frac = frac.Add(one.Lsh(uint(shift - 1)).Sub(one)).Add(frac.Rsh(uint(shift)).And(one)) // round to nearest even
	shift = frac.BitLen() - (shift128 + 1)
	exp += shift - shift128
	if exp >= mask128 {
		// overflow
		return Float128{sign | uvinf128[0], uvinf128[1]}
	}

	frac = frac.Rsh(uint(shift))
	return Float128{
		sign | uint64(exp)<<(shift128-64) | frac[2]&fracMask128[0],
		frac[3],
	}
}

// Quo returns the quotient of a and b.
func (a Float128) Quo(b Float128) Float128 {
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
			return Float128(uvnan128)
		}
		// ±finite / 0 = ±inf
		return Float128{sign | uvinf128[0], uvinf128[1]}
	}
	if a.isZero() {
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
	frac256[3] |= squash256(mod)
	frac := frac256.Uint128()

	if exp <= 0 {
		// the result is subnormal
		shift := -exp + 3 + 1
		frac = frac.Add(ints.Uint128{0, 1}.Lsh(uint(shift - 1)).Sub(ints.Uint128{0, 1})).Add(frac.Rsh(uint(shift)).And(ints.Uint128{0, 1})) // round to nearest even
		frac = frac.Rsh(uint(shift))
		return Float128{sign | frac[0], frac[1]}
	}

	// round-to-nearest-even (guard+round+sticky are in the low 3 bits)
	frac = frac.Add(ints.Uint128{0, 0b11}).Add(frac.Rsh(3).And(ints.Uint128{0, 1}))
	// detect carry-out caused by rounding
	if frac.BitLen() > shift128+3+1 {
		frac = frac.Rsh(1)
		exp++
		if exp >= mask128 { // overflow -> ±Inf
			return Float128{sign | uvinf128[0], uvinf128[1]}
		}
	}
	frac = frac.Rsh(3)
	return Float128{sign | uint64(exp)<<(shift128-64) | frac[0]&fracMask128[0], frac[1] & fracMask128[1]}
}

func (a Float128) split() (sign uint64, exp int, frac ints.Uint128) {
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
