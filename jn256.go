package floats

import (
	"math"

	"github.com/shogo82148/ints"
)

// Jn returns the order-n Bessel function of the first kind.
//
// Special cases are:
//
//	Jn(n, ±Inf) = 0
//	Jn(n, NaN) = NaN
func (a Float256) Jn(n int) Float256 {
	if a.IsNaN() {
		return a
	}
	switch n {
	case 0:
		return a.J0()
	case 1:
		return a.J1()
	case -1:
		return a.J1().Neg()
	}

	neg := false
	if n < 0 {
		n = -n
		neg = !neg
	}
	x := a
	if x.Signbit() {
		x = x.Neg()
		neg = !neg
	}

	var y Float256
	switch {
	case x.IsInf(0), x.IsZero():
		y = Float256{}
	default:
		y = jnPositive256(n, x)
	}
	if neg && n%2 == 1 {
		y = y.Neg()
	}
	return y
}

// jnPositive256 returns Jn(x) for n >= 2 and finite x > 0.
//
// It is calculated in fixed point with 320 fractional bits:
//   - x**2 < 2 (n+1): the Taylor series.
//   - x >= max(110, 2 n**2): Hankel's asymptotic expansion.
//   - otherwise: Miller's algorithm.
//
// The absolute error is about 2**-300 in the scale of the values around the order x, so that Jn(x) may not
// be correctly rounded when it is extremely close to a zero (|Jn(x)| < 2**-70) if x >= 2.
func jnPositive256(n int, x Float256) Float256 {
	_, exp, m := x.normalize()
	x64 := math.Ldexp(float64(m[0]), exp-44)
	if x64*x64 < 2*float64(n+1) {
		return jnTaylor256(n, exp, m)
	}
	if x64 >= 110 && x64 >= 2*float64(n)*float64(n) {
		return jnHankel256(n, x, exp, m)
	}
	return jnMiller256(n, exp, m)
}

// jnTaylor256 returns Jn(x) for n >= 2 and x = m × 2**(exp-236) < sqrt(2 (n+1)) by the Taylor series
//
//	Jn(x) = (x/2)**n/n! sum (-1)**k (x**2/4)**k n!/(k! (n+k)!).
//
// The series decreases (the ratio of the terms is less than 1/2), so that no accuracy is lost.
func jnTaylor256(n, exp int, m ints.Uint256) Float256 {
	// (x/2)**n/n! = exp(n log(x/2) - log(n!)), which is less than 1. It underflows (< 2**-262379 ~ e**-181868) soon,
	// and n < 2**15 otherwise.
	if float64(n)*math.Log(math.Ldexp(float64(m[0]), exp-44)/2)-lgamma64(float64(n+1)) < -190000 {
		return Float256{}
	}
	xm := gammaFix256FromUint256(m, 84)
	logHalfX := lgamma320LogPair(xm, exp-1)
	l := lgammaFix256{logHalfX.neg, logHalfX.v.mulUint(uint64(n))}.add(lgamma320Pos(gammaFix256{uint64(n) + 1}).negate())
	// (x/2)**n/n! < 1 for x**2 < 2 (n+1), so that the logarithm is negative.

	// the series, where u = x**2/4
	x := j0Fix256(exp, m)
	u := gammaMul6(x, x).shr(2)
	sum := lgammaFix256{false, gammaOne256}
	t := lgammaFix256{false, gammaOne256}
	for k := uint64(1); ; k++ {
		t = lgammaFix256{!t.neg, gammaMul6(t.v, u).divUint(k * (k + uint64(n)))}
		if t.v == (gammaFix256{}) {
			break
		}
		sum = sum.add(t)
	}

	mant, e := gammaExpNeg256(l.v)
	return j0Result(lgammaFix256{sum.neg, gammaMul6(mant, sum.v)}, e)
}

// jnMiller256 returns Jn(x) for n >= 2 and x = m × 2**(exp-236) >= sqrt(2 (n+1)) using Miller's algorithm
// (see j0Miller256): a backward recurrence J(k-1) = 2k/x J(k) - J(k+1) from an arbitrary trial value at a high order,
// followed by the normalization with J0 + 2 sum J(2k) = 1. The recurrence is unconditionally stable for the minimal solution,
// whereas the forward recurrence is stable only while k < x. The values are scaled down when they are large,
// so that the result may be extremely small without underflow of the calculation.
func jnMiller256(n, exp int, m ints.Uint256) Float256 {
	// 1/x = r × 2**-exp, where r = 1/xm for the mantissa xm of x in [1, 2).
	r := gammaRecip256(gammaFix256FromUint256(m, 84))

	// The error of the arbitrary starting value is negligible if J(k)(x) ~ (x/2)**k/k! is less than 2**-400 times
	// of Jn(x) and 1, where Jn(x) >= (x/2)**n/n! exp(-1.5 x**2/(4 (n+1))) for n >= x.
	x64 := math.Ldexp(float64(m[0]), exp-44)
	lx := math.Log(x64 / 2)
	logJ := func(k int) float64 { return float64(k)*lx - lgamma64(float64(k+1)) }
	low := -40.0
	if float64(n) >= x64 {
		low = logJ(n) - 1.5*x64*x64/(4*float64(n+1))
	}
	k0 := max(n, int(x64))
	for logJ(k0) > low-400*math.Ln2 {
		k0++
	}
	k0 += k0 % 2

	jn1 := lgammaFix256{} // J(k+1)
	jk := lgammaFix256{false, gammaOne256}
	sum := lgammaFix256{false, gammaOne256.shl(1)} // 2 J(k0)
	var target lgammaFix256
	shifts, targetShifts := 0, 0
	for k := k0; k >= 1; k-- {
		coef := r.mulUint(uint64(2 * k)).shr(uint(exp)) // 2k/x
		jm := lgammaFix256{jk.neg, gammaMul6(jk.v, coef)}.add(jn1.negate())
		if k-1 == n {
			target, targetShifts = jm, shifts
		}
		if (k-1)%2 == 0 {
			sum = sum.add(jm)
			if k != 1 {
				sum = sum.add(jm)
			}
		}
		jn1, jk = jk, jm
		if jk.v[0] >= 1<<48 {
			// scale down not to overflow. The ratio is not changed. target is kept and its scale is recorded.
			jk.v, jn1.v, sum.v = jk.v.shr(64), jn1.v.shr(64), sum.v.shr(64)
			shifts++
		}
	}

	// Jn(x) = target/sum × 2**(-64 (shifts - targetShifts))
	sv, e := gammaNormalize256(sum.v, 0)
	res := lgammaFix256{target.neg, gammaMul6(target.v, gammaRecip256(sv))}
	return j0Result(res, -e-64*(shifts-targetShifts))
}

// jnHankel256 returns Jn(x) for n >= 2 and x = m × 2**(exp-236) >= max(110, 2 n**2) using Hankel's asymptotic expansion
//
//	Jn(x) = sqrt(2/(pi x)) (P(x) cos(chi) - Q(x) sin(chi)), chi = x - (n/2 + 1/4) pi,
//
// where P(x) = sum (-1)**k t(2k), Q(x) = sum (-1)**k t(2k+1), t(k) = prod (4 n**2-(2i-1)**2)/(i 8 x) (see j1Hankel256).
// The terms decrease until k ~ 2x or the ratios (4 n**2-(2k-1)**2)/(8 k x) < 1/(4k), whichever is later, so that
// the minimum term, which is the error of the expansion, is less than 2**-300 for x >= max(110, 2 n**2).
func jnHankel256(n int, x Float256, exp int, m ints.Uint256) Float256 {
	// 1/x = r × 2**-exp, where r = 1/xm for the mantissa xm of x in [1, 2).
	r := gammaRecip256(gammaFix256FromUint256(m, 84))
	w := r.shr(uint(exp + 3)) // 1/(8x)

	mu := 4 * int64(n) * int64(n)
	p := lgammaFix256{false, gammaOne256}
	var q lgammaFix256
	t := lgammaFix256{false, gammaOne256}
	for k := uint64(1); ; k++ {
		// t(k) = t(k-1) (4 n**2-(2k-1)**2) w/k
		c := mu - int64((2*k-1)*(2*k-1))
		mag := gammaMul6(t.v, w).mulUint(uint64(max(c, -c))).divUint(k)
		if mag == (gammaFix256{}) {
			break
		}
		if k > 1 && mag.cmp(t.v) >= 0 {
			break // The terms begin to increase.
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

	// cos(chi) and sin(chi) for chi = x - (2n+1) pi/4 = (j-2n-1) pi/4 + z, where x = j pi/4 + z (mod 2 pi) and |z| <= pi/4.
	j, hi, lo := reduce256(x)
	z := float256ToFix(hi).add(float256ToFix(lo))
	s, c := sincosFix256(z)
	c, s = rotateOctant((int(j)+8-(2*n+1)%8)%8, c, s)

	// P cos(chi) - Q sin(chi)
	q = q.negate()
	val := lgammaFix256{p.neg != c.neg, gammaMul6(p.v, c.v)}.add(lgammaFix256{q.neg != s.neg, gammaMul6(q.v, s.v)})

	amp, e := besselAmp256(r, exp)
	return j0Result(lgammaFix256{val.neg, gammaMul6(val.v, amp)}, -e/2)
}
