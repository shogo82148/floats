package floats

import "math"

// The mathematical functions of BFloat16 are calculated in float64 first.
// The result is rounded to BFloat16 if it is far enough from the midpoint of
// two adjacent BFloat16 values for the error of the float64 calculation.
// Otherwise, they are calculated again in Float256.

// bf16Window is the width of the neighborhood of the midpoints, in ulps of BFloat16,
// in which the rounding of the float64 results is not trusted.
// The errors of the functions of the math package are less than 2^-40 ulps of BFloat16.
const bf16Window = 0x1p-24

// bf16Round rounds y, a result in float64, to BFloat16.
// It returns false if y is too close to the midpoint of two adjacent BFloat16 values.
func bf16Round(y float64) (BFloat16, bool) {
	e := int(math.Float64bits(y)>>52&0x7ff) - 1023
	if e > 127 {
		// Inf, NaN, or overflow.
		return NewBFloat16(y), true
	}
	if e < -126 {
		// subnormal numbers of BFloat16 have a fixed interval.
		e = -126
	}
	// m is y in the unit of the ULP of BFloat16.
	m := math.Abs(y) * math.Float64frombits(uint64(1023+7-e)<<52)
	f := m - math.Floor(m)
	if d := f - 0.5; d < bf16Window && d > -bf16Window {
		return 0, false
	}
	return NewBFloat16(y), true
}

// bf16Result rounds y, a result in float64, to BFloat16.
// If y is too close to the midpoint, it calls slow, which calculates the result in higher precision.
func bf16Result(y float64, slow func() BFloat16) BFloat16 {
	if r, ok := bf16Round(y); ok {
		return r
	}
	return slow()
}

// bf16Unary calculates the function f of a in float64 or, if it is not enough, g in Float256.
func bf16Unary(a BFloat16, f func(float64) float64, g func(Float256) Float256) BFloat16 {
	return bf16Result(f(a.Float64().BuiltIn()), func() BFloat16 { return g(a.Float256()).BFloat16() })
}

// bf16UnaryInt calculates the function f of n and a in float64 or, if it is not enough, g in Float256.
func bf16UnaryInt(a BFloat16, n int, f func(int, float64) float64, g func(Float256, int) Float256) BFloat16 {
	return bf16Result(f(n, a.Float64().BuiltIn()), func() BFloat16 { return g(a.Float256(), n).BFloat16() })
}

// bf16Binary calculates the function f of a and b in float64 or, if it is not enough, g in Float256.
func bf16Binary(a, b BFloat16, f func(x, y float64) float64, g func(x, y Float256) Float256) BFloat16 {
	return bf16Result(f(a.Float64().BuiltIn(), b.Float64().BuiltIn()), func() BFloat16 { return g(a.Float256(), b.Float256()).BFloat16() })
}

// Exp returns e**x, the base-e exponential of a.
//
// Special cases are:
//
//	+Inf.Exp() = +Inf
//	NaN.Exp() = NaN
//
// Very large values overflow to 0 or +Inf.
// Very small values underflow to 1.
func (a BFloat16) Exp() BFloat16 {
	return bf16Unary(a, math.Exp, Float256.Exp)
}

// Exp2 returns 2**x, the base-2 exponential of x.
//
// Special cases are the same as [Exp].
func (a BFloat16) Exp2() BFloat16 {
	return bf16Unary(a, math.Exp2, Float256.Exp2)
}

// Expm1 returns e**a - 1, the base-e exponential of a minus 1.
// It is more accurate than Exp(a) - 1 when a is near zero.
//
// Special cases are:
//
//	+Inf.Expm1() = +Inf
//	-Inf.Expm1() = -1
//	NaN.Expm1() = NaN
//
// Very large values overflow to -1 or +Inf.
func (a BFloat16) Expm1() BFloat16 {
	return bf16Unary(a, math.Expm1, Float256.Expm1)
}

// Log returns the natural logarithm of a.
//
// Special cases are:
//
//	+Inf.Log() = +Inf
//	0.Log() = -Inf
//	(x < 0).Log() = NaN
//	NaN.Log() = NaN
func (a BFloat16) Log() BFloat16 {
	return bf16Unary(a, math.Log, Float256.Log)
}

// Log2 returns the binary logarithm of a.
// The special cases are the same as for [Log].
func (a BFloat16) Log2() BFloat16 {
	return bf16Unary(a, math.Log2, Float256.Log2)
}

// Log10 returns the decimal logarithm of a.
// The special cases are the same as for [Log].
func (a BFloat16) Log10() BFloat16 {
	return bf16Unary(a, math.Log10, Float256.Log10)
}

// Log1p returns the natural logarithm of 1 plus its argument a.
// It is more accurate than [Log](1 + a) when a is near zero.
//
// Special cases are:
//
//	+Inf.Log1p() = +Inf
//	±0.Log1p() = ±0
//	-1.Log1p() = -Inf
//	(a < -1).Log1p() = NaN
//	NaN.Log1p() = NaN
func (a BFloat16) Log1p() BFloat16 {
	return bf16Unary(a, math.Log1p, Float256.Log1p)
}

// Sin returns the sine of the radian argument a.
//
// Special cases are:
//
//	±0.Sin() = ±0
//	±Inf.Sin() = NaN
//	NaN.Sin() = NaN
func (a BFloat16) Sin() BFloat16 {
	return bf16Unary(a, math.Sin, Float256.Sin)
}

// Cos returns the cosine of the radian argument a.
//
// Special cases are:
//
//	±Inf.Cos() = NaN
//	NaN.Cos() = NaN
func (a BFloat16) Cos() BFloat16 {
	return bf16Unary(a, math.Cos, Float256.Cos)
}

// Tan returns the tangent of the radian argument a.
//
// Special cases are:
//
//	±0.Tan() = ±0
//	±Inf.Tan() = NaN
//	NaN.Tan() = NaN
func (a BFloat16) Tan() BFloat16 {
	return bf16Unary(a, math.Tan, Float256.Tan)
}

// Asin returns the arcsine, in radians, of a.
//
// Special cases are:
//
//	±0.Asin() = ±0
//	x.Asin() = NaN if x < -1 or x > 1
func (a BFloat16) Asin() BFloat16 {
	return bf16Unary(a, math.Asin, Float256.Asin)
}

// Acos returns the arccosine, in radians, of a.
//
// Special case is:
//
//	x.Acos() = NaN if x < -1 or x > 1
func (a BFloat16) Acos() BFloat16 {
	return bf16Unary(a, math.Acos, Float256.Acos)
}

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a BFloat16) Atan() BFloat16 {
	return bf16Unary(a, math.Atan, Float256.Atan)
}

// Sinh returns the hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Sinh() = ±0
//	±Inf.Sinh() = ±Inf
//	NaN.Sinh() = NaN
func (a BFloat16) Sinh() BFloat16 {
	return bf16Unary(a, math.Sinh, Float256.Sinh)
}

// Cosh returns the hyperbolic cosine of a.
//
// Special cases are:
//
//	±0.Cosh() = 1
//	±Inf.Cosh() = +Inf
//	NaN.Cosh() = NaN
func (a BFloat16) Cosh() BFloat16 {
	return bf16Unary(a, math.Cosh, Float256.Cosh)
}

// Tanh returns the hyperbolic tangent of a.
//
// Special cases are:
//
//	±0.Tanh() = ±0
//	±Inf.Tanh() = ±1
//	NaN.Tanh() = NaN
func (a BFloat16) Tanh() BFloat16 {
	return bf16Unary(a, math.Tanh, Float256.Tanh)
}

// Asinh returns the inverse hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Asinh() = ±0
//	±Inf.Asinh() = ±Inf
//	NaN.Asinh() = NaN
func (a BFloat16) Asinh() BFloat16 {
	return bf16Unary(a, math.Asinh, Float256.Asinh)
}

// Acosh returns the inverse hyperbolic cosine of a.
//
// Special cases are:
//
//	+Inf.Acosh() = +Inf
//	x.Acosh() = NaN if x < 1
//	NaN.Acosh() = NaN
func (a BFloat16) Acosh() BFloat16 {
	return bf16Unary(a, math.Acosh, Float256.Acosh)
}

// Atanh returns the inverse hyperbolic tangent of a.
//
// Special cases are:
//
//	1.Atanh() = +Inf
//	±0.Atanh() = ±0
//	-1.Atanh() = -Inf
//	x.Atanh() = NaN if x < -1 or x > 1
//	NaN.Atanh() = NaN
func (a BFloat16) Atanh() BFloat16 {
	return bf16Unary(a, math.Atanh, Float256.Atanh)
}

// Cbrt returns the cube root of a.
//
// Special cases are:
//
//	±0.Cbrt() = ±0
//	±Inf.Cbrt() = ±Inf
//	NaN.Cbrt() = NaN
func (a BFloat16) Cbrt() BFloat16 {
	return bf16Unary(a, math.Cbrt, Float256.Cbrt)
}

// Gamma returns the Gamma function of a.
//
// Special cases are:
//
//	+Inf.Gamma() = +Inf
//	+0.Gamma() = +Inf
//	-0.Gamma() = -Inf
//	x.Gamma() = NaN for integer x < 0
//	-Inf.Gamma() = NaN
//	NaN.Gamma() = NaN
func (a BFloat16) Gamma() BFloat16 {
	return bf16Unary(a, math.Gamma, Float256.Gamma)
}

// Erf returns the error function of a.
//
// Special cases are:
//
//	+Inf.Erf() = 1
//	-Inf.Erf() = -1
//	NaN.Erf() = NaN
func (a BFloat16) Erf() BFloat16 {
	return bf16Unary(a, math.Erf, Float256.Erf)
}

// Erfc returns the complementary error function of a.
//
// Special cases are:
//
//	+Inf.Erfc() = 0
//	-Inf.Erfc() = 2
//	NaN.Erfc() = NaN
func (a BFloat16) Erfc() BFloat16 {
	return bf16Unary(a, math.Erfc, Float256.Erfc)
}

// Erfinv returns the inverse error function of a.
//
// Special cases are:
//
//	1.Erfinv() = +Inf
//	-1.Erfinv() = -Inf
//	x.Erfinv() = NaN if x < -1 or x > 1
//	NaN.Erfinv() = NaN
func (a BFloat16) Erfinv() BFloat16 {
	return bf16Unary(a, math.Erfinv, Float256.Erfinv)
}

// Erfcinv returns the inverse of [Erfc](a).
//
// Special cases are:
//
//	0.Erfcinv() = +Inf
//	2.Erfcinv() = -Inf
//	x.Erfcinv() = NaN if x < 0 or x > 2
//	NaN.Erfcinv() = NaN
func (a BFloat16) Erfcinv() BFloat16 {
	if a > 0 && a < 0x3c00 {
		// math.Erfcinv(x) calculates Erfinv(1-x), which loses the precision for small x.
		return a.Float256().Erfcinv().BFloat16()
	}
	if a.Signbit() && !a.IsZero() {
		// 1-x is rounded to 1 for small negative x, and math.Erfcinv returns +Inf.
		return uvnanbf16
	}
	return bf16Unary(a, math.Erfcinv, Float256.Erfcinv)
}

// J0 returns the order-zero Bessel function of the first kind.
//
// Special cases are:
//
//	J0(±Inf) = 0
//	J0(0) = 1
//	J0(NaN) = NaN
func (a BFloat16) J0() BFloat16 {
	return bf16Unary(a, math.J0, Float256.J0)
}

// J1 returns the order-one Bessel function of the first kind.
//
// Special cases are:
//
//	J1(±Inf) = 0
//	J1(NaN) = NaN
func (a BFloat16) J1() BFloat16 {
	if a.Signbit() && !a.IsZero() && !a.IsInf(0) {
		// math.J1(x) loses the sign for small negative x.
		return a.Neg().J1().Neg()
	}
	if abs := a &^ signMaskbf16; abs < 0x0100 {
		// J1(a) = a/2 - a**3/16 + ..., and a**3/16 is much smaller than the ULP.
		// For |a| < 2^-125, a/2 is in the middle of two BFloat16 values if the bits of a are odd.
		// The exact value is a bit smaller than a/2, so the result is not a tie:
		// it is a/2 rounded toward zero.
		return a&signMaskbf16 | abs>>1
	}
	return bf16Unary(a, math.J1, Float256.J1)
}

// Y0 returns the order-zero Bessel function of the second kind.
//
// Special cases are:
//
//	Y0(+Inf) = 0
//	Y0(0) = -Inf
//	Y0(x < 0) = NaN
//	Y0(NaN) = NaN
func (a BFloat16) Y0() BFloat16 {
	return bf16Unary(a, math.Y0, Float256.Y0)
}

// Y1 returns the order-one Bessel function of the second kind.
//
// Special cases are:
//
//	Y1(+Inf) = 0
//	Y1(0) = -Inf
//	Y1(x < 0) = NaN
//	Y1(NaN) = NaN
func (a BFloat16) Y1() BFloat16 {
	return bf16Unary(a, math.Y1, Float256.Y1)
}

// Atan2 returns the arc tangent of a/b, using
// the signs of the two to determine the quadrant
// of the return value.
//
// Special cases are (in order):
//
//	y.Atan2(NaN) = NaN
//	NaN.Atan2(x) = NaN
//	+0.Atan2(x>=0) = +0
//	-0.Atan2(x>=0) = -0
//	+0.Atan2(x<=-0) = +Pi
//	-0.Atan2(x<=-0) = -Pi
//	y>0.Atan2(0) = +Pi/2
//	y<0.Atan2(0) = -Pi/2
//	+Inf.Atan2(+Inf) = +Pi/4
//	-Inf.Atan2(+Inf) = -Pi/4
//	+Inf.Atan2(-Inf) = 3Pi/4
//	-Inf.Atan2(-Inf) = -3Pi/4
//	y.Atan2(+Inf) = 0
//	(y>0).Atan2(-Inf) = +Pi
//	(y<0).Atan2(-Inf) = -Pi
//	+Inf.Atan2(x) = +Pi/2
//	-Inf.Atan2(x) = -Pi/2
func (a BFloat16) Atan2(b BFloat16) BFloat16 {
	x, y := a.Float64().BuiltIn(), b.Float64().BuiltIn()
	if r := math.Abs(x / y); y > 0 && r != 0 && r < 0x1p-110 && !math.IsInf(y, 0) {
		// atan(t) = t - t**3/3 + ... is a bit smaller than t, and the difference is too small for Float256.
		// If a/b is the tie of two adjacent BFloat16 values, the result is rounded toward zero.
		return bf16RoundTieToZero(a.Float256().Quo(b.Float256()))
	}
	return bf16Binary(a, b, math.Atan2, Float256.Atan2)
}

// Pow returns a**b, the base-a exponential of b.
//
// Special cases are (in order):
//
//	a.Pow(±0) = 1 for any a
//	1.Pow(b) = 1 for any b
//	a.Pow(1) = a for any a
//	NaN.Pow(b) = NaN
//	a.Pow(NaN) = NaN
//	±0.Pow(b) = ±Inf for b an odd integer < 0
//	±0.Pow(-Inf) = +Inf
//	±0.Pow(+Inf) = +0
//	±0.Pow(b) = +Inf for finite b < 0 and not an odd integer
//	±0.Pow(b) = ±0 for b an odd integer > 0
//	±0.Pow(b) = +0 for finite b > 0 and not an odd integer
//	-1.Pow(±Inf) = 1
//	a.Pow(+Inf) = +Inf for |a| > 1
//	a.Pow(-Inf) = +0 for |a| > 1
//	a.Pow(+Inf) = +0 for |a| < 1
//	a.Pow(-Inf) = +Inf for |a| < 1
//	+Inf.Pow(b) = +Inf for b > 0
//	+Inf.Pow(b) = +0 for b < 0
//	-Inf.Pow(b) = (-0).Pow(-b)
//	a.Pow(b) = NaN for finite a < 0 and finite non-integer b
func (a BFloat16) Pow(b BFloat16) BFloat16 {
	return bf16Binary(a, b, math.Pow, Float256.Pow)
}

// HypotBF16 returns [Sqrt](p*p + q*q), taking care to avoid
// unnecessary overflow and underflow.
//
// Special cases are:
//
//	HypotBF16(±Inf, q) = +Inf
//	HypotBF16(p, ±Inf) = +Inf
//	HypotBF16(NaN, q) = NaN
//	HypotBF16(p, NaN) = NaN
func HypotBF16(p, q BFloat16) BFloat16 {
	return bf16Binary(p, q, math.Hypot, Hypot256)
}

// Jn returns the order-n Bessel function of the first kind.
//
// Special cases are:
//
//	Jn(n, ±Inf) = 0
//	Jn(n, NaN) = NaN
func (a BFloat16) Jn(n int) BFloat16 {
	switch n {
	case 0:
		return a.J0()
	case 1:
		// math.Jn(1, x) is math.J1(x), which has a problem for the small negative x.
		return a.J1()
	case -1:
		return a.J1().Neg()
	}
	return bf16UnaryInt(a, n, math.Jn, Float256.Jn)
}

// Yn returns the order-n Bessel function of the second kind.
//
// Special cases are:
//
//	Yn(n, +Inf) = 0
//	Yn(n >= 0, 0) = -Inf
//	Yn(n < 0, 0) = +Inf if n is odd, -Inf if n is even
//	Yn(n, x < 0) = NaN
//	Yn(n, NaN) = NaN
func (a BFloat16) Yn(n int) BFloat16 {
	return bf16UnaryInt(a, n, math.Yn, Float256.Yn)
}

// Lgamma returns the natural logarithm and sign (-1 or +1) of Gamma(a).
//
// Special cases are:
//
//	Lgamma(+Inf) = +Inf
//	Lgamma(0) = +Inf
//	Lgamma(-integer) = +Inf
//	Lgamma(-Inf) = -Inf
//	Lgamma(NaN) = NaN
func (a BFloat16) Lgamma() (BFloat16, int) {
	y, sign := math.Lgamma(a.Float64().BuiltIn())
	return bf16Result(y, a.lgammaSlow), sign
}

// lgammaSlow calculates the value of Lgamma in Float256.
func (a BFloat16) lgammaSlow() BFloat16 {
	l, _ := a.Float256().Lgamma()
	return l.BFloat16()
}

// Sincos returns Sin(a), Cos(a).
//
// Special cases are:
//
//	±0.Sincos() = ±0, 1
//	±Inf.Sincos() = NaN, NaN
//	NaN.Sincos() = NaN, NaN
func (a BFloat16) Sincos() (sin, cos BFloat16) {
	return a.Sin(), a.Cos()
}

// NewBFloat16Pow10 returns 10**n, the base-10 exponential of n.
//
// Special cases are:
//
//	NewBFloat16Pow10(n) =    0 for n < -40
//	NewBFloat16Pow10(n) = +Inf for n > 38
func NewBFloat16Pow10(n int) BFloat16 {
	return NewBFloat16(math.Pow10(n))
}

// bf16RoundTieToZero rounds a finite r to BFloat16.
// It is the same as r.BFloat16() except that the tie is rounded toward zero.
func bf16RoundTieToZero(r Float256) BFloat16 {
	bf := r.BFloat16()
	if r.IsZero() || bf.IsInf(0) {
		return bf
	}
	lo := bf
	if bf.Float256().Abs().Gt(r.Abs()) {
		lo = bf - 1 // the adjacent value toward zero
	}
	hi := lo + 1
	mid := lo.Float256().Add(hi.Float256()).Ldexp(-1)
	if r == mid {
		return lo
	}
	return bf
}
