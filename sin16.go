package floats

import "math"

// Sin returns the sine of the radian argument a.
//
// Special cases are:
//
//	±0.Sin() = ±0
//	±Inf.Sin() = NaN
//	NaN.Sin() = NaN
func (a Float16) Sin() Float16 {
	return NewFloat16(math.Sin(a.Float64().BuiltIn()))
}
