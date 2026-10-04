package floats

import "math"

// Sinh returns the hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Sinh() = ±0
//	±Inf.Sinh() = ±Inf
//	NaN.Sinh() = NaN
func (a Float16) Sinh() Float16 {
	ix := a &^ signMask16
	if ix < 0x2800 { // |a| < 2**-5
		// sinh(a) = a + a**3/6 + ... rounds to a.
		return a
	}
	if ix >= 0x4a00 { // |a| >= 12
		if ix > uvinf16 {
			// sinh(NaN) = NaN
			return NewFloat16NaN()
		}
		// sinh(a) overflows.
		return a&signMask16 | uvinf16
	}

	e := exp16(normal16ToFloat64(ix))
	s := (e - 1/e) * 0.5
	if a&signMask16 != 0 {
		s = -s
	}
	return NewFloat16(s)
}

// Cosh returns the hyperbolic cosine of a.
//
// Special cases are:
//
//	±0.Cosh() = 1
//	±Inf.Cosh() = +Inf
//	NaN.Cosh() = NaN
func (a Float16) Cosh() Float16 {
	ix := a &^ signMask16
	if ix < 0x2400 { // |a| < 2**-6
		// cosh(a) = 1 + a**2/2 + ... rounds to 1.
		return uvone16
	}
	if ix >= 0x4a00 { // |a| >= 12
		if ix > uvinf16 {
			// cosh(NaN) = NaN
			return NewFloat16NaN()
		}
		// cosh(a) overflows.
		return uvinf16
	}

	e := exp16(normal16ToFloat64(ix))
	return NewFloat16((e + 1/e) * 0.5)
}

// Tanh returns the hyperbolic tangent of a.
//
// Special cases are:
//
//	±0.Tanh() = ±0
//	±Inf.Tanh() = ±1
//	NaN.Tanh() = NaN
func (a Float16) Tanh() Float16 {
	ix := a &^ signMask16
	if ix < 0x2400 { // |a| < 2**-6
		// tanh(a) = a - a**3/3 + ... rounds to a.
		return a
	}
	if ix >= 0x4800 { // |a| >= 8
		if ix > uvinf16 {
			// tanh(NaN) = NaN
			return NewFloat16NaN()
		}
		// tanh(a) rounds to ±1.
		return a&signMask16 | uvone16
	}

	e := exp16(2 * normal16ToFloat64(ix))
	t := 1 - 2/(e+1)
	if a&signMask16 != 0 {
		t = -t
	}
	return NewFloat16(t)
}

// exp16Table[j] = 2**(j/32)
var exp16Table = [32]float64{
	0x1.0000000000000p+0, 0x1.059b0d3158574p+0, 0x1.0b5586cf9890fp+0, 0x1.11301d0125b51p+0,
	0x1.172b83c7d517bp+0, 0x1.1d4873168b9aap+0, 0x1.2387a6e756238p+0, 0x1.29e9df51fdee1p+0,
	0x1.306fe0a31b715p+0, 0x1.371a7373aa9cbp+0, 0x1.3dea64c123422p+0, 0x1.44e086061892dp+0,
	0x1.4bfdad5362a27p+0, 0x1.5342b569d4f82p+0, 0x1.5ab07dd485429p+0, 0x1.6247eb03a5585p+0,
	0x1.6a09e667f3bcdp+0, 0x1.71f75e8ec5f74p+0, 0x1.7a11473eb0187p+0, 0x1.82589994cce13p+0,
	0x1.8ace5422aa0dbp+0, 0x1.93737b0cdc5e5p+0, 0x1.9c49182a3f090p+0, 0x1.a5503b23e255dp+0,
	0x1.ae89f995ad3adp+0, 0x1.b7f76f2fb5e47p+0, 0x1.c199bdd85529cp+0, 0x1.cb720dcef9069p+0,
	0x1.d5818dcfba487p+0, 0x1.dfc97337b9b5fp+0, 0x1.ea4afa2a490dap+0, 0x1.f50765b6e4540p+0,
}

// exp16 returns e**x for |x| <= 16.
// The relative error is less than 2**-48.
func exp16(x float64) float64 {
	const (
		invLn2N = 0x1.71547652b82fep+5  // 32/ln(2)
		ln2NHi  = 0x1.62e42fefa3800p-6  // ln(2)/32 with the last 11 bits zero, so k*ln2NHi is exact for |k| < 2**11
		ln2NLo  = 0x1.ef35793c76730p-50 // ln(2)/32 - ln2NHi
		shift   = 0x1.8p52
	)

	// x = k*ln(2)/32 + r, |r| <= ln(2)/64
	t := x*invLn2N + shift
	k := int64(math.Float64bits(t) - math.Float64bits(shift))
	kf := t - shift
	r := (x - kf*ln2NHi) - kf*ln2NLo

	// e**r by the Taylor series; the error is r**6/720 < 2**-48.
	p := 1 + r + r*r*(1.0/2+r*(1.0/6+r*(1.0/24+r*(1.0/120))))

	// e**x = 2**(k/32) * e**r
	scale := math.Float64frombits(math.Float64bits(exp16Table[k&31]) + uint64(k>>5)<<52)
	return scale * p
}
