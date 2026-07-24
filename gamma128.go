package floats

// Gamma returns the Gamma function of a.
//
// Special cases are:
//
//	+Inf.Gamma() = +Inf
//	+0.Gamma() = +Inf
//	-0.Gamma() = -Inf
//	x.Gamma() = NaN for integer x < 0
//	-Inf.Gamma() = NaN
//	NaN.Gamma() = NaN
func (a Float128) Gamma() Float128 {
	var (
		// Zero is 0
		Zero = Float128{}

		// Half is 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}

		// One is 1
		One = Float128(uvone128)

		// Two is 2
		Two = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}

		// Three is 3
		Three = Float128{0x4000_8000_0000_0000, 0x0000_0000_0000_0000}

		// Pi is π
		Pi = Float128{0x4000_921f_b544_42d1, 0x8469_898c_c517_01b8}

		// StirlingThreshold is the threshold for using Stirling's approximation
		StirlingThreshold = Float128{0x4004_b800_0000_0000, 0x0000_0000_0000_0000} // 55
	)

	// special cases
	switch {
	case isNegInt128(a) || a.IsInf(-1) || a.IsNaN():
		return NewFloat128NaN()
	case a.IsInf(1):
		return NewFloat128Inf(1)
	case a.IsZero():
		return NewFloat128Inf(1).Copysign(a)
	}

	q := a.Abs()
	p := q.Floor()
	if q.Ge(StirlingThreshold) {
		if !a.Signbit() {
			y1, y2 := stirling128(a)
			return y1.Mul(y2)
		}
		// Note: x is negative but (checked above) not a negative integer,
		// so x must be small enough to be in range for conversion to int128.
		// If |x| were >= 2¹²⁷ it would have to be an integer.
		sign := !isOddInt128(p)
		z := q.Sub(p)
		if z.Gt(Half) {
			p = p.Add(One)
			z = q.Sub(p)
		}
		z = q.Mul(Pi.Mul(z).Sin())
		if z.IsZero() {
			return NewFloat128Inf(1)
		}
		sq1, sq2 := stirling128(q)
		absz := z.Abs()
		d := absz.Mul(sq1).Mul(sq2)
		if d.IsInf(0) {
			z = Pi.Quo(absz).Quo(sq1).Quo(sq2)
		} else {
			z = Pi.Quo(d)
		}
		if sign {
			z = z.Neg()
		}
		return z
	}

	// Reduce argument
	z := One
	for a.Ge(Three) {
		a = a.Sub(One)
		z = z.Mul(a)
	}
	for a.Lt(Zero) {
		z = z.Quo(a)
		a = a.Add(One)
	}
	for a.Lt(Two) {
		z = z.Quo(a)
		a = a.Add(One)
	}
	if a.Eq(Two) {
		return z
	}
	// Now 2 < a < 3.
	return z.Mul(gammaRational128(a))
}

// gammaRational128 returns a degree-16/16 Padé approximant of Gamma(a)
// for 2 < a < 3, expanded around the midpoint a = 2.5.
func gammaRational128(a Float128) Float128 {
	var (
		// Center = 2.5
		Center = Float128{0x4000_4000_0000_0000, 0x0000_0000_0000_0000}

		P0  = Float128{0x3fff_544f_a6d4_7b38, 0xffdb_52c4_3d8a_064b}
		P1  = Float128{0x3fff_4333_5a24_871b, 0xdc22_67fb_74b3_6bbb}
		P2  = Float128{0x3ffe_4b16_9178_daf2, 0x3461_eaff_8620_ce1d}
		P3  = Float128{0x3ffc_d481_9138_270a, 0x788c_bd4c_4764_ed98}
		P4  = Float128{0x3ffa_fb77_43e5_566c, 0xbe39_a38b_fee6_179b}
		P5  = Float128{0x3ff8_bb1b_7fb3_c9b2, 0xde55_01ab_f8b4_481b}
		P6  = Float128{0x3ff6_41d2_689a_9899, 0x9647_a8cb_3c6a_d721}
		P7  = Float128{0x3ff3_8c1e_4193_cc17, 0x55a3_c14d_cc41_d4fe}
		P8  = Float128{0x3ff0_a1db_5d4f_ffdc, 0x84ae_76ba_6f61_7529}
		P9  = Float128{0x3fed_7ba5_6653_ab0a, 0x1d35_e3f8_447b_e664}
		P10 = Float128{0x3fea_2922_fd70_3e16, 0x9ed6_7983_f9d7_4eae}
		P11 = Float128{0x3fe6_8e3e_82b0_b8e3, 0xfab7_d055_b889_1da9}
		P12 = Float128{0x3fe2_c358_0617_618d, 0xfba7_23f7_5061_c8e5}
		P13 = Float128{0x3fde_a5ea_8b25_5679, 0x114e_6df3_46e3_4372}
		P14 = Float128{0x3fda_37fd_8074_83fd, 0xaa3b_fba9_816c_0174}
		P15 = Float128{0x3fd5_4c91_d1a9_4595, 0xbc80_a42c_bc24_f4c9}
		P16 = Float128{0x3fcf_a015_628f_479a, 0xac05_d27a_bb9d_b472}

		Q0  = Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}
		Q1  = Float128{0x3ffc_f8f6_abcc_70de, 0x725a_9f2e_df59_bd0d}
		Q2  = Float128{0xbffc_6f3e_10a2_5a61, 0x4511_7432_c1ee_008b}
		Q3  = Float128{0xbff8_d182_4470_6483, 0x796e_64ae_d3d8_8005}
		Q4  = Float128{0x3ff8_f525_5e25_320d, 0x2d98_33aa_6b98_1712}
		Q5  = Float128{0xbff5_0acc_7ea6_55f3, 0x6590_fcdb_60f2_f7b6}
		Q6  = Float128{0xbff4_1a81_2021_5257, 0x72c7_e221_1ea5_95f1}
		Q7  = Float128{0x3ff1_b5f1_aab2_c47d, 0x1508_e6cc_27cf_217c}
		Q8  = Float128{0x3feb_45af_71fb_1b11, 0x6502_2962_c97e_7bac}
		Q9  = Float128{0xbfec_47bf_047c_71ad, 0x66af_d549_869d_d48b}
		Q10 = Float128{0x3fe9_3213_0da9_9bdd, 0x8082_107c_a1c9_ade2}
		Q11 = Float128{0xbfe0_3000_81a8_df9c, 0x4165_a6c8_0159_4e88}
		Q12 = Float128{0xbfe2_a9d5_c682_acc4, 0x327f_88c2_a0bf_4597}
		Q13 = Float128{0x3fdf_8d19_9899_89cc, 0x26a5_dd00_4c74_c891}
		Q14 = Float128{0xbfdb_6718_0552_5cda, 0xff41_a65b_9fc0_f190}
		Q15 = Float128{0x3fd6_4e8c_b37b_c516, 0x9c0c_c5c8_e952_b5ab}
		Q16 = Float128{0xbfcf_eeaf_366d_8e29, 0xa81c_befc_d2ae_9923}
	)

	t := a.Sub(Center)

	p := P16
	p = FMA128(p, t, P15)
	p = FMA128(p, t, P14)
	p = FMA128(p, t, P13)
	p = FMA128(p, t, P12)
	p = FMA128(p, t, P11)
	p = FMA128(p, t, P10)
	p = FMA128(p, t, P9)
	p = FMA128(p, t, P8)
	p = FMA128(p, t, P7)
	p = FMA128(p, t, P6)
	p = FMA128(p, t, P5)
	p = FMA128(p, t, P4)
	p = FMA128(p, t, P3)
	p = FMA128(p, t, P2)
	p = FMA128(p, t, P1)
	p = FMA128(p, t, P0)

	q := Q16
	q = FMA128(q, t, Q15)
	q = FMA128(q, t, Q14)
	q = FMA128(q, t, Q13)
	q = FMA128(q, t, Q12)
	q = FMA128(q, t, Q11)
	q = FMA128(q, t, Q10)
	q = FMA128(q, t, Q9)
	q = FMA128(q, t, Q8)
	q = FMA128(q, t, Q7)
	q = FMA128(q, t, Q6)
	q = FMA128(q, t, Q5)
	q = FMA128(q, t, Q4)
	q = FMA128(q, t, Q3)
	q = FMA128(q, t, Q2)
	q = FMA128(q, t, Q1)
	q = FMA128(q, t, Q0)

	return p.Quo(q)
}

func isNegInt128(x Float128) bool {
	if x.Lt(Float128{}) {
		_, xf := x.Modf()
		return xf.IsZero()
	}
	return false
}

// Stirling's approximation
func stirling128(x Float128) (Float128, Float128) {
	var (
		// One is 1
		One = Float128(uvone128)

		// Half is 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}

		// MaxStirling
		MaxStirling = Float128{0x4009_b700_0000_0000, 0x0000_0000_0000_0000} // 1756

		// Threshold
		Threshold = Float128{0x4008_f400_0000_0000, 0x0000_0000_0000_0000} // 1000

		// SqrtTwoPi is sqrt(2*pi)
		SqrtTwoPi = Float128{0x4000_40d9_31ff_6270, 0x5965_7ca4_1fae_722d}
	)

	// https://oeis.org/A001163 and https://oeis.org/A001164
	var (
		// P1 = 1/12
		P1 = Float128{0x3ffb_5555_5555_5555, 0x5555_5555_5555_5555}

		// P2 = 1/288
		P2 = Float128{0x3ff6_c71c_71c7_1c71, 0xc71c_71c7_1c71_c71c}

		// P3 = -139/51840
		P3 = Float128{0xbff6_5f72_68ed_ab4c, 0x7be4_300a_1d13_9856}

		// P4 = -571/2488320
		P4 = Float128{0xbff2_e13c_e465_fa85, 0x9562_d16f_75c7_f433}

		// P5 = 163879/209018880
		P5 = Float128{0x3ff4_9b0f_f687_4f2c, 0x41c7_458a_7842_6162}

		// P6 = 5246819/75246796800
		P6 = Float128{0x3ff1_2476_0483_9c03, 0x87e4_c67f_8930_f8c2}

		// P7 = -534703531/902961561600
		P7 = Float128{0xbff4_3677_3bdb_97b4, 0x78ba_4879_f1ee_63fc}

		// P8 = -4483131259/86684309913600
		P8 = Float128{0xbff0_b1d7_5d33_4671, 0x1786_7695_efec_19f8}

		// P9 = 432261921612371/514904800886784000
		P9 = Float128{0x3ff4_b823_9c67_0e69, 0x0242_d838_9578_76ac}

		// P10 = 6232523202521089/86504006548979712000
		P10 = Float128{0x3ff1_2e31_f9b7_913e, 0x9c4c_4f62_d774_84cb}

		// P11 = -25834629665134204969/13494625021640835072000
		P11 = Float128{0xbff5_f5db_caf7_56cd, 0xd9fa_a8e4_e10b_cd80}

		// P12 = -1579029138854919086429/9716130015581401251840000
		P12 = Float128{0xbff2_54d2_4114_4693, 0xe810_4e4b_7424_06f0}

		// P13 = 746590869962651602203151/116593560186976815022080000
		P13 = Float128{0x3ff7_a3a6_99f4_a401, 0xb3f4_2bfb_a624_8fc2}

		// P14 = 1511513601028097903631961/2798245444487443560529920000
		P14 = Float128{0x3ff4_1b33_b019_b3e6, 0xee99_66a2_4aff_0bb9}

		// P15 = -8849272268392873147705987190261/299692087104605205332754432000000
		P15 = Float128{0xbff9_e3c8_e8be_d86b, 0xb66a_b233_673a_fc61}

		// P16 = -142801712490607530608130701097701/57540880724084199423888850944000000
		P16 = Float128{0xbff6_4549_7f33_4cd1, 0xd29b_bf20_7863_2a4b}
	)

	if x.Ge(MaxStirling) {
		return NewFloat128Inf(1), One
	}

	x = x.Sub(One)
	w := One.Quo(x)
	z := P16
	z = FMA128(z, w, P15)
	z = FMA128(z, w, P14)
	z = FMA128(z, w, P13)
	z = FMA128(z, w, P12)
	z = FMA128(z, w, P11)
	z = FMA128(z, w, P10)
	z = FMA128(z, w, P9)
	z = FMA128(z, w, P8)
	z = FMA128(z, w, P7)
	z = FMA128(z, w, P6)
	z = FMA128(z, w, P5)
	z = FMA128(z, w, P4)
	z = FMA128(z, w, P3)
	z = FMA128(z, w, P2)
	z = FMA128(z, w, P1)
	z = FMA128(z, w, One)

	y1 := x.Exp()
	y2 := One
	if x.Ge(Threshold) { // avoid Pow() overflow
		v := x.Pow(x.Add(Half).Mul(Half))
		y1, y2 = v, v.Quo(y1)
	} else {
		y1 = x.Pow(x.Add(Half)).Quo(y1)
	}

	return y1, y2.Mul(z).Mul(SqrtTwoPi)
}
