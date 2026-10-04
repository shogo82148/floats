package floats

import "math"

// Asinh returns the inverse hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Asinh() = ±0
//	±Inf.Asinh() = ±Inf
//	NaN.Asinh() = NaN
func (a Float32) Asinh() Float32 {
	ix := a.Bits() &^ signMask32
	if ix < 0x39800000 || ix >= uvinf32 {
		// |a| < 2**-12: asinh(a) = a - a**3/6 + ... rounds to a.
		// asinh(±Inf) = ±Inf, asinh(NaN) = NaN
		return a
	}

	x := float64(math.Float32frombits(ix))
	x2 := x * x
	var r float64
	switch {
	case x < 0.015:
		// asinh(x) = log(1+u), where u = x + x**2/(1+sqrt(1+x**2)) < 1/64.
		r = log1pKernel(x + x2/(1+math.Sqrt(1+x2)))
	case x < 0.25:
		// asinh(x) = log(1+u) = log(v) + log(1 + d/v),
		// where v = 1+u rounded to float64, and d = 1+u - v is exact.
		u := x + x2/(1+math.Sqrt(1+x2))
		v := 1 + u
		d := (1 - v) + u
		k, l := logKernel64(v)
		r = k*math.Ln2 + l + d/v
	default:
		// asinh(x) = log(x + sqrt(x**2+1)), where the result is greater than 0.24.
		k, l := logKernel64(x + math.Sqrt(x2+1))
		r = k*math.Ln2 + l
	}
	if a < 0 {
		r = -r
	}
	return Float32(r)
}

// Acosh returns the inverse hyperbolic cosine of a.
//
// Special cases are:
//
//	+Inf.Acosh() = +Inf
//	x.Acosh() = NaN if x < 1
//	NaN.Acosh() = NaN
func (a Float32) Acosh() Float32 {
	// https://github.com/chewxy/math32/blob/912ef0b2e4151df0148d7645c92a7b5e22f887f5/acosh.go#L40-L53
	const Ln2 = 6.93147180559945286227e-01 // 0x3FE62E42FEFA39EF
	const Large = 1 << 28                  // 2**28
	// first case is special case
	switch {
	case a < 1 || a.IsNaN():
		return NewFloat32NaN()
	case a == 1:
		return 0
	case a >= Large:
		return a.Log() + Ln2 // a > 2**28
	case a > 2:
		return (2*a - 1./(a+(a*a-1).Sqrt())).Log() // 2**28 > a > 2
	}
	t := a - 1
	return (t + (2*t + t*t).Sqrt()).Log1p() // 2 >= a > 1
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
func (a Float32) Atanh() Float32 {
	// https://github.com/chewxy/math32/blob/912ef0b2e4151df0148d7645c92a7b5e22f887f5/atanh.go#L45-L73
	const NearZero = 1.0 / (1 << 28) // 2**-28
	// special cases
	switch {
	case a < -1 || a > 1 || a.IsNaN():
		return NewFloat32NaN()
	case a == 1:
		return NewFloat32Inf(1)
	case a == -1:
		return NewFloat32Inf(-1)
	}
	sign := false
	if a < 0 {
		a = -a
		sign = true
	}
	var temp Float32
	switch {
	case a < NearZero:
		temp = a
	case a < 0.5:
		temp = a + a
		temp = 0.5 * (temp + temp*a/(1-a)).Log1p()
	default:
		temp = 0.5 * ((a + a) / (1 - a)).Log1p()
	}
	if sign {
		temp = -temp
	}
	return temp
}
