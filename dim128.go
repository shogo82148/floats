package floats

// Dim returns the maximum of a-b or 0.
//
// Special cases are:
//
//	+Inf.Dim(+Inf) = NaN
//	-Inf.Dim(-Inf) = NaN
//	x.Dim(NaN) = NaN.Dim(x) = NaN
func (a Float128) Dim(b Float128) Float128 {
	// The special cases result in NaN after the subtraction:
	//      +Inf - +Inf = NaN
	//      -Inf - -Inf = NaN
	//       NaN - b    = NaN
	//         a - NaN  = NaN
	v := a.Sub(b)
	if v.Le(Float128{}) {
		// v is negative or 0
		return Float128{}
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
func (a Float128) Max(b Float128) Float128 {
	// special cases
	switch {
	case a.IsInf(1) || b.IsInf(1):
		return NewFloat128Inf(1)
	case a.IsNaN() || b.IsNaN():
		return NewFloat128NaN()
	case a.IsZero() && a.Eq(b):
		if a.Signbit() {
			return b
		}
		return a
	}

	if a.Gt(b) {
		return a
	}
	return b
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
func (a Float128) Min(b Float128) Float128 {
	// special cases
	switch {
	case a.IsInf(-1) || b.IsInf(-1):
		return NewFloat128Inf(-1)
	case a.IsNaN() || b.IsNaN():
		return NewFloat128NaN()
	case a.IsZero() && a.Eq(b):
		if a.Signbit() {
			return a
		}
		return b
	}

	if a.Lt(b) {
		return a
	}
	return b
}
