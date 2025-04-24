package floats

import (
	"github.com/shogo82148/ints"
)

const (
	mask128  = 0x7fff       // mask for exponent
	shift128 = 128 - 15 - 1 // shift for exponent
	bias128  = 16383        // bias for exponent
)

// var (
// 	signMask128 = ints.Uint128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000} // mask for sign bit
// 	fracMask128 = ints.Uint128{0x0000_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}
// )

// Float128 is a 128-bit floating-point number.
type Float128 ints.Uint128
