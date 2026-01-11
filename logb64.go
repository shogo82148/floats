package floats

import "math"

// Logb returns the binary exponent of a.
//
// Special cases are:
//
//	±Inf.Logb() = +Inf
//	0.Logb() = -Inf
//	NaN.Logb() = NaN
func (a Float64) Logb() Float64 {
	return NewFloat64(math.Logb(a.Float64().BuiltIn()))
}

// Ilogb returns the binary exponent of a as an integer.
//
// Special cases are:
//
//	±Inf.Ilogb() = MaxInt32
//	0.Ilogb() = MinInt32
//	NaN.Ilogb() = MaxInt32
func (a Float64) Ilogb() int {
	return math.Ilogb(a.Float64().BuiltIn())
}
