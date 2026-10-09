package floats

import "math"

// Jn returns the order-n Bessel function of the first kind.
//
// Special cases are:
//
//	Jn(n, ±Inf) = 0
//	Jn(n, NaN) = NaN
func (a Float32) Jn(n int) Float32 {
	switch n {
	case 0:
		return a.J0()
	case 1:
		return a.J1()
	case -1:
		return -a.J1()
	}

	// J(-n, x) = (-1)**n J(n, x) and J(n, -x) = (-1)**n J(n, x).
	k := n
	if k < 0 {
		k = -k
	}
	x := math.Abs(float64(a))
	neg := (n < 0) != a.Signbit() && k%2 == 1

	// window is the error of y in ulps of float64. If y is closer than window to the midpoint of two adjacent
	// Float32 values, y can not be rounded.
	var y float64
	var window int64
	switch {
	case k > 0 && x > 0 && (x*x < 2*float64(k+1) || (k <= 63 && x < float64(k))):
		// x**2 < 2 (n+1), or x < n <= 63: the Taylor series.
		var ok bool
		y, window, ok = jnTaylor(k, x)
		if !ok {
			return NewFloat32(math.Jn(n, float64(a)))
		}
	case k > 0 && k <= 63 && 2 <= x && x < 64 && float64(k) <= x:
		// 2 <= n <= x < 64: the forward recurrence from J0 and J1, which is stable for n <= x.
		// Its absolute error is less than 2**-46, that is, 2**-46 / |y| × 2**53 ulps. y near zero can not be rounded.
		y = jn32Forward(k, x)
		w := 0x1p7 / math.Abs(y)
		if !(w < 1<<27) {
			return NewFloat256(float64(a)).Jn(n).Float32()
		}
		window = int64(w) + 1
	default:
		return NewFloat32(math.Jn(n, float64(a)))
	}
	if neg {
		y = -y
	}
	if z, ok := float32Round(y, window); ok {
		return z
	}
	// The result is calculated more accurately by Float256.
	return NewFloat256(float64(a)).Jn(n).Float32()
}

// jnInvFact[n] is about 1/n!. The error is a few tens of ulps.
var jnInvFact = func() (t [64]float64) {
	f := 1.0
	for i := range t {
		if i > 0 {
			f *= float64(i)
		}
		t[i] = 1 / f
	}
	return
}()

// jnTaylor returns Jn(x) for n >= 2 and x > 0 by the Taylor series
//
//	Jn(x) = (x/2)**n/n! sum (-1)**k (x**2/4)**k n!/(k! (n+k)!),
//
// and the error of the result in ulps of float64. The series decreases if x**2 < 2 (n+1) (the ratio of the terms is
// less than 1/2). Otherwise the terms increase at first, and ok is false if the cancellation of the sum is larger
// than 2**12, that is, the result loses more than 12 bits.
func jnTaylor(n int, x float64) (y float64, window int64, ok bool) {
	// the prefactor (x/2)**n/n! with the relative error of a few tens of ulps if n <= 63, and
	// exp(n log(x/2) - log(n!)) otherwise, whose relative error is less than 2**-40
	// (the error is about 2**-53 times its exponent, which is at most about 700).
	var pf float64
	if n < len(jnInvFact) {
		h, p := x/2, 1.0
		for i := n; i > 0; i >>= 1 {
			if i&1 != 0 {
				p *= h
			}
			h *= h
		}
		pf = p * jnInvFact[n]
	} else {
		nf := float64(n)
		lg, _ := math.Lgamma(nf + 1)
		l := nf*math.Log(x/2) - lg
		if l < -105 {
			// Jn(x) < exp(-105), which is less than a half of the smallest subnormal number of Float32.
			return 0, 1 << 16, true
		}
		pf = math.Exp(l)
	}

	u := x * x / 4
	t, s, sabs := 1.0, 1.0, 1.0
	for i := 1; ; i++ {
		t *= -u / (float64(i) * float64(n+i))
		s += t
		sabs += math.Abs(t)
		if math.Abs(t) <= 0x1p-60*sabs {
			break
		}
	}
	// The rounding errors of the terms and the sum are about 2**-53 sabs.
	ratio := sabs / math.Abs(s)
	if !(ratio < 1<<12) {
		return 0, 0, false
	}
	return pf * s, 1<<16 + int64(16*ratio), true
}

// jn32Forward returns Jn(x) for 2 <= n <= x < 64 by the forward recurrence J(k+1) = 2k/x J(k) - J(k-1).
func jn32Forward(n int, x float64) float64 {
	i := int(x)
	j0 := jn32Poly(&j032Coeffs[i-2], x, i)
	j1 := jn32Poly(&j132Coeffs[i-2], x, i)
	for k := 1; k < n; k++ {
		j0, j1 = j1, j1*(float64(2*k)/x)-j0
	}
	return j1
}

// jn32Poly returns the value of the polynomial c of the segment [i, i+1) at x.
func jn32Poly(c *[12]float64, x float64, i int) float64 {
	t := x - float64(i) - 0.5
	t2 := t * t
	t4 := t2 * t2
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	return (((c[0] + t*c[1]) + t2*(c[2]+t*c[3])) + t4*((c[4]+t*c[5])+t2*(c[6]+t*c[7]))) + t4*t4*((c[8]+t*c[9])+t2*(c[10]+t*c[11]))
}
