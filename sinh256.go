package floats

// Sinh returns the hyperbolic sine of x.
//
// Special cases are:
//
//	±0.Sinh() = ±0
//	±Inf.Sinh() = ±Inf
//	NaN.Sinh() = NaN
func (a Float256) Sinh() Float256 {
	var (
		// Large = 84
		Large = Float256{
			0x4000_5500_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// One = 1.0
		One = Float256{
			0x3fff_f000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// Half = 0.5
		Half = Float256{
			0x3fff_e000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)

	sign := false
	if a.Signbit() {
		a = a.Neg()
		sign = true
	}

	var temp Float256
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
			temp = temp.Add(power256(a, n).Quo(factorial256(n)))
		}
	}

	if sign {
		temp = temp.Neg()
	}
	return temp
}

// Cosh returns the hyperbolic cosine of x.
//
// Special cases are:
//
//	±0.Cosh() = 1
//	±Inf.Cosh() = +Inf
//	NaN.Cosh() = NaN
func (a Float256) Cosh() Float256 {
	var (
		// Large = 84
		Large = Float256{
			0x4000_5500_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// One = 1.0
		One = Float256{
			0x3fff_f000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// Half = 0.5
		Half = Float256{
			0x3fff_e000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)
	a = a.Abs()
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
func (a Float256) Tanh() Float256 {
	var (
		// ln(max float256 + 0.5ulp)/2 = ln(2²⁶²¹⁴³×(2-2⁻²³⁷))/2
		// ~ 90852.18725035315159593544862376611913079195361086737666810577020431809
		Overflow = Float256{
			0x4000_f62e_42fe_fa39, 0xef35_793c_7673_007e,
			0x5ed5_e81e_6864_ce53, 0x16c5_b141_a2eb_7177,
		}

		// One = 1.0
		One = Float256{
			0x3fff_f000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// Two = 2.0
		Two = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// NearZero = 0.625
		NearZero = Float256{
			0x3fff_e400_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
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
		if a.Signbit() {
			z = z.Neg()
		}
	}
	return z
}
