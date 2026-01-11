package floats

import "math"

// Logb returns the binary exponent of a.
//
// Special cases are:
//
//	±Inf.Logb() = +Inf
//	0.Logb() = -Inf
//	NaN.Logb() = NaN
func (a Float32) Logb() Float32 {
	// special cases
	switch {
	case a.IsZero():
		return NewFloat32Inf(-1)
	case a.IsInf(0):
		return NewFloat32Inf(1)
	case a.IsNaN():
		return NewFloat32NaN()
	}
	_, exp, _ := a.normalize()
	return NewFloat32(float64(exp))
}

// Ilogb returns the binary exponent of x as an integer.
//
// Special cases are:
//
//	±Inf.Ilogb() = MaxInt32
//	0.Ilogb() = MinInt32
//	NaN.Ilogb() = MaxInt32
func (a Float32) Ilogb() int {
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
