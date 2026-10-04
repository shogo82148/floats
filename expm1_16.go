package floats

// Expm1 returns e**a - 1, the base-e exponential of a minus 1.
// It is more accurate than Exp(a) - 1 when a is near zero.
//
// Special cases are:
//
//	+Inf.Expm1() = +Inf
//	-Inf.Expm1() = -1
//	NaN.Expm1() = NaN
//
// Very large values overflow to -1 or +Inf.
func (a Float16) Expm1() Float16 {
	ix := a &^ signMask16
	if ix < 0x1000 { // |a| < 2**-11
		// expm1(a) = a + a**2/2 + ... rounds to a.
		return a
	}
	if a&signMask16 == 0 {
		if ix >= 0x4a00 { // a >= 12 or NaN
			if ix > uvinf16 {
				// expm1(NaN) = NaN
				return NewFloat16NaN()
			}
			// expm1(a) overflows.
			return uvinf16
		}
	} else {
		if ix >= 0x4c00 { // a <= -16 or NaN
			if ix > uvinf16 {
				// expm1(NaN) = NaN
				return NewFloat16NaN()
			}
			// expm1(a) rounds to -1.
			return signMask16 | uvone16
		}
	}
	return NewFloat16(expm1Small(normal16ToFloat64(a)))
}
