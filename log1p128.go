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
func (a Float128) Log1p() Float128 {
	var (
		// One = 1.0
		One = Float128(uvone128)

		// Half = 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}

		// Sqrt2quo2 = sqrt(2)/2 ~ 707106781186547524400844362104849
		Sqrt2quo2 = Float128{0x3ffe_6a09_e667_f3bc, 0xc908_b2fb_1366_ea95}

		// Ln2Hi = ln(2) ~ 0.6931471805599453094172321214581765
		// Ln2Lo = ln(2) - Ln2Hi ~ 8.928835774481220748938623512047474e-35
		Ln2Hi = Float128{0x3ffe_62e4_2fef_a39e, 0xf357_93c7_6730_07e5}
		Ln2Lo = Float128{0x3f8d_dabd_03cd_0c99, 0xca62_d8b6_2834_5d6e}

		// Two = 2.0
		Two = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	// special cases
	switch {
	case a.Lt(One.Neg()) || a.IsNaN(): // includes -Inf
		return NewFloat128NaN()
	case a.Eq(One.Neg()):
		return NewFloat128Inf(-1)
	case a.IsInf(1):
		return NewFloat128Inf(1)
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

	// |a| < 0.5
	// log(1+a) = a - a²/2 + a³/3 - a⁴/4 + ...
	var r Float128
	for n := 100; n > 0; n-- {
		term := power128(a, n).Quo(NewFloat128(float64(n)))
		if n%2 == 0 {
			term = term.Neg()
		}
		r = r.Add(term)
	}
	return r
}
