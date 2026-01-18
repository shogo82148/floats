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
func (a Float128) Exp() Float128 {
	var (
		// Ln2Hi = ln(2) ~ 0.6931471805599453094172321214581765
		// Ln2Lo = ln(2) - Ln2Hi ~ 8.928835774481220748938623512047474e-35
		Ln2Hi = Float128{0x3ffe_62e4_2fef_a39e, 0xf357_93c7_6730_07e5}
		Ln2Lo = Float128{0x3f8d_dabd_03cd_0c99, 0xca62_d8b6_2834_5d6e}

		// log2(e) ~ 1.442695040888963407359924681001892
		Log2e = Float128{0x3fff_7154_7652_b82f, 0xe177_7d0f_fda0_d23a}

		// ln(max float128 + 0.5ulp) = ln(2¹⁶³⁸³×(2-2⁻¹¹³))
		// ~ 11356.523406294143949491931077970765
		Overflow = Float128{0x400c_62e4_2fef_a39e, 0xf357_93c7_6730_07e6}

		// ln(min float128 - 0.5ulp) = ln(2⁻¹⁶⁴⁹⁵)
		// ~ -11433.462743336297878837243843452623
		Underflow = Float128{0xc00c_654b_b3b2_c73e, 0xbb05_9fab_b506_ff34}

		// The upper limit for underflow
		// when exp(a) ~ 1 + a + a²/2! + ...
		// NearZero = 2**-57
		NearZero = Float128{0x3fc6_0000_0000_0000, 0x0000_0000_0000_0000}

		// Half = 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	// special cases
	switch {
	case a.IsNaN():
		return NewFloat128NaN()
	case a.IsInf(1):
		return NewFloat128Inf(1)
	case a.IsInf(-1):
		return Float128{} // 0
	case a.Gt(Overflow):
		return NewFloat128Inf(1)
	case a.Lt(Underflow):
		return Float128{} // 0
	case a.Abs().Lt(NearZero):
		return Float128(uvone128).Add(a)
	}

	// reduce; computed as r = hi - lo for extra precision.
	var k int64
	if a.Signbit() {
		k = Log2e.Mul(a).Sub(Half).Int64()
	} else {
		k = Log2e.Mul(a).Add(Half).Int64()
	}
	fk := NewFloat128(float64(k))
	hi := a.Sub(fk.Mul(Ln2Hi))
	lo := fk.Mul(Ln2Lo)

	// compute
	return expmulti128(hi, lo, k)
}

// Exp2 returns 2**x, the base-2 exponential of x.
//
// Special cases are the same as [Exp].
func (a Float128) Exp2() Float128 {
	var (
		// Ln2Hi = ln(2) ~ 0.6931471805599453094172321214581765
		// Ln2Lo = ln(2) - Ln2Hi ~ 8.928835774481220748938623512047474e-35
		Ln2Hi = Float128{0x3ffe_62e4_2fef_a39e, 0xf357_93c7_6730_07e5}
		Ln2Lo = Float128{0x3f8d_dabd_03cd_0c99, 0xca62_d8b6_2834_5d6e}

		// log2(max float128 + 0.5ulp) = log2(2¹⁶³⁸³×(2-2⁻¹¹³)) ~ 16384
		Overflow = Float128{0x400d_0000_0000_0000, 0x0000_0000_0000_0000}

		// log2(min float128 - 0.5ulp) = log2(2⁻¹⁶⁴⁹⁵) = -16495
		Underflow = Float128{0xc00d_01bc_0000_0000, 0x0000_0000_0000_0000}

		// Half = 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	// special cases
	switch {
	case a.IsNaN():
		return NewFloat128NaN()
	case a.IsInf(1):
		return NewFloat128Inf(1)
	case a.IsInf(-1):
		return Float128{} // 0
	case a.Gt(Overflow):
		return NewFloat128Inf(1)
	case a.Lt(Underflow):
		return Float128{} // 0
	}

	// reduce; computed as r = hi - lo for extra precision.
	var k int64
	if a.Signbit() {
		k = a.Sub(Half).Int64()
	} else {
		k = a.Add(Half).Int64()
	}
	t := a.Sub(NewFloat128(float64(k)))
	hi := t.Mul(Ln2Hi)
	lo := t.Mul(Ln2Lo).Neg()

	// compute
	return expmulti128(hi, lo, k)
}

func expmulti128(hi, lo Float128, k int64) Float128 {
	var (
		One = Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000} // 1.0
		Two = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000} // 2.0
		P0  = Float128{0x3f95_7bef_c047_b03b, 0xb125_e7dd_7ff1_edd1}
		P1  = Float128{0x3ffc_5555_5555_5555, 0x5555_5555_554f_599e}
		P2  = Float128{0xbff6_6c16_c16c_16c1, 0x6c16_c16b_1f26_512a}
		P3  = Float128{0x3ff1_1566_abc0_1156, 0x6abb_f732_a45c_9fab}
		P4  = Float128{0xbfeb_bbd7_7933_4ef0, 0xaa50_c0f2_5388_c992}
		P5  = Float128{0x3fe6_66a8_f2bf_70ea, 0x1f56_2556_a195_3458}
		P6  = Float128{0xbfe1_2280_5d64_3e1c, 0x8e60_87d6_f15c_47b8}
		P7  = Float128{0x3fdb_d6db_2c3f_c331, 0x7985_1841_eb9d_df4f}
		P8  = Float128{0xbfd6_7da4_d23e_aa27, 0x339a_9091_ab9c_241c}
		P9  = Float128{0x3fd1_354d_6ca1_bc9a, 0xf489_11d7_9500_8948}
		P10 = Float128{0xbfcb_ec95_2a19_0369, 0xdfce_a81f_b638_2319}
	)
	r := hi.Sub(lo)
	t := r.Mul(r)
	c := FMA128(t, P10, P9)
	c = FMA128(t, c, P8)
	c = FMA128(t, c, P7)
	c = FMA128(t, c, P6)
	c = FMA128(t, c, P5)
	c = FMA128(t, c, P4)
	c = FMA128(t, c, P3)
	c = FMA128(t, c, P2)
	c = FMA128(t, c, P1)
	c = FMA128(t, c, P0)
	c = r.Sub(c)
	y := One.Sub(lo.Sub(r.Mul(c).Quo(Two.Sub(c))).Sub(hi))
	return y.Ldexp(int(k))
}
