package floats

// Sinh returns the hyperbolic sine of x.
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
		temp = (ex.Sub(One.Quo(ex))).Mul(Half)

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
