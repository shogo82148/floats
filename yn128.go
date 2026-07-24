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
func (a Float128) Yn(n int) Float128 {
	switch {
	case a.IsNaN():
		return a
	case a.IsInf(1):
		return Float128{}
	}

	neg := false
	if n < 0 {
		n = -n
		neg = !neg
	}

	switch {
	case a.IsZero():
		if neg && n%2 == 1 {
			return NewFloat128Inf(1)
		}
		return NewFloat128Inf(-1)
	case a.Signbit():
		return NewFloat128NaN()
	}

	var y Float128
	switch n {
	case 0:
		y = a.Y0()
	case 1:
		y = a.Y1()
	default:
		y = ynForward128(n, a)
	}
	if neg && n%2 == 1 {
		y = y.Neg()
	}
	return y
}

// ynForward128 returns Yn(x) for n >= 2, x > 0 using the forward recurrence
// Y[k+1] = (2k/x)*Y[k] - Y[k-1] starting from Y0(x) and Y1(x). Unlike the
// same recurrence for J (see jnMiller128), this direction is unconditionally
// stable for Y: Y grows (or at worst oscillates with bounded amplitude) as
// the order increases, so it is the dominant solution the recurrence
// naturally tracks, and no Miller-style backward correction is needed.
func ynForward128(n int, x Float128) Float128 {
	var Two = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}

	ykm1 := x.Y0()
	yk := x.Y1()
	for k := 1; k < n; k++ {
		coef := Two.Mul(NewFloat128(float64(k))).Quo(x)
		ykp1 := FMA128(coef, yk, ykm1.Neg())
		ykm1, yk = yk, ykp1
	}
	return yk
}
