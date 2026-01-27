package floats

// Erf returns the error function of a.
//
// Special cases are:
//
//	+Inf.Erf() = 1
//	-Inf.Erf() = -1
//	NaN.Erf() = NaN
func (a Float128) Erf() Float128 {
	var (
		// One is 1
		One = Float128(uvone128)
		// TwoOverSqrtPi is 2/sqrt(π)
		TwoOverSqrtPi = Float128{0x3fff_20dd_7504_29b6, 0xd11a_e3a9_14fe_d7fe}
		// TwoPointFour is 2.4
		TwoPointFour = Float128{0x4000_3333_3333_3333, 0x3333_3333_3333_3333}
		// Nine is 9
		Nine = Float128{0x4002_2000_0000_0000, 0x0000_0000_0000_0000}
		// Sqrt2 is sqrt(2)
		Sqrt2 = Float128{0x3fff_6a09_e667_f3bc, 0xc908_b2fb_1366_ea95}
		// SqrtTwoOverPi is sqrt(2/π)
		SqrtTwoOverPi = Float128{0x3ffe_9884_533d_4365, 0x08d0_fcb3_c500_bab9}
	)

	// special cases
	switch {
	case a.IsInf(0):
		return Float128(uvone128).Copysign(a)
	case a.IsNaN():
		return NewFloat128NaN()
	}

	sign := false
	if a.Signbit() {
		sign = true
		a = a.Neg()
	}

	var y Float128
	switch {
	case a.Lt(TwoPointFour):
		// use Taylor series expansion
		// erf(x) = 2/sqrt(π) * Σ[n=0..∞] (-1)^n * x^(2n+1) / (n! * (2n+1))
		for n := 50; n >= 0; n-- {
			term := power128(a, 2*n+1).Quo(factorial128(n).Mul(NewFloat128(float64(2*n + 1))))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
		y = y.Mul(TwoOverSqrtPi)

	case a.Gt(Nine):
		y = One

	default:
		// use continued fraction expansion
		x := Sqrt2.Mul(a)
		for n := 80; n >= 1; n-- {
			y = NewFloat128(float64(n)).Quo(x.Add(y))
		}
		y = One.Sub(a.Mul(a).Neg().Exp().Quo(x.Add(y)).Mul(SqrtTwoOverPi))
	}
	if sign {
		y = y.Neg()
	}
	return y
}
