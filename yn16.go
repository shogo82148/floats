package floats

import "math"

// Yn returns the order-n Bessel function of the second kind.
//
// Special cases are:
//
//	Yn(n, +Inf) = 0
//	Yn(n >= 0, 0) = -Inf
//	Yn(n < 0, 0) = +Inf if n is odd, -Inf if n is even
//	Yn(n, x < 0) = NaN
//	Yn(n, NaN) = NaN
func (a Float16) Yn(n int) Float16 {
	switch n {
	case 0:
		return a.Y0()
	case 1:
		return a.Y1()
	case -1:
		return a.Y1().Neg()
	}

	// 2 <= x < 64 is 0x4000 <= a < 0x5400. Negative values, NaN and Inf are out of the range.
	k := n
	if k < 0 {
		k = -k
	}
	if k > yn16MaxOrder || a-0x4000 >= 0x1400 {
		return NewFloat16(math.Yn(n, a.Float64().BuiltIn()))
	}
	x := normal16ToFloat64(a)

	// Y(-n, x) = (-1)**n Y(n, x).
	y, m := yn16Forward(k, x)
	if n < 0 && k%2 == 1 {
		y = -y
	}

	// The absolute error of y is less than 2**-29 m, where m is the maximum of the absolute values of the terms of
	// the recurrence, that is, 2**-29 m / |y| × 2**53 ulps of float64.
	// y can be rounded if it is not so close to the midpoint of two adjacent Float16 values, and it is not too close
	// to zero. float16Round requires |y| < 65520, and the results of more than 2**17 overflow.
	ay := math.Abs(y)
	switch {
	case ay > 0x1p17:
		return NewFloat16(math.Copysign(math.Inf(1), y))
	case ay < 65000:
		if w := 0x1p24 * m / ay; w < 1<<40 {
			if z, ok := float16Round(y, int64(w)+1); ok {
				return z
			}
		}
	}
	// The result is calculated more accurately by Float256.
	return NewFloat256(x).Yn(n).Float16()
}

// yn16MaxOrder is the maximum order calculated by the forward recurrence.
const yn16MaxOrder = 100

// yn16Forward returns Yn(x) for 2 <= n <= yn16MaxOrder and 2 <= x < 64 by the forward recurrence
// Y(k+1) = 2k/x Y(k) - Y(k-1) from Y0 and Y1, which is stable, and the maximum absolute value of the terms.
func yn16Forward(n int, x float64) (y, m float64) {
	// The segments of the polynomials of Y0 and Y1 are the same: the width is 1/2 for [2, 3) and 1 for [3, 64).
	var i int
	if x < 3 {
		i = int(x*2) - 4
	} else {
		i = int(x) - 1
	}
	t := x - y016Centers[i]
	t2 := t * t
	t4 := t2 * t2
	y0, y1 := yn16Poly(&y016Coeffs[i], t, t2, t4), yn16Poly(&y116Coeffs[i], t, t2, t4)
	m = max(math.Abs(y0), math.Abs(y1))
	for k := 1; k < n; k++ {
		y0, y1 = y1, y1*(float64(2*k)/x)-y0
		m = max(m, math.Abs(y1))
	}
	return y1, m
}

// yn16Poly returns the value of the polynomial c at t, where t2 = t**2 and t4 = t**4.
func yn16Poly(c *[8]float64, t, t2, t4 float64) float64 {
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	return ((c[0] + t*c[1]) + t2*(c[2]+t*c[3])) + t4*((c[4]+t*c[5])+t2*(c[6]+t*c[7]))
}
