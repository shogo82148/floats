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
		v := gammaMul6(gammaFix256FromUint256(m, 84), gammaMul6(gammaFix256(erf256TwoOverSqrtPi), p))
		return lgamma256FromPair(neg, v, exp)
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

// erf256Round returns the integer nearest to v × 2**(e-320), where v is a 384-bit integer, and the result is
// less than 2**200. If it is exactly half-way between two integers, it returns the even one.
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
	// erf(x0+h) = erf(x0) + 2/sqrt(pi) exp(-x0**2) sum b_n h**(n+1)/(n+1), where x0 = (k + 1/2)/16 is the center of the cell
	// that includes x, and exp(-2 x0 h - h**2) = sum b_n h**n. The recurrence is
	// b_0 = 1, b_1 = -2 x0, and b_(n+1) = -(2 x0 b_n + 2 b_(n-1))/(n+1).
	k := int(x[0]<<4 | x[1]>>60)
	// x0 = (2k+1)/32
	x0 := gammaFix256{uint64(2*k+1) >> 5, uint64(2*k+1) << 59}
	var h lgammaFix256
	if x.cmp(x0) >= 0 {
		h = lgammaFix256{false, x.sub(x0)}
	} else {
		h = lgammaFix256{true, x0.sub(x)}
	}
	n := int(erf256Terms[k])
	var c [80]lgammaFix256 // c[i] = b_i/(i+1)
	bp, b := lgammaFix256{}, lgammaFix256{false, gammaOne256}
	c[0] = b
	for i := range n {
		// b_(i+1) = -(2 x0 b_i + 2 b_(i-1))/(i+1), and 2 x0 = (2k+1)/16.
		t := lgammaFix256{b.neg, b.v.mulUint(uint64(2*k + 1)).shr(4)}
		t = t.add(lgammaFix256{bp.neg, bp.v.shl(1)})
		bp, b = b, lgammaFix256{!t.neg, t.v.divUint(uint64(i + 1))}
		c[i+1] = lgammaFix256{b.neg, b.v.divUint(uint64(i + 2))}
	}
	// sum c_i h**i by Horner's method.
	p := c[n]
	for i := n - 1; i >= 0; i-- {
		p = c[i].add(lgammaFix256{p.neg != h.neg, gammaMul6(p.v, h.v)})
	}
	hp := lgammaFix256{p.neg != h.neg, gammaMul6(p.v, h.v)}

	if k >= 112 {
		// erfc(x) = erfc(x0) - 2/sqrt(pi) exp(-x0**2) h p = erfc(x0) (1 - R h p) is calculated in a relative accuracy,
		// where R = 2/sqrt(pi) exp(-x0**2)/erfc(x0).
		t := lgammaFix256{hp.neg, gammaMul6(hp.v, gammaFix256(erf256R[k-112]))}
		w := lgammaFix256{false, gammaOne256}.add(t.negate())
		m, de := gammaNormalize256(gammaMul6(gammaFix256(erf256Es[k-112]), w.v), 0)
		return m, int(erf256EsExp[k-112]) + de, true
	}

	// the correction is 2/sqrt(pi) exp(-x0**2) h p = m0 × 2**e0 × h p.
	corr := lgammaFix256{hp.neg, gammaMul6(hp.v, gammaFix256(erf256E0[k]))}
	if e0 := int(erf256E0Exp[k]); e0 < 0 {
		corr.v = corr.v.shr(uint(-e0))
	} else {
		corr.v = corr.v.shl(uint(e0))
	}
	return lgammaFix256{false, gammaFix256(erf256Erf[k])}.add(corr).v, 0, false
}

// divUint returns v/d for the fixed point number v.
func (v gammaFix256) divUint(d uint64) gammaFix256 {
	var r gammaFix256
	var rem uint64
	for i := range v {
		r[i], rem = bits.Div64(rem, v[i], d)
	}
	return r
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
		// Zero is 0
		Zero = Float256{}
		// One is 1
		One = Float256(uvone256)
		// Two is 2
		Two = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		// Half is 0.5
		Half = Float256{
			0x3fff_e000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		// SqrtPiOverTwo is sqrt(π)/2
		SqrtPiOverTwo = Float256{
			0x3fff_ec5b_f891_b4ef, 0x6aa7_9c3b_0520_d5db,
			0x9383_fe39_2154_6f63, 0xb252_dca1_00bd_3ea1,
		}
	)

	// special cases
	switch {
	case a.Eq(One):
		return NewFloat256Inf(1)
	case a.Eq(One.Neg()):
		return NewFloat256Inf(-1)
	case a.Lt(One.Neg()) || a.Gt(One):
		return NewFloat256NaN()
	case a.IsNaN():
		return NewFloat256NaN()
	}

	sign := a.Signbit()
	if sign {
		a = a.Neg()
	}

	if a.Gt(Half) {
		// bisection search
		lo := Zero
		hi := One
		for hi.Erf().Lt(a) {
			hi = hi.Mul(Two)
		}
		for range 256 {
			mid := lo.Add(hi).Mul(Half)
			if mid.Erf().Lt(a) {
				lo = mid
			} else {
				hi = mid
			}
		}
		mid := lo.Add(hi).Mul(Half)
		if sign {
			return mid.Neg()
		}
		return mid
	}

	// Initial approximation using built-in math.Erfinv
	fa := a.Float64().BuiltIn()
	x := NewFloat256(math.Erfinv(fa))

	// Newton-Raphson iteration
	for range 700 {
		diff := x.Erf().Sub(a)
		exp := SqrtPiOverTwo.Mul(x.Mul(x).Exp())
		xn := x.Sub(diff.Mul(exp))
		if xn.Eq(x) {
			break
		}
		x = xn
	}
	if sign {
		x = x.Neg()
	}
	return x
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
