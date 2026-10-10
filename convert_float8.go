package floats

import "math"

const (
	uvnane4m3 = 0x7f // NaN value for Float8E4M3
	uvnane5m2 = 0x7e // NaN value for Float8E5M2
	uvinfe5m2 = 0x7c // Infinity value for Float8E5M2
)

// roundFloat32ToFloat8 rounds the magnitude of a finite Float32 to nearest even
// for the 8-bit format that has mbits bits of fraction and
// whose minimum exponent of the normal numbers is emin.
// abs is the bits of the Float32 without the sign bit.
//
// The result is the encoding of the magnitude with the sign bit cleared,
// and it is not checked whether the result is finite:
// the exponent field of the result may exceed the maximum of the format.
// The carries from the fraction to the exponent are the correct encodings
// of the next binade and of the smallest normal number.
func roundFloat32ToFloat8(abs uint32, mbits, emin int) uint32 {
	e := int(abs >> shift32)
	if e == 0 {
		// Zero or a subnormal number of Float32.
		// It is smaller than a half of the smallest subnormal number of Float8.
		return 0
	}
	full := uint64(abs&fracMask32 | 1<<shift32)
	exp := e - bias32

	// number of the bits to be rounded off
	shift := shift32 - mbits
	var base uint32
	if exp >= emin {
		base = uint32(exp-emin) << mbits
	} else {
		// subnormal in Float8
		shift += emin - exp
	}
	if shift > shift32+2 {
		// full < 1<<(shift32+1) is smaller than a half of the unit in the last place.
		return 0
	}

	q := full >> shift
	rem := full & (1<<shift - 1)
	half := uint64(1) << (shift - 1)
	if rem > half || rem == half && q&1 != 0 {
		q++
	}
	return base + uint32(q)
}

// Float8E4M3 converts a to a Float8E4M3.
func (a Float32) Float8E4M3() Float8E4M3 {
	b := math.Float32bits(float32(a))
	sign := uint8(b >> 24 & 0x80)
	abs := b &^ signMask32

	if abs >= uvinf32 {
		// a is NaN or infinity, and Float8E4M3 has no infinity.
		return Float8E4M3(sign | uvnane4m3)
	}
	enc := roundFloat32ToFloat8(abs, 3, -6)
	if enc > 0x7e {
		// overflow
		return Float8E4M3(sign | uvnane4m3)
	}
	return Float8E4M3(sign | uint8(enc))
}

// Float8E5M2 converts a to a Float8E5M2.
func (a Float32) Float8E5M2() Float8E5M2 {
	b := math.Float32bits(float32(a))
	sign := uint8(b >> 24 & 0x80)
	abs := b &^ signMask32

	if abs > uvinf32 {
		// a is NaN
		return Float8E5M2(sign | uvnane5m2)
	}
	if abs == uvinf32 {
		return Float8E5M2(sign | uvinfe5m2)
	}
	enc := roundFloat32ToFloat8(abs, 2, -14)
	if enc >= uvinfe5m2 {
		// overflow
		return Float8E5M2(sign | uvinfe5m2)
	}
	return Float8E5M2(sign | uint8(enc))
}

// Float8E4M3 converts a to a Float8E4M3.
func (a Float16) Float8E4M3() Float8E4M3 {
	return a.Float32().Float8E4M3()
}

// Float8E5M2 converts a to a Float8E5M2.
func (a Float16) Float8E5M2() Float8E5M2 {
	return a.Float32().Float8E5M2()
}

// Float8E4M3 converts a to a Float8E4M3.
func (a BFloat16) Float8E4M3() Float8E4M3 {
	return a.Float32().Float8E4M3()
}

// Float8E5M2 converts a to a Float8E5M2.
func (a BFloat16) Float8E5M2() Float8E5M2 {
	return a.Float32().Float8E5M2()
}

// Float8E4M3 converts a to a Float8E4M3.
func (a Float64) Float8E4M3() Float8E4M3 {
	return a.float32RoundToOdd().Float8E4M3()
}

// Float8E5M2 converts a to a Float8E5M2.
func (a Float64) Float8E5M2() Float8E5M2 {
	return a.float32RoundToOdd().Float8E5M2()
}

// float32RoundToOdd converts a to a Float32 with rounding to odd.
// See [Float64.float32RoundToOdd].
func (a Float128) float32RoundToOdd() Float32 {
	f := a.Float64()
	if f.IsNaN() || f.IsInf(0) || f.Float128() == a {
		return f.float32RoundToOdd()
	}
	// a is not representable in Float64. Round to odd.
	u := math.Float64bits(float64(f))
	if u&1 == 0 {
		if f.Float128().Abs().Gt(a.Abs()) {
			u--
		} else {
			u++
		}
	}
	return Float64(math.Float64frombits(u)).float32RoundToOdd()
}

// float32RoundToOdd converts a to a Float32 with rounding to odd.
// See [Float64.float32RoundToOdd].
func (a Float256) float32RoundToOdd() Float32 {
	f := a.Float64()
	if f.IsNaN() || f.IsInf(0) || f.Float256() == a {
		return f.float32RoundToOdd()
	}
	// a is not representable in Float64. Round to odd.
	u := math.Float64bits(float64(f))
	if u&1 == 0 {
		if f.Float256().Abs().Gt(a.Abs()) {
			u--
		} else {
			u++
		}
	}
	return Float64(math.Float64frombits(u)).float32RoundToOdd()
}

// Float8E4M3 converts a to a Float8E4M3.
func (a Float128) Float8E4M3() Float8E4M3 {
	return a.float32RoundToOdd().Float8E4M3()
}

// Float8E5M2 converts a to a Float8E5M2.
func (a Float128) Float8E5M2() Float8E5M2 {
	return a.float32RoundToOdd().Float8E5M2()
}

// Float8E4M3 converts a to a Float8E4M3.
func (a Float256) Float8E4M3() Float8E4M3 {
	return a.float32RoundToOdd().Float8E4M3()
}

// Float8E5M2 converts a to a Float8E5M2.
func (a Float256) Float8E5M2() Float8E5M2 {
	return a.float32RoundToOdd().Float8E5M2()
}

// Float8E4M3 returns a itself.
func (a Float8E4M3) Float8E4M3() Float8E4M3 {
	return a
}

// Float8E5M2 converts a to a Float8E5M2.
func (a Float8E4M3) Float8E5M2() Float8E5M2 {
	return a.Float32().Float8E5M2()
}

// Float8E5M2 returns a itself.
func (a Float8E5M2) Float8E5M2() Float8E5M2 {
	return a
}

// Float8E4M3 converts a to a Float8E4M3.
func (a Float8E5M2) Float8E4M3() Float8E4M3 {
	return a.Float32().Float8E4M3()
}

// Float16 converts a to a Float16.
// Float8E5M2 is the upper 8 bits of Float16, so the conversion is exact.
func (a Float8E5M2) Float16() Float16 {
	return Float16(a) << 8
}

// Float32 converts a to a Float32.
func (a Float8E5M2) Float32() Float32 {
	return a.Float16().Float32()
}

// Float32 converts a to a Float32.
// The conversion is exact.
func (a Float8E4M3) Float32() Float32 {
	sign := uint32(a&0x80) << 24
	exp := uint32(a>>3) & 0xf
	frac := uint32(a & 0x7)

	switch {
	case exp == 0xf && frac == 0x7:
		return Float32(math.Float32frombits(sign | uvnan32))
	case exp == 0:
		// zero or a subnormal number: frac * 2^-9
		return Float32(math.Float32frombits(math.Float32bits(float32(frac)*0x1p-9) | sign))
	}
	return Float32(math.Float32frombits(sign | (exp+bias32-7)<<shift32 | frac<<(shift32-3)))
}

// Float16 converts a to a Float16.
// The conversion is exact.
func (a Float8E4M3) Float16() Float16 {
	return a.Float32().Float16()
}

// BFloat16 converts a to a BFloat16.
// The conversion is exact.
func (a Float8E4M3) BFloat16() BFloat16 {
	return a.Float32().BFloat16()
}

// BFloat16 converts a to a BFloat16.
// The conversion is exact.
func (a Float8E5M2) BFloat16() BFloat16 {
	return a.Float32().BFloat16()
}

// Float64 converts a to a Float64.
// The conversion is exact.
func (a Float8E4M3) Float64() Float64 {
	return a.Float32().Float64()
}

// Float64 converts a to a Float64.
// The conversion is exact.
func (a Float8E5M2) Float64() Float64 {
	return a.Float32().Float64()
}

// Float128 converts a to a Float128.
// The conversion is exact.
func (a Float8E4M3) Float128() Float128 {
	return a.Float32().Float128()
}

// Float128 converts a to a Float128.
// The conversion is exact.
func (a Float8E5M2) Float128() Float128 {
	return a.Float32().Float128()
}

// Float256 converts a to a Float256.
// The conversion is exact.
func (a Float8E4M3) Float256() Float256 {
	return a.Float32().Float256()
}

// Float256 converts a to a Float256.
// The conversion is exact.
func (a Float8E5M2) Float256() Float256 {
	return a.Float32().Float256()
}
