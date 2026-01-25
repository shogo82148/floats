package floats

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
func (a Float128) Expm1() Float128 {
	var (
		// VerySmall = 2**-113
		VerySmall = Float128{0x3f8e_0000_0000_0000, 0x0000_0000_0000_0000}

		// Small = 0.5
		Small = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}
	)
	absa := a.Abs()
	switch {
	case absa.Lt(VerySmall):
		// exm1(x) ~ x when |x| is very small
		return a
	case absa.Lt(Small):
		// use Taylor expansion when |x| is small
		// exm1(x) = x + x**2/2! + x**3/3! + x**4/4! + x**5/5! + ...
		var y Float128
		for n := 20; n >= 1; n-- {
			y = y.Add(power128(a, n).Quo(factorial128(n)))
		}
		return y
	}

	return a.Exp().Sub(Float128(uvone128))
}
