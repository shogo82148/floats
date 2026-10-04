package floats

// Exp returns e**x, the base-e exponential of a.
//
// Special cases are:
//
//	+Inf.Exp() = +Inf
//	NaN.Exp() = NaN
//
// Very large values overflow to 0 or +Inf.
// Very small values underflow to 1.
func (a Float16) Exp() Float16 {
	ix := a &^ signMask16
	if ix < 0x0c00 { // |a| < 2**-12
		// exp(a) = 1 + a + ... rounds to 1.
		return uvone16
	}
	if a&signMask16 == 0 {
		if ix >= 0x4a00 { // a >= 12 or NaN
			if ix > uvinf16 {
				// exp(NaN) = NaN
				return NewFloat16NaN()
			}
			// exp(a) overflows.
			return uvinf16
		}
	} else {
		if ix >= 0x4c80 { // a <= -18 or NaN
			if ix > uvinf16 {
				// exp(NaN) = NaN
				return NewFloat16NaN()
			}
			// exp(a) underflows.
			return 0
		}
	}
	return NewFloat16(exp16(normal16ToFloat64(a)))
}

// Exp2 returns 2**x, the base-2 exponential of x.
//
// Special cases are the same as [Exp].
func (a Float16) Exp2() Float16 {
	ix := a &^ signMask16
	if ix < 0x0c00 { // |a| < 2**-12
		// 2**a = 1 + a*ln(2) + ... rounds to 1.
		return uvone16
	}
	if a&signMask16 == 0 {
		if ix >= 0x4c00 { // a >= 16 or NaN
			if ix > uvinf16 {
				// exp2(NaN) = NaN
				return NewFloat16NaN()
			}
			// exp2(a) overflows.
			return uvinf16
		}
	} else {
		if ix >= 0x4e40 { // a <= -25 or NaN
			if ix > uvinf16 {
				// exp2(NaN) = NaN
				return NewFloat16NaN()
			}
			// exp2(a) underflows.
			// 2**-25 is the midpoint of 0 and the smallest subnormal, and rounds to even.
			return 0
		}
	}
	scale, p := exp2Reduce(normal16ToFloat64(a))
	return NewFloat16(scale + scale*p)
}
