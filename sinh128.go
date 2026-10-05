package floats

import (
	"math/bits"

	"github.com/shogo82148/ints"
)

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
	// They are in fixed point with 191 fractional bits, relative to 2**k,
	// which cancels out in the division.
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
	sign := a[0] & signMask128[0]
	exp := int((a[0]>>(shift128-64))&mask128) - bias128

	switch {
	case exp < -57:
		// tanh(a) = a - a**3/3 + ... ~ a when |a| < 2**-57,
		// including ±0 and subnormal values.
		return a
	case exp >= 6:
		// 1 - tanh(|a|) ~ 2e**-2|a| < 2**-114 when |a| > 115×ln(2)/2 ~ 39.9,
		// so tanh(a) rounds to ±1.
		if a.IsNaN() {
			return a
		}
		return Float128{sign | uvone128[0], uvone128[1]}
	}

	// a = ±m × 2**(exp-112), where m is a 113-bit integer.
	m1 := a[0]&fracMask128[0] | 1<<(shift128-64)
	m0 := a[1]

	n := expN128(exp, m1, m0)
	if n == 0 {
		// |a| < ln(2)/128. tanh(a) = a × tanh(a)/a.
		// a×2**134 in fixed point, truncated.
		x1, x0 := expFix128(exp, m1, m0)
		z3, z2, _, _ := mul128x128(x1, x0, x1, x0) // a**2 × 2**140
		g1, g0 := tanhPoly128(z3, z2)
		p3, p2, p1, p0 := mul128x128(m1, m0, g1, g0)
		return fixToFloat128(sign, p3, p2, p1, p0 != 0, exp-112-127+64)
	}

	// e**2|a| = 2**k × v, where v = (v2:v1:v0) × 2**-191 is in [0.99, 2).
	// 2|a| = m × 2**(exp+1-112), and k >= 0 because |a| >= ln(2)/128.
	k, v2, v1, v0 := expKernel128(0, exp+1, m1, m0, expN128(exp+1, m1, m0))
	if k == 0 {
		// e**2|a| < 2, so tanh(|a|) < 1/3.
		// tanh(|a|) = (e**2|a| - 1) / (e**2|a| + 1) = (v - 1) / (2 × (v + 1)/2).
		// v - 1 > 0 because e**2|a| >= e**(ln(2)/64).
		q1, q0, e, inexact := quo192(v2-1<<63, v1, v0, v2>>1+1<<62, v2<<63|v1>>1, v1<<63|v0>>1)
		return fixToFloat128(sign, q1, q0, 0, inexact, e-1-128-64)
	}

	// tanh(|a|) = 1 - ε, where ε = 2 / (e**2|a| + 1) = 2**-k / h, and h = (v + 2**-k)/2.
	// ε is computed with a small relative error,
	// so that 1 - ε rounds correctly even when it is very close to a midpoint.
	// h is in [0.49, 1.25) in fixed point with 191 fractional bits. k <= 184 because |a| < 64.
	h2, h1, h0 := v2>>1, v2<<63|v1>>1, v1<<63|v0>>1
	var c uint64
	switch b := uint(190 - k); {
	case b >= 128:
		h2 += 1 << (b - 128)
	case b >= 64:
		h1, c = bits.Add64(h1, 1<<(b-64), 0)
		h2 += c
	default:
		h0, c = bits.Add64(h0, 1<<b, 0)
		h1, c = bits.Add64(h1, 0, c)
		h2 += c
	}
	// h = (d1:d0) × 2**(-127-ld), normalized and truncated to 128 bits.
	ld := uint(bits.LeadingZeros64(h2))
	d1, d0 := h2<<ld|h1>>(64-ld), h1<<ld|h0>>(64-ld)
	// q = 2**254/d = 2**(-127-ld)/h, so ε = q × 2**(ld-k-127).
	q, inexact := quo256by128(ints.Uint128{1 << 62, 0}, ints.Uint128{d1, d0})

	// ε in fixed point with 191 fractional bits, rounded up,
	// so that the result 1 - ε is truncated and the sticky bit is valid.
	// The division is exact only if d and q are powers of two,
	// and then no nonzero bits are shifted out because sh <= 184,
	// so the shifted ε is exact if and only if the division is exact.
	sticky := inexact
	e2, e1, e0 := rsh192(q[0], q[1], 0, uint(k)-ld)
	if sticky {
		e0, c = bits.Add64(e0, 1, 0)
		e1, c = bits.Add64(e1, 0, c)
		e2 += c
	}
	r0, c := bits.Sub64(0, e0, 0)
	r1, c := bits.Sub64(0, e1, c)
	r2, _ := bits.Sub64(1<<63, e2, c)
	return fixToFloat128(sign, r2, r1, r0, sticky, -191)
}

// quo192 returns y/x = (q1:q0) × 2**(e-128) for positive 192-bit fixed point numbers
// y = (y2:y1:y0) and x = (x2:x1:x0) with the same scale.
// y2 and x2 must not be zero. inexact reports whether the quotient of the normalized 128-bit
// operands is not exact; the bits below them are ignored.
func quo192(y2, y1, y0, x2, x1, x0 uint64) (q1, q0 uint64, e int, inexact bool) {
	// normalize x and y so that their most significant bits are set.
	lx := uint(bits.LeadingZeros64(x2))
	x2, x1 = x2<<lx|x1>>(64-lx), x1<<lx|x0>>(64-lx)
	ly := uint(bits.LeadingZeros64(y2))
	y2, y1 = y2<<ly|y1>>(64-ly), y1<<ly|y0>>(64-ly)
	e = int(lx) - int(ly)
	if y2 > x2 || (y2 == x2 && y1 >= x1) {
		y2, y1 = y2>>1, y2<<63|y1>>1
		e++
	}
	q, inexact := quo256by128(ints.Uint128{y2, y1}, ints.Uint128{x2, x1})
	return q[0], q[1], e, inexact
}

// tanhPoly128 returns tanh(r)/r in fixed point with 127 fractional bits,
// where z = r**2 = (z1:z0) × 2**-140 and |r| < 2**-7.
func tanhPoly128(z1, z0 uint64) (g1, g0 uint64) {
	g1, g0 = tanhCoeffs128[0][0], tanhCoeffs128[0][1]
	for _, c := range tanhCoeffs128[1:] {
		// g = c - g×z. The coefficients alternate in sign, and g×z < c.
		p3, p2, _, _ := mul128x128(g1, g0, z1, z0)
		t1, t0 := p3>>12, p3<<52|p2>>12
		var b uint64
		g0, b = bits.Sub64(c[1], t0, 0)
		g1, _ = bits.Sub64(c[0], t1, b)
	}
	return
}
