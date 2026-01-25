package floats

// Asinh returns the inverse hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Asinh() = ±0
//	±Inf.Asinh() = ±Inf
//	NaN.Asinh() = NaN
func (a Float32) Asinh() Float32 {
	// https://github.com/chewxy/math32/blob/912ef0b2e4151df0148d7645c92a7b5e22f887f5/asinh.go#L36-L66
	const (
		Ln2      = 6.93147180559945286227e-01 // 0x3FE62E42FEFA39EF
		NearZero = 1.0 / (1 << 28)            // 2**-28
		Large    = 1 << 28                    // 2**28
	)
	// special cases
	if a.IsNaN() || a.IsInf(0) {
		return a
	}
	sign := false
	if a < 0 {
		a = -a
		sign = true
	}
	var temp Float32
	switch {
	case a > Large:
		temp = a.Log() + Ln2 // |a| > 2**28
	case a > 2:
		temp = (2*a + 1./((a*a+1).Sqrt()+a)).Log() // 2**28 > |a| > 2.0
	case a < NearZero:
		temp = a // |a| < 2**-28
	default:
		temp = (a + a*a/(1+(1+a*a).Sqrt())).Log1p() // 2.0 > |a| > 2**-28
	}
	if sign {
		temp = -temp
	}
	return temp
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
