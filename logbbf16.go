package floats

import "math"

// Logb returns the binary exponent of a.
//
// Special cases are:
//
//	±Inf.Logb() = +Inf
//	0.Logb() = -Inf
//	NaN.Logb() = NaN
func (a BFloat16) Logb() BFloat16 {
	// special cases
	switch {
	case a.IsZero():
		return NewBFloat16Inf(-1)
	case a.IsInf(0):
		return NewBFloat16Inf(1)
	case a.IsNaN():
		return NewBFloat16NaN()
	}
	_, exp, _ := a.normalize()
	return NewBFloat16(float64(exp))
}

// Ilogb returns the binary exponent of a as an integer.
//
// Special cases are:
//
//	±Inf.Ilogb() = MaxInt32
//	0.Ilogb() = MinInt32
//	NaN.Ilogb() = MaxInt32
func (a BFloat16) Ilogb() int {
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
