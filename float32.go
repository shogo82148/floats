package floats

import (
	"math"
	"math/bits"
)

const (
	mask32     = 0xff       // mask for exponent
	shift32    = 32 - 8 - 1 // shift for exponent
	bias32     = 127        // bias for exponent
	signMask32 = 1 << 31    // mask for sign bit
	fracMask32 = 1<<shift32 - 1
)

// Float32 is a 32-bit floating-point number.
type Float32 float32

func (a Float32) split() (sign uint32, exp int, frac uint32) {
	b := math.Float32bits(float32(a))
	sign = uint32(b & signMask32)
	exp = int((b>>shift32)&mask32) - bias32
	frac = uint32(b & fracMask32)

	// normalize a
	if exp == -bias32 {
		// subnormal number
		l := bits.Len32(frac)
		frac <<= shift32 - l + 1
		exp = -(bias32 + shift32) + l
	} else {
		// normal number
		frac |= 1 << shift32
	}
	return
}
