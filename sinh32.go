package floats

// Sinh returns the hyperbolic sine of x.
//
// Special cases are:
//
//	±0.Sinh() = ±0
//	±Inf.Sinh() = ±Inf
//	NaN.Sinh() = NaN
func (a Float32) Sinh() Float32 {
	// The coefficients are #2029 from Hart & Cheney. (20.36D)
	// copy from https://github.com/chewxy/math32/blob/912ef0b2e4151df0148d7645c92a7b5e22f887f5/sinhf.go#L12-L20
	const (
		P0 = -0.6307673640497716991184787251e+6
		P1 = -0.8991272022039509355398013511e+5
		P2 = -0.2894211355989563807284660366e+4
		P3 = -0.2630563213397497062819489e+2
		Q0 = -0.6307673640497716991212077277e+6
		Q1 = 0.1521517378790019070696485176e+5
		Q2 = -0.173678953558233699533450911e+3
	)

	sign := false
	if a < 0 {
		a = -a
		sign = true
	}

	var temp Float32
	switch {
	case a > 21:
		temp = a.Exp() * 0.5

	case a > 0.5:
		ex := a.Exp()
		temp = (ex - 1/ex) * 0.5

	default:
		sq := a * a
		temp = (((P3*sq+P2)*sq+P1)*sq + P0) * a
		temp = temp / ((((sq+Q2)*sq)+Q1)*sq + Q0)
	}

	if sign {
		temp = -temp
	}
	return temp
}
