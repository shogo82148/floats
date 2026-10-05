package floats

import (
	"math"
	"math/bits"

	"github.com/shogo82148/ints"
)

// Cbrt returns the cube root of a.
//
// Special cases are:
//
//	±0.Cbrt() = ±0
//	±Inf.Cbrt() = ±Inf
//	NaN.Cbrt() = NaN
func (a Float128) Cbrt() Float128 {
	// special cases
	switch {
	case a.IsZero() || a.IsInf(0) || a.IsNaN():
		return a
	}

	// |a| = m × 2**(exp-112), where m is a 113-bit integer in [2**112, 2**113).
	sign, exp, m := a.normalize()

	// |a| = 2**(3q) × X, where X = mx × 2**-112 is in [1, 8), and exp = 3q + r with r = 0, 1, 2.
	// cbrt(a) = 2**q × cbrt(X), and cbrt(X) is in [1, 2).
	// exp + 16503 is positive and divisible by 3 for the shift, so that the quotient is floored.
	t := exp + 16503
	q, r := t/3-5501, uint(t%3)
	mx1, mx0 := m[0]<<r|m[1]>>(64-r), m[1]<<r // mx = X × 2**112, 113 to 115 bits

	// the estimate T × 2**-52 of cbrt(X) with an error of about 2**-52, from math.Cbrt.
	xf := float64(mx1<<13|mx0>>51) * 0x1p-61
	est := uint64(math.Cbrt(xf) * (1 << 52))

	// Halley's iteration: t' = t - t × (t³ - X) / (2t³ + X) converges cubically,
	// so that one step makes the error much smaller than 2**-127.
	// t³ = T³ × 2**-156 and X = xs × 2**-156, where xs = mx × 2**44, are computed exactly.
	xs2, xs1, xs0 := mx1>>20, mx1<<44|mx0>>20, mx0<<44
	s1, s0 := bits.Mul64(est, est)
	l1, l0 := bits.Mul64(s0, est)
	h1, h0 := bits.Mul64(s1, est)
	c1, carry := bits.Add64(l1, h0, 0)
	c2, c0 := h1+carry, l0 // c = T³

	// d = |T³ - xs|. T is too large if T³ >= xs.
	over := c2 > xs2 || c2 == xs2 && (c1 > xs1 || c1 == xs1 && c0 >= xs0)
	var d1, d0, b uint64
	if over {
		d0, b = bits.Sub64(c0, xs0, 0)
		d1, _ = bits.Sub64(c1, xs1, b)
	} else {
		d0, b = bits.Sub64(xs0, c0, 0)
		d1, _ = bits.Sub64(xs1, c1, b)
	}
	// d < 2**110, so its upper limb is zero.

	// den = 2T³ + xs
	var k uint64
	den0, k := bits.Add64(c0<<1, xs0, 0)
	den1, k := bits.Add64(c1<<1|c0>>63, xs1, k)
	den2, _ := bits.Add64(c2<<1|c1>>63, xs2, k)

	// the correction T × d / den × 2**75 is about 2**75. It is computed by two steps of the floating-point division.
	// p = T × d
	e1, e0 := bits.Mul64(d0, est)
	f1, f0 := bits.Mul64(d1, est)
	p0 := e0
	p1, carry := bits.Add64(e1, f0, 0)
	p2 := f1 + carry
	denf := float64(den2)*0x1p128 + float64(den1)*0x1p64 + float64(den0)
	pf := float64(p2)*0x1p128 + float64(p1)*0x1p64 + float64(p0)
	u := uint64(pf / denf * 0x1p51) // the first estimate of the correction is u × 2**24

	// the residual p × 2**75 - u × 2**24 × den is less than 2**192 in the absolute value.
	// It is only needed with the absolute error of about den/16 ~ 2**157, so that the lower 128 bits
	// of the products and den0 are ignored.
	a0, a1 := p2<<11|p1>>53, p1<<11|p0>>53 // the upper 128 bits of p × 2**75
	u1h, u1l := bits.Mul64(den1, u)
	u2h, u2l := bits.Mul64(den2, u)
	b2, carry := bits.Add64(u1h, u2l, 0)
	b3 := u2h + carry
	B0, B1 := b3<<24|b2>>40, b2<<24|u1l>>40 // the upper 128 bits of u × den × 2**24
	r1, b := bits.Sub64(a1, B1, 0)
	r0, b := bits.Sub64(a0, B0, b)
	neg := b != 0
	if neg {
		r1, b = bits.Sub64(0, r1, 0)
		r0, _ = bits.Sub64(0, r0, b)
	}
	rf := float64(r0)*0x1p192 + float64(r1)*0x1p128
	n := uint64(rf/denf + 0.5) // the correction of the residual, which is less than 2**28

	// corr = u × 2**24 ± n
	corr1, corr0 := u>>40, u<<24
	if neg {
		corr0, b = bits.Sub64(corr0, n, 0)
		corr1 -= b
	} else {
		corr0, carry = bits.Add64(corr0, n, 0)
		corr1 += carry
	}

	// y = cbrt(X) × 2**127 in [2**127, 2**128), with an error less than 1.
	var y1, y0 uint64
	if over {
		y0, b = bits.Sub64(0, corr0, 0)
		y1, _ = bits.Sub64(est<<11, corr1, b)
	} else {
		y0, carry = bits.Add64(0, corr0, 0)
		y1, _ = bits.Add64(est<<11, corr1, carry)
	}

	// round y to 113 bits: the result significand is in [2**112, 2**113].
	// If the 15 bits below are close to the half, the exact comparison is needed.
	sig1, sig0 := y1>>15, y1<<49|y0>>15
	switch low := y0 & 0x7fff; {
	case low > 0x4000+4:
		sig0, carry = bits.Add64(sig0, 1, 0)
		sig1 += carry
	case low >= 0x4000-4:
		// round up if (sig + 1/2)³ < X × 2**336, i.e., (2 sig + 1)³ < mx × 2**227 (no tie is possible).
		s := ints.Uint512{6: sig1, 7: sig0}.Lsh(1).Add(ints.Uint512{7: 1})
		cube := s.Mul(s).Mul(s)
		if (ints.Uint512{6: mx1, 7: mx0}).Lsh(227).Cmp(cube) > 0 {
			sig0, carry = bits.Add64(sig0, 1, 0)
			sig1 += carry
		}
	}

	// sign, exponent, and fraction. If the significand is rounded up to 2**113, it carries to the exponent.
	hi := sign | uint64(q+bias128)<<(shift128-64)
	hi += sig1 - 1<<(shift128-64)
	return Float128{hi, sig0}
}
