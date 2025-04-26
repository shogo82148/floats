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
