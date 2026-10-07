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
	abs := uint32(a &^ signMask16)
	if abs >= mask16<<shift16 {
		// a is infinity or NaN
		return Float32(math.Float32frombits(sign | mask32<<shift32 | (abs&fracMask16)<<(shift32-shift16)))
	}

	// Place the exponent and the fraction into the Float32 fields and
	// rebias the exponent by a multiplication.
	// It also normalizes subnormal numbers, and the result is exact.
	v := math.Float32frombits(abs<<(shift32-shift16)) * 0x1p112 // 2^(bias32 - bias16)
	return Float32(math.Float32frombits(math.Float32bits(v) | sign))
}

// Float64 converts a to a Float64.
func (a Float16) Float64() Float64 {
	sign := uint64(a&signMask16) << (64 - 16)
	abs := uint64(a &^ signMask16)
	if abs >= mask16<<shift16 {
		// a is infinity or NaN
		return Float64(math.Float64frombits(sign | mask64<<shift64 | (abs&fracMask16)<<(shift64-shift16)))
	}

	// Place the exponent and the fraction into the Float64 fields and
	// rebias the exponent by a multiplication.
	// It also normalizes subnormal numbers, and the result is exact.
	v := math.Float64frombits(abs<<(shift64-shift16)) * 0x1p1008 // 2^(bias64 - bias16)
	return Float64(math.Float64frombits(math.Float64bits(v) | sign))
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
	abs := uint64(a &^ signMask16)
	if abs >= mask16<<shift16 {
		// a is infinity or NaN
		return Float256{sign | mask256<<(shift256-192) | (abs&fracMask16)<<(shift256-192-shift16), 0, 0, 0}
	}

	// Normalize the value in Float64, which is exact, and then
	// move its fields into the Float256 layout.
	v := math.Float64bits(math.Float64frombits(abs<<(shift64-shift16)) * 0x1p1008) // 2^(bias64 - bias16)
	hi := v >> (shift64 - (shift256 - 192))
	if abs != 0 {
		hi += (bias256 - bias64) << (shift256 - 192)
	}
	return Float256{sign | hi, 0, 0, 0}
}

// Float16 converts a to a Float16.
func (a Float32) Float16() Float16 {
	b := math.Float32bits(float32(a))
	sign := uint16((b & signMask32) >> (32 - 16))
	abs := b &^ signMask32

	if abs > uvinf32 {
		// a is NaN
		return Float16(sign | uvnan16)
	}

	if abs < (bias32-bias16+1)<<shift32 {
		// the result is subnormal number.
		// Adding 2^-1 aligns the ULP of Float16 subnormal numbers (2^-24)
		// to the least significant bit of the Float32 fraction,
		// so the FPU rounds to nearest even for us.
		// The carry into the exponent field is the correct encoding of the smallest normal number.
		v := math.Float32bits(math.Float32frombits(abs) + 0x1p-1)
		return Float16(sign | uint16(v-math.Float32bits(0x1p-1)))
	}

	// the result is normal number
	const halfMinusULP = 1<<(shift32-shift16-1) - 1
	abs += halfMinusULP + ((abs >> (shift32 - shift16)) & 1) // round to nearest even
	exp16 := abs>>(shift32-shift16) - (bias32-bias16)<<shift16
	if exp16 >= mask16<<shift16 {
		// overflow or ±infinity
		return Float16(sign | mask16<<shift16)
	}
	return Float16(sign | uint16(exp16))
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
	abs := b &^ signMask64

	if abs > uvinf64 {
		// a is NaN
		return Float16(sign | uvnan16)
	}

	if abs < (bias64-bias16+1)<<shift64 {
		// the result is subnormal number.
		// Adding 2^28 aligns the ULP of Float16 subnormal numbers (2^-24)
		// to the least significant bit of the Float64 fraction,
		// so the FPU rounds to nearest even for us.
		// The carry into the exponent field is the correct encoding of the smallest normal number.
		v := math.Float64bits(math.Float64frombits(abs) + 0x1p28)
		return Float16(sign | uint16(v-math.Float64bits(0x1p28)))
	}

	// the result is normal number
	const halfMinusULP = 1<<(shift64-shift16-1) - 1
	abs += halfMinusULP + ((abs >> (shift64 - shift16)) & 1) // round to nearest even
	exp16 := abs>>(shift64-shift16) - (bias64-bias16)<<shift16
	if exp16 >= mask16<<shift16 {
		// overflow or ±infinity
		return Float16(sign | mask16<<shift16)
	}
	return Float16(sign | uint16(exp16))
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
		frac[0] |= nonzero64(frac[1])
		frac[0] += halfMinusULP + ((frac[0] >> uint(roundBit)) & 1)
		return Float16(sign | uint16(frac[0]>>roundBit))
	}

	// the result is normal number
	// round to nearest even
	const halfMinusULP = 1<<(shift128-shift16-64-1) - 1
	a[0] |= nonzero64(a[1])
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
		frac[0] |= nonzero64(frac[1])
		frac[0] += halfMinusULP + ((frac[0] >> uint(roundBit)) & 1)
		return Float32(math.Float32frombits(sign | uint32(frac[0]>>uint(roundBit))))
	}

	// the result is normal number
	// round to nearest even
	const halfMinusULP = 1<<(shift128-shift32-64-1) - 1
	a[0] |= nonzero64(a[1])
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
	hi, lo := a[0], a[1]
	sign := hi & signMask128[0]
	exp := int(hi>>(shift128-64)) & mask128

	if exp == mask128 {
		// a is ±infinity or NaN
		if hi&(fracMask128[0]) != 0 || lo != 0 {
			return Float64(math.Float64frombits(sign | uvnan64))
		}
		return Float64(math.Float64frombits(sign | mask64<<shift64))
	}

	e := exp - (bias128 - bias64) // the biased exponent of the result
	if e >= mask64 {
		// overflow
		return Float64(math.Float64frombits(sign | mask64<<shift64))
	}

	// sig is the 64 bits significand with the implicit bit:
	// the implicit bit, the top 52 bits of the fraction, and the next 11 bits.
	// The remaining 49 bits are folded into the least significant bit as the sticky bit.
	frac := (hi&fracMask128[0])<<4 | lo>>60
	sig := 1<<63 | frac<<11 | (lo>>49)&(1<<11-1)
	if lo&(1<<49-1) != 0 {
		sig |= 1
	}

	// shift is the number of bits to be discarded.
	shift := uint(11)
	if e <= 0 {
		// the result is subnormal number
		if e < -52 {
			// the result is less than the half of the smallest subnormal number
			return Float64(math.Float64frombits(sign))
		}
		shift += uint(1 - e)
		e = 1
	}

	// round to nearest even
	q := sig >> shift
	rem := sig << (64 - shift)
	if rem > 1<<63 || (rem == 1<<63 && q&1 != 0) {
		q++
	}
	return Float64(math.Float64frombits(sign | (uint64(e-1)<<shift64 + q)))
}

// Float128 returns a itself.
func (a Float128) Float128() Float128 {
	return a
}

// Float256 converts a to a Float256.
func (a Float128) Float256() Float256 {
	b := ints.Uint128(a)
	sign := b[0] & signMask128[0]
	exp := int((b[0] >> (shift128 - 64)) & mask128)
	frac := b.And(fracMask128)

	if exp == mask128 {
		// a is ±infinity or NaN
		frac256 := frac.Uint256().Lsh(shift256 - shift128)
		return Float256{
			sign | mask256<<(shift256-192) | frac256[0],
			frac256[1],
			frac256[2],
			frac256[3],
		}
	} else if exp == 0 {
		// a is subnormal
		if frac.IsZero() {
			// a is zero
			return Float256{sign, 0, 0, 0}
		}

		// normalize a
		l := frac.BitLen()
		exp = l - shift128
		frac = frac.Lsh(shift128 - uint(l) + 1).And(fracMask128)
	}

	exp += bias256 - bias128
	frac256 := frac.Uint256().Lsh(shift256 - shift128)
	return Float256{
		sign | uint64(exp)<<(shift256-192) | frac256[0],
		frac256[1],
		frac256[2],
		frac256[3],
	}
}

// Float16 converts a to a Float16.
func (a Float256) Float16() Float16 {
	sign := uint16((a[0] & signMask256[0]) >> (64 - 16))
	exp := int((a[0] >> (shift256 - 192)) & mask256)

	if exp == mask256 {
		// a is ±infinity or NaN
		frac := ints.Uint256(a).And(fracMask256)
		if frac.IsZero() {
			// a is ±infinity
			return Float16(sign | mask16<<shift16)
		} else {
			// a is NaN
			return Float16(sign | uvnan16)
		}
	}

	exp -= bias256
	if exp <= -bias16 {
		// the result is subnormal number
		frac := ints.Uint256(a).And(fracMask256)
		frac[0] |= (1 << (shift256 - 192))
		// round to nearest even
		roundBit := -exp + shift256 - (bias16 + shift16 - 1) - 192
		halfMinusULP := uint64(1<<(roundBit-1) - 1)
		frac[0] |= nonzero64(frac[1]) | nonzero64(frac[2]) | nonzero64(frac[3])
		frac[0] += halfMinusULP + ((a[0] >> uint(roundBit)) & 1)
		return Float16(sign | uint16(frac[0]>>roundBit))
	}

	// the result is normal number
	// round to nearest even
	const halfMinusULP = 1<<(shift256-shift16-192-1) - 1
	a[0] |= nonzero64(a[1]) | nonzero64(a[2]) | nonzero64(a[3])
	a[0] += halfMinusULP + ((a[0] >> uint(shift256-shift16-192)) & 1)

	exp16 := uint16((a[0]>>(shift256-192))&mask256 - bias256 + bias16)
	if exp16 >= mask16 {
		// overflow
		return Float16(sign | mask16<<shift16)
	}
	frac16 := uint16(a[0]>>(shift256-shift16-192)) & fracMask16
	return Float16(sign | (exp16 << shift16) | frac16)
}

// Float32 converts a to a Float32.
func (a Float256) Float32() Float32 {
	sign := uint32((a[0] & signMask256[0]) >> (64 - 32))
	exp := int((a[0] >> (shift256 - 192)) & mask256)

	if exp == mask256 {
		// a is ±infinity or NaN
		frac := ints.Uint256(a).And(fracMask256)
		if frac.IsZero() {
			// a is ±infinity
			return Float32(math.Float32frombits(sign | mask32<<shift32))
		} else {
			// a is NaN
			return Float32(math.Float32frombits(sign | uvnan32))
		}
	}

	exp -= bias256
	if exp <= -bias32 {
		// the result is subnormal number
		frac := ints.Uint256(a).And(fracMask256)
		frac[0] |= (1 << (shift256 - 192))
		// round to nearest even
		roundBit := -exp + shift256 - (bias32 + shift32 - 1) - 192
		halfMinusULP := uint64(1<<(roundBit-1) - 1)
		frac[0] |= nonzero64(frac[1]) | nonzero64(frac[2]) | nonzero64(frac[3])
		frac[0] += halfMinusULP + ((a[0] >> uint(roundBit)) & 1)
		return Float32(math.Float32frombits(sign | uint32(frac[0]>>uint(roundBit))))
	}

	// the result is normal number
	// round to nearest even
	const halfMinusULP = 1<<(shift256-shift32-192-1) - 1
	a[0] |= nonzero64(a[1]) | nonzero64(a[2]) | nonzero64(a[3])
	a[0] += halfMinusULP + ((a[0] >> uint(shift256-shift32-192)) & 1)

	exp32 := uint32((a[0]>>(shift256-192))&mask256 - bias256 + bias32)
	if exp32 >= mask32 {
		// overflow
		return Float32(math.Float32frombits(sign | mask32<<shift32))
	}
	frac32 := uint32(a[0]>>(shift256-shift32-192)) & fracMask32
	return Float32(math.Float32frombits(sign | (exp32 << shift32) | frac32))
}

// Float64 converts a to a Float64.
func (a Float256) Float64() Float64 {
	b := ints.Uint256(a)
	sign := b[0] & signMask256[0]
	exp := int((b[0] >> (shift256 - 192)) & mask256)

	if exp == mask256 {
		// a is ±infinity or NaN
		frac := b.And(fracMask256)
		if frac.IsZero() {
			// a is ±infinity
			return Float64(math.Float64frombits(sign | mask64<<shift64))
		} else {
			// a is NaN
			return Float64(math.Float64frombits(sign | uvnan64))
		}
	}

	exp -= bias256
	if exp <= -bias64 {
		// the result is subnormal number
		frac := b.And(fracMask256)
		frac[0] |= (1 << (shift256 - 192))
		// round to nearest even
		roundBit := uint(-exp + shift256 - (bias64 + shift64 - 1))
		halfMinusULP := ints.Uint256{0, 0, 0, 1}.Lsh(roundBit - 1).Sub(ints.Uint256{0, 0, 0, 1})
		frac = frac.Add(halfMinusULP).Add(frac.Rsh(roundBit).And(ints.Uint256{0, 0, 0, 1}))
		frac = frac.Rsh(roundBit)
		return Float64(math.Float64frombits(sign | frac[1]))
	}

	// the result is normal number
	// round to nearest even
	halfMinusULP := ints.Uint256{0, 0, 0, 1}.Lsh(shift256 - shift64 - 1).Sub(ints.Uint256{0, 0, 0, 1})
	b = b.Add(halfMinusULP).Add(b.Rsh(shift256 - shift64).And(ints.Uint256{0, 0, 0, 1}))

	exp64 := uint64((b[0]>>(shift256-192))&mask256 - bias256 + bias64)
	if exp64 >= mask64 {
		// overflow
		return Float64(math.Float64frombits(sign | mask64<<shift64))
	}
	frac64 := uint64(b.Rsh(shift256-shift64).Uint64() & fracMask64)
	return Float64(math.Float64frombits(sign | (exp64 << shift64) | frac64))
}

// Float128 converts a to a Float128.
func (a Float256) Float128() Float128 {
	sign := a[0] & signMask256[0]
	exp := int(a[0]>>(shift256-192)) & mask256

	if exp == mask256 {
		// a is ±infinity or NaN
		if a[0]&fracMask256[0] != 0 || a[1] != 0 || a[2] != 0 || a[3] != 0 {
			// a is NaN
			return Float128{sign | mask128<<(shift128-64) | 1<<(shift128-64-1), 0}
		}
		// a is ±infinity
		return Float128{sign | mask128<<(shift128-64), 0}
	}

	e := exp - (bias256 - bias128) // the biased exponent of the result
	if e >= mask128 {
		// overflow
		return Float128{sign | mask128<<(shift128-64), 0}
	}
	if e <= 0 {
		return a.float128Subnormal()
	}

	// the result is normal number
	// the top 112 bits of the fraction are the fraction of the result.
	const fracShift = (shift128 - 64) - (shift256 - 192) // 4
	hi := uint64(e)<<(shift128-64) | (a[0]&fracMask256[0])<<fracShift | a[1]>>(64-fracShift)
	lo := a[1]<<fracShift | a[2]>>(64-fracShift)

	// round to nearest even
	rem := a[2]<<fracShift | nonzero64(a[3])
	if rem > 1<<63 || (rem == 1<<63 && lo&1 != 0) {
		var carry uint64
		lo, carry = bits.Add64(lo, 1, 0)
		hi += carry // a carry into the exponent field is the correct encoding
	}
	return Float128{sign | hi, lo}
}

// float128Subnormal converts a to a Float128 for the case that the result is a subnormal number.
func (a Float256) float128Subnormal() Float128 {
	b := ints.Uint256(a)
	sign := b[0] & signMask256[0]
	exp := int((b[0]>>(shift256-192))&mask256) - bias256

	frac := b.And(fracMask256)
	frac[0] |= (1 << (shift256 - 192))
	// round to nearest even
	roundBit := uint(-exp + shift256 - (bias128 + shift128 - 1))
	halfMinusULP := ints.Uint256{0, 0, 0, 1}.Lsh(roundBit - 1).Sub(ints.Uint256{0, 0, 0, 1})
	frac = frac.Add(halfMinusULP).Add(frac.Rsh(roundBit).And(ints.Uint256{0, 0, 0, 1}))
	frac = frac.Rsh(roundBit)
	return Float128{
		sign | frac[2],
		frac[3],
	}
}

// Float256 returns a itself.
func (a Float256) Float256() Float256 {
	return a
}
