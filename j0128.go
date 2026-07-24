package floats

// J0 returns the order-zero Bessel function of the first kind.
//
// Special cases are:
//
//	J0(±Inf) = 0
//	J0(0) = 1
//	J0(NaN) = NaN
func (a Float128) J0() Float128 {
	switch {
	case a.IsNaN():
		return a
	case a.IsInf(0):
		return Float128{}
	case a.IsZero():
		return Float128(uvone128)
	}

	x := a.Abs()

	var Threshold = Float128{0x4007_f400_0000_0000, 0x0000_0000_0000_0000} // 500

	if x.Lt(Threshold) {
		return j0Miller128(x)
	}
	return j0Asymptotic128(x)
}

// j0Miller128 returns J0(x) for 0 < x < 500 using Miller's algorithm: a
// backward recurrence starting from an arbitrary trial value at a high
// order, which is numerically stable (unlike the power series, which loses
// precision to cancellation once x is more than a handful), followed by
// normalizing against the identity J0(x) + 2*sum(J[2k](x)) = 1.
func j0Miller128(x Float128) Float128 {
	var (
		Zero = Float128{}
		One  = Float128(uvone128)
		Two  = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	// The margin below controls the accuracy: the backward recurrence's
	// sensitivity to the arbitrary starting value decays the further the
	// starting order is past x, so a bigger margin means more correct
	// digits. 400 keeps the error many orders of magnitude below
	// Float128's precision for all x in (0, 500).
	m := x.Ceil().Int64() + 400
	if m%2 != 0 {
		m++
	}

	Jnp1 := Zero
	Jn := One
	sum := Zero
	for n := m; n >= 1; n-- {
		coef := Two.Mul(NewFloat128(float64(n))).Quo(x)
		Jnm1 := FMA128(coef, Jn, Jnp1.Neg())
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
	return Jn.Quo(sum)
}

// j0Asymptotic128 returns J0(x) for x >= 500 using Hankel's asymptotic
// expansion
//
//	J0(x) ~ sqrt(2/(pi*x)) * (P(x)*cos(x-pi/4) - Q(x)*sin(x-pi/4))
//	      = sqrt(1/(pi*x)) * ((P(x)+Q(x))*cos(x) + (P(x)-Q(x))*sin(x))
//
// where P and Q are the (exact, rational) Hankel coefficients for order 0
// (see e.g. Abramowitz & Stegun 9.2.5-9.2.10).
func j0Asymptotic128(x Float128) Float128 {
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
	return amp.Mul(p.Add(q).Mul(cos).Add(p.Sub(q).Mul(sin)))
}
