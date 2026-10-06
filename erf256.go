package floats

import (
	"math"
	"math/bits"

	"github.com/shogo82148/ints"
)

// Erf returns the error function of a.
//
// Special cases are:
//
//	+Inf.Erf() = 1
//	-Inf.Erf() = -1
//	NaN.Erf() = NaN
func (a Float256) Erf() Float256 {
	switch {
	case a.IsInf(0):
		return Float256(uvone256).Copysign(a)
	case a.IsNaN():
		return NewFloat256NaN()
	case a.IsZero():
		return a
	}

	// |a| = m × 2**(exp-236), where m is a 237-bit integer in [2**236, 2**237).
	sign, exp, m := a.normalize()
	neg := sign != 0
	if exp >= 3 && (exp > 3 || a.Abs().Ge(Float256{0x4000_2971_af06_3fb1, 0x3a98_f6a7_2e5f_b545, 0x635d_1d72_2ec8_07a0, 0xaca2_ae63_3fc9_898c})) {
		// erf(x) is rounded to 1 for |x| >= 12.722.
		return Float256(uvone256).Copysign(a)
	}

	if exp < -8 {
		return lgamma256FromPair(neg, erf256SmallSeries(exp, m), exp)
	}

	// |x| in fixed point with 320 fractional bits. It is exact.
	v, e, scaled := erf256Cell(gammaFix256FromUint256(m, uint(exp+84)))
	if !scaled {
		return lgamma256FromFix(lgammaFix256{neg, v})
	}

	// erfc(x) × 2**237 = v × 2**e, where v is in [1, 2). erf(x) = 1 - erfc(x) is rounded to a multiple of 2**-237,
	// that is, 1 - j 2**-237 for the integer j nearest to erfc(x) × 2**237.
	// j is not zero since |x| < 12.722 here and erfc(x) × 2**237 is larger than 1/2.
	j := erf256Round(ints.Uint512{2: v[0], 3: v[1], 4: v[2], 5: v[3], 6: v[4], 7: v[5]}, e)
	// 2**237 - j is the mantissa of the result in [1/2, 1).
	mant := ints.Uint256{1 << 45, 0, 0, 0}.Sub(j)
	return fixToFloat256(sign, mant.Uint512(), false, -237)
}

// erf256SmallSeries returns erf(x) × 2**-exp in fixed point with 320 fractional bits for x = m × 2**(exp-236), where m is
// a 237-bit integer and exp < -8.
func erf256SmallSeries(exp int, m ints.Uint256) gammaFix256 {
	// |x| < 2**-8. erf(x) = 2/sqrt(pi) x sum (-x**2)**n/(n! (2n+1)), and the terms decrease fast.
	// |x| in fixed point with 320 fractional bits, which is not accurate if |x| is tiny, but its square is negligible then.
	var x gammaFix256
	if s := exp + 84; s >= 0 {
		x = gammaFix256FromUint256(m, uint(s))
	} else if s > -384 {
		x = gammaFix256{0, 0, m[0], m[1], m[2], m[3]}.shr(uint(-s))
	}
	u := gammaMul6(x, x)
	c := &erf256Small
	n := len(c)
	p := gammaFix256(c[n-1])
	for i := n - 2; i >= 0; i-- {
		p = gammaFix256(c[i]).sub(gammaMul6(u, p))
	}
	// the product of the mantissa of x and the series is exact except for the last bits.
	return gammaMul6(gammaFix256FromUint256(m, 84), gammaMul6(gammaFix256(erf256TwoOverSqrtPi), p))
}

// erf256Round returns the integer nearest to v × 2**(e-320), where v is a 384-bit integer, and the result is
// less than 2**256. e must be less than 320. If it is exactly half-way between two integers, it returns the even one.
func erf256Round(v ints.Uint512, e int) ints.Uint256 {
	// v × 2**(e-320) = v / 2**sh
	sh := uint(320 - e)
	j := v.Rsh(sh).Uint256()
	half := v.Rsh(sh - 1)[7]&1 != 0
	rest := v.Lsh(512-(sh-1)) != ints.Uint512{}
	if half && (rest || j[3]&1 != 0) {
		j = j.Add(ints.Uint256{3: 1})
	}
	return j
}

// erf256Cell returns erf(x) in fixed point with 320 fractional bits for 2**-8 <= x < 7 (scaled is false),
// and erfc(x) × 2**237 = v × 2**e, where v is in [1, 2), otherwise for x < 12.722 (scaled is true),
// where x is in fixed point with 320 fractional bits. The relative error is less than 2**-310.
func erf256Cell(x gammaFix256) (v gammaFix256, e int, scaled bool) {
	k := int(x[0]<<4 | x[1]>>60)
	if k >= 112 {
		v, e := erf256Scaled(x)
		return v, e + 237, true
	}

	hp := erf256Taylor(x, k, int(erf256Terms[k]))
	// the correction is 2/sqrt(pi) exp(-x0**2) h p = m0 × 2**e0 × h p.
	corr := lgammaFix256{hp.neg, gammaMul6(hp.v, gammaFix256(erf256E0[k]))}
	if e0 := int(erf256E0Exp[k]); e0 < 0 {
		corr.v = corr.v.shr(uint(-e0))
	} else {
		corr.v = corr.v.shl(uint(e0))
	}
	return lgammaFix256{false, gammaFix256(erf256Erf[k])}.add(corr).v, 0, false
}

// erf256Scaled returns erfc(x) = v × 2**e for 4 <= x < 16 in a relative accuracy, where x is in fixed point with
// 320 fractional bits and v is in [1, 2) in fixed point with 320 fractional bits. The relative error is less than 2**-310.
func erf256Scaled(x gammaFix256) (v gammaFix256, e int) {
	k := int(x[0]<<4 | x[1]>>60)
	hp := erf256Taylor(x, k, int(erf256ErfcTerms[k-64]))

	// erfc(x) = erfc(x0) - 2/sqrt(pi) exp(-x0**2) h p = erfc(x0) (1 - R h p) is calculated in a relative accuracy,
	// where R = 2/sqrt(pi) exp(-x0**2)/erfc(x0). 1 - R h p is in [1/e, e].
	t := lgammaFix256{hp.neg, gammaMul6(hp.v, gammaFix256(erf256R[k-64]))}
	w := lgammaFix256{false, gammaOne256}.add(t.negate())
	return gammaNormalize256(gammaMul6(gammaFix256(erf256ErfcMant[k-64]), w.v), int(erf256ErfcExp[k-64]))
}

// erf256Taylor returns h p, where h = x - x0 and p is the sum of c_i h**i with n+1 terms. The cell k includes x.
func erf256Taylor(x gammaFix256, k, n int) lgammaFix256 {
	// erf(x0+h) = erf(x0) + 2/sqrt(pi) exp(-x0**2) sum b_n h**(n+1)/(n+1), where x0 = (k + 1/2)/16 is the center of the cell
	// that includes x, and exp(-2 x0 h - h**2) = sum b_n h**n. The recurrence is
	// b_0 = 1, b_1 = -2 x0, and b_(n+1) = -(2 x0 b_n + 2 b_(n-1))/(n+1).
	// x0 = (2k+1)/32
	x0 := gammaFix256{uint64(2*k+1) >> 5, uint64(2*k+1) << 59}
	var h lgammaFix256
	if x.cmp(x0) >= 0 {
		h = lgammaFix256{false, x.sub(x0)}
	} else {
		h = lgammaFix256{true, x0.sub(x)}
	}
	// c_i = b_i/(i+1) satisfies c_0 = 1, c_1 = -x0, and
	// c_(i+1) = -(2 x0 (i+1) c_i + 2 i c_(i-1))/((i+1) (i+2)) = -((2k+1) (i+1) c_i + 32 i c_(i-1))/d,
	// where d = 16 (i+1) (i+2). |c_i| is less than 2**36, so that the numerator is less than 2**58.
	var c [80]lgammaFix256
	c[0] = lgammaFix256{false, gammaOne256}
	for i := range n {
		t := lgammaFix256{c[i].neg, c[i].v.mulUint(uint64((2*k + 1) * (i + 1)))}
		if i > 0 {
			t = t.add(lgammaFix256{c[i-1].neg, c[i-1].v.mulUint(uint64(32 * i))})
		}
		// t/d = t × erf256InvD[i] × 2**-l, where l is the bit length of d-1.
		l := bits.Len64(uint64(16*(i+1)*(i+2)) - 1)
		c[i+1] = lgammaFix256{!t.neg, gammaMul6(t.v, gammaFix256(erf256InvD[i])).shr(uint(l))}
	}
	// sum c_i h**i by Horner's method.
	p := c[n]
	for i := n - 1; i >= 0; i-- {
		p = c[i].add(lgammaFix256{p.neg != h.neg, gammaMul6(p.v, h.v)})
	}
	return lgammaFix256{p.neg != h.neg, gammaMul6(p.v, h.v)}
}

// Erfc returns the complementary error function of a.
//
// Special cases are:
//
//	+Inf.Erfc() = 0
//	-Inf.Erfc() = 2
//	NaN.Erfc() = NaN
func (a Float256) Erfc() Float256 {
	var (
		one = Float256(uvone256)
		two = Float256{0x4000_0000_0000_0000, 0, 0, 0}
	)

	switch {
	case a.IsInf(1):
		return Float256{}
	case a.IsInf(-1):
		return two
	case a.IsNaN():
		return NewFloat256NaN()
	case a.IsZero():
		return one
	}

	// |a| = m × 2**(exp-236), where m is a 237-bit integer in [2**236, 2**237).
	sign, exp, m := a.normalize()
	neg := sign != 0
	if exp >= 3 {
		abs := a.Abs()
		if !neg && (exp > 8 || abs.Ge(Float256{0x4000_7aa7_382a_14ed, 0x2d77_5059_d7cf_77a2, 0xca42_44d8_9adb_8ce1, 0x0325_ae1c_e38f_78c4})) {
			// erfc(x) is rounded to 0 for x >= 426.45.
			return Float256{}
		}
		if neg && (exp > 3 || abs.Ge(Float256{0x4000_2963_c382_2856, 0xda94_b711_9017_42fd, 0x8025_b5e1_20c3_3205, 0xfccc_8b78_050a_bc50})) {
			// erfc(x) is rounded to 2 for x <= -12.695 (2**-237 > erfc(|x|)).
			return two
		}
	}

	if exp < -8 {
		// erfc(x) = 1 ∓ erf(x). erf(x) is smaller than 2**-238 if |x| < 2**-240.
		if exp < -240 {
			return one
		}
		return erf256OneMinus(neg, erf256SmallSeries(exp, m).shr(uint(-exp)))
	}

	// |x| in fixed point with 320 fractional bits. It is exact.
	x := gammaFix256FromUint256(m, uint(exp+84))
	if exp < 2 {
		// |x| < 4. erfc(x) is larger than 2**-26, so that 1 - erf(x) is accurate enough.
		v, _, _ := erf256Cell(x)
		return erf256OneMinus(neg, v)
	}

	var v gammaFix256
	var e int
	if exp < 4 {
		v, e = erf256Scaled(x)
	} else {
		v, e = erfc256Large(x, exp, m)
	}
	// erfc(|x|) = v × 2**e
	if !neg {
		return lgamma256FromPair(false, v, e)
	}
	// erfc(-x) = 2 - erfc(x), where 2**-237 <= erfc(x) < 2**-26. The result is 2 - j 2**-236 for the integer j nearest to
	// erfc(x) 2**236, which is exactly representable.
	j := erf256Round(ints.Uint512{2: v[0], 3: v[1], 4: v[2], 5: v[3], 6: v[4], 7: v[5]}, e+236)
	mant := ints.Uint256{1 << 45, 0, 0, 0}.Sub(j)
	return fixToFloat256(0, mant.Uint512(), false, -236)
}

// erf256OneMinus returns 1 - v for x > 0 or 1 + v for x < 0, where v is erf(|x|) in fixed point with 320 fractional bits.
func erf256OneMinus(neg bool, v gammaFix256) Float256 {
	if neg {
		return lgamma256FromPair(false, gammaOne256.add(v), 0)
	}
	return lgamma256FromPair(false, gammaOne256.sub(v), 0)
}

// erfc256Large returns erfc(x) = v × 2**e for 16 <= x < 512 by the asymptotic expansion
// erfc(x) = exp(-x**2)/(x sqrt(pi)) sum (-1)**n (2n-1)!!/(2 x**2)**n,
// where x = m × 2**(exp-236) is in fixed point with 320 fractional bits.
// v is in fixed point with 320 fractional bits, and the relative error is less than 2**-275.
func erfc256Large(x gammaFix256, exp int, m ints.Uint256) (v gammaFix256, e int) {
	// exp(-x**2) = em × 2**k
	em, k := gammaExpNeg256(gammaMul6(x, x))

	// 1/x = r × 2**-exp, where r = 1/xm for the mantissa xm of x in [1, 2).
	r := gammaRecip256(gammaFix256FromUint256(m, 84))
	// w = 1/(2 x**2)
	w := gammaMul6(r, r).shr(uint(2*exp + 1))

	// the sum of (-1)**n t_n, where t_n = t_(n-1) (2n-1) w. The terms are decreasing until t_n < 2**-280,
	// because (2n-1) w < 1 for n < x**2.
	s := gammaOne256
	t := gammaOne256
	for n := 1; ; n++ {
		t = gammaMul6(t, w).mulUint(uint64(2*n - 1))
		if t[0]|t[1]|t[2]|t[3]|t[4] == 0 && t[5]>>40 == 0 {
			break
		}
		if n%2 != 0 {
			s = s.sub(t)
		} else {
			s = s.add(t)
		}
	}

	p := gammaMul6(em, r)
	p = gammaMul6(p, gammaFix256(erf256InvSqrtPi))
	return gammaNormalize256(gammaMul6(p, s), k-exp)
}

// Erfinv returns the inverse error function of a.
//
// Special cases are:
//
//	1.Erfinv() = +Inf
//	-1.Erfinv() = -Inf
//	x.Erfinv() = NaN if x < -1 or x > 1
//	NaN.Erfinv() = NaN
func (a Float256) Erfinv() Float256 {
	var (
		one = Float256(uvone256)
		// SqrtPiOverTwo is sqrt(π)/2
		sqrtPiOverTwo = Float128{0x3ffe_c5bf_891b_4ef6, 0xaa79_c3b0_520d_5db9}
	)

	// special cases
	switch {
	case a.IsNaN():
		return NewFloat256NaN()
	case a.IsZero():
		return a
	case a.Eq(one):
		return NewFloat256Inf(1)
	case a.Eq(one.Neg()):
		return NewFloat256Inf(-1)
	case a.Lt(one.Neg()) || a.Gt(one):
		return NewFloat256NaN()
	}

	x := a.Abs()
	_, exp, m := x.normalize()
	var y Float256
	switch {
	case exp < -125:
		// erfinv(x) = sqrt(π)/2 x (1 + π/12 x**2 + ...), and the next term is less than 2**-480 relative.
		xm := gammaFix256FromUint256(m, 84)
		v := gammaMul6(gammaFix256(erf256SqrtPiOverTwo), xm)
		corr := gammaMul6(gammaMul6(gammaFix256(erf256PiOver12), xm), xm).shr(uint(-2 * exp))
		y = lgamma256FromPair(false, v.add(gammaMul6(v, corr)), exp)
	case !x.Gt(Float256{0x3fff_e000_0000_0000}):
		// |a| <= 1/2. Newton's method for erf(y) = x. The initial approximation has about 50 correct bits, and each
		// iteration doubles them. The iterations are calculated in Float128 (the error is about 2**-104),
		// in Float256 (2**-208), and in the fixed point arithmetic (less than 2**-300), with the factor
		// sqrt(π)/2 exp(y**2) whose precision is enough for each iteration.
		y0 := math.Erfinv(x.Float64().BuiltIn())
		y1 := NewFloat128(y0)
		y1 = y1.Sub(y1.Erf().Sub(x.Float128()).Mul(sqrtPiOverTwo).Mul(NewFloat128(math.Exp(y0 * y0))))
		y2 := y1.Float256()
		y2 = y2.Sub(y2.Erf().Sub(x).Mul(sqrtPiOverTwo.Mul(y1.Mul(y1).Exp()).Float256()))
		factor := sqrtPiOverTwo.Mul(y2.Float128().Mul(y2.Float128()).Exp()).Float256()
		y = y2.Sub(erf256Defect(y2, x, false).Mul(factor))
	default:
		// 1/2 < |a| < 1. Newton's method for erfc(y) = 1 - x, which is exact. It keeps the relative accuracy
		// if x is close to 1.
		c := one.Sub(x)
		y0 := erfcinv64(c.Float64().BuiltIn())
		y1 := NewFloat128(y0)
		y1 = y1.Add(y1.Erfc().Sub(c.Float128()).Mul(sqrtPiOverTwo).Mul(NewFloat128(math.Exp(y0 * y0))))
		y2 := y1.Float256()
		y2 = y2.Add(y2.Erfc().Sub(c).Mul(sqrtPiOverTwo.Mul(y1.Mul(y1).Exp()).Float256()))
		factor := sqrtPiOverTwo.Mul(y2.Float128().Mul(y2.Float128()).Exp()).Float256()
		y = y2.Add(erf256Defect(y2, c, true).Mul(factor))
	}
	if a.Signbit() {
		return y.Neg()
	}
	return y
}

// erf256Defect returns erf(y) - target if complement is false, or erfc(y) - target if complement is true,
// which is calculated with the relative accuracy about 2**-300 of the values of the functions,
// for 2**-125 <= y < 0.48 (erf) or 0.47 < y < 13 (erfc), and the target is close to the value of the function.
func erf256Defect(y, target Float256, complement bool) Float256 {
	_, exp, m := y.normalize()
	// the value of the function is v × 2**e.
	var v gammaFix256
	var e int
	switch {
	case !complement && exp < -8:
		v, e = erf256SmallSeries(exp, m), exp
	case !complement:
		v, _, _ = erf256Cell(gammaFix256FromUint256(m, uint(exp+84)))
	case exp < 2:
		// erfc(y) is larger than 2**-26, so that 1 - erf(y) is accurate enough.
		v, _, _ = erf256Cell(gammaFix256FromUint256(m, uint(exp+84)))
		v = gammaOne256.sub(v)
	default:
		v, e = erf256Scaled(gammaFix256FromUint256(m, uint(exp+84)))
	}

	// the target in the same unit, which is close to the value, so that the shift is positive.
	_, texp, tm := target.normalize()
	return erf256FixDiff(v, gammaFix256FromUint256(tm, uint(84+texp-e)), e)
}

// erf256FixDiff returns (v - t) × 2**e as Float256.
func erf256FixDiff(v, t gammaFix256, e int) Float256 {
	switch v.cmp(t) {
	case 0:
		return Float256{}
	case 1:
		return lgamma256FromPair(false, v.sub(t), e)
	}
	return lgamma256FromPair(true, t.sub(v), e)
}

// Erfcinv returns the inverse of [Erfc](a).
//
// Special cases are:
//
//	0.Erfcinv() = +Inf
//	2.Erfcinv() = -Inf
//	x.Erfcinv() = NaN if x < 0 or x > 2
//	NaN.Erfcinv() = NaN
func (a Float256) Erfcinv() Float256 {
	return (Float256(uvone256).Sub(a)).Erfinv()
}
