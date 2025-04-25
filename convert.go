package floats

import (
	"math"
	"math/bits"

	"github.com/shogo82148/ints"
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
	b := math.Float32bits(float32(a))
	sign := uint16((b & signMask32) >> (32 - 16))
	exp := int((b >> shift32) & mask32)

	if exp == mask32 {
		// a is ±infinity or NaN
		frac := b & fracMask32
		if frac == 0 {
			// a is ±infinity
			return Float16(sign | mask16<<shift16)
		} else {
			// a is NaN
			return Float16(sign | uvnan16)
		}
	}

	exp -= bias32
	if exp <= -bias16 {
		// the result is subnormal number
		roundBit := -exp + shift32 - (bias16 + shift16 - 1)
		frac := (b & fracMask32) | (1 << shift32)
		halfMinusULP := uint32(1<<(roundBit-1) - 1)
		frac += halfMinusULP + ((frac >> uint(roundBit)) & 1) // round to nearest even
		return Float16(sign | uint16(frac>>uint(roundBit)))
	}

	// the result is normal number
	const halfMinusULP = 1<<(shift32-shift16-1) - 1
	b += halfMinusULP + ((b >> uint(shift32-shift16)) & 1) // round to nearest even

	exp16 := uint16((b>>shift32)&mask32) - bias32 + bias16
	if exp16 >= mask16 {
		// overflow
		return Float16(sign | mask16<<shift16)
	}
	frac16 := uint16(b>>(shift32-shift16)) & fracMask16
	return Float16(sign | (exp16 << shift16) | frac16)
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
	b := math.Float32bits(float32(a))
	sign := uint64(b&signMask32) << (64 - 32)
	exp := int((b >> shift32) & mask32)
	frac := uint64(b & fracMask32)

	if exp == mask32 {
		// a is ±infinity or NaN
		return Float128{sign | mask128<<(shift128-64) | frac<<(shift128-shift32-64), 0}
	} else if exp == 0 {
		// a is subnormal
		if frac == 0 {
			// a is zero
			return Float128{sign, 0}
		}

		// normalize a
		l := bits.Len64(frac)
		exp = l - shift32
		frac = (frac << (shift32 - l + 1)) & fracMask32
	}

	exp += bias128 - bias32
	frac <<= shift128 - shift32 - 64
	return Float128{
		sign | uint64(exp)<<(shift128-64) | frac,
		0,
	}
}

// Float256 converts a to a Float256.
func (a Float32) Float256() Float256 {
	b := math.Float32bits(float32(a))
	sign := uint64(b&signMask32) << (64 - 32)
	exp := int((b >> shift32) & mask32)
	frac := uint64(b & fracMask32)

	if exp == mask32 {
		// a is ±infinity or NaN
		return Float256{
			sign | mask256<<(shift256-192) | frac<<(shift256-shift32-192),
			0,
			0,
			0,
		}
	} else if exp == 0 {
		// a is subnormal
		if frac == 0 {
			// a is zero
			return Float256{sign, 0, 0, 0}
		}

		// normalize a
		l := bits.Len64(frac)
		exp = l - shift32
		frac = (frac << (shift32 - l + 1)) & fracMask32
	}

	exp += bias256 - bias32
	frac <<= shift256 - shift32 - 192
	return Float256{
		sign | uint64(exp)<<(shift256-192) | frac,
		0,
		0,
		0,
	}
}

// Float16 converts a to a Float16.
func (a Float64) Float16() Float16 {
	b := math.Float64bits(float64(a))
	sign := uint16((b & signMask64) >> (64 - 16))
	exp := int((b >> shift64) & mask64)

	if exp == mask64 {
		// a is ±infinity or NaN
		frac := b & fracMask64
		if frac == 0 {
			// a is ±infinity
			return Float16(sign | mask16<<shift16)
		} else {
			// a is NaN
			return Float16(sign | uvnan16)
		}
	}

	exp -= bias64
	if exp <= -bias16 {
		// the result is subnormal number
		roundBit := -exp + shift64 - (bias16 + shift16 - 1)
		frac := (b & fracMask64) | (1 << shift64)
		halfMinusULP := uint64(1<<(roundBit-1) - 1)
		frac += halfMinusULP + ((frac >> uint(roundBit)) & 1) // round to nearest even
		return Float16(sign | uint16(frac>>roundBit))
	}

	// the result is normal number
	const halfMinusULP = 1<<(shift64-shift16-1) - 1
	b += halfMinusULP + ((b >> uint(shift64-shift16)) & 1) // round to nearest even

	exp16 := uint16((b>>shift64)&mask64) - bias64 + bias16
	if exp16 >= mask16 {
		// overflow
		return Float16(sign | mask16<<shift16)
	}
	frac16 := uint16(b>>(shift64-shift16)) & fracMask16
	return Float16(sign | (exp16 << shift16) | frac16)
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
	b := math.Float64bits(float64(a))
	sign := uint64(b & signMask64)
	exp := int((b >> shift64) & mask64)
	frac := uint64(b & fracMask64)

	if exp == mask64 {
		// a is ±infinity or NaN
		return Float128{
			sign | mask128<<(shift128-64) | frac>>(64-shift128+shift64),
			frac << (shift128 - shift64),
		}
	} else if exp == 0 {
		// a is subnormal
		if frac == 0 {
			// a is zero
			return Float128{sign, 0}
		}

		// normalize a
		l := bits.Len64(frac)
		exp = l - shift64
		frac = (frac << (shift64 - l + 1)) & fracMask64
	}

	exp += bias128 - bias64
	return Float128{
		sign | uint64(exp)<<(shift128-64) | frac>>(64-shift128+shift64),
		frac << (shift128 - shift64),
	}
}

// Float256 converts a to a Float256.
func (a Float64) Float256() Float256 {
	b := math.Float64bits(float64(a))
	sign := uint64(b & signMask64)
	exp := int((b >> shift64) & mask64)
	frac := uint64(b & fracMask64)

	if exp == mask64 {
		// a is ±infinity or NaN
		return Float256{
			sign | mask256<<(shift256-192) | frac>>(192-shift256+shift64),
			frac << (shift256 - shift64 - 128),
			0,
			0,
		}
	} else if exp == 0 {
		// a is subnormal
		if frac == 0 {
			// a is zero
			return Float256{sign, 0, 0, 0}
		}

		// normalize a
		l := bits.Len64(frac)
		exp = l - shift64
		frac = (frac << (shift64 - l + 1)) & fracMask64
	}

	exp += bias256 - bias64
	return Float256{
		sign | uint64(exp)<<(shift256-192) | frac>>(192-shift256+shift64),
		frac << (shift256 - shift64 - 128),
		0,
		0,
	}
}

// Float16 converts a to a Float16.
func (a Float128) Float16() Float16 {
	sign := uint16((a[0] & signMask128[0]) >> (64 - 16))
	exp := int((a[0] >> (shift128 - 64)) & mask128)

	if exp == mask128 {
		// a is ±infinity or NaN
		frac := ints.Uint128(a).And(fracMask128)
		if frac.IsZero() {
			// a is ±infinity
			return Float16(sign | mask16<<shift16)
		} else {
			// a is NaN
			return Float16(sign | uvnan16)
		}
	}

	exp -= bias128
	if exp <= -bias16 {
		// the result is subnormal number
		frac := ints.Uint128(a).And(fracMask128)
		frac[0] |= (1 << (shift128 - 64))
		// round to nearest even
		roundBit := -exp + shift128 - (bias16 + shift16 - 1) - 64
		halfMinusULP := uint64(1<<(roundBit-1) - 1)
		frac[0] |= squash64(frac[1])
		frac[0] += halfMinusULP + ((a[0] >> uint(roundBit)) & 1)
		return Float16(sign | uint16(frac[0]>>roundBit))
	}

	// the result is normal number
	// round to nearest even
	const halfMinusULP = 1<<(shift128-shift16-64-1) - 1
	a[0] |= squash64(a[1])
	a[0] += halfMinusULP + ((a[0] >> uint(shift128-shift16-64)) & 1)

	exp16 := uint16((a[0]>>(shift128-64))&mask128) - bias128 + bias16
	if exp16 >= mask16 {
		// overflow
		return Float16(sign | mask16<<shift16)
	}
	frac16 := uint16(a[0]>>(shift128-shift16-64)) & fracMask16
	return Float16(sign | (exp16 << shift16) | frac16)
}

// Float32 converts a to a Float32.
func (a Float128) Float32() Float32 {
	sign := uint32((a[0] & signMask128[0]) >> (64 - 32))
	exp := int((a[0] >> (shift128 - 64)) & mask128)

	if exp == mask128 {
		// a is ±infinity or NaN
		frac := ints.Uint128(a).And(fracMask128)
		if frac.IsZero() {
			// a is ±infinity
			return Float32(math.Float32frombits(sign | mask32<<shift32))
		} else {
			// a is NaN
			return Float32(math.Float32frombits(sign | uvnan32))
		}
	}

	exp -= bias128
	if exp <= -bias32 {
		// the result is subnormal number
		frac := ints.Uint128(a).And(fracMask128)
		frac[0] |= (1 << (shift128 - 64))
		// round to nearest even
		roundBit := -exp + shift128 - (bias32 + shift32 - 1) - 64
		halfMinusULP := uint64(1<<(roundBit-1) - 1)
		frac[0] |= squash64(frac[1])
		frac[0] += halfMinusULP + ((a[0] >> uint(roundBit)) & 1)
		return Float32(math.Float32frombits(sign | uint32(frac[0]>>uint(roundBit))))
	}

	// the result is normal number
	// round to nearest even
	const halfMinusULP = 1<<(shift128-shift32-64-1) - 1
	a[0] |= squash64(a[1])
	a[0] += halfMinusULP + ((a[0] >> uint(shift128-shift32-64)) & 1)

	exp32 := uint32((a[0]>>(shift128-64))&mask128) - bias128 + bias32
	if exp32 >= mask32 {
		// overflow
		return Float32(math.Float32frombits(sign | mask32<<shift32))
	}
	frac32 := uint32(a[0]>>(shift128-shift32-64)) & fracMask32
	return Float32(math.Float32frombits(sign | (exp32 << shift32) | frac32))
}

// Float64 converts a to a Float64.
func (a Float128) Float64() Float64 {
	b := ints.Uint128(a)
	sign := b[0] & signMask128[0]
	exp := int((b[0] >> (shift128 - 64)) & mask128)

	if exp == mask128 {
		// a is ±infinity or NaN
		frac := b.And(fracMask128)
		if frac.IsZero() {
			// a is ±infinity
			return Float64(math.Float64frombits(sign | mask64<<shift64))
		} else {
			// a is NaN
			return Float64(math.Float64frombits(sign | uvnan64))
		}
	}

	exp -= bias128
	if exp <= -bias64 {
		// the result is subnormal number
		frac := b.And(fracMask128)
		frac[0] |= 1 << (shift128 - 64)
		// round to nearest even
		roundBit := uint(-exp + shift128 - (bias64 + shift64 - 1))
		halfMinusULP := ints.Uint128{0, 1}.Lsh(roundBit - 1).Sub(ints.Uint128{0, 1})
		frac = frac.Add(halfMinusULP).Add(frac.Rsh(roundBit).And(ints.Uint128{0, 1}))
		frac = frac.Rsh(roundBit)
		return Float64(math.Float64frombits(sign | frac[1]))
	}

	// the result is normal number
	// round to nearest even
	const halfMinusULP = 1<<(shift128-shift64-1) - 1
	b = b.Add(ints.Uint128{0, halfMinusULP}).Add(b.Rsh(shift128 - shift64).And(ints.Uint128{0, 1}))

	exp64 := uint64((b[0]>>(shift128-64))&mask128) - bias128 + bias64
	if exp64 >= mask64 {
		// overflow
		return Float64(math.Float64frombits(sign | mask64<<shift64))
	}
	frac64 := uint64(b.Rsh(shift128-shift64).Uint64() & fracMask64)
	return Float64(math.Float64frombits(sign | (exp64 << shift64) | frac64))
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
