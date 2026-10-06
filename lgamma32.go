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
func (a Float32) Lgamma() (Float32, int) {
	b := a.Bits()
	ix := b &^ signMask32
	switch {
	case ix >= 0x7f80_0000:
		return lgamma32Slow(a) // ±Inf and NaN
	case b >= 0x7c44_af8e && b < 0x7f80_0000:
		return NewFloat32Inf(1), 1 // Lgamma(x) overflows for x >= 4.085e36.
	}
	if y, sign, ok := lgamma32Approx(float64(a)); ok {
		if z, ok := gamma32Round(y); ok {
			return z, sign
		}
	}
	return lgamma32Slow(a)
}

// lgamma32Slow returns Lgamma(a) with math.Lgamma.
// math.Lgamma rounded to Float32 is correctly rounded except for the arguments whose results are
// very close to the midpoint of two adjacent Float32 values. lgamma32Hard has the correct results of the ones
// that are not correctly rounded.
func lgamma32Slow(a Float32) (Float32, int) {
	lgamma, sign := math.Lgamma(float64(a))
	b := a.Bits()
	for _, h := range lgamma32Hard {
		if h[0] == b {
			return NewFloat32FromBits(h[1]), sign
		}
	}
	return NewFloat32(lgamma), sign
}

// lgamma32Hard are the arguments whose results are not correctly rounded by math.Lgamma rounded to Float32,
// and the correct results. They were found by checking all Float32 values, and the exact results
// were calculated with mpmath.
var lgamma32Hard = [...][2]uint32{
	{0x3b7c_53aa, 0x40b1_d661},
	{0x77ac_5674, 0x7acf_27b3},
}

// lgamma32Approx returns the approximation of Lgamma(x) and its sign for finite x that does not overflow.
// If ok is true, the relative error is less than 2**-34.
func lgamma32Approx(x float64) (y float64, sign int, ok bool) {
	if x >= 0 {
		switch {
		case x == 0:
			return math.Inf(1), 1, true
		case x < 0x1p-8:
			// Lgamma(x) = -log(x) + Lgamma(1+x)
			return lgamma16Tiny(x) - math.Log(x), 1, true
		case x < 0.5:
			return lgamma16Small(x) - math.Log(x), 1, true
		}
		return lgamma32Pos(x), 1, true
	}

	// Lgamma(x) = log(pi / |sin(pi x)|) - Lgamma(1-x), where |sin(pi x)| = sin(pi |r|) for x = n + r, |r| <= 1/2.
	// x = -|x| in the following. r and 1-x are calculated exactly if |x| < 2**23.
	x = -x
	n := math.RoundToEven(x)
	r := x - n
	if r == 0 {
		return math.Inf(1), 1, true // the negative integers. All Float32 values with |x| >= 2**23 are integers.
	}
	lg := lgamma32Pos(1 + x)
	y = -math.Log(math.Abs(r)*lgamma32Sinc(r)) - lg
	// The absolute error is less than 2**-42 (1+lg). The relative error of y is large if it is close to zero.
	ok = math.Abs(y) >= (1+lg)*(1.0/64)
	// The sign of Gamma(x) is negative if floor(x) = -ceil(|x|) is odd. ceil(|x|) is n+1 if r > 0, and n otherwise.
	c := int64(n) + int64(math.Float64bits(r)>>63^1)
	return y, 1 - int(c&1)<<1, ok
}

// lgamma32Pos returns Lgamma(x) for x >= 1/2 with the relative error less than 2**-38.
func lgamma32Pos(x float64) float64 {
	if x < 8192 {
		return lgamma16Poly(x)
	}
	// Stirling's series. The next term 1/(360 x**3) is less than 2**-60 of the result for x >= 8192,
	// and 1/(12 x) is less than 2**-47 of the result for x >= 2**20.
	const halfLn2Pi = 0.91893853320467274178032973640561763986139747363778
	y := (x-0.5)*math.Log(x) - x + halfLn2Pi
	if x < 0x1p20 {
		y += 1.0 / (12 * x)
	}
	return y
}

// lgamma32Sinc returns sin(pi r)/(pi r) for |r| <= 1/2 with the relative error less than 2**-44.
func lgamma32Sinc(r float64) float64 {
	u := r * r
	u2 := u * u
	c := &lgamma32SincCoeffs
	// Estrin's scheme: the dependency chain is shorter than Horner's method.
	return ((c[0] + u*c[1]) + u2*(c[2]+u*c[3])) + u2*u2*((c[4]+u*c[5])+u2*(c[6]+u*c[7]))
}
