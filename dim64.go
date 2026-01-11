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

// Max returns the larger of a or b.
//
// Special cases are:
//
//	x.Max(+Inf) = +Inf.Max(x) = +Inf
//	x.Max(NaN) = NaN.Max(x) = NaN
//	+0.Max(±0) = ±0.Max(+0) = +0
//	-0.Max(-0) = -0
//
// Note that this differs from the built-in function max when called
// with NaN and +Inf.
func (a Float64) Max(b Float64) Float64 {
	return Float64(math.Max(a.BuiltIn(), b.BuiltIn()))
}

// Min returns the smaller of a or b.
//
// Special cases are:
//
//	x.Min(-Inf) = -Inf.Min(x) = -Inf
//	x.Min(NaN) = NaN.Min(x) = NaN
//	-0.Min(±0) = ±0.Min(-0) = -0
//
// Note that this differs from the built-in function min when called
// with NaN and -Inf.
func (a Float64) Min(b Float64) Float64 {
	return Float64(math.Min(a.BuiltIn(), b.BuiltIn()))
}
