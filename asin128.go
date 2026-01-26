package floats

// Asin returns the arcsine, in radians, of a.
//
// Special cases are:
//
//	±0.Asin() = ±0
//	x.Asin() = NaN if x < -1 or x > 1
func (a Float128) Asin() Float128 {
	switch {
	case a.IsZero():
		return a
	case a.IsNaN():
		return NewFloat128NaN()
	}

	sign := a.Signbit()
	if sign {
		a = a.Neg()
	}

	var Dot7 = Float128{0x3ffe_6666_6666_6666, 0x6666_6666_6666_6666} // 0.7
	temp := Float128(uvone128).Sub(a.Mul(a)).Sqrt()
	if a.Gt(Dot7) {
		// asin(x) = pi/2 - atan(sqrt(1-x²)/x)
		var Pi2 = Float128{0x3fff_921f_b544_42d1, 0x8469_898c_c517_01b8}
		temp = Pi2.Sub(satan128(temp.Quo(a)))
	} else {
		// asin(x) = atan(x/sqrt(1-x²))
		temp = satan128(a.Quo(temp))
	}

	if sign {
		temp = temp.Neg()
	}
	return temp
}

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
		return xatan128(x)
	case x.Gt(Tan3pio8):
		// atan(x) = pi/2 - atan(1/x)
		return Pi2Hi.Sub(xatan128(One.Quo(x))).Add(Pi2Lo)
	default:
		// atan(x) = pi/4 + atan((x-1)/(x+1))
		return Pi4Hi.Add(xatan128((x.Sub(One)).Quo(x.Add(One)))).Add(Pi4Lo)
	}
}

// xatan128 returns the arctangent.
// it is valid in the range [0, 0.66].
func xatan128(x Float128) Float128 {
	var y Float128
	for n := 60; n >= 0; n-- {
		term := power128(x, 2*n+1).Quo(NewFloat128(float64(2*n + 1)))
		if n%2 != 0 {
			term = term.Neg()
		}
		y = y.Add(term)
	}
	return y
}
