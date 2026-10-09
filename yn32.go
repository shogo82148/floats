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
func (a Float32) Yn(n int) Float32 {
	switch n {
	case 0:
		return a.Y0()
	case 1:
		return a.Y1()
	case -1:
		return -a.Y1()
	}

	// 2 <= x < 64 is 0x4000_0000 <= bits < 0x4280_0000. Negative values, NaN and Inf are out of the range.
	k := n
	if k < 0 {
		k = -k
	}
	b := a.Bits()
	if k > yn32MaxOrder || b-0x4000_0000 >= 0x0280_0000 {
		return NewFloat32(math.Yn(n, float64(a)))
	}
	x := float64(math.Float32frombits(b))

	// Y(-n, x) = (-1)**n Y(n, x).
	y, m := yn32Forward(k, x)

	// The absolute error of y is less than 2**-48 (n+1) m, where m is the maximum of the absolute values of the
	// terms of the recurrence, that is, 2**-48 (n+1) m / |y| × 2**53 ulps of float64.
	// If y is too close to the midpoint of two adjacent Float32 values, or too close to zero, y can not be rounded.
	if w := 32 * float64(k+1) * m / math.Abs(y); w < 1<<27 {
		if n < 0 && k%2 == 1 {
			y = -y
		}
		if z, ok := float32Round(y, int64(w)+1); ok {
			return z
		}
	}
	// The result is calculated more accurately by Float256.
	return NewFloat256(x).Yn(n).Float32()
}

// yn32MaxOrder is the maximum order calculated by the forward recurrence. |Yn(x)| < 2**500 for 2 <= x and n <= 100.
const yn32MaxOrder = 100

// yn32Forward returns Yn(x) for 2 <= n <= yn32MaxOrder and 2 <= x < 64 by the forward recurrence
// Y(k+1) = 2k/x Y(k) - Y(k-1) from Y0 and Y1, which is stable, and the maximum absolute value of the terms.
func yn32Forward(n int, x float64) (y, m float64) {
	y0, y1 := y032Poly(x), y132Poly(x)
	m = max(math.Abs(y0), math.Abs(y1))
	for k := 1; k < n; k++ {
		y0, y1 = y1, y1*(float64(2*k)/x)-y0
		m = max(m, math.Abs(y1))
	}
	return y1, m
}
