package floats

// Y1 returns the order-one Bessel function of the second kind.
//
// Special cases are:
//
//	Y1(+Inf) = 0
//	Y1(0) = -Inf
//	Y1(x < 0) = NaN
//	Y1(NaN) = NaN
func (a Float128) Y1() Float128 {
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
		return y1Series128(a)
	case a.Lt(Threshold64):
		return y1Taylor128(a)
	default:
		return y1Asymptotic128(a)
	}
}

// y1Series128 returns Y1(x) for 0 < x < 2 using the power series
//
//	Y1(x) = x/pi * ((ln(x/2) + EulerGamma) * sum (-1)**k z**k / (k! (k+1)!)
//	        - 1/2 * sum (-1)**k (H(k) + H(k+1)) z**k / (k! (k+1)!)) - 2/(pi x),  z = x**2/4,
//
// where H(k) is the k-th harmonic number. Both sums converge quickly and
// without significant cancellation for z < 1.
func y1Series128(x Float128) Float128 {
	var (
		Two     = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
		Quarter = Float128{0x3ffd_0000_0000_0000, 0x0000_0000_0000_0000}
		Euler   = Float128{0x3ffe_2788_cfc6_fb61, 0x8f49_a37c_7f02_02a6}
	)

	z := x.Mul(x).Mul(Quarter)
	j := y1128SeriesJ[len(y1128SeriesJ)-1]
	s := y1128SeriesS[len(y1128SeriesS)-1]
	for k := len(y1128SeriesJ) - 2; k >= 0; k-- {
		j = FMA128(j, z, y1128SeriesJ[k])
		s = FMA128(s, z, y1128SeriesS[k])
	}
	l := x.Quo(Two).Log().Add(Euler)
	return FMA128(FMA128(l, j, s), x.Mul(y1128OneOverPi), y1128TwoOverPi.Quo(x).Neg())
}

// y1Taylor128 returns Y1(x) for 2 <= x < 64 using the Taylor series at the
// nearest center x0 of the table y1128Taylor; see y0Taylor128. The
// coefficients c_n of Y1(x0+h) = sum c_n h**n are given by Y1(x0) and Y1'(x0)
// and the recurrence derived from the differential equation
// x**2 y” + x y' + (x**2-1) y = 0:
//
//	x0**2 (n+2)(n+1) c_(n+2) + x0 (n+1)(2n+1) c_(n+1) + (n**2+x0**2-1) c_n + 2 x0 c_(n-1) + c_(n-2) = 0
func y1Taylor128(x Float128) Float128 {
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
	terms := int(y1128TaylorTerms[i])
	h := x.Sub(x0)
	inv := Float128(uvone128).Quo(x0)

	// c[n+2] is c_n. c_(-2) = c_(-1) = 0.
	var c [y1128MaxTaylorTerms + 3]Float128
	c[2], c[3] = y1128Taylor[i][0], y1128Taylor[i][1]
	for n := 0; n+2 <= terms; n++ {
		t1 := FMA128(NewFloat128(float64((n+1)*(2*n+1))), c[n+3], c[n+1].Add(c[n+1]))
		t2 := FMA128(NewFloat128(float64(n*n-1)), c[n+2], c[n])
		a := FMA128(FMA128(t2, inv, t1), inv, c[n+2])
		c[n+4] = a.Quo(NewFloat128(float64((n + 2) * (n + 1)))).Neg()
	}

	r := c[terms+2]
	for n := terms - 1; n >= 0; n-- {
		r = FMA128(r, h, c[n+2])
	}
	return r
}

// y1Asymptotic128 returns Y1(x) for x >= 64 using Hankel's asymptotic
// expansion
//
//	Y1(x) ~ sqrt(2/(pi*x)) * (P(x)*sin(x-3*pi/4) + Q(x)*cos(x-3*pi/4))
//	      = sqrt(1/(pi*x)) * ((Q(x)-P(x))*sin(x) - (P(x)+Q(x))*cos(x))
//
// using the Hankel coefficients tabulated in y1128_table.go.
func y1Asymptotic128(x Float128) Float128 {
	var (
		One = Float128(uvone128)
		Pi  = Float128{0x4000_921f_b544_42d1, 0x8469_898c_c517_01b8}
	)

	w := One.Quo(x.Mul(x))

	// the number of the terms that makes the truncation error less than 2^-125,
	// where x is at least the lower bound of each band.
	terms := len(y1128HankelP) // 64 <= x
	switch xf := x.Float64().BuiltIn(); {
	case xf >= 4096:
		terms = 6
	case xf >= 512:
		terms = 9
	case xf >= 128:
		terms = 13
	}

	p := y1128HankelP[terms-1]
	q := y1128HankelQ[terms-1]
	for k := terms - 2; k >= 0; k-- {
		p = FMA128(p, w, y1128HankelP[k])
		q = FMA128(q, w, y1128HankelQ[k])
	}
	q = q.Quo(x)

	sin, cos := x.Sincos()
	amp := One.Quo(Pi.Mul(x)).Sqrt()
	return amp.Mul(q.Sub(p).Mul(sin).Sub(p.Add(q).Mul(cos)))
}
