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
