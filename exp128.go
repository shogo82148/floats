package floats

import "math/bits"

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
		// ln(max float128 + 0.5ulp) = ln(2¹⁶³⁸³×(2-2⁻¹¹³))
		// ~ 11356.523406294143949491931077970765
		Overflow = Float128{0x400c_62e4_2fef_a39e, 0xf357_93c7_6730_07e6}

		// ln(min float128 - 0.5ulp) = ln(2⁻¹⁶⁴⁹⁵)
		// ~ -11433.462743336297878837243843452623
		Underflow = Float128{0xc00c_654b_b3b2_c73e, 0xbb05_9fab_b506_ff34}
	)

	sign := a[0] & signMask128[0]
	exp := int((a[0]>>(shift128-64))&mask128) - bias128

	// special cases
	switch {
	case exp < -114:
		// e**a rounds to 1 when |a| < 2**-114, including ±0 and subnormal values.
		return Float128(uvone128)
	case a.IsNaN():
		return a
	case a.Gt(Overflow):
		return NewFloat128Inf(1)
	case a.Lt(Underflow):
		return Float128{} // 0
	}

	// a = ±m × 2**(exp-112), where m is a 113-bit integer.
	m1 := a[0]&fracMask128[0] | 1<<(shift128-64)
	m0 := a[1]

	n := expN128(exp, m1, m0)
	if n == 0 {
		// |a| < ln(2)/128. e**a = 1 + a × (e**a - 1)/a.
		// a × (e**a - 1)/a is computed with the relative precision,
		// so that 1 + a is rounded correctly even when |a| is tiny.
		x1, x0 := expFix128(exp, m1, m0)
		g1, g0 := expm1Poly128(sign != 0, x1, x0)
		p3, p2, p1, p0 := mul128x128(m1, m0, g1, g0)

		// p = |a| × (e**a - 1)/a in fixed point with 191 fractional bits.
		// The product has 112+127-exp fractional bits, and exp <= -8.
		v2, v1, v0, sticky := rsh256Sticky(p3, p2, p1, p0, uint(48-exp))
		var c uint64
		if sign == 0 {
			// e**a = 1 + p
			v2 += 1 << 63
		} else {
			// e**a = 1 - p
			var b uint64
			if sticky {
				b = 1 // 1 - (v + δ) = (1 - v - 1) + (1 - δ) for 0 < δ < 1
			}
			v0, c = bits.Sub64(0, v0, 0)
			v1, c = bits.Sub64(0, v1, c)
			v2, _ = bits.Sub64(1<<63, v2, c)
			v0, c = bits.Sub64(v0, b, 0)
			v1, c = bits.Sub64(v1, 0, c)
			v2, _ = bits.Sub64(v2, 0, c)
		}
		return fixToFloat128(0, v2, v1, v0, sticky, -191)
	}

	k, v2, v1, v0 := expKernel128(sign, exp, m1, m0, n)
	return fixToFloat128(0, v2, v1, v0, false, k-191)
}

// rsh256Sticky returns (x3:x2:x1:x0) >> s for s < 256.
// The result must fit in 192 bits.
// sticky reports whether any of the shifted out bits is nonzero.
func rsh256Sticky(x3, x2, x1, x0 uint64, s uint) (y2, y1, y0 uint64, sticky bool) {
	for ; s >= 64; s -= 64 {
		sticky = sticky || x0 != 0
		x3, x2, x1, x0 = 0, x3, x2, x1
	}
	if s > 0 {
		sticky = sticky || x0<<(64-s) != 0
		x2, x1, x0 = x2>>s|x3<<(64-s), x1>>s|x2<<(64-s), x0>>s|x1<<(64-s)
	}
	return x2, x1, x0, sticky
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
