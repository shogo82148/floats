package floats

// reduce128 reduces the non-negative finite argument a by Pi/4.
// It returns the octant j (0 <= j < 8) and the reduced argument z
// such that a = j * Pi/4 + z (mod 2*Pi) and |z| <= Pi/4.
//
// a is exactly representable in Float256, and reduce256 computes
// the reduced argument with more precision than Float128 has.
func reduce128(a Float128) (j uint64, z Float128) {
	j, z256 := reduce256(a.Float256())
	return j, z256.Float128()
}

// Sin returns the sine of the radian argument a.
//
// Special cases are:
//
//	±0.Sin() = ±0
//	±Inf.Sin() = NaN
//	NaN.Sin() = NaN
func (a Float128) Sin() Float128 {
	// special cases
	switch {
	case a.IsZero():
		return a
	case a.IsNaN() || a.IsInf(0):
		return NewFloat128NaN()
	}

	var Zero = Float128{}

	// make argument positive but save the sign
	sign := false
	if a.Lt(Zero) {
		a = a.Neg()
		sign = true
	}

	j, z := reduce128(a)
	var y Float128

	// reflect in x axis
	if j > 3 {
		sign = !sign
		j -= 4
	}

	if j == 1 || j == 2 {
		// taylor series expansion of cos around 0
		y = Zero
		for n := 20; n >= 0; n-- {
			term := power128(z, 2*n).Quo(factorial128(2 * n))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
	} else {
		// taylor series expansion of sin around 0
		y = Zero
		for n := 20; n >= 0; n-- {
			term := power128(z, 2*n+1).Quo(factorial128(2*n + 1))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
	}
	if sign {
		y = y.Neg()
	}
	return y
}

// Cos returns the cosine of the radian argument a.
//
// Special cases are:
//
//	±Inf.Cos() = NaN
//	NaN.Cos() = NaN
func (a Float128) Cos() Float128 {
	// special cases
	switch {
	case a.IsNaN() || a.IsInf(0):
		return NewFloat128NaN()
	}

	var Zero = Float128{}

	// make argument positive but save the sign
	sign := false
	a = a.Abs()

	j, z := reduce128(a)
	var y Float128

	if j > 3 {
		j -= 4
		sign = !sign
	}
	if j > 1 {
		sign = !sign
	}

	if j == 1 || j == 2 {
		// taylor series expansion of sin around 0
		y = Zero
		for n := 20; n >= 0; n-- {
			term := power128(z, 2*n+1).Quo(factorial128(2*n + 1))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
	} else {
		// taylor series expansion of cos around 0
		y = Zero
		for n := 20; n >= 0; n-- {
			term := power128(z, 2*n).Quo(factorial128(2 * n))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
	}
	if sign {
		y = y.Neg()
	}
	return y
}

// Sincos returns Sin(a), Cos(a).
//
// Special cases are:
//
//	±0.Sincos() = ±0, 1
//	±Inf.Sincos() = NaN, NaN
//	NaN.Sincos() = NaN, NaN
func (a Float128) Sincos() (sin, cos Float128) {
	var (
		Zero = Float128{}
		One  = Float128(uvone128)
	)

	// special cases
	switch {
	case a.IsZero():
		return a, One // return ±0.0, 1.0
	case a.IsNaN() || a.IsInf(0):
		return NewFloat128NaN(), NewFloat128NaN()
	}

	// make argument positive
	sinSign, cosSign := false, false
	if a.Lt(Zero) {
		a = a.Neg()
		sinSign = true
	}

	j, z := reduce128(a)

	if j > 3 { // reflect in x axis
		j -= 4
		sinSign, cosSign = !sinSign, !cosSign
	}
	if j > 1 {
		cosSign = !cosSign
	}

	// taylor series expansion of sin around 0
	sin = Zero
	for n := 20; n >= 0; n-- {
		term := power128(z, 2*n+1).Quo(factorial128(2*n + 1))
		if n%2 != 0 {
			term = term.Neg()
		}
		sin = sin.Add(term)
	}

	// taylor series expansion of cos around 0
	cos = Zero
	for n := 20; n >= 0; n-- {
		term := power128(z, 2*n).Quo(factorial128(2 * n))
		if n%2 != 0 {
			term = term.Neg()
		}
		cos = cos.Add(term)
	}

	if j == 1 || j == 2 {
		sin, cos = cos, sin
	}
	if cosSign {
		cos = cos.Neg()
	}
	if sinSign {
		sin = sin.Neg()
	}
	return
}

// Tan returns the tangent of the radian argument a.
//
// Special cases are:
//
//	±0.Tan() = ±0
//	±Inf.Tan() = NaN
//	NaN.Tan() = NaN
func (a Float128) Tan() Float128 {
	sin, cos := a.Sincos()
	return sin.Quo(cos)
}
