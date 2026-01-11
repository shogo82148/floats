package floats

// Hypot returns [Sqrt](p*p + q*q), taking care to avoid
// unnecessary overflow and underflow.
//
// Special cases are:
//
//	Hypot128(±Inf, q) = +Inf
//	Hypot128(p, ±Inf) = +Inf
//	Hypot128(NaN, q) = NaN
//	Hypot128(p, NaN) = NaN
func Hypot128(p, q Float128) Float128 {
	p = p.Abs()
	q = q.Abs()

	// special cases
	switch {
	case p.IsInf(1) || q.IsInf(1):
		return NewFloat128Inf(1)
	case p.IsNaN() || q.IsNaN():
		return NewFloat128NaN()
	}

	if p.Lt(q) {
		p, q = q, p
	}
	if p.IsZero() {
		return Float128{}
	}
	q = q.Quo(p)
	return p.Mul(Float128(uvone128).Add(q.Mul(q)).Sqrt())
}
