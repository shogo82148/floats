package floats

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float128) Atan() Float128 {
	// special cases
	switch {
	case a.IsZero():
		return a
	case a.IsNaN():
		return NewFloat128NaN()
	}
	if a.Signbit() {
		return satan128(a.Neg()).Neg()
	}
	return satan128(a)
}

// satan128 reduces its argument (known to be positive)
// to the range [0, 0.66] and calls xatan.
func satan128(x Float128) Float128 {
	var (
		One = Float128(uvone128)

		// Dot66 = 0.66
		Dot66 = Float128{0x3ffe_51eb_851e_b851, 0xeb85_1eb8_51eb_851f}

		// Tan3pio8 = tan(3*pi/8) = 1 + sqrt(2)
		Tan3pio8 = Float128{0x4000_3504_f333_f9de, 0x6484_597d_89b3_754b}

		// Pi/2 split into two parts
		Pi2Hi = Float128{0x3fff_921f_b544_42d1, 0x8469_898c_c517_01b8}
		Pi2Lo = Float128{0x3f8c_cd12_9024_e088, 0xa67c_c740_20bb_ea64}

		// Pi/4 split into two parts
		Pi4Hi = Float128{0x3ffe_921f_b544_42d1, 0x8469_898c_c517_01b8}
		Pi4Lo = Float128{0x3f8b_cd12_9024_e088, 0xa67c_c740_20bb_ea64}
	)

	switch {
	case x.Le(Dot66):
		return xatan(x)
	case x.Gt(Tan3pio8):
		// atan(x) = pi/2 - atan(1/x)
		return Pi2Hi.Sub(xatan(One.Quo(x))).Add(Pi2Lo)
	default:
		// atan(x) = pi/4 + atan((x-1)/(x+1))
		return Pi4Hi.Add(xatan((x.Sub(One)).Quo(x.Add(One)))).Add(Pi4Lo)
	}
}

// xatan returns the arctangent.
// it is valid in the range [0, 0.66].
func xatan(x Float128) Float128 {
	var y Float128
	for n := 50; n >= 0; n-- {
		term := power128(x, 2*n+1).Quo(NewFloat128(float64(2*n + 1)))
		if n%2 != 0 {
			term = term.Neg()
		}
		y = y.Add(term)
	}
	return y
}
