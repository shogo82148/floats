package floats

import "math/bits"

// Sinh returns the hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Sinh() = ±0
//	±Inf.Sinh() = ±Inf
//	NaN.Sinh() = NaN
func (a Float128) Sinh() Float128 {
	sign := a[0] & signMask128[0]
	exp := int((a[0]>>(shift128-64))&mask128) - bias128

	switch {
	case exp < -56:
		// sinh(a) = a + a**3/6 + ... ~ a when |a| < 2**-56,
		// including ±0 and subnormal values.
		return a
	case exp >= 14:
		// sinh(a) overflows when |a| > ln(2 × max float128) ~ 11357.2.
		if a.IsNaN() {
			return a
		}
		return Float128{sign | uvinf128[0], uvinf128[1]}
	}

	// a = ±m × 2**(exp-112), where m is a 113-bit integer.
	m1 := a[0]&fracMask128[0] | 1<<(shift128-64)
	m0 := a[1]

	n := expN128(exp, m1, m0)
	if n == 0 {
		// |a| < ln(2)/128. sinh(a) = a × sinh(a)/a.
		// a×2**134 in fixed point, truncated.
		x1, x0 := expFix128(exp, m1, m0)
		z3, z2, _, _ := mul128x128(x1, x0, x1, x0) // a**2 × 2**140
		g1, g0 := sinhPoly128(sinhCoeffs128[:], z3, z2)
		p3, p2, p1, p0 := mul128x128(m1, m0, g1, g0)
		return fixToFloat128(sign, p3, p2, p1, p0 != 0, exp-112-127+64)
	}

	k, rneg, s1, s0, c1, c0, u, w2, w1, w0 := sinhKernel128(exp, m1, m0, n)

	// 2 × sinh(|a|) = e**|a| - e**-|a| = (u - w) × (1 + c) + (u + w) × (±s).
	// They are in fixed point with 191 fractional bits, relative to 2**k.
	d0, c := bits.Sub64(u[2], w0, 0)
	d1, c := bits.Sub64(u[1], w1, c)
	d2, _ := bits.Sub64(u[0], w2, c)
	_, c = bits.Add64(u[2], w0, 0)
	e1, c := bits.Add64(u[1], w1, c)
	e2, c := bits.Add64(u[0], w2, c)
	e2, e1 = c<<63|e2>>1, e2<<63|e1>>1 // (u + w) × 2**126

	// (u - w) × c
	x3, x2, x1, _ := mul128x128(d2, d1, c1, c0) // × 2**266
	x2, x1, x0 := rsh192(x3, x2, x1, 11)
	d0, c = bits.Add64(d0, x0, 0)
	d1, c = bits.Add64(d1, x1, c)
	d2, _ = bits.Add64(d2, x2, c)

	// (u + w) × s
	y3, y2, y1, _ := mul128x128(e2, e1, s1, s0) // × 2**259
	y2, y1, y0 := rsh192(y3, y2, y1, 4)
	if rneg {
		d0, c = bits.Sub64(d0, y0, 0)
		d1, c = bits.Sub64(d1, y1, c)
		d2, _ = bits.Sub64(d2, y2, c)
	} else {
		d0, c = bits.Add64(d0, y0, 0)
		d1, c = bits.Add64(d1, y1, c)
		d2, _ = bits.Add64(d2, y2, c)
	}
	return fixToFloat128(sign, d2, d1, d0, false, k-1-191)
}

// sinhKernel128 computes the parts of e**|a| and e**-|a| for a = ±m × 2**(exp-112),
// where m = (m1:m0) is a 113-bit integer, and n = round(|a| × 64/ln(2)) >= 1 is computed by expN128.
//
//	e**|a| = 2**k × u × (1 + c ± s), e**-|a| = 2**k × w × (1 + c ∓ s),
//
// where u = 2**(n/64 - k) is in [1, 2), w = 2**(-n/64 - k),
// c = cosh(r) - 1, s = |sinh(r)|, |a| = n × ln(2)/64 + r, and rneg reports whether r is negative.
// u and w = (w2:w1:w0) are in fixed point with 191 fractional bits,
// s = (s1:s0) with 133 fractional bits, and c = (c1:c0) with 139 fractional bits.
func sinhKernel128(exp int, m1, m0, n uint64) (k int, rneg bool, s1, s0, c1, c0 uint64, u *[3]uint64, w2, w1, w0 uint64) {
	// |a| = n × ln(2)/64 + r, |r| < 2**-7.
	r1, r0 := expReduce128(exp, m1, m0, n)
	rneg = r1>>63 != 0
	if rneg {
		r1, r0 = neg128(r1, r0)
	}

	// e**r = cosh(r) + sinh(r) = 1 + c + s, e**-r = 1 + c - s,
	// where c = cosh(r) - 1 = z × q, s = sinh(r) = r × p, and z = r**2.
	z3, z2, _, _ := mul128x128(r1, r0, r1, r0) // z × 2**140
	p1, p0 := sinhPoly128(sinhCoeffs128[:], z3, z2)
	q1, q0 := sinhPoly128(coshCoeffs128[:], z3, z2)
	s1, s0, _, _ = mul128x128(r1, r0, p1, p0) // s × 2**133
	c1, c0, _, _ = mul128x128(z3, z2, q1, q0) // c × 2**139

	// u = 2**(n/64 - k) = t[n mod 64], w = 2**(-n/64 - k) = t[-n mod 64] × 2**-d.
	k = int(n >> 6)
	u = &expm1Table128[n&63]
	t := &expm1Table128[-n&63]
	if d := uint(k + int((n+63)>>6)); d < 192 {
		w2, w1, w0 = rsh192(t[0], t[1], t[2], d)
	}
	return
}

// sinhPoly128 evaluates the polynomial with the coefficients coeffs at z by Horner's method.
// z = (z1:z0) × 2**-140, and the coefficients and the result are in fixed point with 127 fractional bits.
func sinhPoly128(coeffs [][2]uint64, z1, z0 uint64) (g1, g0 uint64) {
	g1, g0 = coeffs[0][0], coeffs[0][1]
	for _, c := range coeffs[1:] {
		// g = c + g×z
		p3, p2, _, _ := mul128x128(g1, g0, z1, z0)
		t1, t0 := p3>>12, p3<<52|p2>>12
		var b uint64
		g0, b = bits.Add64(c[1], t0, 0)
		g1, _ = bits.Add64(c[0], t1, b)
	}
	return
}

// Cosh returns the hyperbolic cosine of a.
//
// Special cases are:
//
//	±0.Cosh() = 1
//	±Inf.Cosh() = +Inf
//	NaN.Cosh() = NaN
func (a Float128) Cosh() Float128 {
	exp := int((a[0]>>(shift128-64))&mask128) - bias128

	switch {
	case exp < -56:
		// cosh(a) = 1 + a**2/2 + ... rounds to 1 when |a| < 2**-56,
		// including ±0 and subnormal values.
		return Float128(uvone128)
	case exp >= 14:
		// cosh(a) overflows when |a| > ln(2 × max float128) ~ 11357.2.
		if a.IsNaN() {
			return a
		}
		return Float128(uvinf128)
	}

	// |a| = m × 2**(exp-112), where m is a 113-bit integer.
	m1 := a[0]&fracMask128[0] | 1<<(shift128-64)
	m0 := a[1]

	n := expN128(exp, m1, m0)
	if n == 0 {
		// |a| < ln(2)/128. cosh(a) = 1 + c, where c = z × q and z = a**2.
		// a×2**134 in fixed point, truncated.
		x1, x0 := expFix128(exp, m1, m0)
		z3, z2, _, _ := mul128x128(x1, x0, x1, x0) // z × 2**140
		q1, q0 := sinhPoly128(coshCoeffs128[:], z3, z2)
		c1, c0, _, _ := mul128x128(z3, z2, q1, q0) // c × 2**139
		// 1 + c in fixed point with 191 fractional bits.
		// c is truncated and cosh(a) is never on a midpoint, so the sticky bit is set.
		return fixToFloat128(0, 1<<63|c1>>12, c1<<52|c0>>12, c0<<52, true, -191)
	}

	k, rneg, s1, s0, c1, c0, u, w2, w1, w0 := sinhKernel128(exp, m1, m0, n)

	// cosh(|a|) = (e**|a| + e**-|a|)/2 = (u + w)/2 × (1 + c) + (u - w)/2 × (±s).
	// They are in fixed point with 191 fractional bits, relative to 2**k.
	e0, c := bits.Add64(u[2], w0, 0)
	e1, c := bits.Add64(u[1], w1, c)
	e2, c := bits.Add64(u[0], w2, c)
	e2, e1, e0 = c<<63|e2>>1, e2<<63|e1>>1, e1<<63|e0>>1 // (u + w)/2
	_, c = bits.Sub64(u[2], w0, 0)
	d1, c := bits.Sub64(u[1], w1, c)
	d2, _ := bits.Sub64(u[0], w2, c)

	// (u + w)/2 × c
	x3, x2, x1, _ := mul128x128(e2, e1, c1, c0) // × 2**266
	x2, x1, x0 := rsh192(x3, x2, x1, 11)
	x0, c = bits.Add64(e0, x0, 0)
	x1, c = bits.Add64(e1, x1, c)
	x2, _ = bits.Add64(e2, x2, c)

	// (u - w)/2 × s
	y3, y2, y1, _ := mul128x128(d2, d1, s1, s0) // (u - w) × s × 2**260
	y2, y1, y0 := rsh192(y3, y2, y1, 6)
	if rneg {
		x0, c = bits.Sub64(x0, y0, 0)
		x1, c = bits.Sub64(x1, y1, c)
		x2, _ = bits.Sub64(x2, y2, c)
	} else {
		x0, c = bits.Add64(x0, y0, 0)
		x1, c = bits.Add64(x1, y1, c)
		x2, _ = bits.Add64(x2, y2, c)
	}
	return fixToFloat128(0, x2, x1, x0, false, k-191)
}

// Tanh returns the hyperbolic tangent of a.
//
// Special cases are:
//
//	±0.Tanh() = ±0
//	±Inf.Tanh() = ±1
//	NaN.Tanh() = NaN
func (a Float128) Tanh() Float128 {
	var (
		// Overflow = ln(max float128 + 0.5ulp)/2 = ln(2¹⁶³⁸³×(2-2⁻¹¹³))/2
		// ~ 5678.2617031470719747459655389853825
		Overflow = Float128{0x400b_62e4_2fef_a39e, 0xf357_93c7_6730_07e6}

		// One = 1.0
		One = Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}

		// Two = 2.0
		Two = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}

		// NearZero = 0.625
		NearZero = Float128{0x3ffe_4000_0000_0000, 0x0000_0000_0000_0000}
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
