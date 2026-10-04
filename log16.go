package floats

import "math"

// Log returns the natural logarithm of a.
//
// Special cases are:
//
//	+Inf.Log() = +Inf
//	0.Log() = -Inf
//	(x < 0).Log() = NaN
//	NaN.Log() = NaN
func (a Float16) Log() Float16 {
	if a-1 >= uvinf16-1 { // a <= 0, +Inf, or NaN
		return log16Special(a)
	}
	k, l := logKernel64(positive16ToFloat64(a))
	return NewFloat16(k*math.Ln2 + l)
}

// Log10 returns the decimal logarithm of a.
// The special cases are the same as for [Log].
func (a Float16) Log10() Float16 {
	if a-1 >= uvinf16-1 { // a <= 0, +Inf, or NaN
		return log16Special(a)
	}
	k, l := logKernel64(positive16ToFloat64(a))
	return NewFloat16(k*(math.Ln2/math.Ln10) + l*(1/math.Ln10))
}

// Log2 returns the binary logarithm of a.
// The special cases are the same as for [Log].
func (a Float16) Log2() Float16 {
	if a-1 >= uvinf16-1 { // a <= 0, +Inf, or NaN
		return log16Special(a)
	}
	// l is exactly zero for powers of two, so they give an exact answer.
	k, l := logKernel64(positive16ToFloat64(a))
	return NewFloat16(k + l*(1/math.Ln2))
}

// log16Special handles the special cases of Log, Log10, and Log2.
func log16Special(a Float16) Float16 {
	switch {
	case a.IsNaN() || a == uvinf16:
		return a
	case a&^signMask16 == 0:
		return uvneginf16
	default:
		// a < 0
		return NewFloat16NaN()
	}
}

// positive16ToFloat64 converts a positive finite Float16 value a to float64.
func positive16ToFloat64(a Float16) float64 {
	if a < 0x0400 {
		// a is subnormal.
		return float64(a) * 0x1p-24
	}
	return normal16ToFloat64(a)
}
