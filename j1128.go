package floats

// J1 returns the order-one Bessel function of the first kind.
//
// Special cases are:
//
//	J1(±Inf) = 0
//	J1(NaN) = NaN
func (a Float128) J1() Float128 {
	// Float256 J1 is accurate enough to round the result correctly except for the cases extremely close to
	// a zero of J1, where |J1(x)| < 2**-187.
	return a.Float256().J1().Float128()
}
