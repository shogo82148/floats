package floats

import (
	"math"
	"math/bits"

	"github.com/shogo82148/ints"
)

// J0 returns the order-zero Bessel function of the first kind.
//
// Special cases are:
//
//	J0(±Inf) = 0
//	J0(0) = 1
//	J0(NaN) = NaN
func (a Float256) J0() Float256 {
	switch {
	case a.IsNaN():
		return a
	case a.IsInf(0):
		return Float256{}
	case a.IsZero():
		return Float256(uvone256)
	}

	x := a.Abs()
	_, exp, m := x.normalize()
	switch {
	case exp < -118:
		// J0(x) = 1 - x**2/4 + ... is rounded to 1 for |x| < 2**-118.
		return Float256(uvone256)
	case exp < -100:
		return j0Tiny256(exp, m)
	case exp < 4:
		// x < 16
		return j0Taylor256(exp, m)
	case x.Lt(Float256{0x4000_5b80_0000_0000}): // 110
		return j0Miller256(exp, m)
	}
	return j0Hankel256(x, exp, m)
}

// j0Tiny256 returns J0(x) for x = m × 2**(exp-236), -118 <= exp < -100, which is 1 - x**2/4 with the relative error
// less than 2**-198. The decision of the rounding depends on the low bits of x**2, so that it is calculated exactly.
func j0Tiny256(exp int, m ints.Uint256) Float256 {
	// x**2/4 = T × 2**(2 exp - 474), where T = m**2. 1 - x**2/4 = (2**511 - T 2**(2 exp - 474 + 511)) 2**-511.
	t := m.Mul512(m)
	sh := uint(-(2*exp + 37)) // 2 exp - 474 + 511 = 2 exp + 37 < 0
	ts := t.Rsh(sh)
	sticky := t.Lsh(512-sh) != (ints.Uint512{})
	v := ints.Uint512{1 << 63}.Sub(ts)
	if sticky {
		// the exact value is v - f for 0 < f < 1.
		v = v.Sub(ints.Uint512{7: 1})
	}
	return fixToFloat256(0, v, sticky, -511)
}

// j0Fix256 returns x = m × 2**(exp-236) in fixed point with 320 fractional bits.
// The result must be less than 2**64.
func j0Fix256(exp int, m ints.Uint256) gammaFix256 {
	if s := exp + 84; s >= 0 {
		return gammaFix256FromUint256(m, uint(s))
	} else {
		return gammaFix256{0, 0, m[0], m[1], m[2], m[3]}.shr(uint(-s))
	}
}

// divUint returns a/d.
func (a gammaFix256) divUint(d uint64) gammaFix256 {
	var r gammaFix256
	var rem uint64
	for i := range a {
		r[i], rem = bits.Div64(rem, a[i], d)
	}
	return r
}

// j0Result returns the signed fixed point number v × 2**e as Float256.
func j0Result(v lgammaFix256, e int) Float256 {
	if v.v == (gammaFix256{}) {
		return Float256{}
	}
	return lgamma256FromPair(v.neg, v.v, e)
}

// j0Taylor256 returns J0(x) for x = m × 2**(exp-236) < 16 by the Taylor series
//
//	J0(x) = sum (-1)**k (x**2/4)**k/(k!)**2.
//
// The terms are less than 2**9 times larger than the sum, so that the relative accuracy is lost only by 9 bits.
func j0Taylor256(exp int, m ints.Uint256) Float256 {
	x := j0Fix256(exp, m)
	u := gammaMul6(x, x).shr(2) // x**2/4
	sum := lgammaFix256{false, gammaOne256}
	t := lgammaFix256{false, gammaOne256}
	for k := uint64(1); ; k++ {
		t = lgammaFix256{!t.neg, gammaMul6(t.v, u).divUint(k * k)}
		if t.v == (gammaFix256{}) {
			break
		}
		sum = sum.add(t)
	}
	if j0Small(sum) {
		if r, ok := j0Near256(exp, m); ok {
			return r
		}
	}
	return j0Result(sum, 0)
}

// j0Small reports whether |v| < 2**-60, that is, the fixed point calculation may lose the relative accuracy.
func j0Small(v lgammaFix256) bool {
	return v.v[0] == 0 && v.v[1] < 1<<4
}

// j0Miller256 returns J0(x) for 16 <= x < 110 using Miller's algorithm: a backward recurrence
// J(n-1) = 2n/x J(n) - J(n+1) starting from an arbitrary trial value at a high order, which is numerically stable,
// followed by normalizing against the identity J0(x) + 2 sum J(2k)(x) = 1.
// The calculation is in fixed point with 320 fractional bits, and the values are scaled down when they are large.
func j0Miller256(exp int, m ints.Uint256) Float256 {
	j0, _, sum := millerCore256(exp, m)

	// J0(x) = j0/sum
	sv, e := gammaNormalize256(sum.v, 0)
	res := lgammaFix256{j0.neg, gammaMul6(j0.v, gammaRecip256(sv))}
	if e >= 0 && j0Small(res.shr(e)) {
		if r, ok := j0Near256(exp, m); ok {
			return r
		}
	}
	return j0Result(res, -e)
}

// millerCore256 returns the unnormalized J0(x), J1(x) and the sum J0 + 2 sum J(2k) of Miller's algorithm for
// x = m × 2**(exp-236). The normalized value is j/sum.
func millerCore256(exp int, m ints.Uint256) (j0, j1, sum lgammaFix256) {
	// 1/x = r × 2**-exp, where r = 1/xm for the mantissa xm of x in [1, 2).
	r := gammaRecip256(gammaFix256FromUint256(m, 84))

	// The starting order n is even and J(n)(x) ~ (x/2)**n/n! is less than 2**-345, which is far smaller than
	// the values around the order x, so that the error of the arbitrary starting value is negligible.
	x64 := math.Ldexp(float64(m[0]), exp-44) // the approximate x
	n := int(x64)
	for ; float64(n)*math.Log(x64/2)-lgamma64(float64(n+1)) > -345*math.Ln2; n++ {
	}
	n += n % 2

	jn1 := lgammaFix256{} // J(n+1)
	jn := lgammaFix256{false, gammaOne256}
	sum = lgammaFix256{false, gammaOne256.shl(1)} // 2 J(n)
	for ; n >= 1; n-- {
		coef := r.mulUint(uint64(2 * n)).shr(uint(exp)) // 2n/x
		jm := lgammaFix256{jn.neg, gammaMul6(jn.v, coef)}.add(jn1.negate())
		if (n-1)%2 == 0 {
			sum = sum.add(jm)
			if n != 1 {
				sum = sum.add(jm)
			}
		}
		jn1, jn = jn, jm
		if jn.v[0] >= 1<<48 {
			// scale down not to overflow. The ratio is not changed.
			jn.v, jn1.v, sum.v = jn.v.shr(64), jn1.v.shr(64), sum.v.shr(64)
		}
	}
	return jn, jn1, sum
}

// shr returns v × 2**-s for s >= 0.
func (v lgammaFix256) shr(s int) lgammaFix256 {
	return lgammaFix256{v.neg, v.v.shr(uint(s))}
}

// j0Near256 returns J0(x) for x = m × 2**(exp-236) close to a zero z of J0, by the Taylor series around z:
//
//	J0(z+h) = -J1(z) h (1 + d(2) h + d(3) h**2 + ...), d(n+2) = -((n+1)**2 d(n+1) + z d(n) + d(n-1))/(z (n+2) (n+1)),
//
// where d(0) = 0 and d(1) = 1. h = x - z is calculated exactly (it is accurate to 2**-504), so that the relative
// accuracy is not lost even if J0(x) is extremely small. ok is false if x is not close to a zero in the table.
func j0Near256(exp int, m ints.Uint256) (Float256, bool) {
	// x in fixed point with 504 fractional bits. x < 256.
	x := m.Uint512().Lsh(uint(exp + 268))

	// the k-th zero is about (k-1/4) pi.
	k0 := int(math.Ldexp(float64(m[0]), exp-44)/math.Pi + 0.25)
	best := -1
	var hmag ints.Uint512
	var hneg bool
	for k := k0 - 1; k <= k0+1; k++ {
		if k < 1 || k > len(j0256Zeros) {
			continue
		}
		z := ints.Uint512(j0256Zeros[k-1])
		d, neg := x.Sub(z), false
		if x.Cmp(z) < 0 {
			d, neg = z.Sub(x), true
		}
		if best < 0 || d.Cmp(hmag) < 0 {
			best, hmag, hneg = k, d, neg
		}
	}
	if best < 0 || hmag.IsZero() || hmag.BitLen() > 504-30 {
		return Float256{}, false
	}

	// h = hm × 2**he, where hm is in [1, 2), and hs is h in fixed point.
	bl := hmag.BitLen()
	top := hmag.Rsh(uint(bl - 321)) // bl >= 321 is not guaranteed, but |h| > 2**-183 holds for the Float256 values near a zero
	if bl < 321 {
		top = hmag.Lsh(uint(321 - bl))
	}
	hm := gammaFix256{top[2], top[3], top[4], top[5], top[6], top[7]}
	he := bl - 1 - 504
	hv := hmag.Rsh(184)
	hs := lgammaFix256{hneg, gammaFix256{hv[2], hv[3], hv[4], hv[5], hv[6], hv[7]}}

	zv := ints.Uint512(j0256Zeros[best-1]).Rsh(184)
	z := gammaFix256{zv[2], zv[3], zv[4], zv[5], zv[6], zv[7]}
	zn, ze := gammaNormalize256(z, 0)
	zinv := gammaRecip256(zn).shr(uint(ze)) // 1/z

	// d(n+2) = -((n+1)**2 d(n+1) + z d(n) + d(n-1))/(z (n+2) (n+1))
	const terms = 14
	var d [terms + 1]lgammaFix256
	d[1] = lgammaFix256{false, gammaOne256}
	for n := 0; n+2 <= terms; n++ {
		t := lgammaFix256{d[n+1].neg, d[n+1].v.mulUint(uint64((n + 1) * (n + 1)))}
		t = t.add(lgammaFix256{d[n].neg, gammaMul6(d[n].v, z)})
		if n > 0 {
			t = t.add(d[n-1])
		}
		d[n+2] = lgammaFix256{!t.neg, gammaMul6(t.v, zinv).divUint(uint64((n + 2) * (n + 1)))}
	}
	// the sum of d(n) h**(n-1) by Horner's method
	sum := d[terms]
	for n := terms - 1; n >= 1; n-- {
		sum = d[n].add(lgammaFix256{sum.neg != hs.neg, gammaMul6(sum.v, hs.v)})
	}

	// J0(z+h) = -J1(z) h sum, and the sign of -J1(z) is (-1)**k.
	v := gammaMul6(gammaMul6(gammaFix256(j0256ZeroJ1[best-1]), hm), sum.v)
	neg := best%2 == 1 != (hneg != sum.neg)
	return j0Result(lgammaFix256{neg, v}, he), true
}

// lgamma64 returns the logarithm of Gamma(x) for x > 0.
func lgamma64(x float64) float64 {
	v, _ := math.Lgamma(x)
	return v
}

// j0Hankel256 returns J0(x) for x >= 110 using Hankel's asymptotic expansion
//
//	J0(x) = sqrt(2/(pi x)) (P(x) cos(x-pi/4) - Q(x) sin(x-pi/4)),
//
// where P(x) = sum (-1)**k t(2k), Q(x) = -sum (-1)**k t(2k+1), and t(k) = ((2k-1)!!)**2/(k! (8 x)**k).
// The terms decrease until k ~ 2x, and the minimum, which is less than 2**-300 for x >= 110, is the error.
func j0Hankel256(x Float256, exp int, m ints.Uint256) Float256 {
	// 1/x = r × 2**-exp, where r = 1/xm for the mantissa xm of x in [1, 2).
	r := gammaRecip256(gammaFix256FromUint256(m, 84))
	w := r.shr(uint(exp + 3)) // 1/(8x)

	// P and Qs = -Q
	p := lgammaFix256{false, gammaOne256}
	var qs lgammaFix256
	t := gammaOne256
	for k := uint64(1); ; k++ {
		next := gammaMul6(t, w).mulUint((2*k - 1) * (2*k - 1)).divUint(k)
		if next == (gammaFix256{}) || next.cmp(t) >= 0 {
			break
		}
		t = next
		// the sign is (-1)**(k/2)
		term := lgammaFix256{(k/2)%2 == 1, t}
		if k%2 == 0 {
			p = p.add(term)
		} else {
			qs = qs.add(term)
		}
	}

	// cos(chi) and sin(chi) for chi = x - pi/4 = (j-1) pi/4 + z, where x = j pi/4 + z (mod 2 pi) and |z| <= pi/4.
	j, hi, lo := reduce256(x)
	z := float256ToFix(hi).add(float256ToFix(lo))
	s, c := sincosFix256(z)
	c, s = rotateOctant(int((j+7)%8), c, s)

	// P cos(chi) - Q sin(chi) = P cos(chi) + Qs sin(chi)
	val := lgammaFix256{p.neg != c.neg, gammaMul6(p.v, c.v)}.add(lgammaFix256{qs.neg != s.neg, gammaMul6(qs.v, s.v)})

	if exp < 8 && j0Small(val) {
		// J0(x) is extremely small, and the absolute error 2**-300 of the expansion is not negligible.
		if r, ok := j0Near256(exp, m); ok {
			return r
		}
	}

	amp, e := besselAmp256(r, exp)
	return j0Result(lgammaFix256{val.neg, gammaMul6(val.v, amp)}, -e/2)
}

// besselAmp256 returns amp and e such that sqrt(2/(pi x)) = amp × 2**(-e/2), where e is even,
// for x = xm × 2**exp and r = 1/xm.
func besselAmp256(r gammaFix256, exp int) (amp gammaFix256, e int) {
	// sqrt(2/(pi x)) = sqrt(2/pi/xm 2**-exp). If exp is odd, 2**-exp = 2 × 2**-(exp+1).
	v := gammaMul6(gammaFix256(j0256TwoOverPi), r)
	e = exp
	if e%2 != 0 {
		v = v.shl(1)
		e++
	}
	return gammaMul6(v, j0Rsqrt256(v)), e
}

// float256ToFix returns a as the signed fixed point number with 320 fractional bits. |a| must be less than 2**64.
func float256ToFix(a Float256) lgammaFix256 {
	if a.IsZero() {
		return lgammaFix256{}
	}
	sign, exp, m := a.normalize()
	return lgammaFix256{sign != 0, j0Fix256(exp, m)}
}

// sincosFix256 returns sin(z) and cos(z) for |z| <= 1 in fixed point with 320 fractional bits by the Taylor series.
func sincosFix256(z lgammaFix256) (s, c lgammaFix256) {
	u := gammaMul6(z.v, z.v)
	s = z
	c = lgammaFix256{false, gammaOne256}
	ts, tc := z, c
	for k := uint64(1); ; k++ {
		tc = lgammaFix256{!tc.neg, gammaMul6(tc.v, u).divUint((2*k - 1) * (2 * k))}
		ts = lgammaFix256{!ts.neg, gammaMul6(ts.v, u).divUint((2 * k) * (2*k + 1))}
		if tc.v == (gammaFix256{}) && ts.v == (gammaFix256{}) {
			break
		}
		c = c.add(tc)
		s = s.add(ts)
	}
	return s, c
}

// rotateOctant returns cos(k pi/4 + z) and sin(k pi/4 + z) for 0 <= k < 8, where c = cos(z) and s = sin(z).
func rotateOctant(k int, c, s lgammaFix256) (cos, sin lgammaFix256) {
	// cos(a+z) = cos(a) c - sin(a) s, sin(a+z) = sin(a) c + cos(a) s, and (cos(a), sin(a)) = ±(1, 0), ±(0, 1) or
	// (±1, ±1) sqrt(2)/2.
	h := gammaFix256(j0256HalfSqrt2)
	rot := func(ca, sa lgammaFix256) (lgammaFix256, lgammaFix256) {
		mul := func(a, b lgammaFix256) lgammaFix256 {
			if a.v == (gammaFix256{}) || b.v == (gammaFix256{}) {
				return lgammaFix256{}
			}
			return lgammaFix256{a.neg != b.neg, gammaMul6(a.v, b.v)}
		}
		return mul(ca, c).add(mul(sa, s).negate()), mul(sa, c).add(mul(ca, s))
	}
	one := lgammaFix256{false, gammaOne256}
	zero := lgammaFix256{}
	hp, hn := lgammaFix256{false, h}, lgammaFix256{true, h}
	switch k {
	case 0:
		return rot(one, zero)
	case 1:
		return rot(hp, hp)
	case 2:
		return rot(zero, one)
	case 3:
		return rot(hn, hp)
	case 4:
		return rot(one.negate(), zero)
	case 5:
		return rot(hn, hn)
	case 6:
		return rot(zero, one.negate())
	}
	return rot(hp, hn)
}

// j0Rsqrt256 returns about 1/sqrt(v) for v in (1/4, 2) by Newton's method.
func j0Rsqrt256(v gammaFix256) gammaFix256 {
	// the initial approximation with about 52 bits
	f := float64(v[0]) + float64(v[1])*0x1p-64
	g := 1 / math.Sqrt(f)
	r := gammaFix256{uint64(g), uint64((g - math.Floor(g)) * 0x1p64)}
	// r = r (3 - v r**2)/2 doubles the precision: 52 -> 104 -> 208 -> 416
	three := gammaOne256.shl(1).add(gammaOne256)
	for range 3 {
		t := gammaMul6(v, gammaMul6(r, r))
		r = gammaMul6(r, three.sub(t)).shr(1)
	}
	return r
}
