package floats

// Jn returns the order-n Bessel function of the first kind.
//
// Special cases are:
//
//	Jn(n, ±Inf) = 0
//	Jn(n, NaN) = NaN
func (a Float256) Jn(n int) Float256 {
	if a.IsNaN() {
		return a
	}
	switch n {
	case 0:
		return a.J0()
	case 1:
		return a.J1()
	case -1:
		return a.J1().Neg()
	}

	neg := false
	if n < 0 {
		n = -n
		neg = !neg
	}
	x := a
	if x.Signbit() {
		x = x.Neg()
		neg = !neg
	}

	var y Float256
	switch {
	case x.IsInf(0):
		y = Float256{}
	case x.IsZero():
		y = Float256{}
	default:
		var (
			Two = Float256{
				0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
				0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
			}
			Threshold500 = Float256{
				0x4000_7f40_0000_0000, 0x0000_0000_0000_0000,
				0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
			}
		)
		nn := NewFloat256(float64(n))
		threshold := Threshold500
		if sq := Two.Mul(nn).Mul(nn); sq.Gt(threshold) {
			threshold = sq
		}
		if x.Ge(threshold) {
			y = jnAsymptotic256(n, x)
		} else {
			y = jnMiller256(n, x)
		}
	}
	if neg && n%2 == 1 {
		y = y.Neg()
	}
	return y
}

// jnMiller256 returns Jn(x) for n >= 2, 0 < x < max(500, 2*n**2) using the
// same backward recurrence as jnMiller128/jnMiller256: it is unconditionally
// stable, unlike the forward recurrence J[k+1] = (2k/x)*J[k] - J[k-1], which
// only stays accurate while k < x.
func jnMiller256(n int, x Float256) Float256 {
	var (
		Zero = Float256{}
		One  = Float256(uvone256)
		Two  = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// RescaleThreshold bounds the unnormalized trial values below
		// Float256's overflow point. Deep descents (large n and x both
		// pushing the margin up) can otherwise grow the trial sequence
		// past Float256's range long before normalization brings it back
		// down to Jn(x) itself.
		RescaleThreshold = Float256{
			0x419f_2cf6_c9c9_bc5f, 0x884a_294e_53ed_c955,
			0xf57d_1efa_7827_1816, 0xaafd_3571_b9c9_7763,
		} // 1e2000
	)

	xCeil := x.Ceil().Int64()
	m := xCeil
	if int64(n) > m {
		m = int64(n)
	}
	// The margin below controls the accuracy: the backward recurrence's
	// sensitivity to the arbitrary starting value decays the further the
	// starting order is past max(n, x), so a bigger margin means more
	// correct digits. Near the Miller/asymptotic crossover (x close to
	// max(500, 2*n**2)), a fixed margin like J0/J1 use is not enough once n
	// is large: the required margin grows with x there too, so take the
	// larger of a fixed n-scaled margin and half of x.
	margin := int64(2*n) + 400
	if half := xCeil / 2; half > margin {
		margin = half
	}
	m += margin
	if m%2 != 0 {
		m++
	}

	Jkp1 := Zero
	Jk := One
	sum := Zero
	var target Float256
	for k := m; k >= 1; k-- {
		coef := Two.Mul(NewFloat256(float64(k))).Quo(x)
		Jkm1 := FMA256(coef, Jk, Jkp1.Neg())
		if k-1 == int64(n) {
			target = Jkm1
		}
		if (k-1)%2 == 0 {
			if k == 1 {
				sum = sum.Add(Jkm1)
			} else {
				sum = sum.Add(Jkm1).Add(Jkm1)
			}
		}
		Jkp1 = Jk
		Jk = Jkm1

		if Jk.Abs().Gt(RescaleThreshold) {
			inv := One.Quo(Jk)
			Jkp1 = Jkp1.Mul(inv)
			Jk = Jk.Mul(inv)
			sum = sum.Mul(inv)
			target = target.Mul(inv)
		}
	}
	return target.Quo(sum)
}

// jnAsymptotic256 returns Jn(x) for n >= 2, x >= max(500, 2*n**2) using
// Hankel's asymptotic expansion
//
//	Jn(x) ~ sqrt(2/(pi*x)) * (P(x)*cos(chi) - Q(x)*sin(chi)),  chi = x - n*pi/2 - pi/4
//
// Unlike J0/J1's hardcoded P/Q coefficients, the Hankel coefficients a_k(n)
// here are computed at runtime via their defining recurrence
//
//	a_0(n) = 1, a_k(n) = a_(k-1)(n) * (4*n**2 - (2k-1)**2) / (8k)
//
// since n is arbitrary; this is only reached where x is many times larger
// than n**2, so 40 terms leaves a comfortable margin beyond Float256's
// precision.
func jnAsymptotic256(n int, x Float256) Float256 {
	const M = 40

	var (
		One = Float256(uvone256)
		Two = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Four = Float256{
			0x4000_1000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Eight = Float256{
			0x4000_2000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Pi = Float256{
			0x4000_0921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
		InvSqrt2 = Float256{
			0x3fff_e6a0_9e66_7f3b, 0xcc90_8b2f_b136_6ea9,
			0x57d3_e3ad_ec17_5127, 0x7509_9da2_f590_b066,
		}
	)

	nu := NewFloat256(float64(n))
	fourNuSq := Four.Mul(nu).Mul(nu)

	a := make([]Float256, 2*M+2)
	a[0] = One
	for k := 1; k < len(a); k++ {
		kf := NewFloat256(float64(k))
		twokm1 := Two.Mul(kf).Sub(One)
		a[k] = a[k-1].Mul(fourNuSq.Sub(twokm1.Mul(twokm1))).Quo(Eight.Mul(kf))
	}

	xInv2 := One.Quo(x.Mul(x))
	P := Float256{}
	Q := Float256{}
	pTerm := One
	qTerm := One.Quo(x)
	sign := One
	for k := 0; k <= M; k++ {
		P = P.Add(sign.Mul(a[2*k]).Mul(pTerm))
		Q = Q.Add(sign.Mul(a[2*k+1]).Mul(qTerm))
		pTerm = pTerm.Mul(xInv2)
		qTerm = qTerm.Mul(xInv2)
		sign = sign.Neg()
	}

	// chi = x - n*pi/2 - pi/4. Computing this directly would subtract a
	// value of order n from x, which for large x throws away the low-order
	// bits Sincos needs to reduce the *fractional* part of x accurately
	// (Sincos(x) itself already reduces x mod 2*pi to full precision; a
	// prior subtraction of n*pi/2 here would just re-introduce the very
	// rounding that costs digits). Instead get cos(x-pi/4) and sin(x-pi/4)
	// from Sincos(x) via the sqrt(2) identity, then rotate by n*pi/2 using
	// the fact that cos/sin of a multiple of pi/2 is exactly 0, 1, or -1.
	sinX, cosX := x.Sincos()
	cosXm4 := InvSqrt2.Mul(cosX.Add(sinX))
	sinXm4 := InvSqrt2.Mul(sinX.Sub(cosX))
	var cosChi, sinChi Float256
	switch ((n % 4) + 4) % 4 {
	case 0:
		cosChi, sinChi = cosXm4, sinXm4
	case 1:
		cosChi, sinChi = sinXm4, cosXm4.Neg()
	case 2:
		cosChi, sinChi = cosXm4.Neg(), sinXm4.Neg()
	default: // 3
		cosChi, sinChi = sinXm4.Neg(), cosXm4
	}
	amp := Two.Quo(Pi.Mul(x)).Sqrt()
	return amp.Mul(P.Mul(cosChi).Sub(Q.Mul(sinChi)))
}
