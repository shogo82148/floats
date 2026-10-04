package floats

import "math"

// Exp returns e**x, the base-e exponential of a.
//
// Special cases are:
//
//	+Inf.Exp() = +Inf
//	NaN.Exp() = NaN
//
// Very large values overflow to 0 or +Inf.
// Very small values underflow to 1.
func (a Float32) Exp() Float32 {
	x := float64(a)
	if !(x >= -104 && x <= 89) {
		switch {
		case x != x:
			// exp(NaN) = NaN
			return NewFloat32NaN()
		case x > 0:
			// exp(x) overflows.
			return NewFloat32Inf(1)
		default:
			// exp(x) underflows.
			return 0
		}
	}

	// The result may overflow or underflow in float32,
	// but it is correctly handled by the conversion from float64.
	scale, p := expReduce(x)
	return Float32(scale + scale*p)
}

// Exp2 returns 2**x, the base-2 exponential of x.
//
// Special cases are the same as [Exp].
func (a Float32) Exp2() Float32 {
	x := float64(a)
	if !(x >= -151 && x <= 129) {
		switch {
		case x != x:
			// exp2(NaN) = NaN
			return NewFloat32NaN()
		case x > 0:
			// exp2(x) overflows.
			return NewFloat32Inf(1)
		default:
			// exp2(x) underflows.
			return 0
		}
	}

	// The result may overflow or underflow in float32,
	// but it is correctly handled by the conversion from float64.
	scale, p := exp2Reduce(x)
	return Float32(scale + scale*p)
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
	return expKernel(k, r)
}

// exp2Reduce returns scale = 2**(k/32) and p = e**r - 1 such that
// 2**x = 2**(k/32) * e**r and |r| <= ln(2)/64, for a float32 value x with |x| <= 151.
func exp2Reduce(x float64) (scale, p float64) {
	const shift = 0x1.8p52

	// x = k/32 + f, |f| <= 1/64
	// x*32 and f are exact, because x is a float32 value.
	t := x*32 + shift
	k := int64(math.Float64bits(t) - math.Float64bits(shift))
	f := x - (t-shift)*(1.0/32)
	return expKernel(k, f*math.Ln2)
}

// expKernel returns scale = 2**(k/32) and p = e**r - 1 for |r| <= ln(2)/64.
// The relative error of p is less than 2**-51.
func expKernel(k int64, r float64) (scale, p float64) {
	// e**r - 1 by the Taylor series; the relative error is about r**6/5040 < 2**-51.
	p = r + r*r*(1.0/2+r*(1.0/6+r*(1.0/24+r*(1.0/120+r*(1.0/720)))))

	scale = math.Float64frombits(math.Float64bits(exp2Table[k&31]) + uint64(k>>5)<<52)
	return
}
