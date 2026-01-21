package floats

// Log returns the natural logarithm of a.
//
// Special cases are:
//
//	+Inf.Log() = +Inf
//	0.Log() = -Inf
//	(x < 0).Log() = NaN
//	NaN.Log() = NaN
func (a Float128) Log() Float128 {
	// special cases
	switch {
	case a.IsNaN() || a.IsInf(1):
		return a
	case a.Lt(Float128{}): // a < 0
		return NewFloat128NaN()
	case a.IsZero():
		return NewFloat128Inf(-1)
	}

	var (
		// Sqrt2quo2 = sqrt(2)/2 ~ 707106781186547524400844362104849
		Sqrt2quo2 = Float128{0x3ffe_6a09_e667_f3bc, 0xc908_b2fb_1366_ea95}

		// Ln2Hi = ln(2) ~ 0.6931471805599453094172321214581765
		// Ln2Lo = ln(2) - Ln2Hi ~ 8.928835774481220748938623512047474e-35
		Ln2Hi = Float128{0x3ffe_62e4_2fef_a39e, 0xf357_93c7_6730_07e5}
		Ln2Lo = Float128{0x3f8d_dabd_03cd_0c99, 0xca62_d8b6_2834_5d6e}

		// Two = 2.0
		Two = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	// reduce
	f1, ki := a.Frexp()
	if f1.Lt(Sqrt2quo2) {
		f1 = f1.Add(f1)
		ki--
	}
	f := f1.Sub(Float128(uvone128)) // f := f1 - 1
	k := NewFloat128(float64(ki))

	// compute
	// Let s = f/(2+f); log(1+f) = log((1+s)/(1-s)) = 2s + 2/3 s³ + 2/5 s⁵ + 2/7 s⁷ + ...
	// TODO: use a polynomial approximation
	s := f.Quo(f.Add(Two))
	var r Float128
	for n := 39; n > 0; n -= 2 {
		r = r.Add(power128(s, n).Quo(NewFloat128(float64(n))))
	}
	return k.Mul(Ln2Hi).Add(r.Add(r).Add(k.Mul(Ln2Lo)))
}

// Log10 returns the decimal logarithm of a.
// The special cases are the same as for [Log].
func (a Float128) Log10() Float128 {
	// // 1/ln(10) ~ 0.4342944819032518276511289189166051
	var Ln10Inv = Float128{0x3ffd_bcb7_b152_6e50, 0xe32a_6ab7_555f_5a68}
	return a.Log().Mul(Ln10Inv)
}

// Log2 returns the binary logarithm of a.
// The special cases are the same as for [Log].
func (a Float128) Log2() Float128 {
	var (
		// Half = 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}

		// Ln2Inv = 1/ln(2)
		// ~ 1.442695040888963407359924681001892
		Ln2Inv = Float128{0x3fff_7154_7652_b82f, 0xe177_7d0f_fda0_d23a}
	)

	frac, exp := a.Frexp()
	// Make sure exact powers of two give an exact answer.
	// Don't depend on Log(0.5)*(1/Ln2)+exp being exactly exp-1.
	if frac.Eq(Half) {
		return NewFloat128(float64(exp - 1))
	}
	return frac.Log().Mul(Ln2Inv).Add(NewFloat128(float64(exp)))
}
