package floats

import "math"

// Jn returns the order-n Bessel function of the first kind.
//
// Special cases are:
//
//	Jn(n, ±Inf) = 0
//	Jn(n, NaN) = NaN
func (a Float16) Jn(n int) Float16 {
	switch n {
	case 0:
		return a.J0()
	case 1:
		return a.J1()
	case -1:
		return a.J1().Neg()
	}

	// J(-n, x) = (-1)**n J(n, x) and J(n, -x) = (-1)**n J(n, x).
	k := n
	if k < 0 {
		k = -k
	}
	x := math.Abs(a.Float64().BuiltIn())
	neg := (n < 0) != a.Signbit() && k%2 == 1

	// window is the error of y in ulps of float64. If y is closer than window to the midpoint of two adjacent
	// Float16 values, y can not be rounded.
	var y float64
	var window int64
	switch {
	case k > 0 && x > 0 && (x*x < 2*float64(k+1) || (k <= 63 && x < float64(k))):
		// x**2 < 2 (n+1), or x < n <= 63: the Taylor series.
		var ok bool
		y, window, ok = jnTaylor(k, x)
		if !ok {
			return NewFloat16(math.Jn(n, a.Float64().BuiltIn()))
		}
	case k > 0 && k <= 63 && 2 <= x && x < 64 && float64(k) <= x:
		// 2 <= n <= x < 64: the forward recurrence from J0 and J1, which is stable for n <= x.
		// Its absolute error is less than 2**-29, that is, 2**-29 / |y| × 2**53 ulps. y near zero can not be rounded.
		y = jn16Forward(k, x)
		w := 0x1p24 / math.Abs(y)
		if !(w < 1<<40) {
			return jn16Fallback(n, a)
		}
		window = int64(w) + 1
	default:
		return NewFloat16(math.Jn(n, a.Float64().BuiltIn()))
	}
	if neg {
		y = -y
	}
	if z, ok := float16Round(y, window); ok {
		return z
	}
	return jn16Fallback(n, a)
}

// jn16Fallback returns Jn(a) for the cases which are too close to the midpoint of two adjacent Float16 values,
// or to zero. math.Jn is accurate enough for most of them, and the others are calculated by Float256.
func jn16Fallback(n int, a Float16) Float16 {
	x := a.Float64().BuiltIn()
	// The absolute error of math.Jn is less than 2**-46.
	y := math.Jn(n, x)
	if w := 0x1p7 / math.Abs(y); w < 1<<40 {
		if z, ok := float16Round(y, int64(w)+1); ok {
			return z
		}
	}
	return NewFloat256(x).Jn(n).Float16()
}

// jn16Forward returns Jn(x) for 2 <= n <= x < 64 by the forward recurrence J(k+1) = 2k/x J(k) - J(k-1).
func jn16Forward(n int, x float64) float64 {
	i := int(x)
	j0 := jn16Poly(&j016Coeffs[i-2], x, i)
	j1 := jn16Poly(&j116Coeffs[i-2], x, i)
	for k := 1; k < n; k++ {
		j0, j1 = j1, j1*(float64(2*k)/x)-j0
	}
	return j1
}

// jn16Poly returns the value of the polynomial c of the segment [i, i+1) at x.
func jn16Poly(c *[8]float64, x float64, i int) float64 {
	t := x - float64(i) - 0.5
	t2 := t * t
	t4 := t2 * t2
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	return ((c[0] + t*c[1]) + t2*(c[2]+t*c[3])) + t4*((c[4]+t*c[5])+t2*(c[6]+t*c[7]))
}

// float16Round rounds g to Float16. ok is false if the result may not be correctly rounded, because g is
// closer than the window ulps of float64 to the midpoint of two adjacent Float16 values.
// The window must be larger than the error of g in ulps of float64, and |g| must be less than 65520.
func float16Round(g float64, window int64) (y Float16, ok bool) {
	bits := math.Float64bits(g)
	sign := Float16(bits>>48) & signMask16
	abs := bits &^ (1 << 63)
	if abs < 0x3f10_0000_0000_0000 {
		// 2**-14 > |g|: the result is subnormal. It is m × 2**-24 for the integer m rounded from |g| × 2**24.
		m := math.Float64frombits(abs) * 0x1p24
		k := math.Floor(m)
		if d := m - k - 0.5; -0x1p-10 < d && d < 0x1p-10 {
			return 0, false
		}
		if m-k > 0.5 {
			k++
		}
		return sign | Float16(k), true
	}
	// The midpoint is 2**41 in the lower 42 bits.
	low := abs & (1<<42 - 1)
	if d := int64(low) - 1<<41; -window < d && d < window {
		return 0, false
	}
	// the rounding may carry into the exponent.
	r := abs>>42 + low>>41 - (1023-15)<<10
	return sign | Float16(r), true
}
