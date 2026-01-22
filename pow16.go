package floats

import "math"

// Pow returns a**b, the base-a exponential of b.
//
// Special cases are (in order):
//
//	a.Pow(±0) = 1 for any a
//	1.Pow(b) = 1 for any b
//	a.Pow(1) = a for any a
//	NaN.Pow(b) = NaN
//	a.Pow(NaN) = NaN
//	±0.Pow(b) = ±Inf for b an odd integer < 0
//	±0.Pow(-Inf) = +Inf
//	±0.Pow(+Inf) = +0
//	±0.Pow(b) = +Inf for finite b < 0 and not an odd integer
//	±0.Pow(b) = ±0 for b an odd integer > 0
//	±0.Pow(b) = +0 for finite b > 0 and not an odd integer
//	-1.Pow(±Inf) = 1
//	a.Pow(+Inf) = +Inf for |a| > 1
//	a.Pow(-Inf) = +0 for |a| > 1
//	a.Pow(+Inf) = +0 for |a| < 1
//	a.Pow(-Inf) = +Inf for |a| < 1
//	+Inf.Pow(b) = +Inf for b > 0
//	+Inf.Pow(b) = +0 for b < 0
//	-Inf.Pow(b) = (-0).Pow(-b)
//	a.Pow(b) = NaN for finite a < 0 and finite non-integer b
func (a Float16) Pow(b Float16) Float16 {
	return NewFloat16(math.Pow(a.Float64().BuiltIn(), b.Float64().BuiltIn()))
}
