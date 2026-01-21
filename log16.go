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
	return NewFloat16(math.Log(a.Float64().BuiltIn()))
}

// Log10 returns the decimal logarithm of a.
// The special cases are the same as for [Log].
func (a Float16) Log10() Float16 {
	return NewFloat16(math.Log10(a.Float64().BuiltIn()))
}

// Log2 returns the binary logarithm of a.
// The special cases are the same as for [Log].
func (a Float16) Log2() Float16 {
	return NewFloat16(math.Log2(a.Float64().BuiltIn()))
}
