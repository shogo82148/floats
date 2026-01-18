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

// Exp2 returns 2**x, the base-2 exponential of x.
//
// Special cases are the same as [Exp].
func (a Float256) Exp2() Float256 {
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

		// log2(max float256 + 0.5ulp) = log2(2²⁶²¹⁴³×(2-2⁻²³⁷)) ~ 262144
		Overflow = Float256{
			0x4001_1000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// log2(min float256 - 0.5ulp) = log2(2⁻²⁶²³⁷⁹) = -262379
		Underflow = Float256{
			0xc001_1003_ac00_0000, 0x0000_0000_0000_0000,
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
	}

	// reduce; computed as r = hi - lo for extra precision.
	var k int64
	if a.Signbit() {
		k = a.Sub(Half).Int64()
	} else {
		k = a.Add(Half).Int64()
	}
	t := a.Sub(NewFloat256(float64(k)))
	hi := t.Mul(Ln2Hi)
	lo := t.Mul(Ln2Lo).Neg()

	// compute
	return expmulti256(hi, lo, k)
}

func expmulti256(hi, lo Float256, k int64) Float256 {
	var y Float256
	r := hi.Sub(lo)
	for n := 50; n >= 0; n-- {
		y = y.Add(power256(r, n).Quo(factorial256(n)))
	}
	return y.Ldexp(int(k))
}

func power256(x Float256, n int) Float256 {
	result := Float256(uvone256)
	for n != 0 {
		if n%2 == 1 {
			result = result.Mul(x)
		}
		n /= 2
		x = x.Mul(x)
	}
	return result
}

func factorial256(n int) Float256 {
	if n == 0 {
		return Float256(uvone256)
	}
	result := Float256(uvone256)
	for i := 2; i <= n; i++ {
		result = result.Mul(NewFloat256(float64(i)))
	}
	return result
}
