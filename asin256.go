package floats

// Asin returns the arcsine, in radians, of a.
//
// Special cases are:
//
//	±0.Asin() = ±0
//	x.Asin() = NaN if x < -1 or x > 1
func (a Float256) Asin() Float256 {
	switch {
	case a.IsZero():
		return a
	case a.IsNaN():
		return NewFloat256NaN()
	}

	sign := a.Signbit()
	if sign {
		a = a.Neg()
	}

	var Dot7 = Float256{
		0x3fff_e666_6666_6666, 0x6666_6666_6666_6666,
		0x6666_6666_6666_6666, 0x6666_6666_6666_6666,
	} // 0.7
	temp := Float256(uvone256).Sub(a.Mul(a)).Sqrt()
	if a.Gt(Dot7) {
		// asin(x) = pi/2 - atan(sqrt(1-x²)/x)
		var Pi2 = Float256{
			0x3fff_f921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
		temp = Pi2.Sub(satan256(temp.Quo(a)))
	} else {
		// asin(x) = atan(x/sqrt(1-x²))
		temp = satan256(a.Quo(temp))
	}

	if sign {
		temp = temp.Neg()
	}
	return temp
}

// Acos returns the arccosine, in radians, of a.
//
// Special case is:
//
//	x.Acos() = NaN if x < -1 or x > 1
func (a Float256) Acos() Float256 {
	// acos(x) = pi/2 - asin(x)
	var Pi2 = Float256{
		0x3fff_f921_fb54_442d, 0x1846_9898_cc51_701b,
		0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
	}
	return Pi2.Sub(a.Asin())
}

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float256) Atan() Float256 {
	// special cases
	switch {
	case a.IsZero():
		return a
	case a.IsNaN():
		return NewFloat256NaN()
	}
	if a.Signbit() {
		return satan256(a.Neg()).Neg()
	}
	return satan256(a)
}

// satan256 reduces its argument (known to be positive)
// to the range [0, 0.66] and calls xatan.
func satan256(x Float256) Float256 {
	var (
		One = Float256(uvone256)

		// Dot66 = 0.66
		Dot66 = Float256{
			0x3fff_e51e_b851_eb85, 0x1eb8_51eb_851e_b851,
			0xeb85_1eb8_51eb_851e, 0xb851_eb85_1eb8_51ec,
		}

		// Tan3pio8 = tan(3*pi/8) = 1 + sqrt(2)
		Tan3pio8 = Float256{
			0x4000_0350_4f33_3f9d, 0xe648_4597_d89b_3754,
			0xabe9_f1d6_f60b_a893, 0xba84_ced1_7ac8_5833,
		}

		// Pi/2 split into two parts
		Pi2Hi = Float256{
			0x3fff_f921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
		Pi2Lo = Float256{
			0x3ff1_1cd9_128a_5043, 0xcc71_a026_ef7c_a8cd,
			0x9e69_d218_d981_5853, 0x6f92_fd1a_85bb_f1f6,
		}

		// Pi/4 split into two parts
		Pi4Hi = Float256{
			0x3fff_e921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
		Pi4Lo = Float256{
			0x3ff1_0cd9_128a_5043, 0xcc71_a026_ef7c_a8cd,
			0x9e69_d218_d981_5853, 0x6f92_f43b_26c1_359d,
		}
	)

	switch {
	case x.Le(Dot66):
		return xatan256(x)
	case x.Gt(Tan3pio8):
		// atan(x) = pi/2 - atan(1/x)
		return Pi2Hi.Sub(xatan256(One.Quo(x))).Add(Pi2Lo)
	default:
		// atan(x) = pi/4 + atan((x-1)/(x+1))
		return Pi4Hi.Add(xatan256((x.Sub(One)).Quo(x.Add(One)))).Add(Pi4Lo)
	}
}

// xatan256 returns the arctangent.
// it is valid in the range [0, 0.66].
func xatan256(x Float256) Float256 {
	var y Float256
	for n := 130; n >= 0; n-- {
		term := power256(x, 2*n+1).Quo(NewFloat256(float64(2*n + 1)))
		if n%2 != 0 {
			term = term.Neg()
		}
		y = y.Add(term)
	}
	return y
}

// Atan2 returns the arc tangent of a/b, using
// the signs of the two to determine the quadrant
// of the return value.
//
// Special cases are (in order):
//
//	y.Atan2(NaN) = NaN
//	NaN.Atan2(x) = NaN
//	+0.Atan2(x>=0) = +0
//	-0.Atan2(x>=0) = -0
//	+0.Atan2(x<=-0) = +Pi
//	-0.Atan2(x<=-0) = -Pi
//	y>0.Atan2(0) = +Pi/2
//	y<0.Atan2(0) = -Pi/2
//	+Inf.Atan2(+Inf) = +Pi/4
//	-Inf.Atan2(+Inf) = -Pi/4
//	+Inf.Atan2(-Inf) = 3Pi/4
//	-Inf.Atan2(-Inf) = -3Pi/4
//	y.Atan2(+Inf) = 0
//	(y>0).Atan2(-Inf) = +Pi
//	(y<0).Atan2(-Inf) = -Pi
//	+Inf.Atan2(x) = +Pi/2
//	-Inf.Atan2(x) = -Pi/2
func (a Float256) Atan2(b Float256) Float256 {
	var (
		Zero = Float256{}
		// Pi = Pi
		Pi = Float256{
			0x4000_0921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
		// Pi2 = Pi/2
		Pi2 = Float256{
			0x3fff_f921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
		// Pi4 = Pi/4
		Pi4 = Float256{
			0x3fff_e921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
		// Pi34 = 3Pi/4
		Pi34 = Float256{
			0x4000_02d9_7c7f_3321, 0xd234_f272_993d_1414,
			0xa2b3_9bd8_3750_ccf9, 0xbb2a_e031_19df_958a,
		}
	)

	// special cases
	switch {
	case a.IsNaN() || b.IsNaN():
		return NewFloat256NaN()
	case a.IsZero():
		if b.Ge(Zero) && !b.Signbit() {
			return Zero.Copysign(a)
		}
		return Pi.Copysign(a)
	case b.IsZero():
		return Pi2.Copysign(a)
	case b.IsInf(0):
		if b.IsInf(1) {
			switch {
			case a.IsInf(0):
				return Pi4.Copysign(a)
			default:
				return Zero.Copysign(a)
			}
		}
		switch {
		case a.IsInf(0):
			return Pi34.Copysign(a)
		default:
			return Pi.Copysign(a)
		}
	case a.IsInf(0):
		return Pi2.Copysign(a)
	}

	q := a.Quo(b).Atan()
	if b.Lt(Zero) {
		if q.Le(Zero) {
			return q.Add(Pi)
		}
		return q.Sub(Pi)
	}
	return q
}
