package floats

import "math"

// J0 returns the order-zero Bessel function of the first kind.
//
// Special cases are:
//
//	J0(±Inf) = 0
//	J0(0) = 1
//	J0(NaN) = NaN
func (a Float16) J0() Float16 {
	// 2 <= |x| < 64 is 0x4000 <= ix < 0x5400.
	if ix := a &^ signMask16; ix-0x4000 < 0x1400 {
		// J0 is an even function and it is calculated by the polynomial of the segment [i, i+1).
		x := normal16ToFloat64(ix)
		i := int(x)
		c := &j016Coeffs[i-2]
		t := x - float64(i) - 0.5
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
	return NewFloat16(math.J0(a.Float64().BuiltIn()))
}
