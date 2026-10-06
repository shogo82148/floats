package floats

import (
	"math/bits"

	"github.com/shogo82148/ints"
)

// Gamma returns the Gamma function of a.
//
// Special cases are:
//
//	+Inf.Gamma() = +Inf
//	+0.Gamma() = +Inf
//	-0.Gamma() = -Inf
//	x.Gamma() = NaN for integer x < 0
//	-Inf.Gamma() = NaN
//	NaN.Gamma() = NaN
func (a Float128) Gamma() Float128 {
	// special cases
	switch {
	case isNegInt128(a) || a.IsInf(-1) || a.IsNaN():
		return NewFloat128NaN()
	case a.IsInf(1):
		return NewFloat128Inf(1)
	case a.IsZero():
		return NewFloat128Inf(1).Copysign(a)
	}

	// |a| = m × 2**(exp-112), where m = (m1:m0) is a 113-bit integer in [2**112, 2**113).
	sign, exp, m := a.normalize()
	neg := sign != 0

	if exp >= 11 {
		// |a| >= 2048
		if !neg {
			return NewFloat128Inf(1) // Gamma(x) overflows for x > 1755.6.
		}
		// Gamma(x) underflows for x < -1800, and its sign is that of sin(pi x) = (-1)**floor(x) sin(pi (x - floor(x))).
		if isOddInt128(a.Floor()) {
			return NewFloat128(0).Neg()
		}
		return Float128{}
	}

	// the sign of the result, and the mantissa and the exponent: |Gamma(a)| = mant × 2**(e-192).
	var mant ints.Uint256
	var e int
	if exp < -40 {
		// |a| < 2**-40. Gamma(x) = Gamma(1+x)/x, and Gamma(1+x) = g0 - g1 x + g2 x**2 - ... is calculated by the series.
		mant, e = gammaTiny128(neg, exp, m)
	} else if !neg {
		if exp == 10 && a.Ge(Float128{0x4009_b700_0000_0000, 0}) {
			return NewFloat128Inf(1) // x >= 1756
		}
		mant, e = gammaPos128(m.Uint256().Lsh(uint(80 + exp)))
	} else {
		sign, mant, e = gammaNeg128(exp, m)
	}

	// mant is in [2**192, 2**193).
	return fixToFloat128(sign, mant[1]>>1|mant[0]<<63, mant[2]>>1|mant[1]<<63, mant[3]>>1|mant[2]<<63, mant[3]&1 != 0, e-191)
}

// The numbers of Gamma are in fixed point with 192 fractional bits in ints.Uint256,
// whose first word is the integer part.

// gammaOne is 1.
var gammaOne = ints.Uint256{1, 0, 0, 0}

// gammaMul returns a × b, truncated.
func gammaMul(a, b ints.Uint256) ints.Uint256 {
	return shr512to256(a.Mul512(b), 192)
}

// gammaMul3 returns about a × b from the upper three words of a and b, that is, with 128 fractional bits.
// The absolute error is about 2**-127 × (|a| + |b|).
func gammaMul3(a, b ints.Uint256) ints.Uint256 {
	p := mul192x192([3]uint64{a[0], a[1], a[2]}, [3]uint64{b[0], b[1], b[2]})
	return ints.Uint256{p[1], p[2], p[3], p[4]}
}

// gammaStep64 returns about p × w in fixed point with 64 fractional bits, where p = (p0:p1) is in fixed point with
// 64 fractional bits and w = w1 × 2**-64 < 1.
func gammaStep64(p0, p1, w1 uint64) (hi, lo uint64) {
	hi, lo = bits.Mul64(p0, w1)
	h, _ := bits.Mul64(p1, w1)
	lo, c := bits.Add64(lo, h, 0)
	return hi + c, lo
}

// gammaStep128 returns about p × w in fixed point with 128 fractional bits, where p = (p0:p1:p2) is in fixed point with
// 128 fractional bits and w = (w1:w2) × 2**-128 < 1.
func gammaStep128(p0, p1, p2, w1, w2 uint64) (r0, r1, r2 uint64) {
	// the product is a + b × 2**64, and the lower two words are dropped except for the carry.
	a := mul192x64([3]uint64{p0, p1, p2}, w2)
	b := mul192x64([3]uint64{p0, p1, p2}, w1)
	_, c := bits.Add64(a[2], b[3], 0)
	r2, c = bits.Add64(a[1], b[2], c)
	r1, c = bits.Add64(a[0], b[1], c)
	r0 = b[0] + c
	return
}

// gammaNormalize normalizes (v, e), which represents the number v × 2**(e-192), to v in [2**192, 2**193).
func gammaNormalize(v ints.Uint256, e int) (ints.Uint256, int) {
	sh := v.BitLen() - 193
	if sh >= 0 {
		return v.Rsh(uint(sh)), e + sh
	}
	return v.Lsh(uint(-sh)), e + sh
}

// gammaTiny128 returns |Gamma(x)| = mant × 2**(e-192) for x = ±m × 2**(exp-112), where m is a 113-bit integer, and exp < -40.
func gammaTiny128(neg bool, exp int, m ints.Uint128) (mant ints.Uint256, e int) {
	// |x| in fixed point with 192 fractional bits.
	var x ints.Uint256
	if s := exp + 80; s >= 0 {
		x = m.Uint256().Lsh(uint(s))
	} else if s > -256 {
		x = m.Uint256().Rsh(uint(-s))
	}

	// Gamma(1+x) = sum g_k (-x)**k by Horner's method. g_k is |g_k|(-1)**k, and x < 2**-40.
	n := len(gamma128Gamma1)
	p := ints.Uint256(gamma128Gamma1[n-1])
	for k := n - 2; k >= 0; k-- {
		t := gammaMul(p, x)
		if neg {
			p = ints.Uint256(gamma128Gamma1[k]).Add(t)
		} else {
			p = ints.Uint256(gamma128Gamma1[k]).Sub(t)
		}
	}

	// Gamma(x) = Gamma(1+x)/x, where x = (m × 2**80) × 2**(exp-192).
	q := p.Uint512().Lsh(192).Quo(m.Uint512().Lsh(80))
	return gammaNormalize(q.Uint256(), -exp)
}

// gammaPos128 returns Gamma(y) = mant × 2**(e-192) for y = Y × 2**-192, 2**-40 <= y < 2049, where mant is in [2**192, 2**193).
func gammaPos128(y ints.Uint256) (mant ints.Uint256, e int) {
	z, pm, pe := gammaRecur128(y)

	// log(z) is not accurate for 1 <= z/2**k < 1 + 2**-8, so z is moved out of the range.
	for gammaLogBucket0(z) {
		pm, pe = gammaNormalize(gammaMul(pm, z), pe)
		z = z.Add(gammaOne)
	}

	em, ee := gammaExp192(gammaLogGamma192(z))
	q := em.Uint512().Lsh(192).Quo(pm.Uint512())
	return gammaNormalize(q.Uint256(), ee-pe)
}

// gammaRecur128 returns z >= 24 and P = pm × 2**(pe-192) = y (y+1) ... (z-1) for y = Y × 2**-192,
// where pm is in [2**192, 2**193). Gamma(y) = Gamma(z)/P.
func gammaRecur128(y ints.Uint256) (z, pm ints.Uint256, pe int) {
	n := 0
	if y[0] < 24 {
		n = 24 - int(y[0])
	}
	pm, pe = gammaOne, 0
	y2 := gammaMul(y, y)
	j := 0
	for ; j+4 <= n; j += 4 {
		// The product of the four factors is u (u+2), where u = y (y+3) + j (2y+j+3) = (y+j)(y+j+3).
		u := y2.Add(y.Mul(ints.Uint256{3: uint64(2*j + 3)})).Add(ints.Uint256{uint64(j * (j + 3)), 0, 0, 0})
		g := gammaMul(u, u).Add(u).Add(u)
		pm, pe = gammaNormalize(gammaMul(pm, g), pe)
	}
	z = y.Add(ints.Uint256{uint64(j), 0, 0, 0})
	for ; j < n; j++ {
		pm, pe = gammaNormalize(gammaMul(pm, z), pe)
		z = z.Add(gammaOne)
	}
	return z, pm, pe
}

// gammaLogBucket0 reports whether z is in [2**k, 2**k × (1 + 2**-8)) for an integer k.
func gammaLogBucket0(z ints.Uint256) bool {
	return z.Rsh(uint(z.BitLen() - 193))[1]>>56 == 0
}

// gammaNeg128 returns the sign bit and |Gamma(x)| = mant × 2**(e-192) for x = -m × 2**(exp-112), where m is a 113-bit integer,
// and -40 <= exp < 11. x must not be an integer.
func gammaNeg128(exp int, m ints.Uint128) (sign uint64, mant ints.Uint256, e int) {
	// |x| in fixed point with 192 fractional bits. It is exact.
	x := m.Uint256().Lsh(uint(80 + exp))

	// |x| = n + r, where n is the nearest integer and -1/2 <= r < 1/2. The reflection formula is
	// Gamma(x) = pi / (sin(pi x) Gamma(1-x)), and sin(pi x) = -(-1)**n sin(pi r).
	// |Gamma(x)| = 1 / (|r| sinc(r) Gamma(1+|x|)), where sinc(r) = sin(pi r)/(pi r).
	half := ints.Uint256{0, 1 << 63, 0, 0}
	n := x.Add(half)
	n[1], n[2], n[3] = 0, 0, 0 // the integer part
	var r ints.Uint256
	rneg := x.Cmp(n) < 0
	if rneg {
		r = n.Sub(x)
	} else {
		r = x.Sub(n)
	}
	// sign of Gamma(x) = -(-1)**n sgn(r)
	if (n[0]&1 != 0) == rneg {
		sign = signMask128[0]
	}

	// the mantissa of |r|, which is exact.
	rm, re := gammaNormalize(r, 0)
	gm, ge := gammaPos128(x.Add(gammaOne))

	// 1 / (|r| sinc(r) Gamma(1+|x|))
	dm, de := gammaNormalize(gammaMul(rm, gammaSinc192(r)), re)
	dm, de = gammaNormalize(gammaMul(dm, gm), de+ge)
	q := ints.Uint512{1: 1}.Quo(dm.Uint512()) // 2**384/dm
	mant, e = gammaNormalize(q.Uint256(), -de)
	return
}

// gammaSinc192 returns sin(pi r)/(pi r) for 0 < r <= 1/2 in fixed point with 192 fractional bits.
func gammaSinc192(r ints.Uint256) ints.Uint256 {
	// sinc(r) = sum a_j (-t)**j, where t = r**2 <= 1/4. A step j of Horner's method needs 192 bits for j < 9,
	// and 128 bits for the others.
	t := gammaMul(r, r)
	d := len(gamma128Sinc)
	p := ints.Uint256(gamma128Sinc[d-1])
	for j := d - 2; j >= 0; j-- {
		if j >= 9 {
			p = ints.Uint256(gamma128Sinc[j]).Sub(gammaMul3(p, t))
		} else {
			p = ints.Uint256(gamma128Sinc[j]).Sub(gammaMul(p, t))
		}
	}
	return p
}

// gammaLogGamma192 returns log(Gamma(z)) for 24 <= z < 2049 in fixed point with 192 fractional bits,
// by Stirling's series log(Gamma(z)) = (z-1/2) log(z) - z + log(2 pi)/2 + sum B(2k) / (2k (2k-1) z**(2k-1)).
// z/2**k must not be in [1, 1+2**-8) for an integer k, where the logarithm is not accurate.
func gammaLogGamma192(z ints.Uint256) ints.Uint256 {
	k := z.BitLen() - 193
	l := logKernel128(k, z.Rsh(uint(k)))
	half := ints.Uint256{0, 1 << 63, 0, 0}
	s := gammaMul(z.Sub(half), l).Sub(z).Add(ints.Uint256(gamma128HalfLn2Pi))

	// the series: 1/z × sum c_k (-w)**(k-1), where w = 1/z**2 <= 2**-9.
	// A step of Horner's method needs 192 bits for the last two steps, 128 bits for the 8 steps before them,
	// and 64 bits for the others, since the error is reduced by w for each step.
	r := ints.Uint512{1: 1}.Quo(z.Uint512()).Uint256() // 1/z
	w := gammaMul(r, r)
	c := &gamma128Stirling
	n := len(c)
	p0, p1 := c[n-1][0], c[n-1][1]
	for i := n - 2; i >= 10; i-- {
		h, l := gammaStep64(p0, p1, w[1])
		var b uint64
		p1, b = bits.Sub64(c[i][1], l, 0)
		p0, _ = bits.Sub64(c[i][0], h, b)
	}
	p2 := uint64(0)
	for i := 9; i >= 2; i-- {
		r0, r1, r2 := gammaStep128(p0, p1, p2, w[1], w[2])
		var b uint64
		p2, b = bits.Sub64(c[i][2], r2, 0)
		p1, b = bits.Sub64(c[i][1], r1, b)
		p0, _ = bits.Sub64(c[i][0], r0, b)
	}
	p := ints.Uint256{p0, p1, p2, 0}
	for i := 1; i >= 0; i-- {
		p = ints.Uint256(c[i]).Sub(gammaMul(w, p))
	}
	return s.Add(gammaMul(r, p))
}

// gammaExp192 returns e**s = mant × 2**(k-192) for s >= 0 in fixed point with 192 fractional bits,
// where mant is in [2**192, 2**193).
func gammaExp192(s ints.Uint256) (mant ints.Uint256, k int) {
	q, n := gammaExpReduce(s, false)
	mant = gammaMul(q, ints.Uint256(gamma128Exp2[n&255]))
	return gammaNormalize(mant, int(n>>8))
}

// gammaExpNeg192 returns e**-s = mant × 2**(k-192) for s >= 0 in fixed point with 192 fractional bits,
// where mant is in [2**192, 2**193).
func gammaExpNeg192(s ints.Uint256) (mant ints.Uint256, k int) {
	q, n := gammaExpReduce(s, true)
	// 2**(-n/256) = 2**(-a-1) × 2**((256-b)/256) for n = 256 a + b, 0 < b < 256.
	a, b := int(n>>8), n&255
	if b != 0 {
		a++
	}
	mant = gammaMul(q, ints.Uint256(gamma128Exp2[(256-b)&255]))
	return gammaNormalize(mant, -a)
}

// gammaRecip192 returns 1/m for m in [1, 2) in fixed point with 192 fractional bits.
func gammaRecip192(m ints.Uint256) ints.Uint256 {
	// Newton's method y = y (2 - m y) doubles the number of the correct bits: 52 -> 104 -> 208.
	f := float64(m[0]) + float64(m[1])*0x1p-64
	y := gammaOne
	if r := 1 / f; r < 1 {
		y = ints.Uint256{0, uint64(r * 0x1p64)}
	}
	two := ints.Uint256{2}
	for range 2 {
		y = gammaMul(y, two.Sub(gammaMul(m, y)))
	}
	return y
}

// gammaExpReduce returns q and n such that e**s = 2**(n/256) × q, or e**-s = 2**(-n/256) × q if neg is true,
// where q is about in [0.9986, 1.0014] in fixed point with 192 fractional bits.
func gammaExpReduce(s ints.Uint256, neg bool) (q ints.Uint256, n uint64) {
	// n = round(s × 256/ln(2)), and e**s = 2**(n/256) × e**r, where r = s - n × ln(2)/256.
	p := s.Mul512(ints.Uint256(gamma128InvLn2By256)).Add(ints.Uint512{2: 1 << 63})
	n = shr512to256(p, 384)[3]
	nl2 := shr512to256(ints.Uint256{3: n}.Mul512(ints.Uint256(gamma128Ln2By256)), 64)
	var r ints.Uint256
	rneg := s.Cmp(nl2) < 0
	if rneg {
		r = nl2.Sub(s)
	} else {
		r = s.Sub(nl2)
	}
	if neg {
		// e**-s = 2**(-n/256) × e**-r
		rneg = !rneg
	}

	// e**r = sum r**m / m! for |r| <= ln(2)/512 < 2**-9. A step m of Horner's method needs 192 bits for m < 2,
	// 128 bits for m < 9, and 64 bits for the others.
	c := &gamma128ExpCoeffs
	d := len(c)
	q0, q1 := c[d-1][0], c[d-1][1]
	for i := d - 2; i >= 9; i-- {
		h, l := gammaStep64(q0, q1, r[1])
		if rneg {
			var b uint64
			q1, b = bits.Sub64(c[i][1], l, 0)
			q0, _ = bits.Sub64(c[i][0], h, b)
		} else {
			var cy uint64
			q1, cy = bits.Add64(c[i][1], l, 0)
			q0, _ = bits.Add64(c[i][0], h, cy)
		}
	}
	q2 := uint64(0)
	for i := 8; i >= 2; i-- {
		r0, r1, r2 := gammaStep128(q0, q1, q2, r[1], r[2])
		if rneg {
			var b uint64
			q2, b = bits.Sub64(c[i][2], r2, 0)
			q1, b = bits.Sub64(c[i][1], r1, b)
			q0, _ = bits.Sub64(c[i][0], r0, b)
		} else {
			var cy uint64
			q2, cy = bits.Add64(c[i][2], r2, 0)
			q1, cy = bits.Add64(c[i][1], r1, cy)
			q0, _ = bits.Add64(c[i][0], r0, cy)
		}
	}
	q = ints.Uint256{q0, q1, q2, 0}
	for i := 1; i >= 0; i-- {
		t := gammaMul(q, r)
		if rneg {
			q = ints.Uint256(c[i]).Sub(t)
		} else {
			q = ints.Uint256(c[i]).Add(t)
		}
	}
	return q, n
}

func isNegInt128(x Float128) bool {
	if x.Lt(Float128{}) {
		_, xf := x.Modf()
		return xf.IsZero()
	}
	return false
}
