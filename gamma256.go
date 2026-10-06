package floats

import (
	"math"
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
func (a Float256) Gamma() Float256 {
	// special cases
	switch {
	case isNegInt256(a) || a.IsInf(-1) || a.IsNaN():
		return NewFloat256NaN()
	case a.IsInf(1):
		return NewFloat256Inf(1)
	case a.IsZero():
		return NewFloat256Inf(1).Copysign(a)
	}

	// |a| = m × 2**(exp-236), where m is a 237-bit integer in [2**236, 2**237).
	sign, exp, m := a.normalize()
	neg := sign != 0

	if exp >= 15 {
		// |a| >= 32768
		if !neg {
			return NewFloat256Inf(1) // Gamma(x) overflows for x > 20367.7.
		}
		// Gamma(x) underflows for x < -20400, and its sign is that of sin(pi x) = (-1)**floor(x) sin(pi (x - floor(x))).
		if isOddInt256(a.Floor()) {
			return NewFloat256(0).Neg()
		}
		return Float256{}
	}

	// the sign of the result, and the mantissa and the exponent: |Gamma(a)| = mant × 2**e.
	var mant gammaFix256
	var e int
	if exp < -40 {
		// |a| < 2**-40. Gamma(x) = Gamma(1+x)/x, and Gamma(1+x) = g0 - g1 x + g2 x**2 - ... is calculated by the series.
		mant, e = gammaTiny256(neg, exp, m)
	} else if !neg {
		y := gammaFix256FromUint256(m, uint(exp+84))
		if y[0] >= 20368 {
			return NewFloat256Inf(1)
		}
		mant, e = gammaPos256(y)
	} else {
		sign, mant, e = gammaNeg256(exp, m)
	}

	// mant is in [1, 2) and has 320 fractional bits.
	return fixToFloat256(sign, ints.Uint512{2: mant[0], 3: mant[1], 4: mant[2], 5: mant[3], 6: mant[4], 7: mant[5]}, false, e-320)
}

// gammaFix256 is a number in fixed point with 320 fractional bits in six words.
// The first word is the integer part.
type gammaFix256 [6]uint64

// gammaOne256 is 1.
var gammaOne256 = gammaFix256{1}

// gammaFix256FromUint256 returns m × 2**s as a gammaFix256. The result must be less than 2**64.
func gammaFix256FromUint256(m ints.Uint256, s uint) gammaFix256 {
	return gammaFix256{0, 0, m[0], m[1], m[2], m[3]}.shl(s)
}

// shl returns a × 2**s.
func (a gammaFix256) shl(s uint) gammaFix256 {
	var r gammaFix256
	ws, bs := int(s/64), s%64
	for i := 0; i+ws < 6; i++ {
		r[i] = a[i+ws] << bs
		if i+ws+1 < 6 {
			r[i] |= a[i+ws+1] >> (64 - bs)
		}
	}
	return r
}

// shr returns a / 2**s.
func (a gammaFix256) shr(s uint) gammaFix256 {
	var r gammaFix256
	ws, bs := int(s/64), s%64
	for i := ws; i < 6; i++ {
		r[i] = a[i-ws] >> bs
		if i-ws-1 >= 0 {
			r[i] |= a[i-ws-1] << (64 - bs)
		}
	}
	return r
}

// bitLen returns the bit length of the 384-bit integer a × 2**320.
func (a gammaFix256) bitLen() int {
	for i, v := range a {
		if v != 0 {
			return 64*(6-i) - bits.LeadingZeros64(v)
		}
	}
	return 0
}

// add returns a + b.
func (a gammaFix256) add(b gammaFix256) gammaFix256 {
	var r gammaFix256
	var c uint64
	r[5], c = bits.Add64(a[5], b[5], 0)
	r[4], c = bits.Add64(a[4], b[4], c)
	r[3], c = bits.Add64(a[3], b[3], c)
	r[2], c = bits.Add64(a[2], b[2], c)
	r[1], c = bits.Add64(a[1], b[1], c)
	r[0], _ = bits.Add64(a[0], b[0], c)
	return r
}

// sub returns a - b.
func (a gammaFix256) sub(b gammaFix256) gammaFix256 {
	var r gammaFix256
	var c uint64
	r[5], c = bits.Sub64(a[5], b[5], 0)
	r[4], c = bits.Sub64(a[4], b[4], c)
	r[3], c = bits.Sub64(a[3], b[3], c)
	r[2], c = bits.Sub64(a[2], b[2], c)
	r[1], c = bits.Sub64(a[1], b[1], c)
	r[0], _ = bits.Sub64(a[0], b[0], c)
	return r
}

// cmp compares a and b.
func (a gammaFix256) cmp(b gammaFix256) int {
	for i := range a {
		if a[i] != b[i] {
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}

// mulUint returns a × k, dropping the bits above the integer part.
func (a gammaFix256) mulUint(k uint64) gammaFix256 {
	var r gammaFix256
	var c uint64
	for i := 5; i >= 0; i-- {
		hi, lo := bits.Mul64(a[i], k)
		var cy uint64
		r[i], cy = bits.Add64(lo, c, 0)
		c = hi + cy
	}
	return r
}

// gammaMul256 returns about a × b from the upper n words of a and b, that is, with 64 (n-1) fractional bits.
// The other words of the result are zero. a[0] × b[0] must be less than 2**64.
// The absolute error is about 2**-(64(n-1)) × (|a| + |b| + 2).
func gammaMul256(a, b *gammaFix256, n int) gammaFix256 {
	switch n {
	case 2:
		return gammaMul256n2(a, b)
	case 3:
		return gammaMul256n3(a, b)
	case 4:
		return gammaMul256n4(a, b)
	case 5:
		return gammaMul256n5(a, b)
	}
	return gammaMul256n6(a, b)
}

// gammaMul6 returns about a × b with 320 fractional bits.
func gammaMul6(a, b gammaFix256) gammaFix256 {
	return gammaMul256n6(&a, &b)
}

// gammaNormalize256 normalizes (v, e), which represents the number v × 2**e, to v in [1, 2).
func gammaNormalize256(v gammaFix256, e int) (gammaFix256, int) {
	sh := v.bitLen() - 321
	if sh >= 0 {
		return v.shr(uint(sh)), e + sh
	}
	return v.shl(uint(-sh)), e + sh
}

// gammaRecip256 returns about 1/v for v in [1, 2).
func gammaRecip256(v gammaFix256) gammaFix256 {
	if v[1]|v[2]|v[3]|v[4]|v[5] == 0 {
		return gammaOne256
	}
	// the initial approximation with about 62 bits
	y := 1<<63 | v[1]>>1 | 1
	q, _ := bits.Div64(1<<63, 0, y)
	r := gammaFix256{0, q}

	// Newton's method r = r + r (1 - v r) doubles the precision.
	for _, n := range [...]int{3, 5, 6} {
		t := gammaMul256(&v, &r, n)
		if t[0] >= 1 {
			e := t.sub(gammaOne256)
			r = r.sub(gammaMul256(&r, &e, n))
		} else {
			e := gammaOne256.sub(t)
			r = r.add(gammaMul256(&r, &e, n))
		}
	}
	return r
}

// gammaLimbs256 returns the number of words used in a step of Horner's method,
// where the step needs the absolute precision of the given bits.
func gammaLimbs256(bits int) int {
	return min(6, 1+(bits+8+63)/64)
}

// gammaTiny256 returns |Gamma(x)| = mant × 2**e for x = ±m × 2**(exp-236), where m is a 237-bit integer, and exp < -40.
func gammaTiny256(neg bool, exp int, m ints.Uint256) (mant gammaFix256, e int) {
	// |x| in fixed point with 320 fractional bits.
	var x gammaFix256
	if s := exp + 84; s >= 0 {
		x = gammaFix256FromUint256(m, uint(s))
	} else if s > -384 {
		x = gammaFix256{0, 0, m[0], m[1], m[2], m[3]}.shr(uint(-s))
	}

	// Gamma(1+x) = sum g_k (-x)**k by Horner's method. g_k is |g_k|(-1)**k, and x < 2**-40.
	n := len(gamma256Gamma1)
	p := gammaFix256(gamma256Gamma1[n-1])
	for k := n - 2; k >= 0; k-- {
		t := gammaMul256(&p, &x, gammaLimbs256(320-40*k))
		if neg {
			p = gammaFix256(gamma256Gamma1[k]).add(t)
		} else {
			p = gammaFix256(gamma256Gamma1[k]).sub(t)
		}
	}

	// Gamma(x) = Gamma(1+x)/x, where x = mx × 2**(exp), and mx is in [1, 2).
	mx := gammaFix256{0, 0, m[0], m[1], m[2], m[3]}.shl(84)
	q := gammaMul6(p, gammaRecip256(mx))
	return gammaNormalize256(q, -exp)
}

// gammaPos256 returns Gamma(y) = mant × 2**e for 2**-40 <= y < 2**15 + 1, where mant is in [1, 2).
func gammaPos256(y gammaFix256) (mant gammaFix256, e int) {
	z, pm, pe := gammaRecur256(y)
	em, ee := gammaExp256(gammaLogGamma256(z))
	return gammaNormalize256(gammaMul6(em, gammaRecip256(pm)), ee-pe)
}

// gammaRecur256 returns z >= 48 and P = pm × 2**pe = y (y+1) ... (z-1), where pm is in [1, 2).
// Gamma(y) = Gamma(z)/P.
func gammaRecur256(y gammaFix256) (z, pm gammaFix256, pe int) {
	n := 0
	if y[0] < 48 {
		n = 48 - int(y[0])
	}
	pm = gammaOne256
	y2 := gammaMul256(&y, &y, 6)
	j := 0
	for ; j+4 <= n; j += 4 {
		// The product of the four factors is u (u+2), where u = y (y+3) + j (2y+j+3) = (y+j)(y+j+3).
		u := y2.add(y.mulUint(uint64(2*j + 3))).add(gammaFix256{uint64(j * (j + 3))})
		g := gammaMul256(&u, &u, 6).add(u).add(u)
		pm, pe = gammaNormalize256(gammaMul256(&pm, &g, 6), pe)
	}
	z = y.add(gammaFix256{uint64(j)})
	for ; j < n; j++ {
		pm, pe = gammaNormalize256(gammaMul256(&pm, &z, 6), pe)
		z = z.add(gammaOne256)
	}
	return z, pm, pe
}

// gammaNeg256 returns the sign bit and |Gamma(x)| = mant × 2**e for x = -m × 2**(exp-236), where m is a 237-bit integer,
// and -40 <= exp < 15. x must not be an integer.
func gammaNeg256(exp int, m ints.Uint256) (sign uint64, mant gammaFix256, e int) {
	// |x| in fixed point with 320 fractional bits. It is exact.
	x := gammaFix256FromUint256(m, uint(84+exp))

	// |x| = n + r, where n is the nearest integer and -1/2 <= r < 1/2. The reflection formula is
	// Gamma(x) = pi / (sin(pi x) Gamma(1-x)), and sin(pi x) = -(-1)**n sin(pi r).
	// |Gamma(x)| = 1 / (|r| sinc(r) Gamma(1+|x|)), where sinc(r) = sin(pi r)/(pi r).
	n := x.add(gammaFix256{0, 1 << 63})
	n = gammaFix256{n[0]} // the integer part
	var r gammaFix256
	rneg := x.cmp(n) < 0
	if rneg {
		r = n.sub(x)
	} else {
		r = x.sub(n)
	}
	// sign of Gamma(x) = -(-1)**n sgn(r)
	if (n[0]&1 != 0) == rneg {
		sign = signMask256[0]
	}

	// the mantissa of |r|, which is exact.
	rm, re := gammaNormalize256(r, 0)
	gm, ge := gammaPos256(x.add(gammaOne256))

	// 1 / (|r| sinc(r) Gamma(1+|x|))
	dm, de := gammaNormalize256(gammaMul6(rm, gammaSinc256(r)), re)
	dm, de = gammaNormalize256(gammaMul256(&dm, &gm, 6), de+ge)
	mant, e = gammaNormalize256(gammaRecip256(dm), -de)
	return
}

// gammaSinc256 returns sin(pi r)/(pi r) for 0 < r <= 1/2.
func gammaSinc256(r gammaFix256) gammaFix256 {
	// sinc(r) = sum a_j (-t)**j, where t = r**2 <= 1/4. A step j of Horner's method needs the precision of 312 - 2j bits.
	t := gammaMul256(&r, &r, 6)
	c := &gamma256Sinc
	d := len(c)
	p := gammaFix256(c[d-1])
	for j := d - 2; j >= 0; j-- {
		p = gammaFix256(c[j]).sub(gammaMul256(&p, &t, gammaLimbs256(312-2*j)))
	}
	return p
}

// gammaMulLn2 returns n × c for n < 2**32, where c is log(2)/256 or log(2)/65536.
func gammaMulLn2(c *[7]uint64, n uint64) gammaFix256 {
	var t [7]uint64
	var cy uint64
	for i := 6; i >= 0; i-- {
		hi, lo := bits.Mul64(c[i], n)
		var cc uint64
		t[i], cc = bits.Add64(lo, cy, 0)
		cy = hi + cc
	}
	return gammaFix256{t[0], t[1], t[2], t[3], t[4], t[5]}
}

// gammaLog256 returns log(z) for z >= 1.
func gammaLog256(z gammaFix256) gammaFix256 {
	// z = 2**k × m, where m is in [1, 2). And m = 2**(j/256) × (1+rho), where 0 <= rho < 2**(1/256) - 1.
	k := bits.Len64(z[0]) - 1
	m := z.shr(uint(k))
	lo, hi := 0, 256
	for hi-lo > 1 {
		mid := (lo + hi) / 2
		if gammaFix256(gamma256Exp2[mid]).cmp(m) <= 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	j := lo

	// 2 (1+rho) = m × 2**((256-j)/256)
	var rho gammaFix256
	if t := gammaMul6(m, gamma256Exp2[256-j]); t[0] >= 2 {
		rho = t.sub(gammaFix256{2}).shr(1)
	}

	// 1+rho = 2**(i/65536) × (1+rho2), where 0 <= rho2 < 2**(2/65536) - 1. i is estimated by the floating-point arithmetic,
	// and is decreased by one so that it never exceeds the exact value. The estimate is accurate enough for rho2 to be less than
	// the twice of the width of the interval.
	x := float64(rho[1]) * 0x1p-64
	i := max(min(int(x*(1-x/2+x*x/3)*(65536/math.Ln2)), 255)-1, 0)
	rho2 := gammaMul6(gammaOne256.add(rho), gamma256Exp2Inv[i]).sub(gammaOne256)

	// log(1+rho2) = rho2 × sum (-rho2)**i / (i+1). A step i of Horner's method needs the precision of 312 - 15 i bits.
	c := &gamma256LogCoeffs
	d := len(c)
	p := gammaFix256(c[d-1])
	for i := d - 2; i >= 0; i-- {
		p = gammaFix256(c[i]).sub(gammaMul256(&rho2, &p, gammaLimbs256(312-15*i)))
	}
	return gammaMulLn2(&gamma256Ln2By65536, uint64((256*k+j)*256+i)).add(gammaMul6(rho2, p))
}

// gammaLogGamma256 returns log(Gamma(z)) for 48 <= z < 2**16,
// by Stirling's series log(Gamma(z)) = (z-1/2) log(z) - z + log(2 pi)/2 + sum B(2k) / (2k (2k-1) z**(2k-1)).
func gammaLogGamma256(z gammaFix256) gammaFix256 {
	s := gammaMul6(z.sub(gammaFix256{0, 1 << 63}), gammaLog256(z)).sub(z).add(gammaFix256(gamma256HalfLn2Pi))

	// the series: 1/z × sum b_k (-u)**(k-1), where u = 64/z**2 < 2**-a. a is 5 for z in [48, 64), and 2 floor(log2(z)) - 6 otherwise.
	// A step k of Horner's method needs the precision of 312 - a (k-1) bits.
	lg := bits.Len64(z[0]) - 1
	a := 2*lg - 6
	if lg == 5 {
		a = 5
	}
	r := gammaRecip256(z.shr(uint(lg))).shr(uint(lg)) // 1/z
	u := gammaMul256(&r, &r, 6).mulUint(64)
	c := &gamma256Stirling
	n := gamma256StirlingTerms[lg-5]
	p := gammaFix256(c[n-1])
	for i := n - 2; i >= 0; i-- {
		p = gammaFix256(c[i]).sub(gammaMul256(&u, &p, gammaLimbs256(312-a*i)))
	}
	return s.add(gammaMul256(&r, &p, 6))
}

// gammaExp256 returns e**s = mant × 2**k for s >= 0, where mant is in [1, 2).
func gammaExp256(s gammaFix256) (mant gammaFix256, k int) {
	// n = floor(s × 256/ln(2)), and e**s = 2**(n/256) × e**r, where r = s - n × ln(2)/256 is in [0, ln(2)/256).
	n := gammaMul6(s, gammaFix256(gamma256InvLn2By256))[0]
	nl2 := gammaMulLn2(&gamma256Ln2By256, n)
	if s.cmp(nl2) < 0 {
		// n is too large by one because of the rounding error.
		n--
		nl2 = gammaMulLn2(&gamma256Ln2By256, n)
	}

	// e**r = 2**(1/256) × e**(-x), where x = ln(2)/256 - r is in (0, ln(2)/256].
	// And e**(-x) = 2**(-i/65536) × e**(-y), where i = floor(x × 65536/ln(2)) and y = x - i × ln(2)/65536 is in [0, ln(2)/65536).
	x := gammaFix256{0, gamma256Ln2By256[1], gamma256Ln2By256[2], gamma256Ln2By256[3], gamma256Ln2By256[4], gamma256Ln2By256[5]}.sub(s.sub(nl2))
	i := gammaMul6(x, gammaFix256(gamma256InvLn2By65536))[0]
	// il2 = floor(i × ln(2)/65536) is never larger than x, because the error of the estimate of i is much smaller than
	// the precision of x.
	y := x.sub(gammaMulLn2(&gamma256Ln2By65536, i))

	// e**(-y) = sum (-y)**m / m!. A step m of Horner's method needs the precision of 312 - 16 m bits.
	c := &gamma256ExpCoeffs
	d := len(c)
	q := gammaFix256(c[d-1])
	for m := d - 2; m >= 0; m-- {
		t := gammaMul256(&y, &q, gammaLimbs256(312-16*m))
		q = gammaFix256(c[m]).sub(t)
	}

	// e**s = 2**(n>>8) × 2**((n&255+1)/256) × 2**(-i/65536) × e**(-y)
	q = gammaMul6(q, gamma256Exp2[n&255+1])
	q = gammaMul6(q, gamma256Exp2Inv[i])
	return gammaNormalize256(q, int(n>>8))
}

func isNegInt256(x Float256) bool {
	if x.Lt(Float256{}) {
		_, xf := x.Modf()
		return xf.IsZero()
	}
	return false
}
