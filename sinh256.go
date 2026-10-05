package floats

import (
	"math/bits"

	"github.com/shogo82148/ints"
)

// Sinh returns the hyperbolic sine of x.
//
// Special cases are:
//
//	±0.Sinh() = ±0
//	±Inf.Sinh() = ±Inf
//	NaN.Sinh() = NaN
func (a Float256) Sinh() Float256 {
	sign := a[0] & signMask256[0]
	exp := int((a[0]>>(shift256-192))&mask256) - bias256

	switch {
	case exp < -118:
		// sinh(a) = a + a**3/6 + ... ~ a when |a| < 2**-118,
		// including ±0 and subnormal values.
		return a
	case exp >= 18:
		// sinh(a) overflows when |a| > ln(2 × max float256) ~ 181705.1.
		if a.IsNaN() {
			return a
		}
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}

	// a = ±m × 2**(exp-236), where m is a 237-bit integer.
	m := ints.Uint256{a[0]&fracMask256[0] | 1<<(shift256-192), a[1], a[2], a[3]}

	n := expN256(exp, m)
	if n == 0 {
		// |a| < ln(2)/128. sinh(a) = a × sinh(a)/a.
		x := expFix256(exp, m)
		z := shr512to256(x.Mul512(x), 256) // a**2 × 2**268
		g := sinhPoly256(&sinhCoeffs256, z)
		return fixToFloat256(sign, m.Mul512(g), false, exp-236-255)
	}

	k, rneg, s, c, u, w := sinhKernel256(exp, m, n)

	// 2 × sinh(|a|) = e**|a| - e**-|a| = (u - w) × (1 + c) + (u + w) × (±s).
	// They are in fixed point with 383 fractional bits, relative to 2**k.
	d := u.Sub(w)
	e := u.Add(w)
	v := d.Add(shr512(shr512to256(d, 128).Mul512(c), 139)) // (u - w) × c
	es := shr512(shr512to256(e, 129).Mul512(s), 132)       // (u + w) × s
	if rneg {
		v = v.Sub(es)
	} else {
		v = v.Add(es)
	}
	return fixToFloat256(sign, v, false, k-1-383)
}

// sinhKernel256 computes the parts of e**|a| and e**-|a| for a = ±m × 2**(exp-236),
// where m is a 237-bit integer, and n = round(|a| × 64/ln(2)) >= 1 is computed by expN256.
//
//	e**|a| = 2**k × u × (1 + c ± s), e**-|a| = 2**k × w × (1 + c ∓ s),
//
// where u = 2**(n/64 - k) is in [1, 2), w = 2**(-n/64 - k),
// c = cosh(r) - 1, s = |sinh(r)|, |a| = n × ln(2)/64 + r, and rneg reports whether r is negative.
// u and w are in fixed point with 383 fractional bits,
// s with 261 fractional bits, and c with 267 fractional bits.
func sinhKernel256(exp int, m ints.Uint256, n uint64) (k int, rneg bool, s, c ints.Uint256, u, w ints.Uint512) {
	// |a| = n × ln(2)/64 + r, |r| < 2**-7.
	r := expReduce256(exp, m, n)
	rneg = r[0]>>63 != 0
	if rneg {
		r = r.Neg()
	}

	// e**r = cosh(r) + sinh(r) = 1 + c + s, e**-r = 1 + c - s,
	// where c = cosh(r) - 1 = z × q, s = sinh(r) = r × p, and z = r**2.
	z := shr512to256(r.Mul512(r), 256) // z × 2**268
	p, q := sinhCoshPoly256(z)
	s = shr512to256(r.Mul512(p), 256) // s × 2**261
	c = shr512to256(z.Mul512(q), 256) // c × 2**267

	// u = 2**(n/64 - k) = t[n mod 64], w = 2**(-n/64 - k) = t[-n mod 64] × 2**-d.
	k = int(n >> 6)
	t := &expm1Table256[n&63]
	u = ints.Uint512{0, 0, t[0], t[1], t[2], t[3], t[4], t[5]}
	if d := uint(k + int((n+63)>>6)); d < 384 {
		t := &expm1Table256[-n&63]
		w = shr512(ints.Uint512{0, 0, t[0], t[1], t[2], t[3], t[4], t[5]}, d)
	}
	return
}

// sinhPoly256 evaluates the polynomial with the coefficients coeffs at z by Horner's method.
// z × 2**-268 < 2**-15, and the coefficients and the result are in fixed point with 255 fractional bits.
func sinhPoly256(coeffs *[12][4]uint64, z ints.Uint256) ints.Uint256 {
	// In Horner's method, an error of the partial sum after adding the coefficient of z**i
	// is multiplied by z**i, which is less than 2**-15i.
	// So the partial sums are computed with 128 bits for i >= 9,
	// with 192 bits for i >= 5, and with 256 bits for the rest,
	// keeping the error less than 2**-260.
	const n = len(coeffs)

	// 128 bits: g in fixed point with 127 fractional bits, z with 140 fractional bits.
	g1, g0 := coeffs[0][0], coeffs[0][1]
	for i := 1; i < n-9; i++ {
		p3, p2, _, _ := mul128x128(g1, g0, z[0], z[1])
		t1, t0 := p3>>12, p3<<52|p2>>12
		var b uint64
		g0, b = bits.Add64(coeffs[i][1], t0, 0)
		g1, _ = bits.Add64(coeffs[i][0], t1, b)
	}

	// 192 bits: g in fixed point with 191 fractional bits, z with 204 fractional bits.
	g := [3]uint64{g1, g0, 0}
	for i := n - 9; i < n-5; i++ {
		p := mul192x192(g, [3]uint64{z[0], z[1], z[2]})
		t := [3]uint64{p[0] >> 12, p[0]<<52 | p[1]>>12, p[1]<<52 | p[2]>>12}
		var b uint64
		g[2], b = bits.Add64(coeffs[i][2], t[2], 0)
		g[1], b = bits.Add64(coeffs[i][1], t[1], b)
		g[0], _ = bits.Add64(coeffs[i][0], t[0], b)
	}

	// 256 bits: g in fixed point with 255 fractional bits, z with 268 fractional bits.
	g256 := ints.Uint256{g[0], g[1], g[2], 0}
	for i := n - 5; i < n; i++ {
		g256 = ints.Uint256(coeffs[i]).Add(mulShr268(g256, z))
	}
	return g256
}

// sinhCoshPoly256 returns p = sinh(r)/r and q = (cosh(r) - 1)/r**2 at z = r**2.
// It evaluates the two polynomials of sinhPoly256 together,
// so that the independent multiplications can run in parallel.
func sinhCoshPoly256(z ints.Uint256) (p, q ints.Uint256) {
	const n = len(sinhCoeffs256)
	cp, cq := &sinhCoeffs256, &coshCoeffs256

	// 128 bits: in fixed point with 127 fractional bits, z with 140 fractional bits.
	p1, p0 := cp[0][0], cp[0][1]
	q1, q0 := cq[0][0], cq[0][1]
	for i := 1; i < n-9; i++ {
		x3, x2, _, _ := mul128x128(p1, p0, z[0], z[1])
		y3, y2, _, _ := mul128x128(q1, q0, z[0], z[1])
		var b uint64
		p0, b = bits.Add64(cp[i][1], x3<<52|x2>>12, 0)
		p1, _ = bits.Add64(cp[i][0], x3>>12, b)
		q0, b = bits.Add64(cq[i][1], y3<<52|y2>>12, 0)
		q1, _ = bits.Add64(cq[i][0], y3>>12, b)
	}

	// 192 bits: in fixed point with 191 fractional bits, z with 204 fractional bits.
	z192 := [3]uint64{z[0], z[1], z[2]}
	gp := [3]uint64{p1, p0, 0}
	gq := [3]uint64{q1, q0, 0}
	for i := n - 9; i < n-5; i++ {
		x := mul192x192(gp, z192)
		y := mul192x192(gq, z192)
		var b uint64
		gp[2], b = bits.Add64(cp[i][2], x[1]<<52|x[2]>>12, 0)
		gp[1], b = bits.Add64(cp[i][1], x[0]<<52|x[1]>>12, b)
		gp[0], _ = bits.Add64(cp[i][0], x[0]>>12, b)
		gq[2], b = bits.Add64(cq[i][2], y[1]<<52|y[2]>>12, 0)
		gq[1], b = bits.Add64(cq[i][1], y[0]<<52|y[1]>>12, b)
		gq[0], _ = bits.Add64(cq[i][0], y[0]>>12, b)
	}

	// 256 bits: in fixed point with 255 fractional bits, z with 268 fractional bits.
	p = ints.Uint256{gp[0], gp[1], gp[2], 0}
	q = ints.Uint256{gq[0], gq[1], gq[2], 0}
	for i := n - 5; i < n; i++ {
		x := mulShr268(p, z)
		y := mulShr268(q, z)
		p = ints.Uint256(cp[i]).Add(x)
		q = ints.Uint256(cq[i]).Add(y)
	}
	return
}

// mulShr268 returns about (a × b) >> 268, truncated to 256 bits.
// See mulHi256 for its error.
func mulShr268(a, b ints.Uint256) ints.Uint256 {
	h := mulHi256(a, b)
	return ints.Uint256{h[0] >> 12, h[0]<<52 | h[1]>>12, h[1]<<52 | h[2]>>12, h[2]<<52 | h[3]>>12}
}

// mulHi256 returns about (a × b) >> 256.
// It skips the partial products that only affect the lower 256 bits except for carries,
// so the result may be less than the exact value by at most 7.
func mulHi256(a, b ints.Uint256) ints.Uint256 {
	// a = a[0]:a[1]:a[2]:a[3] in big-endian. The product of a[i] and b[j] is at 2**(64×(6-i-j)).
	// at 2**192: only the higher halves are added, to 2**256.
	h30, _ := bits.Mul64(a[3], b[0])
	h21, _ := bits.Mul64(a[2], b[1])
	h12, _ := bits.Mul64(a[1], b[2])
	h03, _ := bits.Mul64(a[0], b[3])
	// at 2**256
	h20, l20 := bits.Mul64(a[2], b[0])
	h11, l11 := bits.Mul64(a[1], b[1])
	h02, l02 := bits.Mul64(a[0], b[2])
	// at 2**320
	h10, l10 := bits.Mul64(a[1], b[0])
	h01, l01 := bits.Mul64(a[0], b[1])
	// at 2**384
	h00, l00 := bits.Mul64(a[0], b[0])

	var c, c1, c2 uint64
	// the word at 2**256
	r3, c := bits.Add64(h30, h21, 0)
	c1 += c
	r3, c = bits.Add64(r3, h12, 0)
	c1 += c
	r3, c = bits.Add64(r3, h03, 0)
	c1 += c
	r3, c = bits.Add64(r3, l20, 0)
	c1 += c
	r3, c = bits.Add64(r3, l11, 0)
	c1 += c
	r3, c = bits.Add64(r3, l02, 0)
	c1 += c
	// the word at 2**320
	r2, c := bits.Add64(h20, h11, 0)
	c2 += c
	r2, c = bits.Add64(r2, h02, 0)
	c2 += c
	r2, c = bits.Add64(r2, l10, 0)
	c2 += c
	r2, c = bits.Add64(r2, l01, 0)
	c2 += c
	r2, c = bits.Add64(r2, c1, 0)
	c2 += c
	// the word at 2**384
	r1, c := bits.Add64(h10, h01, 0)
	r0 := h00 + c
	r1, c = bits.Add64(r1, l00, 0)
	r0 += c
	r1, c = bits.Add64(r1, c2, 0)
	r0 += c
	return ints.Uint256{r0, r1, r2, r3}
}

// Cosh returns the hyperbolic cosine of x.
//
// Special cases are:
//
//	±0.Cosh() = 1
//	±Inf.Cosh() = +Inf
//	NaN.Cosh() = NaN
func (a Float256) Cosh() Float256 {
	exp := int((a[0]>>(shift256-192))&mask256) - bias256

	switch {
	case exp < -118:
		// cosh(a) = 1 + a**2/2 + ... rounds to 1 when |a| < 2**-118,
		// including ±0 and subnormal values.
		return Float256(uvone256)
	case exp >= 18:
		// cosh(a) overflows when |a| > ln(2 × max float256) ~ 181705.1.
		if a.IsNaN() {
			return a
		}
		return Float256(uvinf256)
	}

	// |a| = m × 2**(exp-236), where m is a 237-bit integer.
	m := ints.Uint256{a[0]&fracMask256[0] | 1<<(shift256-192), a[1], a[2], a[3]}

	n := expN256(exp, m)
	if n == 0 {
		// |a| < ln(2)/128. cosh(a) = 1 + c, where c = z × q and z = a**2.
		x := expFix256(exp, m)
		z := shr512to256(x.Mul512(x), 256) // z × 2**268
		q := sinhPoly256(&coshCoeffs256, z)
		c := mulHi256(z, q) // c × 2**267, truncated
		// 1 + c in fixed point with 383 fractional bits.
		// c is truncated and cosh(a) is never on a midpoint, so the sticky bit is set.
		v := lsh512(ints.Uint512{0, 0, 0, 0, c[0], c[1], c[2], c[3]}, 383-267)
		v[2] |= 1 << 63
		return fixToFloat256(0, v, true, -383)
	}

	k, rneg, s, c, u, w := sinhKernel256(exp, m, n)

	// cosh(|a|) = (e**|a| + e**-|a|)/2 = (u + w)/2 × (1 + c) + (u - w)/2 × (±s).
	// They are in fixed point with 383 fractional bits, relative to 2**k.
	e := shr512(u.Add(w), 1)
	f := shr512(u.Sub(w), 1)
	v := e.Add(shr512(shr512to256(e, 128).Mul512(c), 139)) // (u + w)/2 × c
	fs := shr512(shr512to256(f, 128).Mul512(s), 133)       // (u - w)/2 × s
	if rneg {
		v = v.Sub(fs)
	} else {
		v = v.Add(fs)
	}
	return fixToFloat256(0, v, false, k-383)
}

// Tanh returns the hyperbolic tangent of a.
//
// Special cases are:
//
//	±0.Tanh() = ±0
//	±Inf.Tanh() = ±1
//	NaN.Tanh() = NaN
func (a Float256) Tanh() Float256 {
	var (
		// ln(max float256 + 0.5ulp)/2 = ln(2²⁶²¹⁴³×(2-2⁻²³⁷))/2
		// ~ 90852.18725035315159593544862376611913079195361086737666810577020431809
		Overflow = Float256{
			0x4000_f62e_42fe_fa39, 0xef35_793c_7673_007e,
			0x5ed5_e81e_6864_ce53, 0x16c5_b141_a2eb_7177,
		}

		// One = 1.0
		One = Float256{
			0x3fff_f000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// Two = 2.0
		Two = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// NearZero = 0.625
		NearZero = Float256{
			0x3fff_e400_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)
	z := a.Abs()
	switch {
	case z.Gt(Overflow):
		if a.Signbit() {
			return One.Neg()
		}
		return One
	case z.Ge(NearZero):
		s := z.Add(z).Exp()
		z = One.Sub(Two.Quo(s.Add(One)))
		if a.Signbit() {
			z = z.Neg()
		}
	case z.IsZero():
		return a
	default:
		// TODO: optimize using minimax approximation
		z = z.Sinh().Quo(z.Cosh())
		if a.Signbit() {
			z = z.Neg()
		}
	}
	return z
}
