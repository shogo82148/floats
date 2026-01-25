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
