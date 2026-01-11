package floats

// Hypot returns [Sqrt](p*p + q*q), taking care to avoid
// unnecessary overflow and underflow.
//
// Special cases are:
//
//	Hypot32(±Inf, q) = +Inf
//	Hypot32(p, ±Inf) = +Inf
//	Hypot32(NaN, q) = NaN
//	Hypot32(p, NaN) = NaN
func Hypot32(p, q Float32) Float32 {
	p = p.Abs()
	q = q.Abs()

	// special cases
	switch {
	case p.IsInf(1) || q.IsInf(1):
		return NewFloat32Inf(1)
	case p.IsNaN() || q.IsNaN():
		return NewFloat32NaN()
	}

	if p.Lt(q) {
		p, q = q, p
	}
	if p.IsZero() {
		return 0
	}
	q = q.Quo(p)
	return p.Mul(Float32(1).Add(q.Mul(q)).Sqrt())
}
