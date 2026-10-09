package floats

// Jn returns the order-n Bessel function of the first kind.
//
// Special cases are:
//
//	Jn(n, ±Inf) = 0
//	Jn(n, NaN) = NaN
func (a Float128) Jn(n int) Float128 {
	// Float256 Jn is accurate enough to round the result correctly except for the cases extremely close to
	// a zero of Jn, where |Jn(x)| < 2**-187.
	return a.Float256().Jn(n).Float128()
}
