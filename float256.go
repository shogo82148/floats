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
