package floats

import (
	"math"
	"math/bits"
)

// Expm1 returns e**a - 1, the base-e exponential of a minus 1.
// It is more accurate than Exp(a) - 1 when a is near zero.
//
// Special cases are:
//
//	+Inf.Expm1() = +Inf
//	-Inf.Expm1() = -1
//	NaN.Expm1() = NaN
//
// Very large values overflow to -1 or +Inf.
func (a Float128) Expm1() Float128 {
	var (
		// ln(max float128 + 0.5ulp) = ln(2¹⁶³⁸³×(2-2⁻¹¹³))
		// ~ 11356.523406294143949491931077970765
		Overflow = Float128{0x400c_62e4_2fef_a39e, 0xf357_93c7_6730_07e6}

		// e**a < 2**-114 for a < -80, so e**a - 1 rounds to -1.
		Underflow = Float128{0xc005_4000_0000_0000, 0x0000_0000_0000_0000}

		MinusOne = Float128{0xbfff_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	sign := a[0] & signMask128[0]
	exp := int((a[0]>>(shift128-64))&mask128) - bias128

	// special cases
	switch {
	case exp < -113:
		// expm1(a) ~ a when |a| < 2**-113, including ±0 and subnormal values.
		return a
	case a.IsNaN():
		return a
	case a.Gt(Overflow):
		return NewFloat128Inf(1)
	case a.Lt(Underflow):
		return MinusOne
	}

	// a = ±m × 2**(exp-112), where m is a 113-bit integer.
	m1 := a[0]&fracMask128[0] | 1<<(shift128-64)
	m0 := a[1]

	// reduce: a = n × ln(2)/64 + r, |r| <= ln(2)/128 < 2**-7.
	var n uint64
	if exp >= -8 {
		// n = round(|a| × 64/ln(2)), computed in float64.
		// Its error doesn't matter as long as |r| < 2**-7.
		const InvLn2By64 = 64 / math.Ln2
		top := float64(m1<<15 | m0>>49) // the top 64 bits of m
		scale := math.Float64frombits(uint64(exp-63+1023) << 52)
		n = uint64(top*scale*InvLn2By64 + 0.5)
	}

	if n == 0 {
		// |a| < ln(2)/128. expm1(a) = a × (e**a - 1)/a.
		// a×2**134 in fixed point, truncated.
		var x1, x0 uint64
		if s := exp + 22; s >= 0 {
			x1, x0 = m1<<s|m0>>(64-s), m0<<s
		} else {
			x1, x0 = rsh128(m1, m0, uint(-s))
		}
		g1, g0 := expm1Poly128(sign != 0, x1, x0)
		p3, p2, p1, p0 := mul128x128(m1, m0, g1, g0)
		return fixToFloat128(sign, p3, p2, p1, p0 != 0, exp-112-127+64)
	}

	// r = |a| - n×ln(2)/64 in fixed point with 134 fractional bits.
	// |a|×2**134 and n×ln(2)/64×2**134 may not fit in 128 bits,
	// but their difference does, so they are computed modulo 2**128.
	s := uint(exp + 22) // 14 <= s <= 35
	x1, x0 := m1<<s|m0>>(64-s), m0<<s
	l := &expm1Ln2By64Fix128
	h0, w0 := bits.Mul64(n, l[2])
	h1, w1 := bits.Mul64(n, l[1])
	w2 := n * l[0]
	var c uint64
	w1, c = bits.Add64(w1, h0, 0)
	w2 += h1 + c
	_, c = bits.Add64(w0, 1<<63, 0) // round
	w1, c = bits.Add64(w1, 0, c)
	w2 += c
	r0, c := bits.Sub64(x0, w1, 0)
	r1, _ := bits.Sub64(x1, w2, c)

	// make r positive, and apply the sign of a.
	rneg := r1>>63 != 0
	if rneg {
		r1, r0 = neg128(r1, r0)
	}
	if sign != 0 {
		rneg = !rneg
	}
	k := int(n >> 6)
	j := n & 63
	if sign != 0 {
		// n = -n
		k = -int((n + 63) >> 6)
		j = -n & 63
	}

	// e**a = 2**k × 2**(j/64) × e**r
	//      = 2**k × (t + t×p), where t = 2**(j/64), p = e**r - 1 = r × g.
	g1, g0 := expm1Poly128(rneg, r1, r0)
	p3, p2, p1, _ := mul128x128(r1, r0, g1, g0)
	p1, p0 := p3<<1|p2>>63, p2<<1|p1>>63 // p in fixed point with 134 fractional bits

	t := &expm1Table128[j]
	q3, q2, q1, _ := mul128x128(t[0], t[1], p1, p0)
	// t×p in fixed point with 191 fractional bits
	v2, v1, v0 := q3>>6, q3<<58|q2>>6, q2<<58|q1>>6
	if rneg {
		v0, c = bits.Sub64(t[2], v0, 0)
		v1, c = bits.Sub64(t[1], v1, c)
		v2, _ = bits.Sub64(t[0], v2, c)
	} else {
		v0, c = bits.Add64(t[2], v0, 0)
		v1, c = bits.Add64(t[1], v1, c)
		v2, _ = bits.Add64(t[0], v2, c)
	}

	if k >= 0 {
		// e**a - 1 = 2**k × (v - 2**-k)
		if k <= 191 {
			var s2, s1, s0 uint64
			switch b := uint(191 - k); {
			case b >= 128:
				s2 = 1 << (b - 128)
			case b >= 64:
				s1 = 1 << (b - 64)
			default:
				s0 = 1 << b
			}
			v0, c = bits.Sub64(v0, s0, 0)
			v1, c = bits.Sub64(v1, s1, c)
			v2, _ = bits.Sub64(v2, s2, c)
		}
		return fixToFloat128(0, v2, v1, v0, false, k-191)
	}

	// e**a - 1 = -(1 - 2**k × v)
	v2, v1, v0 = rsh192(v2, v1, v0, uint(-k))
	v0, c = bits.Sub64(0, v0, 0)
	v1, c = bits.Sub64(0, v1, c)
	v2, _ = bits.Sub64(1<<63, v2, c)
	return fixToFloat128(signMask128[0], v2, v1, v0, false, -191)
}

// expm1Poly128 returns g = (e**r - 1)/r in fixed point with 127 fractional bits.
// r = ±(r1:r0) × 2**-134 and |r| < 2**-7. neg is the sign of r.
func expm1Poly128(neg bool, r1, r0 uint64) (g1, g0 uint64) {
	g1, g0 = expm1Coeffs128[0][0], expm1Coeffs128[0][1]
	for _, c := range expm1Coeffs128[1:] {
		// g = c ± g×|r|
		p3, p2, _, _ := mul128x128(g1, g0, r1, r0)
		t1, t0 := p3>>6, p3<<58|p2>>6
		var b uint64
		if neg {
			g0, b = bits.Sub64(c[1], t0, 0)
			g1, _ = bits.Sub64(c[0], t1, b)
		} else {
			g0, b = bits.Add64(c[1], t0, 0)
			g1, _ = bits.Add64(c[0], t1, b)
		}
	}
	return
}

// mul128x128 returns the 256-bit product of (a1:a0) and (b1:b0).
func mul128x128(a1, a0, b1, b0 uint64) (p3, p2, p1, p0 uint64) {
	h00, l00 := bits.Mul64(a0, b0)
	h01, l01 := bits.Mul64(a0, b1)
	h10, l10 := bits.Mul64(a1, b0)
	h11, l11 := bits.Mul64(a1, b1)
	var c uint64
	p0 = l00
	p1, c = bits.Add64(h00, l01, 0)
	p2, c = bits.Add64(h01, l11, c)
	p3 = h11 + c
	p1, c = bits.Add64(p1, l10, 0)
	p2, c = bits.Add64(p2, h10, c)
	p3 += c
	return
}

// neg128 returns -(x1:x0) modulo 2**128.
func neg128(x1, x0 uint64) (uint64, uint64) {
	y0, b := bits.Sub64(0, x0, 0)
	y1, _ := bits.Sub64(0, x1, b)
	return y1, y0
}

// rsh128 returns (x1:x0) >> s for s < 128.
func rsh128(x1, x0 uint64, s uint) (uint64, uint64) {
	if s >= 64 {
		return 0, x1 >> (s - 64)
	}
	return x1 >> s, x0>>s | x1<<(64-s)
}

// rsh192 returns (x2:x1:x0) >> s for s < 192.
func rsh192(x2, x1, x0 uint64, s uint) (uint64, uint64, uint64) {
	for s >= 64 {
		x2, x1, x0 = 0, x2, x1
		s -= 64
	}
	// x << 64 is zero in Go, so s == 0 needs no special case.
	return x2 >> s, x1>>s | x2<<(64-s), x0>>s | x1<<(64-s)
}

// fixToFloat128 returns ±(v2:v1:v0) × 2**exp rounded to nearest even.
// sign is the sign bit, and sticky reports whether there are nonzero bits below v0.
// v2 must not be zero, and the result must not be subnormal.
func fixToFloat128(sign, v2, v1, v0 uint64, sticky bool, exp int) Float128 {
	// normalize so that the most significant bit of v2 is bit 63.
	lz := uint(bits.LeadingZeros64(v2))
	if lz != 0 {
		v2, v1, v0 = v2<<lz|v1>>(64-lz), v1<<lz|v0>>(64-lz), v0<<lz
	}
	exp += 191 - int(lz)

	// take the top 113 bits.
	m1, m0 := v2>>15, v2<<49|v1>>15
	half := v1&(1<<14) != 0
	rest := v1&(1<<14-1) != 0 || v0 != 0 || sticky
	if half && (rest || m0&1 != 0) {
		var c uint64
		m0, c = bits.Add64(m0, 1, 0)
		m1 += c
		if m1>>49 != 0 {
			// m = 2**113. The fraction bits are already zero.
			exp++
		}
	}

	biased := exp + bias128
	if biased >= mask128 {
		return Float128{sign | uvinf128[0], uvinf128[1]}
	}
	return Float128{sign | uint64(biased)<<(shift128-64) | m1&fracMask128[0], m0}
}
