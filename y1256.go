package floats

// Y1 returns the order-one Bessel function of the second kind.
//
// Special cases are:
//
//	Y1(+Inf) = 0
//	Y1(0) = -Inf
//	Y1(x < 0) = NaN
//	Y1(NaN) = NaN
func (a Float256) Y1() Float256 {
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
		return y1Series256(a)
	case a.Lt(Threshold128):
		return y1Taylor256(a)
	default:
		return y1Asymptotic256(a)
	}
}

// y1Series256 returns Y1(x) for 0 < x < 2 using the power series
//
//	Y1(x) = x/pi * ((ln(x/2) + EulerGamma) * sum (-1)**k z**k / (k! (k+1)!)
//	        - 1/2 * sum (-1)**k (H(k) + H(k+1)) z**k / (k! (k+1)!)) - 2/(pi x),  z = x**2/4,
//
// where H(k) is the k-th harmonic number. Both sums converge quickly and
// without significant cancellation for z < 1.
func y1Series256(x Float256) Float256 {
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
	j := y1256SeriesJ[len(y1256SeriesJ)-1]
	s := y1256SeriesS[len(y1256SeriesS)-1]
	for k := len(y1256SeriesJ) - 2; k >= 0; k-- {
		j = FMA256(j, z, y1256SeriesJ[k])
		s = FMA256(s, z, y1256SeriesS[k])
	}
	l := x.Quo(Two).Log().Add(Euler)
	return FMA256(FMA256(l, j, s), x.Mul(y1256OneOverPi), y1256TwoOverPi.Quo(x).Neg())
}

// y1Taylor256 returns Y1(x) for 2 <= x < 128 using the Taylor series at the
// nearest center x0 of the table y1256Taylor; see y0Taylor256. The
// coefficients c_n of Y1(x0+h) = sum c_n h**n are given by Y1(x0) and Y1'(x0)
// and the recurrence derived from the differential equation
// x**2 y” + x y' + (x**2-1) y = 0:
//
//	x0**2 (n+2)(n+1) c_(n+2) + x0 (n+1)(2n+1) c_(n+1) + (n**2+x0**2-1) c_n + 2 x0 c_(n-1) + c_(n-2) = 0
func y1Taylor256(x Float256) Float256 {
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
	terms := int(y1256TaylorTerms[i])
	h := x.Sub(x0)
	inv := Float256(uvone256).Quo(x0)

	// c[n+2] is c_n. c_(-2) = c_(-1) = 0.
	var c [y1256MaxTaylorTerms + 3]Float256
	c[2], c[3] = y1256Taylor[i][0], y1256Taylor[i][1]
	for n := 0; n+2 <= terms; n++ {
		t1 := FMA256(NewFloat256(float64((n+1)*(2*n+1))), c[n+3], c[n+1].Add(c[n+1]))
		t2 := FMA256(NewFloat256(float64(n*n-1)), c[n+2], c[n])
		a := FMA256(FMA256(t2, inv, t1), inv, c[n+2])
		c[n+4] = a.Quo(NewFloat256(float64((n + 2) * (n + 1)))).Neg()
	}

	r := c[terms+2]
	for n := terms - 1; n >= 0; n-- {
		r = FMA256(r, h, c[n+2])
	}
	return r
}

// y1Asymptotic256 returns Y1(x) for x >= 128 using Hankel's asymptotic
// expansion
//
//	Y1(x) ~ sqrt(2/(pi*x)) * (P(x)*sin(x-3*pi/4) + Q(x)*cos(x-3*pi/4))
//	      = sqrt(1/(pi*x)) * ((Q(x)-P(x))*sin(x) - (P(x)+Q(x))*cos(x))
//
// using the Hankel coefficients tabulated in y1256_table.go.
func y1Asymptotic256(x Float256) Float256 {
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
	terms := len(y1256HankelP) // 128 <= x
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

	p := y1256HankelP[terms-1]
	q := y1256HankelQ[terms-1]
	for k := terms - 2; k >= 0; k-- {
		p = FMA256(p, w, y1256HankelP[k])
		q = FMA256(q, w, y1256HankelQ[k])
	}
	q = q.Quo(x)

	sin, cos := x.Sincos()
	amp := One.Quo(Pi.Mul(x)).Sqrt()
	return amp.Mul(q.Sub(p).Mul(sin).Sub(p.Add(q).Mul(cos)))
}
