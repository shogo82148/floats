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
		Two          = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
		Threshold500 = Float128{0x4007_f400_0000_0000, 0x0000_0000_0000_0000}
	)

	switch {
	case a.Lt(Two):
		_, y1 := temmeY0Y1_128(a)
		return y1
	case a.Lt(Threshold500):
		return y1CF2_128(a)
	default:
		return y1Asymptotic128(a)
	}
}

// y1CF2_128 returns Y1(x) for 2 <= x < 500. It reuses the same CF2
// continued fraction and J0(x)/J1(x) as y0CF2_128 to get Y0(x) = J0(x)*gam,
// then gets Y1 directly from Steed's derivative relation
//
//	Y0'(x) = Y0(x)*(p + q/gam) = -Y1(x)
//
// rather than the Wronskian J1(x)*Y0(x) - J0(x)*Y1(x) = 2/(pi*x), which
// would require subtracting two comparable-magnitude quantities.
func y1CF2_128(x Float128) Float128 {
	var (
		Zero = Float128{}
		One  = Float128(uvone128)
		Two  = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
		Four = Float128{0x4001_0000_0000_0000, 0x0000_0000_0000_0000}
	)

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
	y0 := j0.Mul(gam)
	return y0.Mul(p.Add(q.Quo(gam))).Neg()
}

// y1Asymptotic128 returns Y1(x) for x >= 500 using Hankel's asymptotic
// expansion
//
//	Y1(x) ~ sqrt(2/(pi*x)) * (P(x)*sin(x-3pi/4) + Q(x)*cos(x-3pi/4))
//	      = sqrt(1/(pi*x)) * ((Q(x)-P(x))*sin(x) - (P(x)+Q(x))*cos(x))
//
// using the same P/Q Hankel coefficients as j1Asymptotic128 (they depend
// only on the order, which is 1 for both).
func y1Asymptotic128(x Float128) Float128 {
	var (
		One = Float128(uvone128)
		Pi  = Float128{0x4000_921f_b544_42d1, 0x8469_898c_c517_01b8}

		P0  = Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}
		P1  = Float128{0x3ffb_e000_0000_0000, 0x0000_0000_0000_0000}
		P2  = Float128{0xbffc_2750_0000_0000, 0x0000_0000_0000_0000}
		P3  = Float128{0x3ffe_5a6a_5800_0000, 0x0000_0000_0000_0000}
		P4  = Float128{0xc001_b892_0d26_8000, 0x0000_0000_0000_0000}
		P5  = Float128{0x4005_e664_3dc4_a110, 0x0000_0000_0000_0000}
		P6  = Float128{0xc00a_9cc8_b6a2_ea44, 0x9080_0000_0000_0000}
		P7  = Float128{0x400f_f299_45cc_23c3, 0x4f12_1800_0000_0000}
		P8  = Float128{0xc015_9645_bee0_11be, 0x6811_3c15_e800_0000}
		P9  = Float128{0x401b_ad6b_4c84_e170, 0xe388_2c29_a31d_4000}
		P10 = Float128{0xc022_1da5_076c_edb1, 0x5094_f550_610e_709d}
		P11 = Float128{0x4028_d30a_1b77_ee99, 0x4e2c_a2fa_9687_4877}
		P12 = Float128{0xc02f_cc41_8acc_d750, 0x9da0_5b5b_6042_a025}
		P13 = Float128{0x4037_0d08_8ba5_da24, 0x2fb2_38ac_220f_0d33}
		P14 = Float128{0xc03e_701c_3f7b_d61c, 0xb1e7_d572_dad8_1cfa}
		P15 = Float128{0x4046_2361_04ee_d104, 0x5878_2b03_b381_a95d}

		Q0  = Float128{0x3ffd_8000_0000_0000, 0x0000_0000_0000_0000}
		Q1  = Float128{0xbffb_a400_0000_0000, 0x0000_0000_0000_0000}
		Q2  = Float128{0x3ffd_1c3d_0000_0000, 0x0000_0000_0000_0000}
		Q3  = Float128{0xbfff_fe58_1880_0000, 0x0000_0000_0000_0000}
		Q4  = Float128{0x4003_b3fb_3258_c400, 0x0000_0000_0000_0000}
		Q5  = Float128{0xc008_2dec_0ab4_99cb, 0xc000_0000_0000_0000}
		Q6  = Float128{0x400d_3419_80ef_2329, 0xf202_0000_0000_0000}
		Q7  = Float128{0xc012_b2b7_3c0d_fbfb, 0x15bd_6020_0000_0000}
		Q8  = Float128{0x4018_9526_f70e_0a2a, 0x5a22_20dc_6f20_0000}
		Q9  = Float128{0xc01e_e209_7fca_372b, 0x9041_1192_f319_e300}
		Q10 = Float128{0x4025_646b_0f8d_0f71, 0x3f67_9461_6ff6_0485}
		Q11 = Float128{0xc02c_409d_1cc5_06d6, 0x48e1_f251_113f_7a5b}
		Q12 = Float128{0x4033_58c2_b1f8_519a, 0xb4ca_c1df_3f10_a070}
		Q13 = Float128{0xc03a_b4b6_50e8_5536, 0x4ba5_4ce2_1582_35df}
		Q14 = Float128{0x4042_41cc_8e64_bd50, 0x1e8e_dfc4_7961_79e2}
		Q15 = Float128{0xc04a_10f2_8f44_18d3, 0xe08c_f42a_1cbb_75af}
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
	return amp.Mul(q.Sub(p).Mul(sin).Sub(p.Add(q).Mul(cos)))
}
