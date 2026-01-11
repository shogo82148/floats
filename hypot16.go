package floats

// Hypot16 returns [Sqrt](p*p + q*q), taking care to avoid
// unnecessary overflow and underflow.
//
// Special cases are:
//
//	Hypot16(±Inf, q) = +Inf
//	Hypot16(p, ±Inf) = +Inf
//	Hypot16(NaN, q) = NaN
//	Hypot16(p, NaN) = NaN
func Hypot16(p, q Float16) Float16 {
	p = p.Abs()
	q = q.Abs()

	// special cases
	switch {
	case p.IsInf(1) || q.IsInf(1):
		return NewFloat16Inf(1)
	case p.IsNaN() || q.IsNaN():
		return NewFloat16NaN()
	}

	if p.Lt(q) {
		p, q = q, p
	}
	if p.IsZero() {
		return 0
	}
	q = q.Quo(p)
	return p.Mul(Float16(uvone16).Add(q.Mul(q)).Sqrt())
}
