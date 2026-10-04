package floats

import "math"

// Sinh returns the hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Sinh() = ±0
//	±Inf.Sinh() = ±Inf
//	NaN.Sinh() = NaN
func (a Float32) Sinh() Float32 {
	ix := a.Bits() &^ signMask32
	if ix < 0x39800000 { // |a| < 2**-12
		// sinh(a) = a + a**3/6 + ... rounds to a.
		return a
	}
	if ix >= 0x42b30000 { // |a| >= 89.5
		if ix > uvinf32 {
			// sinh(NaN) = NaN
			return NewFloat32NaN()
		}
		// sinh(a) overflows.
		return NewFloat32FromBits(a.Bits()&signMask32 | uvinf32)
	}

	// sinh(x) = (e**x - e**-x)/2 = (E + E/(E+1))/2 where E = e**x - 1
	e := expm1Small(float64(math.Float32frombits(ix)))
	s := 0.5 * (e + e/(e+1))
	if a < 0 {
		s = -s
	}
	return Float32(s)
}

// Cosh returns the hyperbolic cosine of a.
//
// Special cases are:
//
//	±0.Cosh() = 1
//	±Inf.Cosh() = +Inf
//	NaN.Cosh() = NaN
func (a Float32) Cosh() Float32 {
	ix := a.Bits() &^ signMask32
	if ix < 0x39800000 { // |a| < 2**-12
		// cosh(a) = 1 + a**2/2 + ... rounds to 1.
		return 1
	}
	if ix >= 0x42b30000 { // |a| >= 89.5
		if ix > uvinf32 {
			// cosh(NaN) = NaN
			return NewFloat32NaN()
		}
		// cosh(a) overflows.
		return NewFloat32Inf(1)
	}

	e := expm1Small(float64(math.Float32frombits(ix))) + 1
	return Float32(0.5 * (e + 1/e))
}

// Tanh returns the hyperbolic tangent of a.
//
// Special cases are:
//
//	±0.Tanh() = ±0
//	±Inf.Tanh() = ±1
//	NaN.Tanh() = NaN
func (a Float32) Tanh() Float32 {
	ix := a.Bits() &^ signMask32
	if ix < 0x39800000 { // |a| < 2**-12
		// tanh(a) = a - a**3/3 + ... rounds to a.
		return a
	}
	if ix >= 0x41200000 { // |a| >= 10
		if ix > uvinf32 {
			// tanh(NaN) = NaN
			return NewFloat32NaN()
		}
		// tanh(a) rounds to ±1.
		if a < 0 {
			return -1
		}
		return 1
	}

	// tanh(x) = (e**2x - 1)/(e**2x + 1) = E/(E+2) where E = e**2x - 1
	e := expm1Small(2 * float64(math.Float32frombits(ix)))
	t := e / (e + 2)
	if a < 0 {
		t = -t
	}
	return Float32(t)
}

// expm1Small returns e**x - 1 for |x| <= 104.
// The relative error is less than 2**-48.
func expm1Small(x float64) float64 {
	// e**x - 1 = 2**(k/32) * (e**r - 1) + (2**(k/32) - 1)
	scale, p := expReduce(x)
	return scale*p + (scale - 1)
}

// expReduce returns scale = 2**(k/32) and p = e**r - 1 such that
// x = k*ln(2)/32 + r and |r| <= ln(2)/64, for |x| <= 104.
// The relative error of p is less than 2**-51.
func expReduce(x float64) (scale, p float64) {
	const (
		invLn2N = 0x1.71547652b82fep+5  // 32/ln(2)
		ln2NHi  = 0x1.62e42fefa2000p-6  // ln(2)/32 with the last 13 bits zero, so k*ln2NHi is exact for |k| < 2**13
		ln2NLo  = 0x1.9ef35793c7673p-46 // ln(2)/32 - ln2NHi
		shift   = 0x1.8p52
	)

	t := x*invLn2N + shift
	k := int64(math.Float64bits(t) - math.Float64bits(shift))
	kf := t - shift
	r := (x - kf*ln2NHi) - kf*ln2NLo

	// e**r - 1 by the Taylor series; the relative error is about r**6/5040 < 2**-51.
	p = r + r*r*(1.0/2+r*(1.0/6+r*(1.0/24+r*(1.0/120+r*(1.0/720)))))

	scale = math.Float64frombits(math.Float64bits(exp2Table[k&31]) + uint64(k>>5)<<52)
	return
}
