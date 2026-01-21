package floats

// Sinh returns the hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Sinh() = ±0
//	±Inf.Sinh() = ±Inf
//	NaN.Sinh() = NaN
func (a Float128) Sinh() Float128 {
	var (
		// Large = 42
		Large = Float128{0x4004_5000_0000_0000, 0x0000_0000_0000_0000}

		// One = 1.0
		One = Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}

		// Half = 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	sign := false
	if a.Signbit() {
		a = a.Neg()
		sign = true
	}

	var temp Float128
	switch {
	case a.Gt(Large):
		temp = a.Exp().Mul(Half)

	case a.Gt(Half):
		ex := a.Exp()
		temp = ex.Sub(One.Quo(ex)).Mul(Half)

	default:
		// Taylor series expansion
		// TODO: optimize using minimax approximation
		for n := 49; n >= 1; n -= 2 {
			temp = temp.Add(power128(a, n).Quo(factorial128(n)))
		}
	}

	if sign {
		temp = temp.Neg()
	}
	return temp
}

// Cosh returns the hyperbolic cosine of a.
//
// Special cases are:
//
//	±0.Cosh() = 1
//	±Inf.Cosh() = +Inf
//	NaN.Cosh() = NaN
func (a Float128) Cosh() Float128 {
	a = a.Abs()
	var (
		// Large = 42
		Large = Float128{0x4004_5000_0000_0000, 0x0000_0000_0000_0000}

		// One = 1.0
		One = Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}

		// Half = 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	if a.Gt(Large) {
		return a.Exp().Mul(Half)
	}
	ex := a.Exp()
	return ex.Add(One.Quo(ex)).Mul(Half)
}

// Tanh returns the hyperbolic tangent of a.
//
// Special cases are:
//
//	±0.Tanh() = ±0
//	±Inf.Tanh() = ±1
//	NaN.Tanh() = NaN
func (a Float128) Tanh() Float128 {
	var (
		// Overflow = ln(max float128 + 0.5ulp)/2 = ln(2¹⁶³⁸³×(2-2⁻¹¹³))/2
		// ~ 5678.2617031470719747459655389853825
		Overflow = Float128{0x400b_62e4_2fef_a39e, 0xf357_93c7_6730_07e6}

		// One = 1.0
		One = Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}

		// Two = 2.0
		Two = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}

		// NearZero = 0.625
		NearZero = Float128{0x3ffe_4000_0000_0000, 0x0000_0000_0000_0000}
	)
	z := a.Abs()
	switch {
	case z.Gt(Overflow):
		if a.Signbit() {
			return One.Neg()
		}
		return One
	case z.Ge(NearZero):
		s := z.Add(z).Exp()
		z = One.Sub(Two.Quo(s.Add(One)))
		if a.Signbit() {
			z = z.Neg()
		}
	case z.IsZero():
		return a
	default:
		// TODO: optimize using minimax approximation
		z = z.Sinh().Quo(z.Cosh())
	}
	return z
}
