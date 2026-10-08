package floats

// J0 returns the order-zero Bessel function of the first kind.
//
// Special cases are:
//
//	J0(±Inf) = 0
//	J0(0) = 1
//	J0(NaN) = NaN
func (a Float128) J0() Float128 {
	// Float256 J0 is accurate enough to round the result correctly except for the cases extremely close to
	// a zero of J0, where |J0(x)| < 2**-187.
	return a.Float256().J0().Float128()
}
