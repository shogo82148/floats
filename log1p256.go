package floats

// Log1p returns the natural logarithm of 1 plus its argument a.
// It is more accurate than [Log](1 + a) when a is near zero.
//
// Special cases are:
//
//	+Inf.Log1p() = +Inf
//	±0.Log1p() = ±0
//	-1.Log1p() = -Inf
//	(a < -1).Log1p() = NaN
//	NaN.Log1p() = NaN
func (a Float256) Log1p() Float256 {
	var (
		// Sqrt2quo2 = sqrt(2)/2 ~ 0.707106781186547524400844362104849039284835937688474036588339868995366237
		Sqrt2quo2 = Float256{
			0x3fff_e6a0_9e66_7f3b, 0xcc90_8b2f_b136_6ea9,
			0x57d3_e3ad_ec17_5127, 0x7509_9da2_f590_b066,
		}

		// Ln2Hi = ln(2) ~ 0.69314718055994530941723212145817656807550013436025525412068000949339362
		// Ln2Lo = ln(2) - Ln2Hi ~ 1.68505384472783385591763704017160191700974868940429492048047E-72
		Ln2Hi = Float256{
			0x3fff_e62e_42fe_fa39, 0xef35_793c_7673_007e,
			0x5ed5_e81e_6864_ce53, 0x16c5_b141_a2eb_7175,
		}
		Ln2Lo = Float256{
			0x3ff1_07d1_5f3d_c3b1, 0x036f_5d64_c2ac_aa97,
			0xda57_d0d8_8769_7571, 0xae09_c0c7_cb80_70d0,
		}

		// One = 1.0
		One = Float256(uvone256)

		// Half = 0.5
		Half = Float256{
			0x3fff_e000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// Two = 2.0
		Two = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)

	// special cases
	switch {
	case a.Lt(One.Neg()) || a.IsNaN(): // includes -Inf
		return NewFloat256NaN()
	case a.Eq(One.Neg()):
		return NewFloat256Inf(-1)
	case a.IsInf(1):
		return NewFloat256Inf(1)
	case a.IsZero():
		return a
	}

	if a.Abs().Gt(Half) {
		// reduce
		f1, ki := a.Add(One).Frexp()
		if f1.Lt(Sqrt2quo2) {
			f1 = f1.Add(f1)
			ki--
		}
		f := f1.Sub(One) // f := f1 - 1
		k := NewFloat256(float64(ki))

		// compute
		// Let s = f/(2+f); log(1+f) = log((1+s)/(1-s)) = 2s + 2/3 s³ + 2/5 s⁵ + 2/7 s⁷ + ...
		// TODO: use a polynomial approximation
		s := f.Quo(f.Add(Two))
		var r Float256
		for n := 79; n > 0; n -= 2 {
			r = r.Add(power256(s, n).Quo(NewFloat256(float64(n))))
		}
		return k.Mul(Ln2Hi).Add(r.Add(r).Add(k.Mul(Ln2Lo)))
	}

	// |a| < 0.5
	// log(1+a) = a - a²/2 + a³/3 - a⁴/4 + ...
	// TODO: use a polynomial approximation
	var r Float256
	for n := 210; n > 0; n-- {
		term := power256(a, n).Quo(NewFloat256(float64(n)))
		if n%2 == 0 {
			term = term.Neg()
		}
		r = r.Add(term)
	}
	return r
}
