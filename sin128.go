package floats

// Sin returns the sine of the radian argument a.
//
// Special cases are:
//
//	±0.Sin() = ±0
//	±Inf.Sin() = NaN
//	NaN.Sin() = NaN
func (a Float128) Sin() Float128 {
	// special cases
	switch {
	case a.IsZero():
		return a
	case a.IsNaN() || a.IsInf(0):
		return NewFloat128NaN()
	}

	var (
		Zero = Float128{}

		One = Float128(uvone128)

		// MPI4 = 4/pi
		MPI4 = Float128{0x3fff_45f3_06dc_9c88, 0x2a53_f84e_afa3_ea6a}

		// Pi/4 split into three parts
		PI4A = Float128{0x3ffe_921f_b544_42d1, 0x8400_0000_0000_0000}
		PI4B = Float128{0x3fc4_a626_3314_5c06, 0xe000_0000_0000_0000}
		PI4C = Float128{0x3f8b_cd12_9024_e088, 0xa67c_c740_20bb_ea64}
	)

	// make argument positive but save the sign
	sign := false
	if a.Lt(Zero) {
		a = a.Neg()
		sign = true
	}

	var j uint64
	var y, z Float128
	// TODO: use Payne-Hanek reduction for large arguments
	j = a.Mul(MPI4).Uint64()
	y = NewFloat128(float64(j))

	// map zeros to origin
	if j&1 == 1 {
		j++
		y = y.Add(One)
	}
	j &= 7 // octant modulo 2Pi radians (360 degrees)

	// Extended precision modular arithmetic
	y = y.Neg()
	z = FMA128(y, PI4A, a)
	z = FMA128(y, PI4B, z)
	z = FMA128(y, PI4C, z)

	// reflect in x axis
	if j > 3 {
		sign = !sign
		j -= 4
	}

	if j == 1 || j == 2 {
		// taylor series expansion of cos around 0
		y = Zero
		for n := 20; n >= 0; n-- {
			term := power128(z, 2*n).Quo(factorial128(2 * n))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
	} else {
		// taylor series expansion of sin around 0
		y = Zero
		for n := 20; n >= 0; n-- {
			term := power128(z, 2*n+1).Quo(factorial128(2*n + 1))
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
func (a Float128) Cos() Float128 {
	// special cases
	switch {
	case a.IsNaN() || a.IsInf(0):
		return NewFloat128NaN()
	}

	var (
		Zero = Float128{}

		One = Float128(uvone128)

		// MPI4 = 4/pi
		MPI4 = Float128{0x3fff_45f3_06dc_9c88, 0x2a53_f84e_afa3_ea6a}

		// Pi/4 split into three parts
		PI4A = Float128{0x3ffe_921f_b544_42d1, 0x8400_0000_0000_0000}
		PI4B = Float128{0x3fc4_a626_3314_5c06, 0xe000_0000_0000_0000}
		PI4C = Float128{0x3f8b_cd12_9024_e088, 0xa67c_c740_20bb_ea64}
	)

	// make argument positive but save the sign
	sign := false
	a = a.Abs()

	var j uint64
	var y, z Float128
	// TODO: use Payne-Hanek reduction for large arguments
	j = a.Mul(MPI4).Uint64()
	y = NewFloat128(float64(j))

	// map zeros to origin
	if j&1 == 1 {
		j++
		y = y.Add(One)
	}
	j &= 7 // octant modulo 2Pi radians (360 degrees)

	// Extended precision modular arithmetic
	y = y.Neg()
	z = FMA128(y, PI4A, a)
	z = FMA128(y, PI4B, z)
	z = FMA128(y, PI4C, z)

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
		for n := 20; n >= 0; n-- {
			term := power128(z, 2*n+1).Quo(factorial128(2*n + 1))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
	} else {
		// taylor series expansion of cos around 0
		y = Zero
		for n := 20; n >= 0; n-- {
			term := power128(z, 2*n).Quo(factorial128(2 * n))
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
