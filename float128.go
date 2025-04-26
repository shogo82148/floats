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
	// uvnan128    = ints.Uint128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000} // NaN value for Float128
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
	}

	// a is normal
	frac[0] = frac[0] | (1 << (shift128 - 64))
	return
}
