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
	ix := a.Bits()
	if ix-uvone32 >= uvinf32-uvone32 { // a < 1, +Inf, or NaN
		if a > 1 {
			// acosh(+Inf) = +Inf
			return a
		}
		// a < 1 or NaN
		return NewFloat32NaN()
	}

	x := float64(a)
	t := x - 1 // exact
	var r float64
	switch {
	case t < 1e-4:
		// acosh(x) = log(1+u), where u = t + sqrt(t*(2+t)) < 1/64.
		r = log1pKernel(t + math.Sqrt(t*(2+t)))
	case t < 0.03:
		// acosh(x) = log(1+u) = log(v) + log(1 + d/v),
		// where v = 1+u rounded to float64, and d = 1+u - v is exact.
		u := t + math.Sqrt(t*(2+t))
		v := 1 + u
		d := (1 - v) + u
		k, l := logKernel64(v)
		r = k*math.Ln2 + l + d/v
	default:
		// acosh(x) = log(x + sqrt(x**2-1)), where the result is greater than 0.24.
		k, l := logKernel64(x + math.Sqrt(x*x-1))
		r = k*math.Ln2 + l
	}
	return Float32(r)
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
	ix := a.Bits() &^ signMask32
	if ix < 0x39800000 {
		// |a| < 2**-12: atanh(a) = a + a**3/3 + ... rounds to a.
		return a
	}
	if ix >= uvone32 {
		switch a {
		case 1:
			return NewFloat32Inf(1)
		case -1:
			return NewFloat32Inf(-1)
		default:
			// |a| > 1 or NaN
			return NewFloat32NaN()
		}
	}

	// atanh(x) = log((1+x)/(1-x))/2 = log(1+w)/2, where w = 2x/(1-x).
	// 1-x is exact.
	x := float64(math.Float32frombits(ix))
	var r float64
	switch {
	case x < 0.0075:
		// w < 1/64
		r = log1pKernel(2 * x / (1 - x))
	case x < 0.25:
		// log(1+w) = log(v) + log(1 + d/v),
		// where v = 1+w rounded to float64, and d = 1+w - v is exact.
		w := 2 * x / (1 - x)
		v := 1 + w
		d := (1 - v) + w
		k, l := logKernel64(v)
		r = k*math.Ln2 + l + d/v
	default:
		// The result is greater than 0.25.
		k, l := logKernel64((1 + x) / (1 - x))
		r = k*math.Ln2 + l
	}
	r *= 0.5
	if a < 0 {
		r = -r
	}
	return Float32(r)
}
