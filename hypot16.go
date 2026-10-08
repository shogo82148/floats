package floats

import "math"

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

	// Float16 products are exact in float64, and the sum and square root
	// are each rounded once with 53 bits of precision, so the result is
	// correctly rounded in all but astronomically rare double-rounding cases.
	x, y := p.Float64().BuiltIn(), q.Float64().BuiltIn()
	return NewFloat16(math.Sqrt(x*x + y*y))
}
