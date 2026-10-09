package floats

import "math"

// Y1 returns the order-one Bessel function of the second kind.
//
// Special cases are:
//
//	Y1(+Inf) = 0
//	Y1(0) = -Inf
//	Y1(x < 0) = NaN
//	Y1(NaN) = NaN
func (a Float16) Y1() Float16 {
	// 2 <= x < 64 is 0x4000 <= a < 0x5400. Negative values, NaN and Inf are out of the range.
	if a-0x4000 < 0x1400 {
		// Y1 is calculated by the polynomial of the segment: the width is 1/2 for [2, 3) and 1 for [3, 64).
		x := normal16ToFloat64(a)
		var i int
		if x < 3 {
			i = int(x*2) - 4
		} else {
			i = int(x) - 1
		}
		c := &y116Coeffs[i]
		t := x - y116Centers[i]
		t2 := t * t
		t4 := t2 * t2
		// Estrin's scheme: the dependency chain is shorter than Horner's method.
		y := ((c[0] + t*c[1]) + t2*(c[2]+t*c[3])) + t4*((c[4]+t*c[5])+t2*(c[6]+t*c[7]))

		// The absolute error of y is less than 2**-30, that is, 2**-30 / |y| × 2**53 ulps of float64.
		// y can be rounded if it is not so close to the midpoint of two adjacent Float16 values, which is 2**41 in
		// the lower 42 bits, and it is not too close to zero (the result may be subnormal).
		ay := math.Abs(y)
		if ay >= 0x1p-13 {
			d := int64(math.Float64bits(ay)&(1<<42-1)) - 1<<41
			if w := int64(0x1p23/ay) + 1; d <= -w || w <= d {
				return NewFloat16(y)
			}
		}
	}
	return NewFloat16(math.Y1(a.Float64().BuiltIn()))
}
