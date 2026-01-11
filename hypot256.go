package floats

// Hypot256 returns [Sqrt](p*p + q*q), taking care to avoid
// unnecessary overflow and underflow.
//
// Special cases are:
//
//	Hypot256(±Inf, q) = +Inf
//	Hypot256(p, ±Inf) = +Inf
//	Hypot256(NaN, q) = NaN
//	Hypot256(p, NaN) = NaN
func Hypot256(p, q Float256) Float256 {
	p = p.Abs()
	q = q.Abs()

	// special cases
	switch {
	case p.IsInf(1) || q.IsInf(1):
		return NewFloat256Inf(1)
	case p.IsNaN() || q.IsNaN():
		return NewFloat256NaN()
	}

	if p.Lt(q) {
		p, q = q, p
	}
	if p.IsZero() {
		return Float256{}
	}
	q = q.Quo(p)
	return p.Mul(Float256(uvone256).Add(q.Mul(q)).Sqrt())
}
