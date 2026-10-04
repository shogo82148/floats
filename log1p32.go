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
func (a Float32) Log1p() Float32 {
	ix := a.Bits() &^ signMask32
	if ix < 0x33800000 { // |a| < 2**-24
		// log1p(a) = a - a**2/2 + ... rounds to a.
		return a
	}
	if !(a > -1 && ix < uvinf32) {
		switch {
		case a == -1:
			return NewFloat32Inf(-1)
		case a > 0:
			// log1p(+Inf) = +Inf
			return a
		default:
			// a < -1 or NaN
			return NewFloat32NaN()
		}
	}

	// 1+a is exact for |a| < 2**29, and its relative error is less than 2**-53 otherwise.
	k, l := logKernel64(1 + float64(a))
	return Float32(k*math.Ln2 + l)
}
