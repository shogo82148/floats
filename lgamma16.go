package floats

import "math"

// Lgamma returns the natural logarithm and sign (-1 or +1) of Gamma(a).
//
// Special cases are:
//
//	Lgamma(+Inf) = +Inf
//	Lgamma(0) = +Inf
//	Lgamma(-integer) = +Inf
//	Lgamma(-Inf) = -Inf
//	Lgamma(NaN) = NaN
func (a Float16) Lgamma() (Float16, int) {
	ix := a &^ signMask16
	if ix >= uvinf16 {
		return lgamma16Slow(a) // ±Inf and NaN
	}
	if a&signMask16 == 0 && ix >= 0x6ffe {
		return uvinf16, 1 // Lgamma(x) overflows for x >= 8184.
	}
	y, sign, ok := lgamma16Approx(a, ix)
	if !ok {
		return lgamma16Slow(a)
	}
	return lgamma16Round(a, y), sign
}

// lgamma16Approx returns the approximation of Lgamma(a) and its sign.
// If ok is true, the relative error is less than 2**-28.
// ix = |a| must be finite, and a must not overflow.
func lgamma16Approx(a Float16, ix Float16) (y float64, sign int, ok bool) {
	var x float64
	if ix < 0x0400 {
		x = float64(ix) * 0x1p-24 // subnormal
	} else {
		x = normal16ToFloat64(ix)
	}

	if a&signMask16 == 0 {
		switch {
		case ix == 0:
			return math.Inf(1), 1, true
		case x < 0x1p-8:
			// Lgamma(x) = -log(x) + Lgamma(1+x)
			return lgamma16Tiny(x) - math.Log(x), 1, true
		case x < 0.5:
			return lgamma16Small(x) - math.Log(x), 1, true
		}
		return lgamma16Poly(x), 1, true
	}

	// Lgamma(x) = log(pi / |sin(pi x)|) - Lgamma(1-x), where |sin(pi x)| = sin(pi |r|) for x = n + r, |r| <= 1/2.
	// x = -|x| in the following. r and 1-x are calculated exactly.
	n := math.RoundToEven(x)
	r := x - n
	if r == 0 {
		return math.Inf(1), 1, true // the negative integers and -0
	}
	lg := lgamma16Poly(1 + x)
	y = -math.Log(math.Abs(r)*gamma16Sinc(r)) - lg
	// The absolute error is less than 2**-36 (1+lg). The relative error of y is large if it is close to zero.
	ok = math.Abs(y) >= (1+lg)*(1.0/256)
	// The sign of Gamma(x) is negative if floor(x) = -ceil(|x|) is odd. ceil(|x|) is n+1 if r > 0, and n otherwise.
	c := int64(n) + int64(math.Float64bits(r)>>63^1)
	return y, 1 - int(c&1)<<1, ok
}

// lgamma16Slow returns Lgamma(a) calculated by the math package.
func lgamma16Slow(a Float16) (Float16, int) {
	lgamma, sign := math.Lgamma(a.Float64().BuiltIn())
	return NewFloat16(lgamma), sign
}

// lgamma16Round rounds y, which is the approximation of Lgamma(a) with the relative error less than 2**-28,
// to Float16. If the rounding is not determined because y is too close to the midpoint of two adjacent Float16 values,
// it falls back to the math package.
func lgamma16Round(a Float16, y float64) Float16 {
	// the rounding position is the bit 41 of the float64 representation, and the midpoint is 0b1000....
	const window = 1 << 28
	low := math.Float64bits(y) & (1<<42 - 1)
	if low-(1<<41-window) < 2*window {
		z, _ := lgamma16Slow(a)
		return z
	}
	return NewFloat16(y)
}

// lgamma16Poly returns Lgamma(x) for x in [1/2, 8184] with the relative error less than 2**-39,
// by the polynomial of the segment that includes x.
func lgamma16Poly(x float64) float64 {
	b := math.Float64bits(x)
	c := &lgamma16Coeffs[b>>49-8176]
	t := x - math.Float64frombits(b&^(1<<49-1)|1<<48) // the center of the segment
	p := c[0] + t*(c[1]+t*(c[2]+t*(c[3]+t*(c[4]+t*(c[5]+t*(c[6]+t*c[7]))))))
	// Lgamma(x) = (x-1)(x-2) p, where p is smooth at the zeros x = 1 and x = 2 of Lgamma.
	// + 0 makes the sign of the result at the zeros positive.
	return (x-1)*(x-2)*p + 0
}

// lgamma16Small returns Lgamma(1+x) for x in [0, 1/2) with the absolute error less than 2**-36.
func lgamma16Small(x float64) float64 {
	s := int(x * 8)
	c := &lgamma16SmallCoeffs[s]
	t := x - float64(2*s+1)*(1.0/16)
	return c[0] + t*(c[1]+t*(c[2]+t*(c[3]+t*(c[4]+t*(c[5]+t*c[6])))))
}

// lgamma16Tiny returns Lgamma(1+x) for x in [0, 2**-8) with the absolute error less than 2**-42.
func lgamma16Tiny(x float64) float64 {
	c := &lgamma16TinyCoeffs
	return x * (c[0] + x*(c[1]+x*(c[2]+x*c[3])))
}
