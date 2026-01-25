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
func (a Float256) Expm1() Float256 {
	var (
		// VerySmall = 2**-237
		VerySmall = Float256{
			0x3ff1_2000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// Small = 0.5
		Small = Float256{
			0x3fff_e000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)
	absa := a.Abs()
	switch {
	case absa.Lt(VerySmall):
		// exm1(x) ~ x when |x| is very small
		return a
	case absa.Lt(Small):
		// use Taylor expansion when |x| is small
		// exm1(x) = x + x**2/2! + x**3/3! + x**4/4! + x**5/5! + ...
		// TODO: optimize using minimax approximation
		var y Float256
		for n := 40; n >= 1; n-- {
			y = y.Add(power256(a, n).Quo(factorial256(n)))
		}
		return y
	}

	return a.Exp().Sub(Float256(uvone256))
}
