package floats

import (
	"math"

	"github.com/shogo82148/ints"
)

// J1 returns the order-one Bessel function of the first kind.
//
// Special cases are:
//
//	J1(±Inf) = 0
//	J1(NaN) = NaN
func (a Float256) J1() Float256 {
	switch {
	case a.IsNaN():
		return a
	case a.IsInf(0):
		return Float256{}
	case a.IsZero():
		return a
	}

	x := a.Abs()
	_, exp, m := x.normalize()
	var y Float256
	switch {
	case exp < -118:
		// J1(x) = x/2 (1 - x**2/8 + ...) is a little less than x/2, which can be a tie for the subnormal numbers.
		// The result is x/2 rounded as the value which is a little less than x/2.
		v := m.Uint512().Lsh(200).Sub(ints.Uint512{7: 1})
		y = fixToFloat256(0, v, true, exp-237-200)
	case exp < 4:
		// x < 16
		y = j1Taylor256(exp, m)
	case x.Lt(Float256{0x4000_5b80_0000_0000}): // 110
		y = j1Miller256(exp, m)
	default:
		y = j1Hankel256(x, exp, m)
	}
	if a.Signbit() {
		y = y.Neg()
	}
	return y
}

// j1Taylor256 returns J1(x) for x = m × 2**(exp-236) < 16 by the Taylor series
//
//	J1(x) = x/2 sum (-1)**k (x**2/4)**k/(k! (k+1)!).
//
// The terms are less than 2**9 times larger than the sum, so that the relative accuracy is lost only by 9 bits.
func j1Taylor256(exp int, m ints.Uint256) Float256 {
	x := j0Fix256(exp, m)
	u := gammaMul6(x, x).shr(2) // x**2/4
	sum := lgammaFix256{false, gammaOne256}
	t := lgammaFix256{false, gammaOne256}
	for k := uint64(1); ; k++ {
		t = lgammaFix256{!t.neg, gammaMul6(t.v, u).divUint(k * (k + 1))}
		if t.v == (gammaFix256{}) {
			break
		}
		sum = sum.add(t)
	}
	if j0Small(sum) {
		if r, ok := j1Near256(exp, m); ok {
			return r
		}
	}
	// x/2 = xm × 2**(exp-1), where xm is the mantissa of x in [1, 2).
	xm := gammaFix256FromUint256(m, 84)
	return j0Result(lgammaFix256{sum.neg, gammaMul6(xm, sum.v)}, exp-1)
}

// j1Miller256 returns J1(x) for 16 <= x < 110 using Miller's algorithm (see j0Miller256), where the unnormalized
// value J1 is read off at n=1 instead of n=0.
func j1Miller256(exp int, m ints.Uint256) Float256 {
	_, j1, sum := millerCore256(exp, m)

	// J1(x) = j1/sum
	sv, e := gammaNormalize256(sum.v, 0)
	res := lgammaFix256{j1.neg, gammaMul6(j1.v, gammaRecip256(sv))}
	if e >= 0 && j0Small(res.shr(e)) {
		if r, ok := j1Near256(exp, m); ok {
			return r
		}
	}
	return j0Result(res, -e)
}

// j1Hankel256 returns J1(x) for x >= 110 using Hankel's asymptotic expansion
//
//	J1(x) = sqrt(2/(pi x)) (P(x) cos(x-3pi/4) - Q(x) sin(x-3pi/4)),
//
// where P(x) = sum (-1)**k t(2k), Q(x) = sum (-1)**k t(2k+1), t(k) = prod (4-(2i-1)**2)/(i 8 x) (see j0Hankel256).
// The terms decrease until k ~ 2x, and the minimum, which is less than 2**-300 for x >= 110, is the error.
func j1Hankel256(x Float256, exp int, m ints.Uint256) Float256 {
	// 1/x = r × 2**-exp, where r = 1/xm for the mantissa xm of x in [1, 2).
	r := gammaRecip256(gammaFix256FromUint256(m, 84))
	w := r.shr(uint(exp + 3)) // 1/(8x)

	p := lgammaFix256{false, gammaOne256}
	var q lgammaFix256
	t := lgammaFix256{false, gammaOne256}
	for k := uint64(1); ; k++ {
		// t(k) = t(k-1) (4-(2k-1)**2) w/k
		c := int64(4) - int64((2*k-1)*(2*k-1))
		mag := gammaMul6(t.v, w).mulUint(uint64(max(c, -c))).divUint(k)
		if mag == (gammaFix256{}) || (k > 1 && mag.cmp(t.v) >= 0) {
			break
		}
		t = lgammaFix256{t.neg != (c < 0), mag}
		// the sign is (-1)**(k/2)
		term := lgammaFix256{t.neg != ((k/2)%2 == 1), mag}
		if k%2 == 0 {
			p = p.add(term)
		} else {
			q = q.add(term)
		}
	}

	// cos(chi) and sin(chi) for chi = x - 3pi/4 = (j-3) pi/4 + z, where x = j pi/4 + z (mod 2 pi) and |z| <= pi/4.
	j, hi, lo := reduce256(x)
	z := float256ToFix(hi).add(float256ToFix(lo))
	s, c := sincosFix256(z)
	c, s = rotateOctant(int((j+5)%8), c, s)

	// P cos(chi) - Q sin(chi)
	q = q.negate()
	val := lgammaFix256{p.neg != c.neg, gammaMul6(p.v, c.v)}.add(lgammaFix256{q.neg != s.neg, gammaMul6(q.v, s.v)})

	if exp < 8 && j0Small(val) {
		// J1(x) is extremely small, and the absolute error 2**-300 of the expansion is not negligible.
		if r, ok := j1Near256(exp, m); ok {
			return r
		}
	}

	amp, e := besselAmp256(r, exp)
	return j0Result(lgammaFix256{val.neg, gammaMul6(val.v, amp)}, -e/2)
}

// j1Near256 returns J1(x) for x = m × 2**(exp-236) close to a zero z of J1, by the Taylor series around z:
//
//	J1(z+h) = J0(z) h (1 + d(2) h + d(3) h**2 + ...),
//	d(n+2) = -(d(n+1) z (n+1) (2n+1) + d(n) (n**2+z**2-1) + 2 z d(n-1) + d(n-2))/(z**2 (n+2) (n+1)),
//
// where d(0) = 0 and d(1) = 1, which is derived from the differential equation of J1. h = x - z is calculated exactly
// (it is accurate to 2**-504), so that the relative accuracy is not lost even if J1(x) is extremely small.
// ok is false if x is not close to a zero in the table.
func j1Near256(exp int, m ints.Uint256) (Float256, bool) {
	// x in fixed point with 504 fractional bits. x < 256.
	x := m.Uint512().Lsh(uint(exp + 268))

	// the k-th zero is about (k+1/4) pi.
	k0 := int(math.Ldexp(float64(m[0]), exp-44)/math.Pi - 0.25)
	best := -1
	var hmag ints.Uint512
	var hneg bool
	for k := k0 - 1; k <= k0+1; k++ {
		if k < 1 || k > len(j1256Zeros) {
			continue
		}
		z := ints.Uint512(j1256Zeros[k-1])
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
	top := hmag.Rsh(uint(bl - 321))
	if bl < 321 {
		top = hmag.Lsh(uint(321 - bl))
	}
	hm := gammaFix256{top[2], top[3], top[4], top[5], top[6], top[7]}
	he := bl - 1 - 504
	hv := hmag.Rsh(184)
	hs := lgammaFix256{hneg, gammaFix256{hv[2], hv[3], hv[4], hv[5], hv[6], hv[7]}}

	zv := ints.Uint512(j1256Zeros[best-1]).Rsh(184)
	z := gammaFix256{zv[2], zv[3], zv[4], zv[5], zv[6], zv[7]}
	zn, ze := gammaNormalize256(z, 0)
	zinv := gammaRecip256(zn).shr(uint(ze)) // 1/z
	zinv2 := gammaMul6(zinv, zinv)
	z2 := gammaMul6(z, z)

	const terms = 14
	var d [terms + 1]lgammaFix256
	d[1] = lgammaFix256{false, gammaOne256}
	for n := 0; n+2 <= terms; n++ {
		// d(n+1) z (n+1) (2n+1)
		t := lgammaFix256{d[n+1].neg, gammaMul6(d[n+1].v, z).mulUint(uint64((n + 1) * (2*n + 1)))}
		// d(n) (n**2+z**2-1)
		coef := lgammaFix256{false, z2}.add(lgammaFix256{false, gammaFix256{uint64(n * n)}}).add(lgammaFix256{true, gammaOne256})
		t = t.add(lgammaFix256{d[n].neg != coef.neg, gammaMul6(d[n].v, coef.v)})
		if n > 0 {
			// 2 z d(n-1)
			t = t.add(lgammaFix256{d[n-1].neg, gammaMul6(d[n-1].v, z).shl(1)})
		}
		if n > 1 {
			t = t.add(d[n-2])
		}
		d[n+2] = lgammaFix256{!t.neg, gammaMul6(t.v, zinv2).divUint(uint64((n + 2) * (n + 1)))}
	}
	// the sum of d(n) h**(n-1) by Horner's method
	sum := d[terms]
	for n := terms - 1; n >= 1; n-- {
		sum = d[n].add(lgammaFix256{sum.neg != hs.neg, gammaMul6(sum.v, hs.v)})
	}

	// J1(z+h) = J0(z) h sum, and the sign of J0(z) is (-1)**k.
	v := gammaMul6(gammaMul6(gammaFix256(j1256ZeroJ0[best-1]), hm), sum.v)
	neg := best%2 == 1 != (hneg != sum.neg)
	return j0Result(lgammaFix256{neg, v}, he), true
}
