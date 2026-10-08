package floats

// Y0 returns the order-zero Bessel function of the second kind.
//
// Special cases are:
//
//	Y0(+Inf) = 0
//	Y0(0) = -Inf
//	Y0(x < 0) = NaN
//	Y0(NaN) = NaN
func (a Float128) Y0() Float128 {
	switch {
	case a.IsNaN():
		return a
	case a.IsInf(1):
		return Float128{}
	case a.IsZero():
		return NewFloat128Inf(-1)
	case a.Signbit():
		return NewFloat128NaN()
	}

	var (
		Two         = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
		Threshold64 = Float128{0x4005_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	switch {
	case a.Lt(Two):
		return y0Series128(a)
	case a.Lt(Threshold64):
		return y0Taylor128(a)
	default:
		return y0Asymptotic128(a)
	}
}

// y0Series128 returns Y0(x) for 0 < x < 2 using the power series
//
//	Y0(x) = 2/pi * ((ln(x/2) + EulerGamma) * J0(x) + sum (-1)**(k+1) H(k) z**k / (k!)**2),  z = x**2/4,
//
// where H(k) is the k-th harmonic number. Both sums converge quickly and
// without significant cancellation for z < 1.
func y0Series128(x Float128) Float128 {
	var (
		Two     = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
		Quarter = Float128{0x3ffd_0000_0000_0000, 0x0000_0000_0000_0000}
		Euler   = Float128{0x3ffe_2788_cfc6_fb61, 0x8f49_a37c_7f02_02a6}
	)

	z := x.Mul(x).Mul(Quarter)
	j := y0128SeriesJ[len(y0128SeriesJ)-1]
	s := y0128SeriesS[len(y0128SeriesS)-1]
	for k := len(y0128SeriesJ) - 2; k >= 0; k-- {
		j = FMA128(j, z, y0128SeriesJ[k])
		s = FMA128(s, z, y0128SeriesS[k])
	}
	l := x.Quo(Two).Log().Add(Euler)
	return FMA128(l, j, s).Mul(y0128TwoOverPi)
}

// y0Taylor128 returns Y0(x) for 2 <= x < 64 using the Taylor series at the
// nearest center x0 of the table y0128Taylor. The coefficients c_n of
// Y0(x0+h) = sum c_n h**n are given by Y0(x0) and Y0'(x0) = -Y1(x0) and the
// recurrence derived from the differential equation x*y” + y' + x*y = 0:
//
//	x0 (n+2)(n+1) c_(n+2) + (n+1)**2 c_(n+1) + x0 c_n + c_(n-1) = 0
func y0Taylor128(x Float128) Float128 {
	// The centers are 2.25, 2.75, ..., 7.75 (step 1/2) and 8.5, 9.5, ..., 63.5 (step 1).
	xf := x.Float64().BuiltIn()
	var i int
	var x0 Float128
	if xf < 8 {
		i = int((xf - 2) * 2)
		x0 = NewFloat128(2.25 + 0.5*float64(i))
	} else {
		k := min(int(xf)-8, 55) // x < 64, but xf may round up to 64
		i = 12 + k
		x0 = NewFloat128(8.5 + float64(k))
	}
	terms := int(y0128TaylorTerms[i])
	h := x.Sub(x0)
	invX0 := Float128(uvone128).Quo(x0)

	var c [40]Float128
	c[0], c[1] = y0128Taylor[i][0], y0128Taylor[i][1]
	// n = 0: c_(-1) = 0.
	c[2] = c[1].Mul(invX0).Add(c[0]).Quo(Float128{0x4000_0000_0000_0000, 0}).Neg()
	for n := 1; n+2 <= terms; n++ {
		t := FMA128(NewFloat128(float64((n+1)*(n+1))), c[n+1], c[n-1])
		t = FMA128(t, invX0, c[n])
		c[n+2] = t.Quo(NewFloat128(float64((n + 2) * (n + 1)))).Neg()
	}

	r := c[terms]
	for n := terms - 1; n >= 0; n-- {
		r = FMA128(r, h, c[n])
	}
	return r
}

// temmeY0Y1_128 returns Y0(x) and Y1(x) for 0 < x < 2 using Temme's series,
// which is built to remain accurate through x's log singularity at 0
// (unlike the textbook power series for Y0/Y1, which shares J0's power
// series' cancellation problem: see j0Miller128).
//
// This is the nu=0 specialization of Temme's general-order series (Numerical
// Recipes §6.7): several of the general recurrences collapse at nu=0 (e.g.
// the p/q sequences coincide, and the term that is singular as nu -> 0
// vanishes identically rather than needing a limit), leaving
//
//	p_0 = q_0 = 1/pi
//	f_0 = -(2/pi)*(EulerGamma + ln(x/2))
//	p_k = p_(k-1)/k                        (k >= 1)
//	f_k = (k*f_(k-1) + 2*p_(k-1)) / k**2    (k >= 1)
//	c_k = (-x**2/4)**k / k!
//	h_k = -k*f_k + p_k
//
//	Y0(x) = -sum(c_k * f_k)
//	Y1(x) = -(2/x) * sum(c_k * h_k)
func temmeY0Y1_128(x Float128) (y0, y1 Float128) {
	var (
		One   = Float128(uvone128)
		Two   = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
		Four  = Float128{0x4001_0000_0000_0000, 0x0000_0000_0000_0000}
		Pi    = Float128{0x4000_921f_b544_42d1, 0x8469_898c_c517_01b8}
		Euler = Float128{0x3ffe_2788_cfc6_fb61, 0x8f49_a37c_7f02_02a6}
	)

	negQuarterX2 := x.Mul(x).Quo(Four).Neg()

	p := One.Quo(Pi)
	f := Two.Quo(Pi).Mul(Euler.Add(x.Quo(Two).Log())).Neg()
	c := One

	sumG := c.Mul(f)
	sumH := c.Mul(p)

	const N = 40
	for k := 1; k <= N; k++ {
		kf := NewFloat128(float64(k))
		pPrev, fPrev := p, f
		p = pPrev.Quo(kf)
		f = kf.Mul(fPrev).Add(Two.Mul(pPrev)).Quo(kf.Mul(kf))
		c = c.Mul(negQuarterX2).Quo(kf)
		h := kf.Mul(f).Neg().Add(p)
		sumG = sumG.Add(c.Mul(f))
		sumH = sumH.Add(c.Mul(h))
	}

	y0 = sumG.Neg()
	y1 = Two.Quo(x).Mul(sumH).Neg()
	return y0, y1
}

// y0Asymptotic128 returns Y0(x) for x >= 64 using Hankel's asymptotic
// expansion
//
//	Y0(x) ~ sqrt(2/(pi*x)) * (P(x)*sin(x-pi/4) + Q(x)*cos(x-pi/4))
//	      = sqrt(1/(pi*x)) * ((P(x)+Q(x))*sin(x) + (Q(x)-P(x))*cos(x))
//
// using the Hankel coefficients tabulated in y0128_table.go.
func y0Asymptotic128(x Float128) Float128 {
	var (
		One = Float128(uvone128)
		Pi  = Float128{0x4000_921f_b544_42d1, 0x8469_898c_c517_01b8}
	)

	w := One.Quo(x.Mul(x))

	// the number of the terms that makes the truncation error less than 2^-125,
	// where x is at least the lower bound of each band.
	terms := len(y0128HankelP) // 64 <= x
	switch xf := x.Float64().BuiltIn(); {
	case xf >= 4096:
		terms = 6
	case xf >= 512:
		terms = 9
	case xf >= 128:
		terms = 13
	}

	p := y0128HankelP[terms-1]
	q := y0128HankelQ[terms-1]
	for k := terms - 2; k >= 0; k-- {
		p = FMA128(p, w, y0128HankelP[k])
		q = FMA128(q, w, y0128HankelQ[k])
	}
	q = q.Quo(x)

	sin, cos := x.Sincos()
	amp := One.Quo(Pi.Mul(x)).Sqrt()
	return amp.Mul(p.Add(q).Mul(sin).Add(q.Sub(p).Mul(cos)))
}
