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
		v := gammaMul(m.Uint256().Lsh(80), gammaMul(ints.Uint256(erf128TwoOverSqrtPi), p))
		return erf128FromFix(sign, v, exp)
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
	// erf(x0+h) = erf(x0) + 2/sqrt(pi) exp(-x0**2) sum b_n h**(n+1)/(n+1), where x0 = (k + 1/2)/16 is the center of the cell
	// that includes x, and exp(-2 x0 h - h**2) = sum b_n h**n. The recurrence is
	// b_0 = 1, b_1 = -2 x0, and b_(n+1) = -(2 x0 b_n + 2 b_(n-1))/(n+1).
	k := int(x[0]<<4 | x[1]>>60)
	x0 := ints.Uint256{3: uint64(2*k + 1)}.Lsh(187)
	var h lgammaFix
	if x.Cmp(x0) >= 0 {
		h = lgammaFix{false, x.Sub(x0)}
	} else {
		h = lgammaFix{true, x0.Sub(x)}
	}
	n := int(erf128Terms[k])
	var c [40]lgammaFix // c[i] = b_i/(i+1)
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
	hp := lgammaFix{p.neg != h.neg, gammaMul(p.v, h.v)}

	if k >= 112 {
		// erfc(x) = erfc(x0) - 2/sqrt(pi) exp(-x0**2) h p = erfc(x0) (1 - R h p) is calculated in a relative accuracy,
		// where R = 2/sqrt(pi) exp(-x0**2)/erfc(x0). erfc(x) × 2**113 is less than 2**40 here.
		t := lgammaFix{hp.neg, gammaMul(hp.v, ints.Uint256(erf128R[k-112]))}
		es0 := ints.Uint256(erf128ErfcScaled[k-112])
		return lgammaFix{false, es0}.add(lgammaFix{!t.neg, gammaMul(es0, t.v)}).v, true
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
		// Zero is 0
		Zero = Float128{}
		// One is 1
		One = Float128(uvone128)
		// Two is 2
		Two = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
		// TwoOverSqrtPi is 2/sqrt(π)
		TwoOverSqrtPi = Float128{0x3fff_20dd_7504_29b6, 0xd11a_e3a9_14fe_d7fe}
		// TwoPointFour is 2.4
		TwoPointFour = Float128{0x4000_3333_3333_3333, 0x3333_3333_3333_3333}
		// Sqrt2 is sqrt(2)
		Sqrt2 = Float128{0x3fff_6a09_e667_f3bc, 0xc908_b2fb_1366_ea95}
		// SqrtTwoOverPi is sqrt(2/π)
		SqrtTwoOverPi = Float128{0x3ffe_9884_533d_4365, 0x08d0_fcb3_c500_bab9}
	)

	// special cases
	switch {
	case a.IsInf(1):
		return Zero
	case a.IsInf(-1):
		return Two
	case a.IsNaN():
		return NewFloat128NaN()
	}

	sign := false
	if a.Signbit() {
		sign = true
		a = a.Neg()
	}

	var y Float128
	switch {
	case a.Lt(TwoPointFour):
		// use Taylor series expansion
		// erf(x) = 2/sqrt(π) * Σ[n=0..∞] (-1)^n * x^(2n+1) / (n! * (2n+1))
		for n := 50; n >= 0; n-- {
			term := power128(a, 2*n+1).Quo(factorial128(n).Mul(NewFloat128(float64(2*n + 1))))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
		y = One.Sub(y.Mul(TwoOverSqrtPi))

	default:
		// use continued fraction expansion
		x := Sqrt2.Mul(a)
		for n := 90; n >= 1; n-- {
			y = NewFloat128(float64(n)).Quo(x.Add(y))
		}
		y = a.Mul(a).Neg().Exp().Quo(x.Add(y)).Mul(SqrtTwoOverPi)
	}
	if sign {
		y = Two.Sub(y)
	}
	return y
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
