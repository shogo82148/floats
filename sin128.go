package floats

func bitOf4PiLong128(k int) uint64 {
	if k < 0 {
		return 0
	}
	if k == 0 {
		return 1
	}
	word := 1 + (k-1)/64
	if word >= len(mPi4Long) {
		return 0
	}
	bit := 63 - ((k - 1) % 64)
	return (mPi4Long[word] >> uint(bit)) & 1
}

func reducePowerOfTwoLarge128(a Float128) (j uint64, z Float128, ok bool) {
	_, exp, frac := a.normalize()
	if frac[0] != (uint64(1)<<(shift128-64)) || frac[1] != 0 {
		return 0, Float128{}, false
	}

	// Need enough 4/pi bits for fractional reconstruction.
	const fracBits = 300
	maxBit := 64 * (len(mPi4Long) - 1)
	if exp+fracBits+2 > maxBit {
		return 0, Float128{}, false
	}

	// j0 = floor(2^exp * 4/pi) mod 8 from the 4/pi bit stream.
	j0 := bitOf4PiLong128(exp) | (bitOf4PiLong128(exp-1) << 1) | (bitOf4PiLong128(exp-2) << 2)

	one256 := Float256(uvone256)
	frac256 := Float256{}
	for k := exp + fracBits; k >= exp+1; k-- {
		if bitOf4PiLong128(k) == 1 {
			frac256 = frac256.Add(one256)
		}
		frac256 = frac256.Ldexp(-1)
	}

	if j0&1 == 1 {
		j = (j0 + 1) & 7
		frac256 = frac256.Sub(one256)
	} else {
		j = j0
	}

	pi4 := Float256{
		0x3fff_e921_fb54_442d, 0x1846_9898_cc51_701b,
		0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
	}
	z = frac256.Mul(pi4).Float128()
	return j, z, true
}

func modTauLarge128(a Float128) Float128 {
	// Tau = 2*pi
	var Tau256 = Float256{
		0x4001_0921_fb54_442d, 0x1846_9898_cc51_701b,
		0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
	}
	var Zero256 = Float256{}
	var One256 = Float256(uvone256)

	_, exp, frac := a.normalize()
	k0 := exp - shift128

	var pow Float256
	if k0 >= 0 {
		pow = One256
		for i := 0; i < k0; i++ {
			pow = pow.Add(pow)
			if pow.Ge(Tau256) {
				pow = pow.Sub(Tau256)
			}
		}
	} else {
		pow = One256.Ldexp(k0)
	}

	r := Zero256
	for p := 0; p <= shift128; p++ {
		var bit uint64
		if p < 64 {
			bit = (frac[1] >> uint(p)) & 1
		} else {
			bit = (frac[0] >> uint(p-64)) & 1
		}
		if bit != 0 {
			r = r.Add(pow)
			if r.Ge(Tau256) {
				r = r.Sub(Tau256)
			}
		}

		pow = pow.Add(pow)
		if pow.Ge(Tau256) {
			pow = pow.Sub(Tau256)
		}
	}

	return r.Float128()
}

func reduce128(a Float128) (j uint64, z Float128) {
	var (
		One = Float128(uvone128)

		// Tau = 2*pi
		// Largest consecutive integer in float64. Beyond this, converting
		// octant counters through float64 loses integer precision.
		TwoPow53 = NewFloat128(0x1p53)

		// MPI4 = 4/pi
		MPI4 = Float128{0x3fff_45f3_06dc_9c88, 0x2a53_f84e_afa3_ea6a}

		// Pi/4 split into three parts
		PI4A = Float128{0x3ffe_921f_b544_42d1, 0x8400_0000_0000_0000}
		PI4B = Float128{0x3fc4_a626_3314_5c06, 0xe000_0000_0000_0000}
		PI4C = Float128{0x3f8b_cd12_9024_e088, 0xa67c_c740_20bb_ea64}
	)

	if a.Mul(MPI4).Gt(TwoPow53) {
		if j0, z0, ok := reducePowerOfTwoLarge128(a); ok {
			return j0, z0
		}
		// Keep the reduced argument bounded so the octant index stays accurate.
		a = modTauLarge128(a)
	}

	y := NewFloat128(float64(a.Mul(MPI4).Uint64()))
	j = y.Uint64()

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
	return
}

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

	var Zero = Float128{}

	// make argument positive but save the sign
	sign := false
	if a.Lt(Zero) {
		a = a.Neg()
		sign = true
	}

	j, z := reduce128(a)
	var y Float128

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

	var Zero = Float128{}

	// make argument positive but save the sign
	sign := false
	a = a.Abs()

	j, z := reduce128(a)
	var y Float128

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

// Sincos returns Sin(a), Cos(a).
//
// Special cases are:
//
//	±0.Sincos() = ±0, 1
//	±Inf.Sincos() = NaN, NaN
//	NaN.Sincos() = NaN, NaN
func (a Float128) Sincos() (sin, cos Float128) {
	var (
		Zero = Float128{}
		One  = Float128(uvone128)
	)

	// special cases
	switch {
	case a.IsZero():
		return a, One // return ±0.0, 1.0
	case a.IsNaN() || a.IsInf(0):
		return NewFloat128NaN(), NewFloat128NaN()
	}

	// make argument positive
	sinSign, cosSign := false, false
	if a.Lt(Zero) {
		a = a.Neg()
		sinSign = true
	}

	j, z := reduce128(a)

	if j > 3 { // reflect in x axis
		j -= 4
		sinSign, cosSign = !sinSign, !cosSign
	}
	if j > 1 {
		cosSign = !cosSign
	}

	// taylor series expansion of sin around 0
	sin = Zero
	for n := 20; n >= 0; n-- {
		term := power128(z, 2*n+1).Quo(factorial128(2*n + 1))
		if n%2 != 0 {
			term = term.Neg()
		}
		sin = sin.Add(term)
	}

	// taylor series expansion of cos around 0
	cos = Zero
	for n := 20; n >= 0; n-- {
		term := power128(z, 2*n).Quo(factorial128(2 * n))
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
func (a Float128) Tan() Float128 {
	sin, cos := a.Sincos()
	return sin.Quo(cos)
}
