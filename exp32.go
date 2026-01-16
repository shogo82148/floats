package floats

// Exp returns e**x, the base-e exponential of a.
//
// Special cases are:
//
//	+Inf.Exp() = +Inf
//	NaN.Exp() = NaN
//
// Very large values overflow to 0 or +Inf.
// Very small values underflow to 1.
func (a Float32) Exp() Float32 {
	const (
		// ln(2) split into high and low parts
		Ln2Hi = Float32(6.9313812256e-01)
		Ln2Lo = Float32(9.0580006145e-06)

		// log2(e)
		Log2e = Float32(1.4426950216e+00)

		// ln(max float32 + 0.5ulp) = ln(2¹²⁷×(2-2⁻²⁴))
		Overflow = Float32(88.7228390818706768)

		// ln(min float32 - 0.5ulp) = ln(2⁻¹⁵⁰)
		Underflow = Float32(-103.972077083991796)

		// The upper limit for underflow
		// when exp(a) ~ 1 + a + a²/2! + ...
		NearZero = Float32(0x1p-14) // 2**-14
	)

	// special cases
	switch {
	case a.IsNaN():
		return NewFloat32NaN()
	case a.IsInf(1):
		return NewFloat32Inf(1)
	case a.IsInf(-1):
		return 0
	case a > Overflow:
		return NewFloat32Inf(1)
	case a < Underflow:
		return 0
	case a.Abs() < NearZero:
		return 1 + a
	}

	// reduce; computed as r = hi - lo for extra precision.
	var k int
	switch {
	case a > 0:
		k = int(Log2e*a - 0.5)
	case a < 0:
		k = int(Log2e*a + 0.5)
	}
	hi := a - Float32(k)*Ln2Hi
	lo := Float32(k) * Ln2Lo

	// compute
	return expmulti32(hi, lo, k)
}

func expmulti32(hi, lo Float32, k int) Float32 {
	const (
		P1 = 1.66666666666666657415e-01
		P2 = -2.77777777770155933842e-03
		P3 = 6.61375632143793436117e-05
		P4 = -1.65339022054652515390e-06
		P5 = 4.13813679705723846039e-08
	)

	r := hi - lo
	t := r * r
	c := r - t*(P1+t*(P2+t*(P3+t*(P4+t*P5))))
	y := 1 - ((lo - (r*c)/(2-c)) - hi)
	return y.Ldexp(k)
}
