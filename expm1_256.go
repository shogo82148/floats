package floats

import (
	"math"
	"math/bits"

	"github.com/shogo82148/ints"
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
func (a Float256) Expm1() Float256 {
	var (
		// ln(max float256 + 0.5ulp) = ln(2²⁶²¹⁴³×(2-2⁻²³⁷))
		// ~ 181704.374500706303191870897247532238261583907221734753336211540408636177
		Overflow = Float256{
			0x4001_062e_42fe_fa39, 0xef35_793c_7673_007e,
			0x5ed5_e81e_6864_ce53, 0x16c5_b141_a2eb_7175,
		}

		// e**a < 2**-238 for a < -166, so e**a - 1 rounds to -1.
		Underflow = Float256{
			0xc000_64c0_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		MinusOne = Float256{
			0xbfff_f000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)

	sign := a[0] & signMask256[0]
	exp := int((a[0]>>(shift256-192))&mask256) - bias256

	// special cases
	switch {
	case exp < -237:
		// expm1(a) ~ a when |a| < 2**-237, including ±0 and subnormal values.
		return a
	case a.IsNaN():
		return a
	case a.Gt(Overflow):
		return NewFloat256Inf(1)
	case a.Lt(Underflow):
		return MinusOne
	}

	// a = ±m × 2**(exp-236), where m is a 237-bit integer.
	m := ints.Uint256{a[0]&fracMask256[0] | 1<<(shift256-192), a[1], a[2], a[3]}

	n := expN256(exp, m)
	if n == 0 {
		// |a| < ln(2)/128. expm1(a) = a × (e**a - 1)/a.
		x := expFix256(exp, m)
		g := expm1Poly256(sign != 0, x)
		return fixToFloat256(sign, m.Mul512(g), false, exp-236-255)
	}

	k, v := expKernel256(sign, exp, m, n)
	if k >= 0 {
		// e**a - 1 = 2**k × (v - 2**-k)
		if k <= 383 {
			var t ints.Uint512
			b := uint(383 - k)
			t[7-b/64] = 1 << (b % 64)
			v = v.Sub(t)
		}
		return fixToFloat256(0, v, false, k-383)
	}

	// e**a - 1 = -(1 - 2**k × v)
	v = ints.Uint512{0, 0, 1 << 63, 0, 0, 0, 0, 0}.Sub(shr512(v, uint(-k)))
	return fixToFloat256(signMask256[0], v, false, -383)
}

// expN256 returns n = round(|a| × 64/ln(2)) for a = ±m × 2**(exp-236),
// where m is a 237-bit integer and exp <= 18.
// It returns 0 if exp < -8, that is |a| < 2**-7.
// n is computed in float64. Its error doesn't matter as long as |a - n × ln(2)/64| < 2**-7.
func expN256(exp int, m ints.Uint256) uint64 {
	if exp < -8 {
		return 0
	}
	const InvLn2By64 = 64 / math.Ln2
	top := float64(m[0]<<19 | m[1]>>45) // the top 64 bits of m
	scale := math.Float64frombits(uint64(exp-63+1023) << 52)
	return uint64(top*scale*InvLn2By64 + 0.5)
}

// expFix256 returns |a|×2**262 modulo 2**256, truncated, for a = ±m × 2**(exp-236).
func expFix256(exp int, m ints.Uint256) ints.Uint256 {
	s := exp + 26
	if s < 0 {
		return m.Rsh(uint(-s))
	}
	return m.Lsh(uint(s))
}

// expKernel256 returns e**a = 2**k × v × 2**-383 for a = ±m × 2**(exp-236),
// where sign is the sign bit of a, m is a 237-bit integer,
// and n = round(|a| × 64/ln(2)) is computed by expN256.
// v × 2**-383 is in [0.99, 2), and its absolute error is about 2**-258.
func expKernel256(sign uint64, exp int, m ints.Uint256, n uint64) (k int, v ints.Uint512) {
	return expScale256(sign, n, expReduce256(exp, m, n))
}

// expReduce256 returns r = |a| - n × ln(2)/64 for a = ±m × 2**(exp-236),
// where m is a 237-bit integer, and n = round(|a| × 64/ln(2)) is computed by expN256.
// r × 2**-262 is a signed 256-bit fixed point number with |r| < 2**-7.
func expReduce256(exp int, m ints.Uint256, n uint64) ints.Uint256 {
	// reduce: |a| = n × ln(2)/64 + r, |r| <= ln(2)/128 < 2**-7.
	// r = |a| - n×ln(2)/64 in fixed point with 262 fractional bits.
	// |a|×2**262 and n×ln(2)/64×2**262 may not fit in 256 bits,
	// but their difference does, so they are computed modulo 2**256.
	x := expFix256(exp, m)
	l := &expm1Ln2By64Fix256
	h4, w4 := bits.Mul64(n, l[4])
	h3, w3 := bits.Mul64(n, l[3])
	h2, w2 := bits.Mul64(n, l[2])
	h1, w1 := bits.Mul64(n, l[1])
	w0 := n * l[0]
	var c uint64
	w3, c = bits.Add64(w3, h4, 0)
	w2, c = bits.Add64(w2, h3, c)
	w1, c = bits.Add64(w1, h2, c)
	w0 += h1 + c
	_, c = bits.Add64(w4, 1<<63, 0) // round
	w3, c = bits.Add64(w3, 0, c)
	w2, c = bits.Add64(w2, 0, c)
	w1, c = bits.Add64(w1, 0, c)
	w0 += c
	return x.Sub(ints.Uint256{w0, w1, w2, w3})
}

// expScale256 returns e**a = 2**k × v × 2**-383
// for a = ±(n × ln(2)/64 + r), where sign is the sign bit of a,
// and r × 2**-262 is a signed 256-bit fixed point number with |r| < 2**-7.
// v × 2**-383 is in [0.99, 2), and its absolute error is about 2**-258.
func expScale256(sign, n uint64, r ints.Uint256) (k int, v ints.Uint512) {
	// make r positive, and apply the sign of a.
	rneg := r[0]>>63 != 0
	if rneg {
		r = r.Neg()
	}
	if sign != 0 {
		rneg = !rneg
	}
	k = int(n >> 6)
	j := n & 63
	if sign != 0 {
		// n = -n
		k = -int((n + 63) >> 6)
		j = -n & 63
	}

	// e**a = 2**k × 2**(j/64) × e**r
	//      = 2**k × (t + t×p), where t = 2**(j/64), p = e**r - 1 = r × g.
	g := expm1Poly256(rneg, r)
	p := shr512to256(r.Mul512(g), 255) // p in fixed point with 262 fractional bits

	t := &expm1Table256[j]
	tp := shr512(ints.Uint256{t[0], t[1], t[2], t[3]}.Mul512(p), 134) // t×p in fixed point with 383 fractional bits
	v = ints.Uint512{0, 0, t[0], t[1], t[2], t[3], t[4], t[5]}
	if rneg {
		v = v.Sub(tp)
	} else {
		v = v.Add(tp)
	}
	return
}

// expm1Poly256 returns g = (e**r - 1)/r in fixed point with 255 fractional bits.
// r = ±r × 2**-262 and |r| < 2**-7. neg is the sign of r.
func expm1Poly256(neg bool, r ints.Uint256) ints.Uint256 {
	// In Horner's method, an error of the partial sum after adding the coefficient of r**i
	// is multiplied by r**i, which is less than 2**-7.5i.
	// So the partial sums are computed with 128 bits for i >= 18,
	// with 192 bits for i >= 9, and with 256 bits for the rest,
	// keeping the error less than 2**-258.
	c := &expm1Coeffs256
	const n = len(expm1Coeffs256)

	// 128 bits: g in fixed point with 127 fractional bits, r with 134 fractional bits.
	g1, g0 := c[0][0], c[0][1]
	for i := 1; i < n-18; i++ {
		p3, p2, _, _ := mul128x128(g1, g0, r[0], r[1])
		t1, t0 := p3>>6, p3<<58|p2>>6
		var b uint64
		if neg {
			g0, b = bits.Sub64(c[i][1], t0, 0)
			g1, _ = bits.Sub64(c[i][0], t1, b)
		} else {
			g0, b = bits.Add64(c[i][1], t0, 0)
			g1, _ = bits.Add64(c[i][0], t1, b)
		}
	}

	// 192 bits: g in fixed point with 191 fractional bits, r with 198 fractional bits.
	g := [3]uint64{g1, g0, 0}
	for i := n - 18; i < n-9; i++ {
		p := mul192x192(g, [3]uint64{r[0], r[1], r[2]})
		t := [3]uint64{p[0] >> 6, p[0]<<58 | p[1]>>6, p[1]<<58 | p[2]>>6}
		var b uint64
		if neg {
			g[2], b = bits.Sub64(c[i][2], t[2], 0)
			g[1], b = bits.Sub64(c[i][1], t[1], b)
			g[0], _ = bits.Sub64(c[i][0], t[0], b)
		} else {
			g[2], b = bits.Add64(c[i][2], t[2], 0)
			g[1], b = bits.Add64(c[i][1], t[1], b)
			g[0], _ = bits.Add64(c[i][0], t[0], b)
		}
	}

	// 256 bits: g in fixed point with 255 fractional bits, r with 262 fractional bits.
	g256 := ints.Uint256{g[0], g[1], g[2], 0}
	for i := n - 9; i < n; i++ {
		t := shr512to256(g256.Mul512(r), 262)
		if neg {
			g256 = ints.Uint256(c[i]).Sub(t)
		} else {
			g256 = ints.Uint256(c[i]).Add(t)
		}
	}
	return g256
}

// mul192x192 returns the 384-bit product of a and b.
// The limbs are in the big-endian order.
func mul192x192(a, b [3]uint64) (p [6]uint64) {
	// a × b[2], a × b[1], and a × b[0]
	r2 := mul192x64(a, b[2])
	r1 := mul192x64(a, b[1])
	r0 := mul192x64(a, b[0])

	var c uint64
	p[5] = r2[3]
	p[4], c = bits.Add64(r2[2], r1[3], 0)
	p[3], c = bits.Add64(r2[1], r1[2], c)
	p[2], c = bits.Add64(r2[0], r1[1], c)
	p[1] = r1[0] + c

	p[3], c = bits.Add64(p[3], r0[3], 0)
	p[2], c = bits.Add64(p[2], r0[2], c)
	p[1], c = bits.Add64(p[1], r0[1], c)
	p[0] = r0[0] + c
	return
}

// mul192x64 returns the 256-bit product of a and b.
func mul192x64(a [3]uint64, b uint64) (r [4]uint64) {
	h2, l2 := bits.Mul64(a[2], b)
	h1, l1 := bits.Mul64(a[1], b)
	h0, l0 := bits.Mul64(a[0], b)
	var c uint64
	r[3] = l2
	r[2], c = bits.Add64(l1, h2, 0)
	r[1], c = bits.Add64(l0, h1, c)
	r[0] = h0 + c
	return
}

// shr512 returns p >> s for s < 512.
func shr512(p ints.Uint512, s uint) ints.Uint512 {
	w, q := 7-int(s/64), s%64 // the bit 0 of the result is the bit q of p[w].
	var r ints.Uint512
	for i := 7; w >= 0; i-- {
		hi := uint64(0)
		if w > 0 {
			hi = p[w-1]
		}
		r[i] = p[w]>>q | hi<<(64-q)
		w--
	}
	return r
}

// shr512to256 returns the lower 256 bits of p >> s for s <= 512.
func shr512to256(p ints.Uint512, s uint) ints.Uint256 {
	w, q := 7-int(s/64), s%64 // the bit 0 of the result is the bit q of p[w].
	var r ints.Uint256
	for i := 3; i >= 0 && w >= 0; i-- {
		hi := uint64(0)
		if w > 0 {
			hi = p[w-1]
		}
		r[i] = p[w]>>q | hi<<(64-q)
		w--
	}
	return r
}

// expOnePlus256 returns 1 ± m × 2**(exp-236) × g rounded to nearest even,
// where sign is the sign bit, m is a 237-bit integer, exp <= -8,
// and g × 2**-255 < 2.
// m × g is computed with the relative precision,
// so that the result is rounded correctly even when it is very close to 1.
func expOnePlus256(sign uint64, exp int, m, g ints.Uint256) Float256 {
	p := m.Mul512(g)

	// p = m × 2**(exp-236) × g in fixed point with 383 fractional bits.
	// The product has 236+255-exp fractional bits.
	v, sticky := rsh512Sticky(p, uint(108-exp))
	one := ints.Uint512{0, 0, 1 << 63, 0, 0, 0, 0, 0}
	if sign == 0 {
		// 1 + p
		v = v.Add(one)
	} else {
		// 1 - p = (1 - v - 1) + (1 - δ) for 0 < δ < 1, where v + δ is the true value of p.
		v = one.Sub(v)
		if sticky {
			v = v.Sub(ints.Uint512{0, 0, 0, 0, 0, 0, 0, 1})
		}
	}
	return fixToFloat256(0, v, sticky, -383)
}

// rsh512Sticky returns p >> s for s < 512.
// sticky reports whether any of the shifted out bits is nonzero.
func rsh512Sticky(p ints.Uint512, s uint) (ints.Uint512, bool) {
	var sticky bool
	for ; s >= 64; s -= 64 {
		sticky = sticky || p[7] != 0
		p = ints.Uint512{0, p[0], p[1], p[2], p[3], p[4], p[5], p[6]}
	}
	if s > 0 {
		sticky = sticky || p[7]<<(64-s) != 0
		p = ints.Uint512{
			p[0] >> s,
			p[1]>>s | p[0]<<(64-s),
			p[2]>>s | p[1]<<(64-s),
			p[3]>>s | p[2]<<(64-s),
			p[4]>>s | p[3]<<(64-s),
			p[5]>>s | p[4]<<(64-s),
			p[6]>>s | p[5]<<(64-s),
			p[7]>>s | p[6]<<(64-s),
		}
	}
	return p, sticky
}

// bit512 returns the bit i of p.
func bit512(p ints.Uint512, i uint) uint64 {
	return p[7-i/64] >> (i % 64) & 1
}

// lowBits512 reports whether any of the lowest n bits of p is nonzero, for n < 512.
func lowBits512(p ints.Uint512, n uint) bool {
	w := 7 - int(n/64)
	for i := 7; i > w; i-- {
		if p[i] != 0 {
			return true
		}
	}
	return p[w]&(1<<(n%64)-1) != 0
}

// fixToFloat256 returns ±v × 2**exp rounded to nearest even.
// sign is the sign bit, and sticky reports whether there are nonzero bits below v.
// v must not be zero.
func fixToFloat256(sign uint64, v ints.Uint512, sticky bool, exp int) Float256 {
	// the most significant bit of v is the bit 511-lz.
	lz := v.LeadingZeros()
	exp += 511 - lz

	// take the top 237 bits, or fewer bits if the result is subnormal.
	s := 511 - lz - 236
	subnormal := exp < 1-bias256
	if subnormal {
		s += 1 - bias256 - exp
	}
	var m ints.Uint256
	half, rest := false, sticky
	switch {
	case s <= 0:
		m = v.Uint256().Lsh(uint(-s))
	case s <= 512:
		m = shr512to256(v, uint(s))
		half = bit512(v, uint(s-1)) != 0
		rest = rest || lowBits512(v, uint(s-1))
	}
	if half && (rest || m[3]&1 != 0) {
		m = m.Add(ints.Uint256{0, 0, 0, 1})
		if m[0]>>45 != 0 {
			// m = 2**237. The fraction bits are already zero.
			exp++
		}
	}

	if subnormal {
		// If m is rounded up to 2**236, it becomes the smallest normal number.
		return Float256{sign | m[0], m[1], m[2], m[3]}
	}

	biased := exp + bias256
	if biased >= mask256 {
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}
	return Float256{sign | uint64(biased)<<(shift256-192) | m[0]&fracMask256[0], m[1], m[2], m[3]}
}
