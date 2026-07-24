package floats

// J1 returns the order-one Bessel function of the first kind.
//
// Special cases are:
//
//	J1(±Inf) = 0
//	J1(NaN) = NaN
func (a Float128) J1() Float128 {
	switch {
	case a.IsNaN():
		return a
	case a.IsInf(0):
		return Float128{}
	case a.IsZero():
		return a
	}

	x := a.Abs()

	var Threshold = Float128{0x4007_f400_0000_0000, 0x0000_0000_0000_0000} // 500

	var y Float128
	if x.Lt(Threshold) {
		y = j1Miller128(x)
	} else {
		y = j1Asymptotic128(x)
	}
	if a.Signbit() {
		y = y.Neg()
	}
	return y
}

// j1Miller128 returns J1(x) for 0 < x < 500 using the same backward
// recurrence as j0Miller128, but reading off the unnormalized value at
// n=1 instead of n=0.
func j1Miller128(x Float128) Float128 {
	var (
		Zero = Float128{}
		One  = Float128(uvone128)
		Two  = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	m := x.Ceil().Int64() + 400
	if m%2 != 0 {
		m++
	}

	Jnp1 := Zero
	Jn := One
	sum := Zero
	var j1unnorm Float128
	for n := m; n >= 1; n-- {
		coef := Two.Mul(NewFloat128(float64(n))).Quo(x)
		Jnm1 := FMA128(coef, Jn, Jnp1.Neg())
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

// j1Asymptotic128 returns J1(x) for x >= 500 using Hankel's asymptotic
// expansion
//
//	J1(x) ~ sqrt(2/(pi*x)) * (P(x)*cos(x-3pi/4) - Q(x)*sin(x-3pi/4))
//	      = sqrt(1/(pi*x)) * ((P(x)+Q(x))*sin(x) + (Q(x)-P(x))*cos(x))
//
// where P and Q are the (exact, rational) Hankel coefficients for order 1
// (see e.g. Abramowitz & Stegun 9.2.5-9.2.10 with nu=1).
func j1Asymptotic128(x Float128) Float128 {
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
	return amp.Mul(p.Add(q).Mul(sin).Add(q.Sub(p).Mul(cos)))
}
