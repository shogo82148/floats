package floats

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
