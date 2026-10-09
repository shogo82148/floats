package floats

import "math"

// Y0 returns the order-zero Bessel function of the second kind.
//
// Special cases are:
//
//	Y0(+Inf) = 0
//	Y0(0) = -Inf
//	Y0(x < 0) = NaN
//	Y0(NaN) = NaN
func (a Float32) Y0() Float32 {
	// 2 <= x < 64 is 0x4000_0000 <= bits < 0x4280_0000. Negative values, NaN and Inf are out of the range.
	b := a.Bits()
	if b-0x4000_0000 >= 0x0280_0000 {
		return NewFloat32(math.Y0(float64(a)))
	}
	x := float64(math.Float32frombits(b))

	// 2 <= x < 64. Y0 is calculated by the polynomial of the segment: the width is 1/4 for [2, 4), 1/2 for [4, 6),
	// and 1 for [6, 64).
	var i int
	switch {
	case x < 4:
		i = int(x*4) - 8
	case x < 6:
		i = 8 + int((x-4)*2)
	default:
		i = 12 + int(x) - 6
	}
	c := &y032Coeffs[i]
	t := x - y032Centers[i]
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
	return NewFloat256(x).Y0().Float32()
}
