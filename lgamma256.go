package floats

import (
	"github.com/shogo82148/ints"
)

// Lgamma returns the natural logarithm and sign (-1 or +1) of Gamma(a).
//
// Special cases are:
//
//	Lgamma(+Inf) = +Inf
//	Lgamma(0) = +Inf
//	Lgamma(-integer) = +Inf
//	Lgamma(-Inf) = -Inf
//	Lgamma(NaN) = NaN
//
// The result is correctly rounded except for the arguments extremely close to the zeros of Lgamma at the negative numbers,
// whose results have the absolute error less than 2**-285.
func (a Float256) Lgamma() (Float256, int) {
	switch {
	case a.IsNaN() || a.IsInf(0):
		return a, 1
	case a.IsZero() || isNegInt256(a):
		return NewFloat256Inf(1), 1
	}

	// |a| = m × 2**(exp-236), where m is a 237-bit integer in [2**236, 2**237).
	sign, exp, m := a.normalize()
	neg := sign != 0
	switch {
	case exp >= 15:
		return lgamma256Large(a, neg, exp, m)
	case exp < -40:
		return lgamma256Tiny(neg, exp, m)
	case !neg && (a == Float256(uvone256) || a == Float256{0x4000_0000_0000_0000, 0, 0, 0}):
		return Float256{}, 1 // the zeros of Lgamma
	}

	// |a| in fixed point with 320 fractional bits. It is exact.
	x := gammaFix256FromUint256(m, uint(exp+84))
	if neg {
		sign, v := lgamma320NegFix(x)
		return lgamma256FromFix(v), sign
	}

	// Lgamma(x) = t × sum (-1)**k c_k t**(k-1) for x = 1+t, and t × sum ... for x = 2+t, where |t| is small.
	// They keep the relative accuracy near the zeros of Lgamma.
	if t, tneg, ok := lgamma320Near(x, gammaOne256); ok {
		return lgamma256Near(t, tneg, lgamma256Taylor1[:], false), 1
	}
	if t, tneg, ok := lgamma320Near(x, gammaFix256{2}); ok {
		return lgamma256Near(t, tneg, lgamma256Taylor2[:], true), 1
	}
	return lgamma256FromFix(lgamma320Pos(x)), 1
}

// lgamma256FromFix returns v as Float256. v must not be zero.
func lgamma256FromFix(v lgammaFix256) Float256 {
	return lgamma256FromPair(v.neg, v.v, 0)
}

// lgamma256FromPair returns ±mant × 2**e as Float256. mant must not be zero.
func lgamma256FromPair(neg bool, mant gammaFix256, e int) Float256 {
	var sign uint64
	if neg {
		sign = signMask256[0]
	}
	return fixToFloat256(sign, ints.Uint512{2: mant[0], 3: mant[1], 4: mant[2], 5: mant[3], 6: mant[4], 7: mant[5]}, false, e-320)
}

// lgamma320Near returns t = |x - c|, and whether x < c, and reports whether t < 2**-18.
func lgamma320Near(x, c gammaFix256) (t gammaFix256, tneg, ok bool) {
	if x.cmp(c) >= 0 {
		t = x.sub(c)
	} else {
		t = c.sub(x)
		tneg = true
	}
	return t, tneg, t[0] == 0 && t[1] < 1<<46
}

// lgamma320Bracket returns the sum of coefficient(k) t**(k-1), where t = ±s, and coefficient(k) is (-1)**k c[k-1],
// except that the coefficient(1) is c[0] if first is true. s must be less than 2**-18.
func lgamma320Bracket(s gammaFix256, tneg bool, c [][6]uint64, n int, first bool) lgammaFix256 {
	coef := func(k int) lgammaFix256 {
		return lgammaFix256{k%2 == 1 && (!first || k != 1), gammaFix256(c[k-1])}
	}
	// A step k of Horner's method needs the precision of 312 - 18 (k-1) bits.
	p := coef(n)
	for k := n - 1; k >= 1; k-- {
		prod := lgammaFix256{p.neg != tneg, gammaMul256(&s, &p.v, gammaLimbs256(312-18*(k-1)))}
		p = coef(k).add(prod)
	}
	return p
}

// lgamma256Near returns Lgamma(x) for x = c + t where |t| < 2**-18 and c is 1 or 2.
func lgamma256Near(t gammaFix256, tneg bool, c [][6]uint64, first bool) Float256 {
	b := lgamma320Bracket(t, tneg, c, len(c), first)
	// t is exact, and its relative accuracy must be kept.
	tm, te := gammaNormalize256(t, 0)
	mant, e := gammaNormalize256(gammaMul6(tm, b.v), te)
	return lgamma256FromPair(b.neg != tneg, mant, e)
}

// lgamma256Tiny returns Lgamma(x) for x = ±m × 2**(exp-236), where m is a 237-bit integer, and exp < -40.
func lgamma256Tiny(neg bool, exp int, m ints.Uint256) (Float256, int) {
	// |x| in fixed point with 320 fractional bits.
	var x gammaFix256
	if s := exp + 84; s > -384 {
		x = gammaFix256{0, 0, m[0], m[1], m[2], m[3]}
		if s >= 0 {
			x = x.shl(uint(s))
		} else {
			x = x.shr(uint(-s))
		}
	}

	// Lgamma(x) = -log|x| + Lgamma(1+x), and -log|x| = -exp log(2) - log(m/2**236) is positive.
	mv := gammaFix256FromUint256(m, 84)
	r := lgammaFix256{false, gammaMulLn2(&gamma256Ln2By256, uint64(-exp)*256)}.add(lgammaFix256{true, gammaLog256(mv)})

	// Lgamma(1+x) = x sum (-1)**k c_k x**(k-1).
	b := lgamma320Bracket(x, neg, lgamma256Taylor1[:], 8, false)
	r = r.add(lgammaFix256{b.neg != neg, gammaMul6(x, b.v)})
	sign := 1
	if neg {
		sign = -1
	}
	return lgamma256FromFix(r), sign
}

// lgamma256Large returns Lgamma(a) for |a| >= 2**15 with the calculation in fixed point.
func lgamma256Large(a Float256, neg bool, exp int, m ints.Uint256) (Float256, int) {
	if !neg {
		ym, ye := lgamma320Stirling(gammaFix256FromUint256(m, 84), exp)
		return lgamma256FromPair(false, ym, ye), 1
	}

	// Lgamma(x) = log(pi / |sin(pi x)|) - Lgamma(1-x), where |sin(pi x)| = sin(pi |r|) for x = n + r, |r| <= 1/2.
	// n and r are calculated exactly, since a is not an integer here.
	half := Float256{0x3fff_e000_0000_0000, 0, 0, 0}
	one := Float256(uvone256)
	q := a.Abs()
	n := q.Floor()
	r := q.Sub(n)
	if r.Gt(half) {
		n = n.Add(one)
		r = q.Sub(n)
	}
	// The sign of Gamma(x) is negative if floor(x) = -ceil(|x|) is odd. ceil(|x|) is n+1 if r > 0, and n otherwise.
	sign := 1
	if isOddInt256(n) != !r.Signbit() {
		sign = -1
	}

	// |r| in fixed point with 320 fractional bits. It is exact since |r| >= 2**-221.
	_, rexp, rm := r.Abs().normalize()
	rfix := gammaFix256FromUint256(rm, 84).shr(uint(-rexp))
	rn, re := gammaNormalize256(rfix, 0)
	dm, de := gammaNormalize256(gammaMul6(rn, gammaSinc256(rfix)), re)
	lnD := lgamma320LogPair(dm, de) // negative, and lnD.v is its absolute value

	// log(|x|-1) = -log(D) - Lgamma(1+|x|), where Lgamma(1+|x|) = lm × 2**le is larger than 2**17.
	_, e1, m1 := q.Add(one).normalize()
	lm, le := lgamma320Stirling(gammaFix256FromUint256(m1, 84), e1)
	if le < 330 {
		lm = lm.sub(lgamma320Shr(lnD.v, le))
	}
	return lgamma256FromPair(true, lm, le), sign
}

// lgamma320Shr returns v / 2**s.
func lgamma320Shr(v gammaFix256, s int) gammaFix256 {
	return v.shr(uint(s))
}

// lgamma320Stirling returns Lgamma(z) = ym × 2**ye for z = zm × 2**ze >= 2**15, where zm is in [1, 2) and ym is in [1, 2).
func lgamma320Stirling(zm gammaFix256, ze int) (ym gammaFix256, ye int) {
	// Lgamma(z) = z B, where B = log(z) - 1 + (log(2 pi)/2 - log(z)/2)/z + P(u)/z**2 with the series P of Stirling's series.
	lnz := lgamma320LogPair(zm, ze).v // positive
	b := lnz.sub(gammaOne256)
	if ze < 330 {
		invz := gammaRecip256(zm).shr(uint(ze))
		b = b.sub(gammaMul6(lnz.shr(1).sub(gammaFix256(gamma256HalfLn2Pi)), invz))

		// the series P(u) = sum b_k (-u)**(k-1), where u = 64/z**2 < 2**-5. A step k of Horner's method needs
		// the precision of 312 - a (k-1) bits, where a = 2 ze - 6.
		w := gammaMul6(invz, invz)
		u := w.mulUint(64)
		lg := min(ze, 15)
		a := 2*ze - 6
		c := &gamma256Stirling
		n := gamma256StirlingTerms[lg-5]
		p := gammaFix256(c[n-1])
		for i := n - 2; i >= 0; i-- {
			p = gammaFix256(c[i]).sub(gammaMul256(&u, &p, gammaLimbs256(312-a*i)))
		}
		b = b.add(gammaMul6(p, w))
	}
	return gammaNormalize256(gammaMul6(zm, b), ze)
}
