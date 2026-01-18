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

	// TODO: optimize for large and small values
	var temp Float256
	ex := a.Exp()
	temp = (ex.Sub(One.Quo(ex))).Mul(Half)

	if sign {
		temp = temp.Neg()
	}
	return temp
}
