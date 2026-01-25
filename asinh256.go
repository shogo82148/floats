package floats

// Asinh returns the inverse hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Asinh() = ±0
//	±Inf.Asinh() = ±Inf
//	NaN.Asinh() = NaN
func (a Float256) Asinh() Float256 {
	var (
		// Ln2 = ln(2)
		Ln2 = Float256{
			0x3fff_e62e_42fe_fa39, 0xef35_793c_7673_007e,
			0x5ed5_e81e_6864_ce53, 0x16c5_b141_a2eb_7175,
		}
		// Zero = 0.0
		Zero = Float256{}
		// One = 1.0
		One = Float256(uvone256)
		// Two = 2.0
		Two = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		// NeearZero = 2**-170
		NearZero = Float256{
			0x3ff5_5000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		// Large = 2**170
		Large = Float256{
			0x400a_9000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)
	// special cases
	if a.IsNaN() || a.IsInf(0) {
		return a
	}
	sign := false
	if a.Lt(Zero) {
		a = a.Neg()
		sign = true
	}
	var temp Float256
	switch {
	case a.Gt(Large):
		temp = a.Log().Add(Ln2) // |a| > 2**58
	case a.Gt(Two):
		temp = (a.Add(a).Add(One.Quo((a.Mul(a).Add(One)).Sqrt().Add(a)))).Log() // 2**58 > |a| > 2.0
	case a.Lt(NearZero):
		temp = a // |a| < 2**-58
	default:
		a2 := a.Mul(a)
		temp = (a.Add(a2.Quo(One.Add(One.Add(a2).Sqrt())))).Log1p() // 2.0 > |a| > 2**-58
	}
	if sign {
		temp = temp.Neg()
	}
	return temp
}
