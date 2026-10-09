package floats

import "math"

// BFloat16 returns a itself.
func (a BFloat16) BFloat16() BFloat16 {
	return a
}

// Float16 converts a to a Float16.
func (a BFloat16) Float16() Float16 {
	return a.Float32().Float16()
}

// Float32 converts a to a Float32.
// BFloat16 is the upper 16 bits of Float32, so the conversion is exact.
func (a BFloat16) Float32() Float32 {
	return Float32(math.Float32frombits(uint32(a) << 16))
}

// Float64 converts a to a Float64.
func (a BFloat16) Float64() Float64 {
	return a.Float32().Float64()
}

// Float128 converts a to a Float128.
func (a BFloat16) Float128() Float128 {
	return a.Float32().Float128()
}

// Float256 converts a to a Float256.
func (a BFloat16) Float256() Float256 {
	return a.Float32().Float256()
}

// BFloat16 converts a to a BFloat16.
func (a Float16) BFloat16() BFloat16 {
	return a.Float32().BFloat16()
}

// BFloat16 converts a to a BFloat16.
func (a Float32) BFloat16() BFloat16 {
	b := math.Float32bits(float32(a))
	sign := uint16(b >> 16 & signMaskbf16)
	abs := b &^ signMask32

	if abs > uvinf32 {
		// a is NaN
		return BFloat16(sign | uvnanbf16)
	}

	// Round to nearest even.
	// The encodings of Float32 and BFloat16 have the same exponent field,
	// so the carry into the exponent field is the correct encoding of the next binade,
	// of the smallest normal number, and of infinity.
	const halfMinusULP = 1<<15 - 1
	abs += halfMinusULP + (abs>>16)&1
	return BFloat16(sign | uint16(abs>>16))
}

// BFloat16 converts a to a BFloat16.
func (a Float64) BFloat16() BFloat16 {
	return a.float32RoundToOdd().BFloat16()
}

// float32RoundToOdd converts a to a Float32 with rounding to odd:
// if a is not representable in Float32, the result is one of the two adjacent Float32 values
// whose least significant bit of the fraction is 1.
// Rounding the result to a format that has at most 22 bits of the fraction
// is the same as rounding a directly.
func (a Float64) float32RoundToOdd() Float32 {
	f := float32(a)
	if f != f || math.IsInf(float64(f), 0) || float64(f) == float64(a) {
		// NaN, overflow, or exact
		return Float32(f)
	}
	u := math.Float32bits(f)
	if u&1 == 0 {
		// move f toward a
		if math.Abs(float64(f)) > math.Abs(float64(a)) {
			u--
		} else {
			u++
		}
	}
	return Float32(math.Float32frombits(u))
}

// BFloat16 converts a to a BFloat16.
func (a Float128) BFloat16() BFloat16 {
	f := a.Float64()
	if f.IsNaN() || f.IsInf(0) || f.Float128() == a {
		return f.BFloat16()
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
	return Float64(math.Float64frombits(u)).BFloat16()
}

// BFloat16 converts a to a BFloat16.
func (a Float256) BFloat16() BFloat16 {
	f := a.Float64()
	if f.IsNaN() || f.IsInf(0) || f.Float256() == a {
		return f.BFloat16()
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
	return Float64(math.Float64frombits(u)).BFloat16()
}
