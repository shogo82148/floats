package floats

import "math"

// Dim returns the maximum of a-b or 0.
//
// Special cases are:
//
//	+Inf.Dim(+Inf) = NaN
//	-Inf.Dim(-Inf) = NaN
//	x.Dim(NaN) = NaN.Dim(x) = NaN
func (a Float64) Dim(b Float64) Float64 {
	return Float64(math.Dim(a.BuiltIn(), b.BuiltIn()))
}
