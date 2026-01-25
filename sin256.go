package floats

// Sin returns the sine of the radian argument a.
//
// Special cases are:
//
//	±0.Sin() = ±0
//	±Inf.Sin() = NaN
//	NaN.Sin() = NaN
func (a Float256) Sin() Float256 {
	// special cases
	switch {
	case a.IsZero():
		return a
	case a.IsNaN() || a.IsInf(0):
		return NewFloat256NaN()
	}

	var (
		Zero = Float256{}
		One  = Float256(uvone256)

		// MPI4 = 4/pi
		MPI4 = Float256{
			0x3fff_f45f_306d_c9c8, 0x82a5_3f84_eafa_3ea6,
			0x9bb8_1b6c_52b3_2788, 0x7208_3fca_2c75_7bd7,
		}

		// Pi/4 split into three parts
		PI4A = Float256{
			0x3fff_e921_fb54_442d, 0x1846_9898_cc51_701b, 0x8300_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		PI4B = Float256{
			0x3ff8_9344_a409_3822, 0x299f_31d0_082e_fa98,
			0xec00_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		PI4C = Float256{
			0x3ff1_339b_2251_4a08, 0x798e_3404_ddef_9519,
			0xb3cd_3a43_1b30_2b0a, 0x6df2_5f79_0a22_9257,
		}
	)

	// make argument positive but save the sign
	sign := false
	if a.Lt(Zero) {
		a = a.Neg()
		sign = true
	}

	var j uint64
	var y, z Float256
	// TODO: use Payne-Hanek reduction for large arguments
	j = a.Mul(MPI4).Uint64()
	y = NewFloat256(float64(j))

	// map zeros to origin
	if j&1 == 1 {
		j++
		y = y.Add(One)
	}
	j &= 7 // octant modulo 2Pi radians (360 degrees)

	// Extended precision modular arithmetic
	y = y.Neg()
	z = FMA256(y, PI4A, a)
	z = FMA256(y, PI4B, z)
	z = FMA256(y, PI4C, z)

	// reflect in x axis
	if j > 3 {
		sign = !sign
		j -= 4
	}

	if j == 1 || j == 2 {
		// taylor series expansion of cos around 0
		y = Zero
		for n := 50; n >= 0; n-- {
			term := power256(z, 2*n).Quo(factorial256(2 * n))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
	} else {
		// taylor series expansion of sin around 0
		y = Zero
		for n := 50; n >= 0; n-- {
			term := power256(z, 2*n+1).Quo(factorial256(2*n + 1))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
	}
	if sign {
		y = y.Neg()
	}
	return y
}

// Cos returns the cosine of the radian argument a.
//
// Special cases are:
//
//	±Inf.Cos() = NaN
//	NaN.Cos() = NaN
func (a Float256) Cos() Float256 {
	// special cases
	switch {
	case a.IsNaN() || a.IsInf(0):
		return NewFloat256NaN()
	}

	var (
		Zero = Float256{}
		One  = Float256(uvone256)

		// MPI4 = 4/pi
		MPI4 = Float256{
			0x3fff_f45f_306d_c9c8, 0x82a5_3f84_eafa_3ea6,
			0x9bb8_1b6c_52b3_2788, 0x7208_3fca_2c75_7bd7,
		}

		// Pi/4 split into three parts
		PI4A = Float256{
			0x3fff_e921_fb54_442d, 0x1846_9898_cc51_701b, 0x8300_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		PI4B = Float256{
			0x3ff8_9344_a409_3822, 0x299f_31d0_082e_fa98,
			0xec00_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		PI4C = Float256{
			0x3ff1_339b_2251_4a08, 0x798e_3404_ddef_9519,
			0xb3cd_3a43_1b30_2b0a, 0x6df2_5f79_0a22_9257,
		}
	)

	// make argument positive but save the sign
	sign := false
	a = a.Abs()

	var j uint64
	var y, z Float256
	// TODO: use Payne-Hanek reduction for large arguments
	j = a.Mul(MPI4).Uint64()
	y = NewFloat256(float64(j))

	// map zeros to origin
	if j&1 == 1 {
		j++
		y = y.Add(One)
	}
	j &= 7 // octant modulo 2Pi radians (360 degrees)

	// Extended precision modular arithmetic
	y = y.Neg()
	z = FMA256(y, PI4A, a)
	z = FMA256(y, PI4B, z)
	z = FMA256(y, PI4C, z)

	if j > 3 {
		j -= 4
		sign = !sign
	}
	if j > 1 {
		sign = !sign
	}

	if j == 1 || j == 2 {
		// taylor series expansion of sin around 0
		y = Zero
		for n := 50; n >= 0; n-- {
			term := power256(z, 2*n+1).Quo(factorial256(2*n + 1))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
	} else {
		// taylor series expansion of cos around 0
		y = Zero
		for n := 50; n >= 0; n-- {
			term := power256(z, 2*n).Quo(factorial256(2 * n))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
	}
	if sign {
		y = y.Neg()
	}
	return y
}

// Sincos returns Sin(a), Cos(a).
//
// Special cases are:
//
//	±0.Sincos() = ±0, 1
//	±Inf.Sincos() = NaN, NaN
//	NaN.Sincos() = NaN, NaN
func (a Float256) Sincos() (sin, cos Float256) {
	var (
		Zero = Float256{}
		One  = Float256(uvone256)

		// MPI4 = 4/pi
		MPI4 = Float256{
			0x3fff_f45f_306d_c9c8, 0x82a5_3f84_eafa_3ea6,
			0x9bb8_1b6c_52b3_2788, 0x7208_3fca_2c75_7bd7,
		}

		// Pi/4 split into three parts
		PI4A = Float256{
			0x3fff_e921_fb54_442d, 0x1846_9898_cc51_701b, 0x8300_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		PI4B = Float256{
			0x3ff8_9344_a409_3822, 0x299f_31d0_082e_fa98,
			0xec00_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		PI4C = Float256{
			0x3ff1_339b_2251_4a08, 0x798e_3404_ddef_9519,
			0xb3cd_3a43_1b30_2b0a, 0x6df2_5f79_0a22_9257,
		}
	)

	// special cases
	switch {
	case a.IsZero():
		return a, One // return ±0.0, 1.0
	case a.IsNaN() || a.IsInf(0):
		return NewFloat256NaN(), NewFloat256NaN()
	}

	// make argument positive
	sinSign, cosSign := false, false
	if a.Lt(Zero) {
		a = a.Neg()
		sinSign = true
	}

	var j uint64
	var y, z Float256
	// TODO: use Payne-Hanek reduction for large arguments
	j = a.Mul(MPI4).Uint64()
	y = NewFloat256(float64(j))

	// map zeros to origin
	if j&1 == 1 {
		j++
		y = y.Add(One)
	}
	j &= 7 // octant modulo 2Pi radians (360 degrees)

	// Extended precision modular arithmetic
	y = y.Neg()
	z = FMA256(y, PI4A, a)
	z = FMA256(y, PI4B, z)
	z = FMA256(y, PI4C, z)

	if j > 3 { // reflect in x axis
		j -= 4
		sinSign, cosSign = !sinSign, !cosSign
	}
	if j > 1 {
		cosSign = !cosSign
	}

	// taylor series expansion of sin around 0
	sin = Zero
	for n := 50; n >= 0; n-- {
		term := power256(z, 2*n+1).Quo(factorial256(2*n + 1))
		if n%2 != 0 {
			term = term.Neg()
		}
		sin = sin.Add(term)
	}

	// taylor series expansion of cos around 0
	cos = Zero
	for n := 50; n >= 0; n-- {
		term := power256(z, 2*n).Quo(factorial256(2 * n))
		if n%2 != 0 {
			term = term.Neg()
		}
		cos = cos.Add(term)
	}

	if j == 1 || j == 2 {
		sin, cos = cos, sin
	}
	if cosSign {
		cos = cos.Neg()
	}
	if sinSign {
		sin = sin.Neg()
	}
	return
}

// Tan returns the tangent of the radian argument a.
//
// Special cases are:
//
//	±0.Tan() = ±0
//	±Inf.Tan() = NaN
//	NaN.Tan() = NaN
func (a Float256) Tan() Float256 {
	sin, cos := a.Sincos()
	return sin.Quo(cos)
}
