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
func (a Float32) Y1() Float32 {
	// 2 <= x < 64 is 0x4000_0000 <= bits < 0x4280_0000. Negative values, NaN and Inf are out of the range.
	b := a.Bits()
	if b-0x4000_0000 >= 0x0280_0000 {
		return NewFloat32(math.Y1(float64(a)))
	}
	x := float64(math.Float32frombits(b))

	// 2 <= x < 64. Y1 is calculated by the polynomial of the segment: the width is 1/4 for [2, 3), 1/2 for [3, 5),
	// and 1 for [5, 64).
	var i int
	switch {
	case x < 3:
		i = int(x*4) - 8
	case x < 5:
		i = 4 + int((x-3)*2)
	default:
		i = 8 + int(x) - 5
	}
	c := &y132Coeffs[i]
	t := x - y132Centers[i]
	t2 := t * t
	t4 := t2 * t2
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	y := (((c[0] + t*c[1]) + t2*(c[2]+t*c[3])) + t4*((c[4]+t*c[5])+t2*(c[6]+t*c[7]))) + t4*t4*((c[8]+t*c[9])+t2*(c[10]+t*c[11]))

	// The absolute error of y is less than 2**-48, that is, 2**-48 / |y| × 2**53 ulps of float64.
	// If y is too close to the midpoint of two adjacent Float32 values, or too close to zero, y can not be rounded.
	if w := 0x1p5 / math.Abs(y); w < 1<<27 {
		if z, ok := float32Round(y, int64(w)+1); ok {
			return z
		}
	}
	// The result is calculated more accurately by Float256.
	return NewFloat256(x).Y1().Float32()
}

// y132Poly returns the value of the polynomial of the segment of Y1 for 2 <= x < 64 at x: the width of the segments is
// 1/4 for [2, 3), 1/2 for [3, 5), and 1 for [5, 64). Its absolute error is less than 2**-48.
// Y1 itself does not call it, because the call slows Y1 down by about 10%.
func y132Poly(x float64) float64 {
	var i int
	switch {
	case x < 3:
		i = int(x*4) - 8
	case x < 5:
		i = 4 + int((x-3)*2)
	default:
		i = 8 + int(x) - 5
	}
	c := &y132Coeffs[i]
	t := x - y132Centers[i]
	t2 := t * t
	t4 := t2 * t2
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	return (((c[0] + t*c[1]) + t2*(c[2]+t*c[3])) + t4*((c[4]+t*c[5])+t2*(c[6]+t*c[7]))) + t4*t4*((c[8]+t*c[9])+t2*(c[10]+t*c[11]))
}
