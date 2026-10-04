package floats

import "math"

// Log1p returns the natural logarithm of 1 plus its argument a.
// It is more accurate than [Log](1 + a) when a is near zero.
//
// Special cases are:
//
//	+Inf.Log1p() = +Inf
//	±0.Log1p() = ±0
//	-1.Log1p() = -Inf
//	(a < -1).Log1p() = NaN
//	NaN.Log1p() = NaN
func (a Float16) Log1p() Float16 {
	ix := a &^ signMask16
	if ix < 0x1000 { // |a| < 2**-11
		// log1p(a) = a - a**2/2 + ... rounds to a.
		return a
	}
	if a >= uvinf16 && ix >= uvone16 { // a == +Inf, a <= -1, or NaN
		switch {
		case a == uvinf16:
			// log1p(+Inf) = +Inf
			return a
		case a == signMask16|uvone16:
			// log1p(-1) = -Inf
			return uvneginf16
		default:
			// a < -1 or NaN
			return NewFloat16NaN()
		}
	}

	// 1+a is exact, because a is a normal Float16 value.
	k, l := logKernel64(1 + normal16ToFloat64(a))
	return NewFloat16(k*math.Ln2 + l)
}
