package floats

// Exp returns e**x, the base-e exponential of a.
//
// Special cases are:
//
//	+Inf.Exp() = +Inf
//	NaN.Exp() = NaN
//
// Very large values overflow to 0 or +Inf.
// Very small values underflow to 1.
func (a Float256) Exp() Float256 {
	var (
		// Ln2Hi = ln(2) ~ 0.69314718055994530941723212145817656807550013436025525412068000949339362
		// Ln2Lo = ln(2) - Ln2Hi ~ 1.68505384472783385591763704017160191700974868940429492048047E-72
		Ln2Hi = Float256{
			0x3fff_e62e_42fe_fa39, 0xef35_793c_7673_007e,
			0x5ed5_e81e_6864_ce53, 0x16c5_b141_a2eb_7175,
		}
		Ln2Lo = Float256{
			0x3ff1_07d1_5f3d_c3b1, 0x036f_5d64_c2ac_aa97,
			0xda57_d0d8_8769_7571, 0xae09_c0c7_cb80_70d0,
		}

		// log2(e) ~ 1.44269504088896340735992468100189213742664595415298593413544940693110922
		Log2e = Float256{
			0x3fff_f715_4765_2b82, 0xfe17_77d0_ffda_0d23,
			0xa7d1_1d6a_ef55_1bad, 0x2b4b_1164_a2cd_9a34,
		}

		// ln(max float256 + 0.5ulp) = ln(2²⁶²¹⁴³×(2-2⁻²³⁷))
		// ~ 181704.374500706303191870897247532238261583907221734753336211540408636177
		Overflow = Float256{
			0x4001_062e_42fe_fa39, 0xef35_793c_7673_007e,
			0x5ed5_e81e_6864_ce53, 0x16c5_b141_a2eb_7175,
		}

		// ln(min float256 - 0.5ulp) = ln(2⁻²⁶²³⁷⁹)
		// ~ 181867.264088137890339583946796074909755081649753309413320929900210867126
		Underflow = Float256{
			0xc001_0633_5a1c_da3d, 0xdc01_1ec2_0913_2f62,
			0xbbd6_2bb5_6d5f_4375, 0x7057_2b20_10bb_94fe,
		}

		// The upper limit for underflow
		// when exp(a) ~ 1 + a + a²/2! + ...
		// NearZero = 2**-119
		NearZero = Float256{
			0x3ff8_8000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// Half = 0.5
		Half = Float256{
			0x3fff_e000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)

	// special cases
	switch {
	case a.IsNaN():
		return NewFloat256NaN()
	case a.IsInf(1):
		return NewFloat256Inf(1)
	case a.IsInf(-1):
		return Float256{} // 0
	case a.Gt(Overflow):
		return NewFloat256Inf(1)
	case a.Lt(Underflow):
		return Float256{} // 0
	case a.Abs().Lt(NearZero):
		return Float256(uvone256).Add(a)
	}

	// reduce; computed as r = hi - lo for extra precision.
	var k int64
	if a.Signbit() {
		k = Log2e.Mul(a).Sub(Half).Int64()
	} else {
		k = Log2e.Mul(a).Add(Half).Int64()
	}
	fk := NewFloat256(float64(k))
	hi := a.Sub(fk.Mul(Ln2Hi))
	lo := fk.Mul(Ln2Lo)

	// compute
	return expmulti256(hi, lo, k)
}

func expmulti256(hi, lo Float256, k int64) Float256 {
	var (
		One = Float256{
			0x3fff_f000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Two = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		P1 = Float256{
			0x3fffc55555555555, 0x5555555555555555,
			0x5555555555555555, 0x5555555555555555,
		}
		P2 = Float256{
			0xbfff66c16c16c16c, 0x16c16c16c16c16c1,
			0x6c16c16c16c16c16, 0xc16c16c16c16c16c,
		}
		P3 = Float256{
			0x3fff11566abc0115, 0x66abc011566abc01,
			0x1566abc011566abc, 0x011566abc011566b,
		}
		P4 = Float256{
			0xbffebbbd779334ef, 0x0aac668223ddf99b,
			0x557112cce88a4460, 0x01bbd779334ef0ab,
		}
		P5 = Float256{
			0x3ffe666a8f2bf70e, 0xbda296113e9983e2,
			0x5ee7024351e41838, 0xe4f4e1d64da9cf68,
		}
		P6 = Float256{
			0xbffe122805d64426, 0x7ef74ac66ffe4f96,
			0xfb48f7e49b4ce8d2, 0x38f0a7ea87bf4a68,
		}
		P7 = Float256{
			0x3ffdbd6db2c4e091, 0x61fb84aeed184e9e,
			0x066f99f1fa9b097b, 0x4435abc4022f0329,
		}
		P8 = Float256{
			0xbffd67da4e1f7995, 0x5c4bfe25340de85a,
			0x1b8843952d475d5c, 0xd39c6dcad7ce5566,
		}
		P9 = Float256{
			0x3ffd1355871d652e, 0x9d9dcacccbafaf78,
			0x3b7f40ce17da75ac, 0x5a335f1795976a93,
		}
		P10 = Float256{
			0xbffcbf57d968caac, 0xf0cc79c1ea15b541,
			0x70424a575dfd1b2b, 0x1809e9a481006df6,
		}
		P11 = Float256{
			0x3ffc6967e1f09c37, 0x6efb05695b6091d5,
			0x4bcf6a6161dd22ec, 0xc4bb7a36aba7b96c,
		}
		P12 = Float256{
			0xbffc1497d9033a2b, 0x5c6e10fccaab4d90,
			0xed5eca9c39810d3c, 0x4dbd570c96d05531,
		}
		P13 = Float256{
			0x3ffbc0b132d7c6ad, 0x064075149b239d76,
			0xd4d8f247c70bb516, 0xa70d11dee0e0dd2d,
		}
		P14 = Float256{
			0xbffb6b0f72d59f1c, 0x167cc2dd227ed9e4,
			0x0ad942efb09a5396, 0xd5384522de856610,
		}
		P15 = Float256{
			0x3ffb15ef2da4cca2, 0x6d5ae64eb7f7519c,
			0xda1da319063baac7, 0x956ed49a2d37e0ad,
		}
		P16 = Float256{
			0xbffac1c77df96de3, 0x8afc4a74c45e598a,
			0x7ed8e44f7af332d8, 0xa0d1917809da6a86,
		}
		P17 = Float256{
			0x3ffa6cd299de521b, 0x61afe281e5f785f2,
			0xc83b854131ff0921, 0xa5549e67c001e3f0,
		}
		P18 = Float256{
			0xbffa175cde656574, 0xa6cec60c69655d10,
			0x25d93865d5ba3100, 0x9b43dbe56e7bda3f,
		}
		P19 = Float256{
			0x3ff9c2efe8db3b4a, 0xdec67e2d31c24582,
			0x3ef11fe81fa34df7, 0x2177b9431d8bd518,
		}
		P20 = Float256{
			0xbff96eb322904761, 0xfeecf7d20d16b312,
			0x6bc740e6bcbe5901, 0x2cb30a988a3162c8,
		}
		P21 = Float256{
			0x3ff86eae02f6528e, 0x246cdda9507c08db,
			0xb36a0808645ef26a, 0xc9a1415f2aaa8379,
		}
		P22 = Float256{
			0xbff7bb8b1d55007d, 0x9223c6bd551d95ea,
			0xfce3c29088853835, 0x0d568621bbfc1dbb,
		}
		P23 = Float256{
			0x3ff7007cfdf27c56, 0xb3ae0920c29d7f08,
			0xd09495ebbf514763, 0x334b44387dab38d4,
		}
		P24 = Float256{
			0xbff6e4c2d3d9d15f, 0x81c5f7dcad8023c0,
			0xc7e5098cb50237ee, 0x803c97e7de618244,
		}
		P25 = Float256{
			0x3ff686584a910fad, 0xeb83156863328f90,
			0xa3abdcb46599cf53, 0x01dd2be2eb39e318,
		}
	)
	r := hi.Sub(lo)
	t := r.Mul(r)
	c := FMA256(t, P25, P24)
	c = FMA256(t, c, P23)
	c = FMA256(t, c, P22)
	c = FMA256(t, c, P21)
	c = FMA256(t, c, P20)
	c = FMA256(t, c, P19)
	c = FMA256(t, c, P18)
	c = FMA256(t, c, P17)
	c = FMA256(t, c, P16)
	c = FMA256(t, c, P15)
	c = FMA256(t, c, P14)
	c = FMA256(t, c, P13)
	c = FMA256(t, c, P12)
	c = FMA256(t, c, P11)
	c = FMA256(t, c, P10)
	c = FMA256(t, c, P9)
	c = FMA256(t, c, P8)
	c = FMA256(t, c, P7)
	c = FMA256(t, c, P6)
	c = FMA256(t, c, P5)
	c = FMA256(t, c, P4)
	c = FMA256(t, c, P3)
	c = FMA256(t, c, P2)
	c = FMA256(t, c, P1)
	c = t.Mul(c)
	c = r.Sub(c)
	y := One.Sub(lo.Sub(r.Mul(c).Quo(Two.Sub(c))).Sub(hi))
	return y.Ldexp(int(k))
}
