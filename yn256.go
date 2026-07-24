package floats

// Yn returns the order-n Bessel function of the second kind.
//
// Special cases are:
//
//	Yn(n, +Inf) = 0
//	Yn(n >= 0, 0) = -Inf
//	Yn(n < 0, 0) = +Inf if n is odd, -Inf if n is even
//	Yn(n, x < 0) = NaN
//	Yn(n, NaN) = NaN
func (a Float256) Yn(n int) Float256 {
	switch {
	case a.IsNaN():
		return a
	case a.IsInf(1):
		return Float256{}
	}

	neg := false
	if n < 0 {
		n = -n
		neg = !neg
	}

	switch {
	case a.IsZero():
		if neg && n%2 == 1 {
			return NewFloat256Inf(1)
		}
		return NewFloat256Inf(-1)
	case a.Signbit():
		return NewFloat256NaN()
	}

	var y Float256
	switch n {
	case 0:
		y = a.Y0()
	case 1:
		y = a.Y1()
	default:
		y = ynForward256(n, a)
	}
	if neg && n%2 == 1 {
		y = y.Neg()
	}
	return y
}

// ynForward256 returns Yn(x) for n >= 2, x > 0; see ynForward128 for the
// derivation.
func ynForward256(n int, x Float256) Float256 {
	var Two = Float256{
		0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
		0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
	}

	ykm1 := x.Y0()
	yk := x.Y1()
	for k := 1; k < n; k++ {
		coef := Two.Mul(NewFloat256(float64(k))).Quo(x)
		ykp1 := FMA256(coef, yk, ykm1.Neg())
		ykm1, yk = yk, ykp1
	}
	return yk
}
