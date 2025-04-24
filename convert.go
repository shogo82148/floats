package floats

import (
	"math"
	"math/bits"
)

// Float16 returns a itself.
func (a Float16) Float16() Float16 {
	return a
}

// Float32 converts a to a Float32.
func (a Float16) Float32() Float32 {
	sign := uint32(a&signMask16) << (32 - 16)
	exp := uint32(a>>shift16) & mask16
	frac := uint32(a & fracMask16)

	if exp == 0 {
		// a is subnormal number
		if frac == 0 {
			// a is zero
			return Float32(math.Float32frombits(sign))
		} else {
			l := bits.Len32(frac)
			frac = (frac << (shift16 - l + 1)) & fracMask16
			exp = bias32 - (bias16 + shift16) + uint32(l)
		}
	} else if exp == mask16 {
		// a is infinity or NaN
		exp = mask32
	} else {
		// a is normal number
		exp += bias32 - bias16
	}
	exp <<= shift32
	frac <<= shift32 - shift16
	return Float32(math.Float32frombits(sign | exp | frac))
}

// Float64 converts a to a Float64.
func (a Float16) Float64() Float64 {
	sign := uint64(a&signMask16) << (64 - 16)
	exp := uint64(a>>shift16) & mask16
	frac := uint64(a & fracMask16)

	if exp == 0 {
		// a is subnormal number
		if frac == 0 {
			// a is zero
			return Float64(math.Float64frombits(sign))
		} else {
			l := bits.Len64(frac)
			frac = (frac << (shift16 - l + 1)) & fracMask16
			exp = bias64 - (bias16 + shift16) + uint64(l)
		}
	} else if exp == mask16 {
		// a is infinity or NaN
		exp = mask64
	} else {
		// a is normal number
		exp += bias64 - bias16
	}
	exp <<= shift64
	frac <<= shift64 - shift16
	return Float64(math.Float64frombits(sign | exp | frac))
}

// Float128 converts a to a Float128.
func (a Float16) Float128() Float128 {
	sign := uint64(a&signMask16) << (64 - 16)
	exp := uint64(a>>shift16) & mask16
	frac := uint64(a & fracMask16)

	if exp == 0 {
		// a is subnormal number
		if frac == 0 {
			// a is zero
			return Float128{sign, 0}
		} else {
			l := bits.Len64(frac)
			frac = (frac << (shift16 - l + 1)) & fracMask16
			exp = bias128 - (bias16 + shift16) + uint64(l)
		}
	} else if exp == mask16 {
		// a is infinity or NaN
		exp = mask128
	} else {
		// a is normal number
		exp += bias128 - bias16
	}
	exp <<= shift128 - 64
	frac <<= shift128 - 64 - shift16
	return Float128{sign | exp | frac, 0}
}

// Float256 converts a to a Float256.
func (a Float16) Float256() Float256 {
	sign := uint64(a&signMask16) << (64 - 16)
	exp := uint64(a>>shift16) & mask16
	frac := uint64(a & fracMask16)

	if exp == 0 {
		// a is subnormal number
		if frac == 0 {
			// a is zero
			return Float256{sign, 0, 0, 0}
		} else {
			l := bits.Len64(frac)
			frac = (frac << (shift16 - l + 1)) & fracMask16
			exp = bias256 - (bias16 + shift16) + uint64(l)
		}
	} else if exp == mask16 {
		// a is infinity or NaN
		exp = mask256
	} else {
		// a is normal number
		exp += bias256 - bias16
	}

	exp <<= shift256 - 192
	frac <<= shift256 - 192 - shift16
	return Float256{sign | exp | frac, 0, 0, 0}
}

// Float16 converts a to a Float16.
func (a Float32) Float16() Float16 {
	return Float16(0) // TODO: implement
}

// Float32 returns a itself.
func (a Float32) Float32() Float32 {
	return a
}

// Float64 converts a to a Float64.
func (a Float32) Float64() Float64 {
	return Float64(a)
}

// Float128 converts a to a Float128.
func (a Float32) Float128() Float128 {
	return Float128{0, 0} // TODO: implement
}

// Float256 converts a to a Float256.
func (a Float32) Float256() Float256 {
	return Float256{0, 0, 0, 0} // TODO: implement
}

// Float16 converts a to a Float16.
func (a Float64) Float16() Float16 {
	return Float16(0) // TODO: implement
}

// Float32 converts a to a Float32.
func (a Float64) Float32() Float32 {
	return Float32(a)
}

// Float64 returns a itself.
func (a Float64) Float64() Float64 {
	return a
}

// Float128 converts a to a Float128.
func (a Float64) Float128() Float128 {
	return Float128{0, 0} // TODO: implement
}

// Float256 converts a to a Float256.
func (a Float64) Float256() Float256 {
	return Float256{0, 0, 0, 0} // TODO: implement
}

// Float16 converts a to a Float16.
func (a Float128) Float16() Float16 {
	return Float16(0) // TODO: implement
}

// Float32 converts a to a Float32.
func (a Float128) Float32() Float32 {
	return Float32(0) // TODO: implement
}

// Float64 converts a to a Float64.
func (a Float128) Float64() Float64 {
	return Float64(0) // TODO: implement
}

// Float128 returns a itself.
func (a Float128) Float128() Float128 {
	return a
}

// Float256 converts a to a Float256.
func (a Float128) Float256() Float256 {
	return Float256{0, 0, 0, 0} // TODO: implement
}

// Float16 converts a to a Float16.
func (a Float256) Float16() Float16 {
	return Float16(0) // TODO: implement
}

// Float32 converts a to a Float32.
func (a Float256) Float32() Float32 {
	return Float32(0) // TODO: implement
}

// Float64 converts a to a Float64.
func (a Float256) Float64() Float64 {
	return Float64(0) // TODO: implement
}

// Float128 converts a to a Float128.
func (a Float256) Float128() Float128 {
	return Float128{0, 0} // TODO: implement
}

// Float256 returns a itself.
func (a Float256) Float256() Float256 {
	return a
}
