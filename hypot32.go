package floats

import "math"

// Hypot32 returns [Sqrt](p*p + q*q), taking care to avoid
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

	// Float32 squares are exact in float64, and the sum and square root
	// are each rounded once with 53 bits of precision, so the result is
	// correctly rounded in all but astronomically rare double-rounding cases.
	// float64 has enough range that no scaling is needed.
	x, y := p.Float64().BuiltIn(), q.Float64().BuiltIn()
	return NewFloat32(math.Sqrt(x*x + y*y))
}
