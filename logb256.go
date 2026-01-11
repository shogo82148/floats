package floats

import "math"

// Logb returns the binary exponent of a.
//
// Special cases are:
//
//	±Inf.Logb() = +Inf
//	0.Logb() = -Inf
//	NaN.Logb() = NaN
func (a Float256) Logb() Float256 {
	// special cases
	switch {
	case a.IsZero():
		return NewFloat256Inf(-1)
	case a.IsInf(0):
		return NewFloat256Inf(1)
	case a.IsNaN():
		return NewFloat256NaN()
	}
	_, exp, _ := a.normalize()
	return NewFloat256(float64(exp))
}

// Ilogb returns the binary exponent of a as an integer.
//
// Special cases are:
//
//	±Inf.Ilogb() = MaxInt32
//	0.Ilogb() = MinInt32
//	NaN.Ilogb() = MaxInt32
func (a Float256) Ilogb() int {
	// special cases
	switch {
	case a.IsZero():
		return math.MinInt32
	case a.IsInf(0):
		return math.MaxInt32
	case a.IsNaN():
		return math.MaxInt32
	}
	_, exp, _ := a.normalize()
	return exp
}
