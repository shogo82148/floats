package floats

import "math"

// J1 returns the order-one Bessel function of the first kind.
//
// Special cases are:
//
//	J1(±Inf) = 0
//	J1(NaN) = NaN
func (a Float32) J1() Float32 {
	// 2 <= |x| < 64 is 0x4000_0000 <= bits < 0x4280_0000. NaN and Inf are out of the range.
	b := a.Bits() &^ signMask32
	if b-0x4000_0000 >= 0x0280_0000 {
		if b == 0 {
			return a // ±0. J1 is odd.
		}
		return NewFloat32(math.J1(float64(a)))
	}
	x := float64(math.Float32frombits(b))

	// 2 <= |x| < 64. J1 is an odd function and it is calculated by the polynomial of the segment [i, i+1).
	i := int(x)
	c := &j132Coeffs[i-2]
	t := x - float64(i) - 0.5
	t2 := t * t
	t4 := t2 * t2
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	y := (((c[0] + t*c[1]) + t2*(c[2]+t*c[3])) + t4*((c[4]+t*c[5])+t2*(c[6]+t*c[7]))) + t4*t4*((c[8]+t*c[9])+t2*(c[10]+t*c[11]))
	if a.Signbit() {
		y = -y
	}

	// The absolute error of y is less than 2**-48, that is, 2**-48 / |y| × 2**53 ulps of float64.
	// If y is too close to the midpoint of two adjacent Float32 values, or too close to zero, y can not be rounded.
	if w := 0x1p5 / math.Abs(y); w < 1<<27 {
		if z, ok := float32Round(y, int64(w)+1); ok {
			return z
		}
	}
	// math.J1 rounded to Float32 is correctly rounded for all these inputs, which was checked with mpmath.
	return NewFloat32(math.J1(float64(a)))
}
