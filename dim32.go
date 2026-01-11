package floats

import "math"

// Dim returns the maximum of a-b or 0.
//
// Special cases are:
//
//	+Inf.Dim(+Inf) = NaN
//	-Inf.Dim(-Inf) = NaN
//	x.Dim(NaN) = NaN.Dim(x) = NaN
func (a Float32) Dim(b Float32) Float32 {
	// The special cases result in NaN after the subtraction:
	//      +Inf - +Inf = NaN
	//      -Inf - -Inf = NaN
	//       NaN - b    = NaN
	//         a - NaN  = NaN
	v := a.Sub(b)
	if v.Le(0) {
		// v is negative or 0
		return 0
	}
	// v is positive or NaN
	return v
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
func (a Float32) Max(b Float32) Float32 {
	return NewFloat32(math.Max(a.Float64().BuiltIn(), b.Float64().BuiltIn()))
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
func (a Float32) Min(b Float32) Float32 {
	return NewFloat32(math.Min(a.Float64().BuiltIn(), b.Float64().BuiltIn()))
}
