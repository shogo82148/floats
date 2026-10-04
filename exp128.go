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
		return expOnePlus128(sign, exp, m1, m0, g1, g0)
	}

	k, v2, v1, v0 := expKernel128(sign, exp, m1, m0, n)
	return fixToFloat128(0, v2, v1, v0, false, k-191)
}

// expOnePlus128 returns 1 ± m × 2**(exp-112) × g rounded to nearest even,
// where sign is the sign bit, m = (m1:m0) is a 113-bit integer, exp <= -8,
// and g = (g1:g0) × 2**-127 < 2.
// m × g is computed with the relative precision,
// so that the result is rounded correctly even when it is very close to 1.
func expOnePlus128(sign uint64, exp int, m1, m0, g1, g0 uint64) Float128 {
	p3, p2, p1, p0 := mul128x128(m1, m0, g1, g0)

	// p = m × 2**(exp-112) × g in fixed point with 191 fractional bits.
	// The product has 112+127-exp fractional bits.
	v2, v1, v0, sticky := rsh256Sticky(p3, p2, p1, p0, uint(48-exp))
	var c uint64
	if sign == 0 {
		// 1 + p
		v2 += 1 << 63
	} else {
		// 1 - p
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
		// log2(max float128 + 0.5ulp) = log2(2¹⁶³⁸³×(2-2⁻¹¹³)) ~ 16384
		Overflow = Float128{0x400d_0000_0000_0000, 0x0000_0000_0000_0000}

		// log2(min float128 - 0.5ulp) = log2(2⁻¹⁶⁴⁹⁵) = -16495
		Underflow = Float128{0xc00d_01bc_0000_0000, 0x0000_0000_0000_0000}
	)

	sign := a[0] & signMask128[0]
	exp := int((a[0]>>(shift128-64))&mask128) - bias128

	// special cases
	switch {
	case exp < -114:
		// 2**a rounds to 1 when |a| < 2**-114, including ±0 and subnormal values.
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

	// |a|×2**134 modulo 2**128, that is, the fraction part of |a|×64 in fixed point
	// with 128 fractional bits. It is exact for exp >= -22.
	f1, f0 := expFix128(exp, m1, m0)

	if exp < -7 {
		// |a| < 1/128. 2**a = 1 + a × ln(2) × (e**b - 1)/b, where b = a × ln(2).
		l := &expm1Ln2Fix128
		b1, b0, _, _ := mul128x128(f1, f0, l[0], l[1]) // |b|×2**134
		g1, g0 := expm1Poly128(sign != 0, b1, b0)
		h1, h0, _, _ := mul128x128(g1, g0, l[0], l[1]) // ln(2) × g × 2**127
		return expOnePlus128(sign, exp, m1, m0, h1, h0)
	}

	// reduce: |a|×64 = n + f, where n is an integer and |f| <= 1/2.
	// n = round(|a|×64). |a|×64 = m × 2**(exp-106), and -7 <= exp <= 14.
	n := m1 >> (42 - exp)
	n += f1 >> 63 // round up if f >= 1/2; then f is negative in two's complement.

	// r = f × ln(2)/64 in fixed point with 134 fractional bits.
	fneg := f1>>63 != 0
	if fneg {
		f1, f0 = neg128(f1, f0)
	}
	l := &expm1Ln2Fix128
	r1, r0, _, _ := mul128x128(f1, f0, l[0], l[1])
	if fneg {
		r1, r0 = neg128(r1, r0)
	}

	k, v2, v1, v0 := expScale128(sign, n, r1, r0)
	return fixToFloat128(0, v2, v1, v0, false, k-191)
}
