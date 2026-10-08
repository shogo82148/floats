package floats

// Floor returns the greatest integer value less than or equal to a.
//
// Special cases are:
//
//	±0.Floor() = ±0
//	±Inf.Floor() = ±Inf
//	NaN.Floor() = NaN
func (a Float16) Floor() Float16 {
	exp := int(a>>shift16) & mask16
	if exp < bias16 {
		// |a| < 1
		if a&^signMask16 == 0 {
			return a // ±0
		}
		if a&signMask16 != 0 {
			return 0xbc00 // -1
		}
		return 0
	}
	if exp >= bias16+shift16 {
		// already an integer, or ±Inf, or NaN
		return a
	}
	mask := Float16(fracMask16 >> (exp - bias16))
	if a&mask == 0 {
		return a
	}
	if a&signMask16 != 0 {
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
func (a Float16) Ceil() Float16 {
	exp := int(a>>shift16) & mask16
	if exp < bias16 {
		// |a| < 1
		if a&^signMask16 == 0 {
			return a // ±0
		}
		if a&signMask16 != 0 {
			return signMask16 // -0
		}
		return 0x3c00 // 1
	}
	if exp >= bias16+shift16 {
		// already an integer, or ±Inf, or NaN
		return a
	}
	mask := Float16(fracMask16 >> (exp - bias16))
	if a&mask == 0 {
		return a
	}
	if a&signMask16 != 0 {
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
func (a Float16) Trunc() Float16 {
	exp := int(a>>shift16) & mask16
	if exp < bias16 {
		// |a| < 1: ±0 with the sign of a
		return a & signMask16
	}
	if exp >= bias16+shift16 {
		// already an integer, or ±Inf, or NaN
		return a
	}
	return a &^ Float16(fracMask16>>(exp-bias16))
}

// Round returns the nearest integer, rounding half away from zero.
//
// Special cases are:
//
//	±0.Round() = ±0
//	±Inf.Round() = ±Inf
//	NaN.Round() = NaN
func (a Float16) Round() Float16 {
	exp := int(a>>shift16) & mask16
	if exp < bias16-1 {
		// |a| < 0.5
		return a & signMask16
	}
	if exp == bias16-1 {
		// 0.5 <= |a| < 1
		return a&signMask16 | 0x3c00
	}
	if exp >= bias16+shift16 {
		// already an integer, or ±Inf, or NaN
		return a
	}
	e := exp - bias16
	half := Float16(1 << (shift16 - 1) >> e)
	mask := Float16(fracMask16 >> e)
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
func (a Float16) RoundToEven() Float16 {
	exp := int(a>>shift16) & mask16
	if exp < bias16-1 {
		// |a| < 0.5
		return a & signMask16
	}
	if exp == bias16-1 {
		// 0.5 <= |a| < 1
		if a&fracMask16 == 0 {
			return a & signMask16 // ±0.5 rounds to ±0
		}
		return a&signMask16 | 0x3c00
	}
	if exp >= bias16+shift16 {
		// already an integer, or ±Inf, or NaN
		return a
	}
	e := exp - bias16
	half := Float16(1 << (shift16 - 1) >> e)
	mask := Float16(fracMask16 >> e)
	// add half-1, plus 1 if the integer part is odd (ties go to even).
	// a carry into the exponent is correct.
	odd := (a >> (shift16 - e)) & 1
	return (a + half - 1 + odd) &^ mask
}
