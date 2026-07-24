package floats

// Lgamma returns the natural logarithm and sign (-1 or +1) of Gamma(a).
//
// Special cases are:
//
//	Lgamma(+Inf) = +Inf
//	Lgamma(0) = +Inf
//	Lgamma(-integer) = +Inf
//	Lgamma(-Inf) = -Inf
//	Lgamma(NaN) = NaN
func (a Float256) Lgamma() (Float256, int) {
	var (
		// Half is 0.5
		Half = Float256{
			0x3fff_e000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// One is 1
		One = Float256(uvone256)

		// Pi is π
		Pi = Float256{
			0x4000_0921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}

		// LnPi is ln(π)
		LnPi = Float256{
			0x3fff_f250_d048_e7a1, 0xbd0b_d5f9_56c6_a843,
			0xf499_85e6_ddbf_3b3f, 0x2606_e338_02ec_aefb,
		}
	)

	switch {
	case a.IsNaN() || a.IsInf(0):
		return a, 1
	case a.IsZero() || isNegInt256(a):
		return NewFloat256Inf(1), 1
	}

	if g := a.Gamma(); !g.IsInf(0) && !g.IsZero() {
		// Gamma(a) didn't overflow (or underflow) here, so just take its log directly.
		sign := 1
		if g.Signbit() {
			sign = -1
		}
		return g.Abs().Log(), sign
	}

	// Gamma(a) over/underflowed, so work in log space instead. Note that
	// |a| need not exceed MaxStirling for this to happen: Gamma is
	// increasing quickly enough near its overflow point that some
	// non-integer a below MaxStirling already overflow Gamma(a).
	q := a.Abs()
	if !a.Signbit() {
		return lnStirling256(a), 1
	}

	// Reflection formula in log space, mirroring the one used by Gamma:
	//
	//	|Gamma(a)| = pi / (|q * sin(pi*frac)| * Gamma(q))
	p := q.Floor()
	z := q.Sub(p)
	if z.Gt(Half) {
		p = p.Add(One)
		z = q.Sub(p)
	}
	s := q.Mul(Pi.Mul(z).Sin())
	sign := -1
	if isOddInt256(p) {
		sign = 1
	}
	if s.IsZero() {
		return NewFloat256Inf(1), sign
	}
	return LnPi.Sub(s.Abs().Log()).Sub(lnStirling256(q)), sign
}

// lnStirling256 returns ln(Gamma(x)) for x large enough that Gamma(x)
// itself would overflow. It uses the same Stirling asymptotic series as
// stirling256, but evaluated directly in log space so that it stays finite
// well beyond the point where Gamma(x) would overflow.
func lnStirling256(x Float256) Float256 {
	var (
		Half = Float256{
			0x3fff_e000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		One = Float256(uvone256)

		// HalfLn2Pi is ln(sqrt(2*pi))
		HalfLn2Pi = Float256{
			0x3fff_ed67_f1c8_64be, 0xb4a6_9297_9200_2883,
			0x2404_79f6_11f1_a268, 0xb169_bbd8_d462_67b5,
		}
	)

	// Coefficients of the asymptotic series
	//
	//	ln Gamma(n+1) = 0.5*ln(2*pi*n) + n*ln(n) - n + M1/n + M2/n**3 + ...
	//
	// i.e. the Bernoulli-derived log-domain Stirling series (OEIS A046968/A046969).
	var (
		M1 = Float256{
			0x3fff_b555_5555_5555, 0x5555_5555_5555_5555,
			0x5555_5555_5555_5555, 0x5555_5555_5555_5555,
		}
		M2 = Float256{
			0xbfff_66c1_6c16_c16c, 0x16c1_6c16_c16c_16c1,
			0x6c16_c16c_16c1_6c16, 0xc16c_16c1_6c16_c16c,
		}
		M3 = Float256{
			0x3fff_4a01_a01a_01a0, 0x1a01_a01a_01a0_1a01,
			0xa01a_01a0_1a01_a01a, 0x01a0_1a01_a01a_01a0,
		}
		M4 = Float256{
			0xbfff_4381_3813_8138, 0x1381_3813_8138_1381,
			0x3813_8138_1381_3813, 0x8138_1381_3813_8138,
		}
		M5 = Float256{
			0x3fff_4b95_1e2b_18ff, 0x2357_0ea7_3806_e547,
			0x8ac6_3fc8_d5c3_a9ce, 0x01b9_51e2_b18f_f235,
		}
		M6 = Float256{
			0xbfff_5f6a_b0d9_993c, 0x7c81_f6ab_0d99_93c7,
			0xc81f_6ab0_d999_3c7c, 0x81f6_ab0d_9993_c7c8,
		}
		M7 = Float256{
			0x3fff_7a41_a41a_41a4, 0x1a41_a41a_41a4_1a41,
			0xa41a_41a4_1a41_a41a, 0x41a4_1a41_a41a_41a4,
		}
		M8 = Float256{
			0xbfff_9e42_86cb_0f53, 0x97dc_2064_a8ed_3175,
			0xb9fe_4286_cb0f_5397, 0xdc20_64a8_ed31_75ba,
		}
		M9 = Float256{
			0x3fff_c6fe_9638_1e06, 0x7ffa_1876_fe96_381e,
			0x067f_fa18_76fe_9638, 0x1e06_7ffa_1876_fe96,
		}
		M10 = Float256{
			0xbfff_f647_6701_181f, 0x39ed_bdb9_ce62_5987,
			0xd4c0_e916_2983_e954, 0xe044_7670_1181_f39f,
		}
	)

	n := x.Sub(One)
	w := One.Quo(n)
	w2 := w.Mul(w)

	mu := M10
	mu = FMA256(mu, w2, M9)
	mu = FMA256(mu, w2, M8)
	mu = FMA256(mu, w2, M7)
	mu = FMA256(mu, w2, M6)
	mu = FMA256(mu, w2, M5)
	mu = FMA256(mu, w2, M4)
	mu = FMA256(mu, w2, M3)
	mu = FMA256(mu, w2, M2)
	mu = FMA256(mu, w2, M1)
	mu = mu.Mul(w)

	return x.Sub(Half).Mul(n.Log()).Sub(n).Add(HalfLn2Pi).Add(mu)
}
