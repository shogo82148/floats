package floats

// Dim returns the maximum of a-b or 0.
//
// Special cases are:
//
//	+Inf.Dim(+Inf) = NaN
//	-Inf.Dim(-Inf) = NaN
//	x.Dim(NaN) = NaN.Dim(x) = NaN
func (a Float256) Dim(b Float256) Float256 {
	// The special cases result in NaN after the subtraction:
	//      +Inf - +Inf = NaN
	//      -Inf - -Inf = NaN
	//       NaN - b    = NaN
	//         a - NaN  = NaN
	v := a.Sub(b)
	if v.Le(Float256{}) {
		// v is negative or 0
		return Float256{}
	}
	// v is positive or NaN
	return v
}
