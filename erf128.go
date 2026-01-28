package floats

import (
	"math"
)

// Erf returns the error function of a.
//
// Special cases are:
//
//	+Inf.Erf() = 1
//	-Inf.Erf() = -1
//	NaN.Erf() = NaN
func (a Float128) Erf() Float128 {
	var (
		// One is 1
		One = Float128(uvone128)
		// TwoOverSqrtPi is 2/sqrt(π)
		TwoOverSqrtPi = Float128{0x3fff_20dd_7504_29b6, 0xd11a_e3a9_14fe_d7fe}
		// TwoPointFour is 2.4
		TwoPointFour = Float128{0x4000_3333_3333_3333, 0x3333_3333_3333_3333}
		// Nine is 9
		Nine = Float128{0x4002_2000_0000_0000, 0x0000_0000_0000_0000}
		// Sqrt2 is sqrt(2)
		Sqrt2 = Float128{0x3fff_6a09_e667_f3bc, 0xc908_b2fb_1366_ea95}
		// SqrtTwoOverPi is sqrt(2/π)
		SqrtTwoOverPi = Float128{0x3ffe_9884_533d_4365, 0x08d0_fcb3_c500_bab9}
	)

	// special cases
	switch {
	case a.IsInf(0):
		return One.Copysign(a)
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
		y = y.Mul(TwoOverSqrtPi)

	case a.Gt(Nine):
		y = One

	default:
		// use continued fraction expansion
		x := Sqrt2.Mul(a)
		for n := 80; n >= 1; n-- {
			y = NewFloat128(float64(n)).Quo(x.Add(y))
		}
		y = One.Sub(a.Mul(a).Neg().Exp().Quo(x.Add(y)).Mul(SqrtTwoOverPi))
	}
	if sign {
		y = y.Neg()
	}
	return y
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
