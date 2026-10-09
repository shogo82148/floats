package floats

// Floor returns the greatest integer value less than or equal to a.
//
// Special cases are:
//
//	±0.Floor() = ±0
//	±Inf.Floor() = ±Inf
//	NaN.Floor() = NaN
func (a BFloat16) Floor() BFloat16 {
	exp := int(a>>shiftbf16) & maskbf16
	if exp < biasbf16 {
		// |a| < 1
		if a&^signMaskbf16 == 0 {
			return a // ±0
		}
		if a&signMaskbf16 != 0 {
			return 0xbf80 // -1
		}
		return 0
	}
	if exp >= biasbf16+shiftbf16 {
		// already an integer, or ±Inf, or NaN
		return a
	}
	mask := BFloat16(fracMaskbf16 >> (exp - biasbf16))
	if a&mask == 0 {
		return a
	}
	if a&signMaskbf16 != 0 {
		// round away from zero; a carry into the exponent is correct.
		return (a | mask) + 1
	}
	return a &^ mask
}

// Ceil returns the least integer value greater than or equal to a.
//
// Special cases are:
//
//	±0.Ceil() = ±0
//	±Inf.Ceil() = ±Inf
//	NaN.Ceil() = NaN
func (a BFloat16) Ceil() BFloat16 {
	exp := int(a>>shiftbf16) & maskbf16
	if exp < biasbf16 {
		// |a| < 1
		if a&^signMaskbf16 == 0 {
			return a // ±0
		}
		if a&signMaskbf16 != 0 {
			return signMaskbf16 // -0
		}
		return 0x3f80 // 1
	}
	if exp >= biasbf16+shiftbf16 {
		// already an integer, or ±Inf, or NaN
		return a
	}
	mask := BFloat16(fracMaskbf16 >> (exp - biasbf16))
	if a&mask == 0 {
		return a
	}
	if a&signMaskbf16 != 0 {
		return a &^ mask
	}
	// round away from zero; a carry into the exponent is correct.
	return (a | mask) + 1
}

// Trunc returns the integer value of a.
//
// Special cases are:
//
//	±0.Trunc() = ±0
//	±Inf.Trunc() = ±Inf
//	NaN.Trunc() = NaN
func (a BFloat16) Trunc() BFloat16 {
	exp := int(a>>shiftbf16) & maskbf16
	if exp < biasbf16 {
		// |a| < 1: ±0 with the sign of a
		return a & signMaskbf16
	}
	if exp >= biasbf16+shiftbf16 {
		// already an integer, or ±Inf, or NaN
		return a
	}
	return a &^ BFloat16(fracMaskbf16>>(exp-biasbf16))
}

// Round returns the nearest integer, rounding half away from zero.
//
// Special cases are:
//
//	±0.Round() = ±0
//	±Inf.Round() = ±Inf
//	NaN.Round() = NaN
func (a BFloat16) Round() BFloat16 {
	exp := int(a>>shiftbf16) & maskbf16
	if exp < biasbf16-1 {
		// |a| < 0.5
		return a & signMaskbf16
	}
	if exp == biasbf16-1 {
		// 0.5 <= |a| < 1
		return a&signMaskbf16 | 0x3f80
	}
	if exp >= biasbf16+shiftbf16 {
		// already an integer, or ±Inf, or NaN
		return a
	}
	e := exp - biasbf16
	half := BFloat16(1 << (shiftbf16 - 1) >> e)
	mask := BFloat16(fracMaskbf16 >> e)
	// a carry into the exponent is correct.
	return (a + half) &^ mask
}

// RoundToEven returns the nearest integer, rounding ties to even.
//
// Special cases are:
//
//	±0.RoundToEven() = ±0
//	±Inf.RoundToEven() = ±Inf
//	NaN.RoundToEven() = NaN
func (a BFloat16) RoundToEven() BFloat16 {
	exp := int(a>>shiftbf16) & maskbf16
	if exp < biasbf16-1 {
		// |a| < 0.5
		return a & signMaskbf16
	}
	if exp == biasbf16-1 {
		// 0.5 <= |a| < 1
		if a&fracMaskbf16 == 0 {
			return a & signMaskbf16 // ±0.5 rounds to ±0
		}
		return a&signMaskbf16 | 0x3f80
	}
	if exp >= biasbf16+shiftbf16 {
		// already an integer, or ±Inf, or NaN
		return a
	}
	e := exp - biasbf16
	half := BFloat16(1 << (shiftbf16 - 1) >> e)
	mask := BFloat16(fracMaskbf16 >> e)
	// add half-1, plus 1 if the integer part is odd (ties go to even).
	// a carry into the exponent is correct.
	odd := (a >> (shiftbf16 - e)) & 1
	return (a + half - 1 + odd) &^ mask
}
