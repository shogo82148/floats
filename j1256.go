package floats

// J1 returns the order-one Bessel function of the first kind.
//
// Special cases are:
//
//	J1(±Inf) = 0
//	J1(NaN) = NaN
func (a Float256) J1() Float256 {
	switch {
	case a.IsNaN():
		return a
	case a.IsInf(0):
		return Float256{}
	case a.IsZero():
		return a
	}

	x := a.Abs()

	var Threshold = Float256{
		0x4000_7f40_0000_0000, 0x0000_0000_0000_0000,
		0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
	} // 500

	var y Float256
	if x.Lt(Threshold) {
		y = j1Miller256(x)
	} else {
		y = j1Asymptotic256(x)
	}
	if a.Signbit() {
		y = y.Neg()
	}
	return y
}

// j1Miller256 returns J1(x) for 0 < x < 500 using the same backward
// recurrence as j0Miller256, but reading off the unnormalized value at
// n=1 instead of n=0.
func j1Miller256(x Float256) Float256 {
	var (
		Zero = Float256{}
		One  = Float256(uvone256)
		Two  = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)

	m := x.Ceil().Int64() + 400
	if m%2 != 0 {
		m++
	}

	Jnp1 := Zero
	Jn := One
	sum := Zero
	var j1unnorm Float256
	for n := m; n >= 1; n-- {
		coef := Two.Mul(NewFloat256(float64(n))).Quo(x)
		Jnm1 := FMA256(coef, Jn, Jnp1.Neg())
		if n-1 == 1 {
			j1unnorm = Jnm1
		}
		if (n-1)%2 == 0 {
			if n == 1 {
				sum = sum.Add(Jnm1)
			} else {
				sum = sum.Add(Jnm1).Add(Jnm1)
			}
		}
		Jnp1 = Jn
		Jn = Jnm1
	}
	return j1unnorm.Quo(sum)
}

// j1Asymptotic256 returns J1(x) for x >= 500 using Hankel's asymptotic
// expansion
//
//	J1(x) ~ sqrt(2/(pi*x)) * (P(x)*cos(x-3pi/4) - Q(x)*sin(x-3pi/4))
//	      = sqrt(1/(pi*x)) * ((P(x)+Q(x))*sin(x) + (Q(x)-P(x))*cos(x))
//
// where P and Q are the (exact, rational) Hankel coefficients for order 1
// (see e.g. Abramowitz & Stegun 9.2.5-9.2.10 with nu=1).
func j1Asymptotic256(x Float256) Float256 {
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
			0x3fff_be00_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P2 = Float256{
			0xbfff_c275_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P3 = Float256{
			0x3fff_e5a6_a580_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P4 = Float256{
			0xc000_1b89_20d2_6800, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P5 = Float256{
			0x4000_5e66_43dc_4a11, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P6 = Float256{
			0xc000_a9cc_8b6a_2ea4, 0x4908_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P7 = Float256{
			0x4000_ff29_945c_c23c, 0x34f1_2180_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P8 = Float256{
			0xc001_5964_5bee_011b, 0xe681_13c1_5e80_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P9 = Float256{
			0x4001_bad6_b4c8_4e17, 0x0e38_82c2_9a31_d400,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P10 = Float256{
			0xc002_21da_5076_cedb, 0x1509_4f55_0610_e709,
			0xd800_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P11 = Float256{
			0x4002_8d30_a1b7_7ee9, 0x94e2_ca2f_a968_7487,
			0x71bf_4000_0000_0000, 0x0000_0000_0000_0000,
		}
		P12 = Float256{
			0xc002_fcc4_18ac_cd75, 0x09da_05b5_b604_2a02,
			0x55c8_e052_4000_0000, 0x0000_0000_0000_0000,
		}
		P13 = Float256{
			0x4003_70d0_88ba_5da2, 0x42fb_238a_c220_f0d3,
			0x2d6f_6f05_04b2_8000, 0x0000_0000_0000_0000,
		}
		P14 = Float256{
			0xc003_e701_c3f7_bd61, 0xcb1e_7d57_2dad_81cf,
			0xa7fb_3aab_5862_6479, 0x0000_0000_0000_0000,
		}
		P15 = Float256{
			0x4004_6236_104e_ed10, 0x4587_82b0_3b38_1a95,
			0xcd3c_efdf_4d8d_83ac, 0xec66_0000_0000_0000,
		}
		P16 = Float256{
			0xc004_e083_7cd4_f1b4, 0x899c_e721_b03f_036d,
			0xafc3_d974_36a5_5f96, 0xfa9f_cea7_4000_0000,
		}
		P17 = Float256{
			0x4005_6101_9694_dd1b, 0x226b_07c7_8b22_0028,
			0xc6a4_9318_f33c_5935, 0x70f7_94a7_edae_8000,
		}
		P18 = Float256{
			0xc005_e3bd_25bb_7c10, 0x6d61_17ff_69c1_3c5d,
			0x407a_71c7_93c0_fd0e, 0x5421_8121_f386_a40b,
		}
		P19 = Float256{
			0x4006_69a5_d585_f4db, 0x13a0_b0de_0c0c_c14c,
			0xa0ce_82ba_8f9d_7c8a, 0x9f3e_6e37_8091_e3ae,
		}
		P20 = Float256{
			0xc006_f28a_90bd_e902, 0xaee9_c554_5ade_10ee,
			0x5f2f_d120_1bd3_cf74, 0x4ca1_1767_675f_c893,
		}
		P21 = Float256{
			0x4007_7dab_229e_6606, 0x2af3_429c_3795_6bae,
			0x9bec_7816_c65b_1cd9, 0x5be0_60a6_84f8_4428,
		}
		P22 = Float256{
			0xc008_0a24_3a56_8cfc, 0x9615_9dcb_2c0f_fbd3,
			0x4c11_919c_2c06_f869, 0x1139_6ef5_af9c_3f9c,
		}
		P23 = Float256{
			0x4008_9941_3b1d_204d, 0x7092_18ae_753d_5d1a,
			0x30c3_40ab_b8e8_5a0a, 0x67c5_6825_aa9f_89b8,
		}
		P24 = Float256{
			0xc009_2aa4_79e1_3a94, 0xcbff_dc34_a212_528b,
			0x6c30_24c1_0793_d5d3, 0xbe63_2d34_895f_1b40,
		}
		P25 = Float256{
			0x4009_be93_fee7_9350, 0x5a9b_d899_851c_c07b,
			0xf2e8_ed7e_cca0_59c6, 0x2170_a9a9_7021_8324,
		}
		P26 = Float256{
			0xc00a_5306_9329_0238, 0x9a7f_8e72_5671_9973,
			0xf67c_2544_3f84_7ec8, 0x07de_d720_22be_bf88,
		}
		P27 = Float256{
			0x400a_e996_d8d5_2018, 0xa4b1_5718_689c_aa56,
			0xa73a_1c90_7fcb_db26, 0x9f8a_b567_e300_31c8,
		}
		P28 = Float256{
			0xc00b_828b_a5df_db4a, 0x7aa9_f08a_5eaf_da34,
			0xf7ab_a129_8e56_3e44, 0x60d7_99e6_4afb_432e,
		}
		P29 = Float256{
			0x400c_1ce4_3ccb_bcbd, 0x8d5f_2518_5bba_880d,
			0xa8d7_ca58_4e4f_f0de, 0x6793_b3ab_6a89_e180,
		}
		P30 = Float256{
			0xc00c_b820_8486_4fbf, 0x9625_9ff3_aa7c_9e0e,
			0x7046_fc3f_bdf8_f870, 0x6392_1a1e_880e_72c1,
		}

		Q0 = Float256{
			0x3fff_d800_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q1 = Float256{
			0xbfff_ba40_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q2 = Float256{
			0x3fff_d1c3_d000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q3 = Float256{
			0xbfff_ffe5_8188_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q4 = Float256{
			0x4000_3b3f_b325_8c40, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q5 = Float256{
			0xc000_82de_c0ab_499c, 0xbc00_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q6 = Float256{
			0x4000_d341_980e_f232, 0x9f20_2000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q7 = Float256{
			0xc001_2b2b_73c0_dfbf, 0xb15b_d602_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q8 = Float256{
			0x4001_8952_6f70_e0a2, 0xa5a2_220d_c6f2_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q9 = Float256{
			0xc001_ee20_97fc_a372, 0xb904_1119_2f31_9e30,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q10 = Float256{
			0x4002_5646_b0f8_d0f7, 0x13f6_7946_16ff_6048,
			0x4860_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Q11 = Float256{
			0xc002_c409_d1cc_506d, 0x648e_1f25_1113_f7a5,
			0xacf4_6b80_0000_0000, 0x0000_0000_0000_0000,
		}
		Q12 = Float256{
			0x4003_358c_2b1f_8519, 0xab4c_ac1d_f3f1_0a07,
			0x0775_44d4_a680_0000, 0x0000_0000_0000_0000,
		}
		Q13 = Float256{
			0xc003_ab4b_650e_8553, 0x64ba_54ce_2158_235d,
			0xe86b_8d8c_2567_1600, 0x0000_0000_0000_0000,
		}
		Q14 = Float256{
			0x4004_241c_c8e6_4bd5, 0x01e8_edfc_4796_179e,
			0x26ea_b9d7_07ad_bcef, 0x9200_0000_0000_0000,
		}
		Q15 = Float256{
			0xc004_a10f_28f4_418d, 0x3e08_cf42_a1cb_b75a,
			0xe80f_db47_9095_24b3, 0x7806_8400_0000_0000,
		}
		Q16 = Float256{
			0x4005_2080_7c32_9c71, 0x5183_e1e0_644d_839b,
			0x9bd2_7006_dbb2_b5e2, 0xb0a0_c8de_c480_0000,
		}
		Q17 = Float256{
			0xc005_a20e_c5a9_d6e4, 0x2f00_d91e_1c22_aca0,
			0x5361_d605_6837_cb4b, 0xc62e_33f6_327d_eb00,
		}
		Q18 = Float256{
			0x4006_2631_9705_0485, 0x47f4_9309_493b_4493,
			0x02ed_ff2c_5805_79ee, 0xde0d_abd7_4f0e_401f,
		}
		Q19 = Float256{
			0xc006_ae70_fb6e_393c, 0x059a_91ec_7874_ad6d,
			0x41f0_2f29_6dd7_61d1, 0x7415_545a_cc72_143c,
		}
		Q20 = Float256{
			0x4007_372a_7e4e_d714, 0xc465_272e_e85c_93e2,
			0xcddb_1c11_5647_d007, 0x4ec4_64e6_6a47_d22a,
		}
		Q21 = Float256{
			0xc007_c376_3cd3_4d1c, 0x8ebb_86d5_9758_6973,
			0x7254_361f_deb8_be9f, 0x3bd3_b8b3_ee3f_99eb,
		}
		Q22 = Float256{
			0x4008_51f7_29f5_7b28, 0x8546_fd2c_e503_c8dd,
			0x4747_01de_75fd_971e, 0x12b4_3430_71f3_faea,
		}
		Q23 = Float256{
			0xc008_e225_45d0_4f66, 0x9e85_39c6_b5b9_e5fa,
			0xab3f_b182_30eb_8a65, 0x79eb_dbb3_be87_6fc8,
		}
		Q24 = Float256{
			0x4009_73f9_b9d4_65fb, 0xaf3c_8c58_afcc_f641,
			0x8dd4_574f_d3fc_c7a5, 0x4ee3_4c28_8d9b_a76e,
		}
		Q25 = Float256{
			0xc00a_07e1_d2ac_831d, 0x2b79_2176_7d8d_3873,
			0xa14e_4631_cc2b_970e, 0x732d_1c09_ec39_3b8b,
		}
		Q26 = Float256{
			0x400a_9ee8_87bf_ddc2, 0x5eea_77d3_efd4_3f37,
			0xbf2f_0a47_0fca_ae1d, 0x56bf_906d_604b_640e,
		}
		Q27 = Float256{
			0xc00b_3595_e1a2_2ddb, 0x3003_eb98_6d0b_47c2,
			0x70ce_bf1d_0487_36f3, 0xe849_dc42_44c6_4144,
		}
		Q28 = Float256{
			0x400b_d039_3743_90ba, 0xae25_d1f3_2684_0b92,
			0x0514_ee16_d39e_eb35, 0x2f74_2f86_757c_f5fb,
		}
		Q29 = Float256{
			0xc00c_6a2d_5f04_308d, 0x083e_1668_1ee8_4fea,
			0x57a4_31f6_47e7_728f, 0xbf4d_bd9e_d201_0762,
		}
		Q30 = Float256{
			0x400d_069d_4c7a_8c07, 0x08d9_441a_bbfa_9bb0,
			0x4134_a22c_3478_fc24, 0x7d97_38a5_4e50_4bec,
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
