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
func (a Float128) Erf() Float128 {
	switch {
	case a.IsInf(0):
		return Float128(uvone128).Copysign(a)
	case a.IsNaN():
		return NewFloat128NaN()
	case a.IsZero():
		return a
	}

	// |a| = m × 2**(exp-112), where m is a 113-bit integer in [2**112, 2**113).
	sign, exp, m := a.normalize()
	neg := sign != 0
	if exp >= 3 && (exp > 3 || a.Abs().Ge(Float128{0x4002_1778_42bc_e674, 0x48bc_471e_ea54_0735})) {
		// erf(x) is rounded to 1 for |x| >= 8.7334.
		return Float128(uvone128).Copysign(a)
	}

	if exp < -8 {
		return erf128FromFix(sign, erf128Small128(exp, m), exp)
	}

	// |x| in fixed point with 192 fractional bits. It is exact.
	v, scaled := erf128Cell(m.Uint256().Lsh(uint(80 + exp)))
	if !scaled {
		return lgamma128FromFix(lgammaFix{neg, v})
	}

	// v = erfc(x) × 2**113 is less than 2**40. erf(x) = 1 - erfc(x) is rounded to a multiple of 2**-113,
	// that is, 1 - j 2**-113 for the integer j nearest to v. j is not zero since |x| < 8.7334 here and v is larger than 1/2.
	j := erf128Round(v)
	// 2**113 - j is the mantissa of the result in [1/2, 1).
	mant := ints.Uint128{1 << 49, 0}.Sub(ints.Uint128{0, j})
	return fixToFloat128(sign, mant[0]<<15|mant[1]>>49, mant[1]<<15, 0, false, -192)
}

// erf128Small128 returns erf(x) × 2**-exp in fixed point with 192 fractional bits for x = m × 2**(exp-112), where m is
// a 113-bit integer and exp < -8.
func erf128Small128(exp int, m ints.Uint128) ints.Uint256 {
	// |x| < 2**-8. erf(x) = 2/sqrt(pi) x sum (-x**2)**n/(n! (2n+1)), and the terms decrease fast.
	// |x| in fixed point with 192 fractional bits, which is not accurate if |x| is tiny, but its square is negligible then.
	var x ints.Uint256
	if s := exp + 80; s >= 0 {
		x = m.Uint256().Lsh(uint(s))
	} else if s > -256 {
		x = m.Uint256().Rsh(uint(-s))
	}
	u := gammaMul(x, x)
	c := &erf128Small
	n := len(c)
	p := ints.Uint256(c[n-1])
	for i := n - 2; i >= 0; i-- {
		p = ints.Uint256(c[i]).Sub(gammaMul(u, p))
	}
	// the product of the mantissa of x and the series is exact except for the last bits.
	return gammaMul(m.Uint256().Lsh(80), gammaMul(ints.Uint256(erf128TwoOverSqrtPi), p))
}

// erf128Round returns the integer nearest to v in fixed point with 192 fractional bits, which is less than 2**64.
// If v is exactly half-way between two integers, it returns the even one.
func erf128Round(v ints.Uint256) uint64 {
	j := v[0]
	if v[1] > 1<<63 || v[1] == 1<<63 && (v[2]|v[3] != 0 || j&1 != 0) {
		j++
	}
	return j
}

// erf128Cell returns erf(x) in fixed point with 192 fractional bits for 2**-8 <= x < 7, and erfc(x) × 2**113 otherwise
// for x < 8.74, where x is in fixed point with 192 fractional bits. scaled reports the latter.
// The relative error is less than 2**-170.
func erf128Cell(x ints.Uint256) (v ints.Uint256, scaled bool) {
	k := int(x[0]<<4 | x[1]>>60)
	hp := erf128Taylor(x, k)
	if k >= 112 {
		// erfc(x) = erfc(x0) (1 - R h p) × 2**113 is less than 2**40 here.
		e := int(erf128ErfcExp[k-64]) + 113
		es0 := ints.Uint256(erf128ErfcMant[k-64])
		if e < 0 {
			es0 = es0.Rsh(uint(-e))
		} else {
			es0 = es0.Lsh(uint(e))
		}
		return erf128ScaledSum(es0, hp, k), true
	}

	// the correction is 2/sqrt(pi) exp(-x0**2) h p = m0 × 2**e0 × h p.
	corr := hp
	corr.v = gammaMul(corr.v, ints.Uint256(erf128E0[k]))
	if e := int(erf128E0Exp[k]); e < 0 {
		corr.v = corr.v.Rsh(uint(-e))
	} else {
		corr.v = corr.v.Lsh(uint(e))
	}
	return lgammaFix{false, ints.Uint256(erf128Erf[k])}.add(corr).v, false
}

// erf128Scaled returns erfc(x) = v × 2**e for 4 <= x < 16 in a relative accuracy, where x is in fixed point with
// 192 fractional bits and v is about in [1, 2) in fixed point with 192 fractional bits. The relative error is less than 2**-170.
func erf128Scaled(x ints.Uint256) (v ints.Uint256, e int) {
	k := int(x[0]<<4 | x[1]>>60)
	hp := erf128Taylor(x, k)
	return erf128ScaledSum(ints.Uint256(erf128ErfcMant[k-64]), hp, k), int(erf128ErfcExp[k-64])
}

// erf128ScaledSum returns es0 (1 - R h p), where es0 is erfc(x0) in any scale.
func erf128ScaledSum(es0 ints.Uint256, hp lgammaFix, k int) ints.Uint256 {
	// erfc(x) = erfc(x0) - 2/sqrt(pi) exp(-x0**2) h p = erfc(x0) (1 - R h p) is calculated in a relative accuracy,
	// where R = 2/sqrt(pi) exp(-x0**2)/erfc(x0).
	t := lgammaFix{hp.neg, gammaMul(hp.v, ints.Uint256(erf128R[k-64]))}
	return lgammaFix{false, es0}.add(lgammaFix{!t.neg, gammaMul(es0, t.v)}).v
}

// erf128Taylor returns h p, where p is the sum of c_i h**i and h = x - x0. The cell k includes x.
func erf128Taylor(x ints.Uint256, k int) lgammaFix {
	// erf(x0+h) = erf(x0) + 2/sqrt(pi) exp(-x0**2) sum b_n h**(n+1)/(n+1), where x0 = (k + 1/2)/16 is the center of the cell
	// that includes x, and exp(-2 x0 h - h**2) = sum b_n h**n. The recurrence is
	// b_0 = 1, b_1 = -2 x0, and b_(n+1) = -(2 x0 b_n + 2 b_(n-1))/(n+1).
	x0 := ints.Uint256{3: uint64(2*k + 1)}.Lsh(187)
	var h lgammaFix
	if x.Cmp(x0) >= 0 {
		h = lgammaFix{false, x.Sub(x0)}
	} else {
		h = lgammaFix{true, x0.Sub(x)}
	}
	n := int(erf128Terms[k])
	var c [42]lgammaFix // c[i] = b_i/(i+1), and the number of the terms is at most 41
	bp, b := lgammaFix{}, lgammaFix{false, gammaOne}
	c[0] = b
	for i := range n {
		// b_(i+1) = -(2 x0 b_i + 2 b_(i-1))/(i+1), and 2 x0 = (2k+1)/16.
		t := lgammaFix{b.neg, b.v.Mul(ints.Uint256{3: uint64(2*k + 1)}).Rsh(4)}
		t = t.add(lgammaFix{bp.neg, bp.v.Lsh(1)})
		bp, b = b, lgammaFix{!t.neg, gammaDivUint(t.v, uint64(i+1))}
		c[i+1] = lgammaFix{b.neg, gammaDivUint(b.v, uint64(i+2))}
	}
	// sum c_i h**i by Horner's method.
	p := c[n]
	for i := n - 1; i >= 0; i-- {
		p = c[i].add(lgammaFix{p.neg != h.neg, gammaMul(p.v, h.v)})
	}
	return lgammaFix{p.neg != h.neg, gammaMul(p.v, h.v)}
}

// gammaDivUint returns v/d for the fixed point number v.
func gammaDivUint(v ints.Uint256, d uint64) ints.Uint256 {
	var r ints.Uint256
	var rem uint64
	for i := range v {
		r[i], rem = bits.Div64(rem, v[i], d)
	}
	return r
}

// erf128FromFix returns ±v × 2**e as Float128 for v in fixed point with 192 fractional bits. v must not be zero.
func erf128FromFix(sign uint64, v ints.Uint256, e int) Float128 {
	lz := v.LeadingZeros()
	v = v.Lsh(uint(lz))
	return fixToFloat128(sign, v[0], v[1], v[2], v[3] != 0, e-128-int(lz))
}

// Erfc returns the complementary error function of x.
//
// Special cases are:
//
//	+Inf.Erfc() = 0
//	-Inf.Erfc() = 2
//	NaN.Erfc() = NaN
func (a Float128) Erfc() Float128 {
	var (
		one = Float128(uvone128)
		two = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	switch {
	case a.IsInf(1):
		return Float128{}
	case a.IsInf(-1):
		return two
	case a.IsNaN():
		return NewFloat128NaN()
	case a.IsZero():
		return one
	}

	// |a| = m × 2**(exp-112), where m is a 113-bit integer in [2**112, 2**113).
	sign, exp, m := a.normalize()
	neg := sign != 0
	if exp >= 3 {
		abs := a.Abs()
		if !neg && abs.Ge(Float128{0x4005_ab9c_8393_ddd2, 0x517e_5404_d6dd_3c9d}) {
			// erfc(x) is rounded to 0 for x >= 106.9, including +Inf.
			return Float128{}
		}
		if neg && abs.Ge(Float128{0x4002_1634_8a58_5939, 0x8f7a_8603_4cb5_b0b4}) {
			// erfc(x) is rounded to 2 for x <= -8.7334 (2**-113 > erfc(|x|)).
			return two
		}
	}

	if exp < -8 {
		// erfc(x) = 1 ∓ erf(x). erf(x) is smaller than 2**-113 if |x| < 2**-114.
		if exp < -120 {
			return one
		}
		return erf128OneMinus(neg, erf128Small128(exp, m).Rsh(uint(-exp)))
	}

	// |x| in fixed point with 192 fractional bits. It is exact.
	x := m.Uint256().Lsh(uint(80 + exp))
	if exp < 2 {
		// |x| < 4. erfc(x) is larger than 2**-26, so that 1 - erf(x) is accurate enough.
		v, _ := erf128Cell(x)
		return erf128OneMinus(neg, v)
	}

	var v ints.Uint256
	var e int
	if exp < 4 {
		v, e = erf128Scaled(x)
	} else {
		v, e = erfc128Large(x, exp, m)
	}
	// erfc(|x|) = v × 2**e
	if !neg {
		return erf128FromFix(0, v, e)
	}
	// erfc(-x) = 2 - erfc(x), where erfc(x) < 2**-26. The result is 2 - j 2**-112 for the integer j nearest to
	// erfc(x) 2**112 = v 2**(e+112-192).
	j := erf128RoundShift(v, uint(80-e))
	// 2 - j 2**-112 = 1 + (2**112 - j) 2**-112. If j = 0, the carry makes the exponent 2.
	f := ints.Uint128{1 << 48, 0}.Sub(j)
	return Float128{0x3fff<<48 | f[0], f[1]}
}

// erf128RoundShift returns the integer nearest to v/2**sh, which must be less than 2**128, for 2 <= sh < 256.
// If v/2**sh is exactly half-way between two integers, it returns the even one.
func erf128RoundShift(v ints.Uint256, sh uint) ints.Uint128 {
	r := v.Rsh(sh - 1)
	j := r.Rsh(1)
	if r[3]&1 != 0 && (!v.Lsh(257-sh).IsZero() || j[3]&1 != 0) {
		j = j.Add(ints.Uint256{3: 1})
	}
	return ints.Uint128{j[2], j[3]}
}

// erf128OneMinus returns 1 - v for x > 0 or 1 + v for x < 0, where v is erf(|x|) in fixed point with 192 fractional bits.
func erf128OneMinus(neg bool, v ints.Uint256) Float128 {
	if neg {
		return erf128FromFix(0, gammaOne.Add(v), 0)
	}
	return erf128FromFix(0, gammaOne.Sub(v), 0)
}

// erfc128Large returns erfc(x) = v × 2**e for x >= 16 by the asymptotic expansion
// erfc(x) = exp(-x**2)/(x sqrt(pi)) sum (-1)**n (2n-1)!!/(2 x**2)**n,
// where x = m × 2**(exp-112) is in fixed point with 192 fractional bits.
// v is in fixed point with 192 fractional bits, and the relative error is less than 2**-135.
func erfc128Large(x ints.Uint256, exp int, m ints.Uint128) (v ints.Uint256, e int) {
	// exp(-x**2) = em × 2**(k-192)
	em, k := gammaExpNeg192(gammaMul(x, x))

	// 1/x = r × 2**-exp, where r = 1/xm for the mantissa xm of x in [1, 2).
	r := gammaRecip192(m.Uint256().Lsh(80))
	// w = 1/(2 x**2)
	w := gammaMul(r, r).Rsh(uint(2*exp + 1))

	// the sum of (-1)**n t_n, where t_n = t_(n-1) (2n-1) w. The terms are decreasing until t_n < 2**-140.
	s := gammaOne
	t := gammaOne
	for n := 1; ; n++ {
		t = gammaMul(t, w).Mul(ints.Uint256{3: uint64(2*n - 1)})
		if t[0]|t[1]|t[2] == 0 && t[3]>>52 == 0 {
			break
		}
		if n%2 != 0 {
			s = s.Sub(t)
		} else {
			s = s.Add(t)
		}
	}

	p := gammaMul(em, r)
	p = gammaMul(p, ints.Uint256(erf128InvSqrtPi))
	return gammaMul(p, s), k - exp
}

// Erfinv returns the inverse error function of a.
//
// Special cases are:
//
//	1.Erfinv() = +Inf
//	-1.Erfinv() = -Inf
//	x.Erfinv() = NaN if x < -1 or x > 1
//	NaN.Erfinv() = NaN
func (a Float128) Erfinv() Float128 {
	var (
		// Zero is 0
		Zero = Float128{}
		// One is 1
		One = Float128(uvone128)
		// Two is 2
		Two = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
		// Half is 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}
		// SqrtPiOverTwo is sqrt(π)/2
		SqrtPiOverTwo = Float128{0x3ffe_c5bf_891b_4ef6, 0xaa79_c3b0_520d_5db9}
	)

	// special cases
	switch {
	case a.Eq(One):
		return NewFloat128Inf(1)
	case a.Eq(One.Neg()):
		return NewFloat128Inf(-1)
	case a.Lt(One.Neg()) || a.Gt(One):
		return NewFloat128NaN()
	case a.IsNaN():
		return NewFloat128NaN()
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
		for range 128 {
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
	x := NewFloat128(math.Erfinv(fa))

	// Newton-Raphson iteration
	for range 128 {
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
func (a Float128) Erfcinv() Float128 {
	return (Float128(uvone128).Sub(a)).Erfinv()
}
