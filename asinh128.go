package floats

// Asinh returns the inverse hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Asinh() = ±0
//	±Inf.Asinh() = ±Inf
//	NaN.Asinh() = NaN
func (a Float128) Asinh() Float128 {
	// https://github.com/chewxy/math32/blob/912ef0b2e4151df0148d7645c92a7b5e22f887f5/asinh.go#L36-L66
	var (
		Ln2      = Float128{0x3ffe_62e4_2fef_a39e, 0xf357_93c7_6730_07e6} // 6.93147180559945286227e-01
		Zero     = Float128{}                                             // 0.0
		One      = Float128(uvone128)                                     // 1.0
		Two      = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000} // 2.0
		NearZero = Float128{0x3fc5_0000_0000_0000, 0x0000_0000_0000_0000} // 2**-58
		Large    = Float128{0x4039_0000_0000_0000, 0x0000_0000_0000_0000} // 2**58
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
	var temp Float128
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

// Acosh returns the inverse hyperbolic cosine of a.
//
// Special cases are:
//
//	+Inf.Acosh() = +Inf
//	x.Acosh() = NaN if x < 1
//	NaN.Acosh() = NaN
func (a Float128) Acosh() Float128 {
	var (
		Ln2   = Float128{0x3ffe_62e4_2fef_a39e, 0xf357_93c7_6730_07e6} // 6.93147180559945286227e-01
		One   = Float128(uvone128)                                     // 1.0
		Two   = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000} // 2.0
		Large = Float128{0x4039_0000_0000_0000, 0x0000_0000_0000_0000} // 2**58
	)
	switch {
	case a.Lt(One) || a.IsNaN():
		return NewFloat128NaN()
	case a.Eq(One):
		return Float128{}
	case a.Ge(Large):
		return a.Log().Add(Ln2) // a > 2**58
	case a.Gt(Two):
		return (a.Add((a.Mul(a).Sub(One)).Sqrt())).Log() // 2**58 > a > 2.0
	}
	t := a.Sub(One)
	return (t.Add((t.Mul(t).Add(Two.Mul(t))).Sqrt())).Log1p() // 2 >= a > 1
}
