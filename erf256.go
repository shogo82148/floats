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
func (a Float256) Erf() Float256 {
	var (
		// One is 1
		One = Float256(uvone256)
		// TwoOverSqrtPi is 2/sqrt(π)
		TwoOverSqrtPi = Float256{
			0x3fff_f20d_d750_429b, 0x6d11_ae3a_914f_ed7f,
			0xd868_8281_341d_7587, 0xcea2_e734_2b06_199d,
		}
		// TwoPointFour is 2.4
		TwoPointFour = Float256{
			0x4000_0333_3333_3333, 0x3333_3333_3333_3333,
			0x3333_3333_3333_3333, 0x3333_3333_3333_3333,
		}
		// Thirteen is 13
		Thirteen = Float256{
			0x4000_2a00_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		// Sqrt2 is sqrt(2)
		Sqrt2 = Float256{
			0x3fff_f6a0_9e66_7f3b, 0xcc90_8b2f_b136_6ea9,
			0x57d3_e3ad_ec17_5127, 0x7509_9da2_f590_b066,
		}
		// SqrtTwoOverPi is sqrt(2/π)
		SqrtTwoOverPi = Float256{
			0x3fff_e988_4533_d436, 0x508d_0fcb_3c50_0bab,
			0x8e2f_f4e0_a7cc_1449, 0xc1b6_3011_c6d2_0099,
		}
	)

	// special cases
	switch {
	case a.IsInf(0):
		return Float256(uvone256).Copysign(a)
	case a.IsNaN():
		return NewFloat256NaN()
	}

	sign := false
	if a.Signbit() {
		sign = true
		a = a.Neg()
	}

	var y Float256
	switch {
	case a.Lt(TwoPointFour):
		// use Taylor series expansion
		// erf(x) = 2/sqrt(π) * Σ[n=0..∞] (-1)^n * x^(2n+1) / (n! * (2n+1))
		for n := 80; n >= 0; n-- {
			term := power256(a, 2*n+1).Quo(factorial256(n).Mul(NewFloat256(float64(2*n + 1))))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
		y = y.Mul(TwoOverSqrtPi)

	case a.Gt(Thirteen):
		y = One

	default:
		// use continued fraction expansion
		x := Sqrt2.Mul(a)
		for n := 400; n >= 1; n-- {
			y = NewFloat256(float64(n)).Quo(x.Add(y))
		}
		y = One.Sub(a.Mul(a).Neg().Exp().Quo(x.Add(y)).Mul(SqrtTwoOverPi))
	}
	if sign {
		y = y.Neg()
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
			0x3fff_c5bf_891b_4ef6, 0xaa79_c3b0_520d_5db9,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
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
