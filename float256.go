package floats

import "github.com/shogo82148/ints"

const (
	mask256  = 0x7_ffff     // mask for exponent
	shift256 = 256 - 19 - 1 // shift for exponent
	bias256  = 262143       // bias for exponent
)

var (
	// signMask256 = ints.Uint256{0x8000_0000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000} // mask for sign bit
	fracMask256 = ints.Uint256{0x0000_0fff_ffff_ffff, 0xffff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}
)

// Float256 is a 256-bit floating-point number.
type Float256 ints.Uint256

// IsNaN reports whether a is an IEEE 754 “not-a-number” value.
func (a Float256) IsNaN() bool {
	return a[0]&(mask256<<(shift256-192)) == (mask256<<(shift256-192)) &&
		!ints.Uint256(a).And(fracMask256).IsZero()
}
