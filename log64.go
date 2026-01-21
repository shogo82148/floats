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
func (a Float64) Log() Float64 {
	return NewFloat64(math.Log(a.BuiltIn()))
}

// Log10 returns the decimal logarithm of x.
// The special cases are the same as for [Log].
func (a Float64) Log10() Float64 {
	return NewFloat64(math.Log10(a.BuiltIn()))
}

// Log2 returns the binary logarithm of x.
// The special cases are the same as for [Log].
func (a Float64) Log2() Float64 {
	return NewFloat64(math.Log2(a.BuiltIn()))
}
