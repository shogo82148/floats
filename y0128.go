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
		Two          = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
		Threshold500 = Float128{0x4007_f400_0000_0000, 0x0000_0000_0000_0000}
	)

	switch {
	case a.Lt(Two):
		y0, _ := temmeY0Y1_128(a)
		return y0
	case a.Lt(Threshold500):
		return y0CF2_128(a)
	default:
		return y0Asymptotic128(a)
	}
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

// y0CF2_128 returns Y0(x) for 2 <= x < 500. It uses J0(x) and J1(x)
// (already computed accurately via Miller's algorithm, see j0Miller128) plus
// Temme's CF2 continued fraction
//
//	p+iq = J0'(x)+iY0'(x))/(J0(x)+iY0(x)) = -1/(2x) + i + (i/x)*K(x)
//
// where K is the complex continued fraction with terms a_j=(2j-1)**2/4,
// b_j=2(x+j*i) (Numerical Recipes §6.7), evaluated here by backward
// truncation rather than the forward (modified Lentz) form, mirroring how
// j0Miller128 evaluates its own backward recurrence. Given p, q and
// f = J0'(x)/J0(x) = -J1(x)/J0(x), Y0(x) = J0(x) * (p-f)/q.
func y0CF2_128(x Float128) Float128 {
	var (
		Zero = Float128{}
		One  = Float128(uvone128)
		Two  = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
		Four = Float128{0x4001_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	// 400 backward terms keeps CF2's error many orders of magnitude below
	// Float128's precision even at x=2, the slowest-converging point in
	// this branch's range (larger x converges faster).
	const maxit = 400

	twoX := Two.Mul(x)
	kRe, kIm := Zero, Zero
	for j := maxit; j >= 1; j-- {
		aj := NewFloat128(float64((2*j - 1) * (2*j - 1))).Quo(Four)
		bIm := NewFloat128(float64(2 * j))

		denomRe := twoX.Add(kRe)
		denomIm := bIm.Add(kIm)
		denomSq := denomRe.Mul(denomRe).Add(denomIm.Mul(denomIm))

		kRe = aj.Mul(denomRe).Quo(denomSq)
		kIm = aj.Mul(denomIm).Quo(denomSq).Neg()
	}

	p := One.Neg().Quo(Two.Mul(x)).Sub(kIm.Quo(x))
	q := One.Add(kRe.Quo(x))

	j0 := x.J0()
	j1 := x.J1()
	f := j1.Quo(j0).Neg()
	gam := p.Sub(f).Quo(q)
	return j0.Mul(gam)
}

// y0Asymptotic128 returns Y0(x) for x >= 500 using Hankel's asymptotic
// expansion
//
//	Y0(x) ~ sqrt(2/(pi*x)) * (P(x)*sin(x-pi/4) + Q(x)*cos(x-pi/4))
//	      = sqrt(1/(pi*x)) * ((P(x)+Q(x))*sin(x) + (Q(x)-P(x))*cos(x))
//
// using the same P/Q Hankel coefficients as j0Asymptotic128 (they depend
// only on the order, which is 0 for both).
func y0Asymptotic128(x Float128) Float128 {
	var (
		One = Float128(uvone128)
		Pi  = Float128{0x4000_921f_b544_42d1, 0x8469_898c_c517_01b8}

		P0  = Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}
		P1  = Float128{0xbffb_2000_0000_0000, 0x0000_0000_0000_0000}
		P2  = Float128{0x3ffb_cb60_0000_0000, 0x0000_0000_0000_0000}
		P3  = Float128{0xbffe_251e_e800_0000, 0x0000_0000_0000_0000}
		P4  = Float128{0x4001_84bd_1aa9_8000, 0x0000_0000_0000_0000}
		P5  = Float128{0xc005_b811_8d37_ff70, 0x0000_0000_0000_0000}
		P6  = Float128{0x400a_7bc2_e577_2972, 0x4780_0000_0000_0000}
		P7  = Float128{0xc00f_d036_6d1f_2a1f, 0xc534_2800_0000_0000}
		P8  = Float128{0x4015_7da6_5df9_46f8, 0xaf56_0224_1800_0000}
		P9  = Float128{0xc01b_9635_1108_1386, 0x765e_379d_0214_c000}
		P10 = Float128{0x4022_0fb5_f454_e219, 0x0e36_4701_8807_7ddb}
		P11 = Float128{0xc028_be48_3c61_88f8, 0xe44c_c93f_185f_231c}
		P12 = Float128{0x402f_b978_561d_4bea, 0x0f5b_1e28_a03f_e7fa}
		P13 = Float128{0xc037_02e1_94de_62d0, 0xb524_4031_b682_6879}
		P14 = Float128{0x403e_6331_b684_f705, 0x3b61_f1e3_98c7_8c3e}
		P15 = Float128{0xc046_19d3_58b4_a032, 0x5df6_5392_44b3_f7bf}

		Q0  = Float128{0xbffc_0000_0000_0000, 0x0000_0000_0000_0000}
		Q1  = Float128{0x3ffb_2c00_0000_0000, 0x0000_0000_0000_0000}
		Q2  = Float128{0xbffc_d11e_0000_0000, 0x0000_0000_0000_0000}
		Q3  = Float128{0x3fff_ba4c_5980_0000, 0x0000_0000_0000_0000}
		Q4  = Float128{0xc003_8616_a64f_6c00, 0x0000_0000_0000_0000}
		Q5  = Float128{0x4008_13aa_fea4_e577, 0x4000_0000_0000_0000}
		Q6  = Float128{0xc00d_1d47_059b_0d98, 0x9db6_0000_0000_0000}
		Q7  = Float128{0x4012_96ab_69ba_805e, 0x7fb1_2860_0000_0000}
		Q8  = Float128{0xc018_7e00_2ac4_1836, 0x8f7f_438e_0260_0000}
		Q9  = Float128{0x401e_c951_3798_75fb, 0x6178_cf08_21b6_1900}
		Q10 = Float128{0xc025_53d7_328c_73ee, 0xf503_8192_7c9d_2dfb}
		Q11 = Float128{0x402c_32f8_7824_21c7, 0xb82a_08af_a906_4f05}
		Q12 = Float128{0xc033_4b3d_91e4_8aa3, 0xb2b8_c44e_f651_3ac6}
		Q13 = Float128{0x403a_a4d4_ec38_521d, 0x0c62_c7c2_970d_c434}
		Q14 = Float128{0xc042_36e3_feb8_1ab1, 0x3337_9718_eec1_f7ec}
		Q15 = Float128{0x404a_0848_51d4_388c, 0x2290_9b24_b63b_9a95}
	)

	w := One.Quo(x.Mul(x))

	p := P15
	p = FMA128(p, w, P14)
	p = FMA128(p, w, P13)
	p = FMA128(p, w, P12)
	p = FMA128(p, w, P11)
	p = FMA128(p, w, P10)
	p = FMA128(p, w, P9)
	p = FMA128(p, w, P8)
	p = FMA128(p, w, P7)
	p = FMA128(p, w, P6)
	p = FMA128(p, w, P5)
	p = FMA128(p, w, P4)
	p = FMA128(p, w, P3)
	p = FMA128(p, w, P2)
	p = FMA128(p, w, P1)
	p = FMA128(p, w, P0)

	q := Q15
	q = FMA128(q, w, Q14)
	q = FMA128(q, w, Q13)
	q = FMA128(q, w, Q12)
	q = FMA128(q, w, Q11)
	q = FMA128(q, w, Q10)
	q = FMA128(q, w, Q9)
	q = FMA128(q, w, Q8)
	q = FMA128(q, w, Q7)
	q = FMA128(q, w, Q6)
	q = FMA128(q, w, Q5)
	q = FMA128(q, w, Q4)
	q = FMA128(q, w, Q3)
	q = FMA128(q, w, Q2)
	q = FMA128(q, w, Q1)
	q = FMA128(q, w, Q0)
	q = q.Quo(x)

	sin, cos := x.Sincos()
	amp := One.Quo(Pi.Mul(x)).Sqrt()
	return amp.Mul(p.Add(q).Mul(sin).Add(q.Sub(p).Mul(cos)))
}
