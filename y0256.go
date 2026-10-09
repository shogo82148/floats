package floats

// Y0 returns the order-zero Bessel function of the second kind.
//
// Special cases are:
//
//	Y0(+Inf) = 0
//	Y0(0) = -Inf
//	Y0(x < 0) = NaN
//	Y0(NaN) = NaN
func (a Float256) Y0() Float256 {
	switch {
	case a.IsNaN():
		return a
	case a.IsInf(1):
		return Float256{}
	case a.IsZero():
		return NewFloat256Inf(-1)
	case a.Signbit():
		return NewFloat256NaN()
	}

	var (
		Two = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Threshold128 = Float256{
			0x4000_6000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)

	switch {
	case a.Lt(Two):
		return y0Series256(a)
	case a.Lt(Threshold128):
		return y0Taylor256(a)
	default:
		return y0Asymptotic256(a)
	}
}

// y0Series256 returns Y0(x) for 0 < x < 2 using the power series
//
//	Y0(x) = 2/pi * ((ln(x/2) + EulerGamma) * J0(x) + sum (-1)**(k+1) H(k) z**k / (k!)**2),  z = x**2/4,
//
// where H(k) is the k-th harmonic number; see y0Series128.
func y0Series256(x Float256) Float256 {
	var (
		Two = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Quarter = Float256{
			0x3fff_d000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Euler = Float256{
			0x3fff_e278_8cfc_6fb6, 0x18f4_9a37_c7f0_202a,
			0x596a_d439_d987_5ecb, 0x9803_2180_7be6_8e13,
		}
	)

	z := x.Mul(x).Mul(Quarter)
	j := y0256SeriesJ[len(y0256SeriesJ)-1]
	s := y0256SeriesS[len(y0256SeriesS)-1]
	for k := len(y0256SeriesJ) - 2; k >= 0; k-- {
		j = FMA256(j, z, y0256SeriesJ[k])
		s = FMA256(s, z, y0256SeriesS[k])
	}
	l := x.Quo(Two).Log().Add(Euler)
	return FMA256(l, j, s).Mul(y0256TwoOverPi)
}

// y0Taylor256 returns Y0(x) for 2 <= x < 128 using the Taylor series at the
// nearest center x0 of the table y0256Taylor; see y0Taylor128.
func y0Taylor256(x Float256) Float256 {
	// The centers are 2.25, 2.75, ..., 7.75 (step 1/2) and 8.5, 9.5, ..., 127.5 (step 1).
	xf := x.Float64().BuiltIn()
	var i int
	var x0 Float256
	if xf < 8 {
		i = int((xf - 2) * 2)
		x0 = NewFloat256(2.25 + 0.5*float64(i))
	} else {
		k := min(int(xf)-8, 119) // x < 128, but xf may round up to 128
		i = 12 + k
		x0 = NewFloat256(8.5 + float64(k))
	}
	terms := int(y0256TaylorTerms[i])
	h := x.Sub(x0)
	invX0 := Float256(uvone256).Quo(x0)

	var c [y0256MaxTaylorTerms + 1]Float256
	c[0], c[1] = y0256Taylor[i][0], y0256Taylor[i][1]
	// n = 0: c_(-1) = 0.
	c[2] = c[1].Mul(invX0).Add(c[0]).Quo(NewFloat256(2)).Neg()
	for n := 1; n+2 <= terms; n++ {
		t := FMA256(NewFloat256(float64((n+1)*(n+1))), c[n+1], c[n-1])
		t = FMA256(t, invX0, c[n])
		c[n+2] = t.Quo(NewFloat256(float64((n + 2) * (n + 1)))).Neg()
	}

	r := c[terms]
	for n := terms - 1; n >= 0; n-- {
		r = FMA256(r, h, c[n])
	}
	return r
}

// y0Asymptotic256 returns Y0(x) for x >= 128 using Hankel's asymptotic
// expansion
//
//	Y0(x) ~ sqrt(2/(pi*x)) * (P(x)*sin(x-pi/4) + Q(x)*cos(x-pi/4))
//	      = sqrt(1/(pi*x)) * ((P(x)+Q(x))*sin(x) + (Q(x)-P(x))*cos(x))
//
// using the Hankel coefficients tabulated in y0256_table.go.
func y0Asymptotic256(x Float256) Float256 {
	var (
		One = Float256(uvone256)
		Pi  = Float256{
			0x4000_0921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
	)

	w := One.Quo(x.Mul(x))

	// the number of the terms that makes the truncation error less than 2^-245,
	// where x is at least the lower bound of each band.
	terms := len(y0256HankelP) // 128 <= x
	switch xf := x.Float64().BuiltIn(); {
	case xf >= 1e5:
		terms = 9
	case xf >= 4096:
		terms = 13
	case xf >= 1024:
		terms = 17
	case xf >= 512:
		terms = 20
	case xf >= 256:
		terms = 26
	}

	p := y0256HankelP[terms-1]
	q := y0256HankelQ[terms-1]
	for k := terms - 2; k >= 0; k-- {
		p = FMA256(p, w, y0256HankelP[k])
		q = FMA256(q, w, y0256HankelQ[k])
	}
	q = q.Quo(x)

	sin, cos := x.Sincos()
	amp := One.Quo(Pi.Mul(x)).Sqrt()
	return amp.Mul(p.Add(q).Mul(sin).Add(q.Sub(p).Mul(cos)))
}
