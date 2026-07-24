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
		Threshold500 = Float256{
			0x4000_7f40_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)

	switch {
	case a.Lt(Two):
		y0, _ := temmeY0Y1_256(a)
		return y0
	case a.Lt(Threshold500):
		return y0CF2_256(a)
	default:
		return y0Asymptotic256(a)
	}
}

// temmeY0Y1_256 returns Y0(x) and Y1(x) for 0 < x < 2 using Temme's series;
// see temmeY0Y1_128 for the derivation (the nu=0 specialization of Numerical
// Recipes §6.7's general-order series).
func temmeY0Y1_256(x Float256) (y0, y1 Float256) {
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
		Pi = Float256{
			0x4000_0921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
		Euler = Float256{
			0x3fff_e278_8cfc_6fb6, 0x18f4_9a37_c7f0_202a,
			0x596a_d439_d987_5ecb, 0x9803_2180_7be6_8e13,
		}
	)

	negQuarterX2 := x.Mul(x).Quo(Four).Neg()

	p := One.Quo(Pi)
	f := Two.Quo(Pi).Mul(Euler.Add(x.Quo(Two).Log())).Neg()
	c := One

	sumG := c.Mul(f)
	sumH := c.Mul(p)

	const N = 40
	for k := 1; k <= N; k++ {
		kf := NewFloat256(float64(k))
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

// y0CF2_256 returns Y0(x) for 2 <= x < 500; see y0CF2_128 for the derivation.
func y0CF2_256(x Float256) Float256 {
	var (
		Zero = Float256{}
		One  = Float256(uvone256)
		Two  = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Four = Float256{
			0x4000_1000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)

	// Float256 needs more backward terms than Float128 for the same margin
	// below its (much deeper) precision at the slowest-converging point,
	// x=2.
	const maxit = 1200

	twoX := Two.Mul(x)
	kRe, kIm := Zero, Zero
	for j := maxit; j >= 1; j-- {
		aj := NewFloat256(float64((2*j - 1) * (2*j - 1))).Quo(Four)
		bIm := NewFloat256(float64(2 * j))

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

// y0Asymptotic256 returns Y0(x) for x >= 500 using Hankel's asymptotic
// expansion
//
//	Y0(x) ~ sqrt(2/(pi*x)) * (P(x)*sin(x-pi/4) + Q(x)*cos(x-pi/4))
//	      = sqrt(1/(pi*x)) * ((P(x)+Q(x))*sin(x) + (Q(x)-P(x))*cos(x))
//
// using the same P/Q Hankel coefficients as j0Asymptotic256 (they depend
// only on the order, which is 0 for both).
func y0Asymptotic256(x Float256) Float256 {
	var (
		One = Float256(uvone256)
		Pi  = Float256{
			0x4000_0921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}

		P0 = Float256{
			0x3fff_f000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P1 = Float256{
			0xbfff_b200_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P2 = Float256{
			0x3fff_bcb6_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P3 = Float256{
			0xbfff_e251_ee80_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P4 = Float256{
			0x4000_184b_d1aa_9800, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P5 = Float256{
			0xc000_5b81_18d3_7ff7, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P6 = Float256{
			0x4000_a7bc_2e57_7297, 0x2478_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P7 = Float256{
			0xc000_fd03_66d1_f2a1, 0xfc53_4280_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P8 = Float256{
			0x4001_57da_65df_946f, 0x8af5_6022_4180_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P9 = Float256{
			0xc001_b963_5110_8138, 0x6765_e379_d021_4c00,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P10 = Float256{
			0x4002_20fb_5f45_4e21, 0x90e3_6470_1880_77dd,
			0xa800_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P11 = Float256{
			0xc002_8be4_83c6_188f, 0x8e44_cc93_f185_f231,
			0xc7b6_c000_0000_0000, 0x0000_0000_0000_0000,
		}
		P12 = Float256{
			0x4002_fb97_8561_d4be, 0xa0f5_b1e2_8a03_fe7f,
			0xa0a6_8e05_c000_0000, 0x0000_0000_0000_0000,
		}
		P13 = Float256{
			0xc003_702e_194d_e62d, 0x0b52_4403_1b68_2687,
			0x95fc_2260_9a41_8000, 0x0000_0000_0000_0000,
		}
		P14 = Float256{
			0x4003_e633_1b68_4f70, 0x53b6_1f1e_398c_78c3,
			0xe0f6_e7c4_c590_57f7, 0x0000_0000_0000_0000,
		}
		P15 = Float256{
			0xc004_619d_358b_4a03, 0x25df_6539_244b_3f7b,
			0xe815_2b27_b81f_f4dd, 0xcfaa_0000_0000_0000,
		}
		P16 = Float256{
			0x4004_e001_693c_ab40, 0xb4a3_e3f5_542d_5217,
			0x9a9a_608c_34f6_ee5f, 0x222c_9cf8_c000_0000,
		}
		P17 = Float256{
			0xc005_6083_65b1_f0ab, 0x0b28_db08_409b_732b,
			0x4dde_e074_fb06_a83f, 0x05cf_035c_9175_8000,
		}
		P18 = Float256{
			0x4005_e332_b47a_bb4b, 0x97f8_bc29_82e9_87dc,
			0x7350_83b4_14f7_514d, 0x0f32_1ee8_e958_dead,
		}
		P19 = Float256{
			0xc006_68fb_4afd_7d74, 0xf885_3e8f_2302_7322,
			0xbde0_6b62_992f_019e, 0x4b50_c1cb_abc6_9e99,
		}
		P20 = Float256{
			0x4006_f215_5dc2_b3d9, 0x87d4_323c_260e_4fb9,
			0x1425_279d_bc53_523b, 0xb308_9215_d699_77c5,
		}
		P21 = Float256{
			0xc007_7cf8_6d19_2a66, 0x662c_c892_8494_eaa7,
			0x7d26_26f5_1c0a_a9b9, 0x2380_bec0_b80d_87d3,
		}
		P22 = Float256{
			0x4008_098d_d73a_b7d7, 0x450c_805c_2df2_dc47,
			0x5b9e_1e28_7b93_c1ed, 0xe5b0_f3a8_413c_b13c,
		}
		P23 = Float256{
			0xc008_98b6_3195_9e3b, 0x421b_5776_67b5_298d,
			0x3d7a_3c86_ff40_f241, 0x3c3f_c901_124f_0823,
		}
		P24 = Float256{
			0x4009_2a17_d904_2c2d, 0x70b3_53b4_de11_f1d5,
			0x1580_f721_5698_b463, 0x2bf2_494d_d85a_819e,
		}
		P25 = Float256{
			0xc009_bdf8_fc64_41d2, 0x9093_b0d3_4fc8_89fa,
			0xc302_bb2b_2df5_e366, 0xf5b0_5531_7d1e_5063,
		}
		P26 = Float256{
			0x400a_52a9_cd4f_3cb1, 0x6e1b_9a5c_a7b1_43a2,
			0x8413_6182_56b0_4b9a, 0xc5e4_591f_860b_926f,
		}
		P27 = Float256{
			0xc00a_e91e_a5e1_a7ba, 0x3f07_557f_4cdb_8d60,
			0xce6f_104c_161c_aa81, 0x82c7_938b_8efb_7e60,
		}
		P28 = Float256{
			0x400b_8237_9e68_5ace, 0xd31c_bcb5_3b07_5c02,
			0x3074_7c53_dd60_0b53, 0x06c6_313a_9226_63f9,
		}
		P29 = Float256{
			0xc00c_1c65_ce5a_da55, 0xa996_684a_4447_c0cd,
			0xf91a_1e68_4cf9_3db5, 0x67fe_9427_6457_631e,
		}
		P30 = Float256{
			0x400c_b7ba_6d2d_5902, 0x3ceb_e102_ae04_1409,
			0xf7f1_2f16_7d7a_24fe, 0x6628_0487_cfdb_6ebe,
		}

		Q0 = Float256{
			0xbfff_c000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q1 = Float256{
			0x3fff_b2c0_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q2 = Float256{
			0xbfff_cd11_e000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q3 = Float256{
			0x3fff_fba4_c598_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q4 = Float256{
			0xc000_3861_6a64_f6c0, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q5 = Float256{
			0x4000_813a_afea_4e57, 0x7400_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q6 = Float256{
			0xc000_d1d4_7059_b0d9, 0x89db_6000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q7 = Float256{
			0x4001_296a_b69b_a805, 0xe7fb_1286_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q8 = Float256{
			0xc001_87e0_02ac_4183, 0x68f7_f438_e026_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q9 = Float256{
			0x4001_ec95_1379_875f, 0xb617_8cf0_821b_6190,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q10 = Float256{
			0xc002_553d_7328_c73e, 0xef50_3819_27c9_d2df,
			0xb620_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q11 = Float256{
			0x4002_c32f_8782_421c, 0x7b82_a08a_fa90_64f0,
			0x53e4_9280_0000_0000, 0x0000_0000_0000_0000,
		}
		Q12 = Float256{
			0xc003_34b3_d91e_48aa, 0x3b2b_8c44_ef65_13ac,
			0x6689_c4a4_2780_0000, 0x0000_0000_0000_0000,
		}
		Q13 = Float256{
			0x4003_aa4d_4ec3_8521, 0xd0c6_2c7c_2970_dc43,
			0x3867_a453_d991_e200, 0x0000_0000_0000_0000,
		}
		Q14 = Float256{
			0xc004_236e_3feb_81ab, 0x1333_7971_8eec_1f7e,
			0xc1cd_12ff_783b_5fc0, 0x6600_0000_0000_0000,
		}
		Q15 = Float256{
			0x4004_a084_851d_4388, 0xc229_09b2_4b63_b9a9,
			0x4a58_7efc_2667_c613, 0x5bd5_8c00_0000_0000,
		}
		Q16 = Float256{
			0xc005_2002_6190_9f6a, 0x1d69_0529_eeae_8b24,
			0x5637_2415_efe2_de4e, 0x8cc9_d5fa_8180_0000,
		}
		Q17 = Float256{
			0x4005_a18c_8d9d_d80c, 0xa10f_3f2b_af2c_83b5,
			0x0c86_bdf6_d50e_8f7c, 0x1e58_2b49_6727_7100,
		}
		Q18 = Float256{
			0xc006_259a_14b2_f6bf, 0x2aba_da35_6969_795b,
			0xe421_ad46_77ce_b788, 0xec9c_aaac_0544_7bda,
		}
		Q19 = Float256{
			0x4006_adab_b103_bca5, 0x7066_0c9b_f3c9_379e,
			0x56f3_d33b_d2c1_ba13, 0x7124_fdf7_48e3_d965,
		}
		Q20 = Float256{
			0xc007_369b_9706_0029, 0xdb6b_f7f9_57f7_a5e9,
			0xad22_f035_ee6b_181f, 0xce69_4ce0_dcec_a501,
		}
		Q21 = Float256{
			0x4007_c303_b412_1f33, 0x70f7_f38d_00bd_5e32,
			0xffe2_759d_aa85_6aca, 0xa753_44a4_0632_5e79,
		}
		Q22 = Float256{
			0xc008_5192_1558_2c7e, 0xd6be_6553_4ae1_f175,
			0xf420_e011_d5db_e290, 0xc0b5_dea5_88c4_6b8e,
		}
		Q23 = Float256{
			0x4008_e1c3_7a3d_1a87, 0x7d8a_8151_58f9_5a66,
			0x9215_ee71_f9fe_d5a1, 0x51a0_d461_d013_680a,
		}
		Q24 = Float256{
			0xc009_7392_6b15_ecf9, 0x2f93_3e85_70c6_3c45,
			0x66c3_1f3e_b342_a9c0, 0xffb7_eaf1_6e4d_8505,
		}
		Q25 = Float256{
			0x400a_076b_1ba1_b4c3, 0x20af_f196_fc5b_414c,
			0x1a72_092e_585c_73d0, 0x088d_2a68_2d3a_9b51,
		}
		Q26 = Float256{
			0xc00a_9e54_a1ec_2163, 0xd260_1aac_1669_66b5,
			0x8230_8dac_9ca7_c9f6, 0x8295_b402_0f8a_946e,
		}
		Q27 = Float256{
			0x400b_3532_50e2_23ce, 0x0354_9204_61da_a506,
			0x6ec6_68a4_9121_9b70, 0xaa72_0b07_6abe_11f7,
		}
		Q28 = Float256{
			0xc00b_cfe1_f8d9_5f34, 0xfd31_d679_ab69_e5c3,
			0xb0f1_7ada_7e7d_5192, 0xd103_7ec9_e91d_a513,
		}
		Q29 = Float256{
			0x400c_69bc_be39_e697, 0x91c8_e03f_a5ea_dc8e,
			0x33bf_8da0_64d0_323f, 0xe2cf_a939_316e_b9d0,
		}
		Q30 = Float256{
			0xc00d_063f_29ef_3041, 0x3254_b14e_5505_1a32,
			0x978f_5ada_50c1_f1cc, 0x7b8c_6feb_76aa_9174,
		}
	)

	w := One.Quo(x.Mul(x))

	p := P30
	p = FMA256(p, w, P29)
	p = FMA256(p, w, P28)
	p = FMA256(p, w, P27)
	p = FMA256(p, w, P26)
	p = FMA256(p, w, P25)
	p = FMA256(p, w, P24)
	p = FMA256(p, w, P23)
	p = FMA256(p, w, P22)
	p = FMA256(p, w, P21)
	p = FMA256(p, w, P20)
	p = FMA256(p, w, P19)
	p = FMA256(p, w, P18)
	p = FMA256(p, w, P17)
	p = FMA256(p, w, P16)
	p = FMA256(p, w, P15)
	p = FMA256(p, w, P14)
	p = FMA256(p, w, P13)
	p = FMA256(p, w, P12)
	p = FMA256(p, w, P11)
	p = FMA256(p, w, P10)
	p = FMA256(p, w, P9)
	p = FMA256(p, w, P8)
	p = FMA256(p, w, P7)
	p = FMA256(p, w, P6)
	p = FMA256(p, w, P5)
	p = FMA256(p, w, P4)
	p = FMA256(p, w, P3)
	p = FMA256(p, w, P2)
	p = FMA256(p, w, P1)
	p = FMA256(p, w, P0)

	q := Q30
	q = FMA256(q, w, Q29)
	q = FMA256(q, w, Q28)
	q = FMA256(q, w, Q27)
	q = FMA256(q, w, Q26)
	q = FMA256(q, w, Q25)
	q = FMA256(q, w, Q24)
	q = FMA256(q, w, Q23)
	q = FMA256(q, w, Q22)
	q = FMA256(q, w, Q21)
	q = FMA256(q, w, Q20)
	q = FMA256(q, w, Q19)
	q = FMA256(q, w, Q18)
	q = FMA256(q, w, Q17)
	q = FMA256(q, w, Q16)
	q = FMA256(q, w, Q15)
	q = FMA256(q, w, Q14)
	q = FMA256(q, w, Q13)
	q = FMA256(q, w, Q12)
	q = FMA256(q, w, Q11)
	q = FMA256(q, w, Q10)
	q = FMA256(q, w, Q9)
	q = FMA256(q, w, Q8)
	q = FMA256(q, w, Q7)
	q = FMA256(q, w, Q6)
	q = FMA256(q, w, Q5)
	q = FMA256(q, w, Q4)
	q = FMA256(q, w, Q3)
	q = FMA256(q, w, Q2)
	q = FMA256(q, w, Q1)
	q = FMA256(q, w, Q0)
	q = q.Quo(x)

	sin, cos := x.Sincos()
	amp := One.Quo(Pi.Mul(x)).Sqrt()
	return amp.Mul(p.Add(q).Mul(sin).Add(q.Sub(p).Mul(cos)))
}
