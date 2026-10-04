package floats

import "math"

// The kernels and the argument reduction are shared with Float32. See sin32.go.

// Sin returns the sine of the radian argument a.
//
// Special cases are:
//
//	±0.Sin() = ±0
//	±Inf.Sin() = NaN
//	NaN.Sin() = NaN
func (a Float16) Sin() Float16 {
	ix := a &^ signMask16
	if ix < 0x3a49 { // |a| < Pi/4
		if ix < 0x2400 { // |a| < 2**-6
			// sin(a) = a - a**3/6 + ... rounds to a.
			return a
		}
		x := normal16ToFloat64(a)
		return NewFloat16(sinKernel32(x, x*x))
	}
	if ix >= uvinf16 {
		// sin(±Inf) = NaN, sin(NaN) = NaN
		return NewFloat16NaN()
	}

	n, r := trigReduceSmall(normal16ToFloat64(ix))
	z := r * r
	var s float64
	if n&1 == 0 {
		s = sinKernel32(r, z)
	} else {
		s = cosKernel32(z)
	}
	if n&2 != 0 {
		s = -s
	}
	if a&signMask16 != 0 {
		s = -s
	}
	return NewFloat16(s)
}

// Cos returns the cosine of the radian argument a.
//
// Special cases are:
//
//	±Inf.Cos() = NaN
//	NaN.Cos() = NaN
func (a Float16) Cos() Float16 {
	ix := a &^ signMask16
	if ix < 0x3a49 { // |a| < Pi/4
		if ix < 0x2400 { // |a| < 2**-6
			// cos(a) = 1 - a**2/2 + ... rounds to 1.
			return uvone16
		}
		x := normal16ToFloat64(a)
		return NewFloat16(cosKernel32(x * x))
	}
	if ix >= uvinf16 {
		// cos(±Inf) = NaN, cos(NaN) = NaN
		return NewFloat16NaN()
	}

	n, r := trigReduceSmall(normal16ToFloat64(ix))
	z := r * r
	var c float64
	if n&1 == 0 {
		c = cosKernel32(z)
	} else {
		c = sinKernel32(r, z)
	}
	if (n+1)&2 != 0 {
		c = -c
	}
	return NewFloat16(c)
}

// Sincos returns Sin(a), Cos(a).
//
// Special cases are:
//
//	±0.Sincos() = ±0, 1
//	±Inf.Sincos() = NaN, NaN
//	NaN.Sincos() = NaN, NaN
func (a Float16) Sincos() (sin, cos Float16) {
	ix := a &^ signMask16
	if ix < 0x3a49 { // |a| < Pi/4
		if ix < 0x2400 { // |a| < 2**-6
			return a, uvone16
		}
		x := normal16ToFloat64(a)
		z := x * x
		return NewFloat16(sinKernel32(x, z)), NewFloat16(cosKernel32(z))
	}
	if ix >= uvinf16 {
		nan := NewFloat16NaN()
		return nan, nan
	}

	n, r := trigReduceSmall(normal16ToFloat64(ix))
	z := r * r
	s, c := sinKernel32(r, z), cosKernel32(z)
	if n&1 != 0 {
		s, c = c, s
	}
	if n&2 != 0 {
		s = -s
	}
	if (n+1)&2 != 0 {
		c = -c
	}
	if a&signMask16 != 0 {
		s = -s
	}
	return NewFloat16(s), NewFloat16(c)
}

// Tan returns the tangent of the radian argument a.
//
// Special cases are:
//
//	±0.Tan() = ±0
//	±Inf.Tan() = NaN
//	NaN.Tan() = NaN
func (a Float16) Tan() Float16 {
	ix := a &^ signMask16
	if ix < 0x3a49 { // |a| < Pi/4
		if ix < 0x2400 { // |a| < 2**-6
			// tan(a) = a + a**3/3 + ... rounds to a.
			return a
		}
		x := normal16ToFloat64(a)
		p, q := tanKernel32(x, x*x)
		return NewFloat16(p / q)
	}
	if ix >= uvinf16 {
		// tan(±Inf) = NaN, tan(NaN) = NaN
		return NewFloat16NaN()
	}

	n, r := trigReduceSmall(normal16ToFloat64(ix))
	p, q := tanKernel32(r, r*r)
	var t float64
	if n&1 == 0 {
		t = p / q
	} else {
		t = -q / p
	}
	if a&signMask16 != 0 {
		t = -t
	}
	return NewFloat16(t)
}

// normal16ToFloat64 converts a normal Float16 value a to float64.
func normal16ToFloat64(a Float16) float64 {
	sign := uint64(a&signMask16) << (64 - 16)
	return math.Float64frombits(sign | (uint64(a&^signMask16)<<(shift64-shift16) + (bias64-bias16)<<shift64))
}
