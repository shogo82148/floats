package floats

import "math"

// Pow returns a**b, the base-a exponential of b.
//
// Special cases are (in order):
//
//	a.Pow(±0) = 1 for any a
//	1.Pow(b) = 1 for any b
//	a.Pow(1) = a for any a
//	NaN.Pow(b) = NaN
//	a.Pow(NaN) = NaN
//	±0.Pow(b) = ±Inf for b an odd integer < 0
//	±0.Pow(-Inf) = +Inf
//	±0.Pow(+Inf) = +0
//	±0.Pow(b) = +Inf for finite b < 0 and not an odd integer
//	±0.Pow(b) = ±0 for b an odd integer > 0
//	±0.Pow(b) = +0 for finite b > 0 and not an odd integer
//	-1.Pow(±Inf) = 1
//	a.Pow(+Inf) = +Inf for |a| > 1
//	a.Pow(-Inf) = +0 for |a| > 1
//	a.Pow(+Inf) = +0 for |a| < 1
//	a.Pow(-Inf) = +Inf for |a| < 1
//	+Inf.Pow(b) = +Inf for b > 0
//	+Inf.Pow(b) = +0 for b < 0
//	-Inf.Pow(b) = (-0).Pow(-b)
//	a.Pow(b) = NaN for finite a < 0 and finite non-integer b
func (a Float32) Pow(b Float32) Float32 {
	ia, ib := a.Bits(), b.Bits()
	if ia-1 < 0x7f7fffff && (ib&^signMask32)-1 < 0x7f7fffff {
		// fast path: a is positive finite, and b is finite non-zero.
		return Float32(pow32(ia, float64(b)))
	}
	return powSpecial32(a, b)
}

// powSpecial32 handles the special cases of Pow, and negative a.
func powSpecial32(a, b Float32) Float32 {
	switch {
	case b == 0 || a == 1:
		return 1
	case b == 1:
		return a
	case a.IsNaN() || b.IsNaN():
		return NewFloat32NaN()
	case a == 0:
		switch {
		case b < 0:
			if isOddInt32(b) {
				return NewFloat32Inf(1).Copysign(a)
			}
			return NewFloat32Inf(1)
		case b > 0:
			if isOddInt32(b) {
				return a
			}
			return 0
		}
	case b.IsInf(0):
		switch {
		case a == -1:
			return 1
		case (a.Abs() < 1) == b.IsInf(1):
			return 0
		default:
			return NewFloat32Inf(1)
		}
	case a.IsInf(0):
		if a.IsInf(-1) {
			return (1. / a).Pow(-b) // Pow(-0, -b)
		}
		switch {
		case b < 0:
			return 0
		case b > 0:
			return NewFloat32Inf(1)
		}
	}

	// a is negative finite, and b is finite non-zero.
	_, frac := b.Modf()
	if frac != 0 {
		return NewFloat32NaN()
	}
	r := pow32(a.Bits()&^signMask32, float64(b))
	if isOddInt32(b) {
		r = -r
	}
	return Float32(r)
}

// pow32 returns x**y, where ix is the bits of a positive finite Float32 value x.
// The relative error is less than 2**-40,
// so the result rounded to Float32 is within 1 ulp, and correctly rounded except for rare cases.
// Exact results on the midpoint of two adjacent Float32 values are rounded to even,
// if y is an integer in [1, 64] or x is a power of two.
func pow32(ix uint32, y float64) float64 {
	// x**y = e**(y*log(x))
	k, l := logKernel32(ix)
	t := y * (k*math.Ln2 + l)
	if !(t >= -104 && t <= 89) {
		if t > 0 {
			// overflow
			return math.Inf(1)
		}
		// underflow
		return 0
	}
	scale, p := expReduce(t)
	r := scale + scale*p

	// An exact result may be the midpoint of two adjacent Float32 values,
	// and it must be rounded to even. Such results are close to the midpoint,
	// so check them again if r is within 2**-38 of the midpoint, or in the subnormal range.
	const mask = 1<<(shift64-shift32) - 1
	const half = 1 << (shift64 - shift32 - 1)
	if d := math.Float64bits(r)&mask - half; d+1<<14 < 1<<15 || r < 0x1p-126 {
		return powExact32(ix, y, k, l, r)
	}
	return r
}

// powExact32 returns x**y exactly if it is easily computed, and r otherwise.
// ix is the bits of a positive finite Float32 value x, and log(x) = k*ln(2) + l.
func powExact32(ix uint32, y, k, l, r float64) float64 {
	if l == 0 {
		// x is a power of two.
		if e := k * y; e == math.Trunc(e) {
			return math.Ldexp(1, int(e))
		}
		return r
	}
	if y == math.Trunc(y) && y > 0 && y <= 64 {
		// If x**y is the midpoint of two adjacent Float32 values,
		// it has 25 significant bits, and so does x**i for i <= y.
		// So the following multiplications are exact.
		x := float64(math.Float32frombits(ix))
		ret := 1.0
		for n := int(y); n > 0; n >>= 1 {
			if n&1 != 0 {
				ret *= x
			}
			x *= x
		}
		return ret
	}
	return r
}

func isOddInt32(x Float32) bool {
	if x.Abs() >= 1<<24 {
		// 1 << 24 is the largest exact integer in the float32 format.
		// Any number outside this range will be truncated before the decimal point and therefore will always be
		// an even integer.
		return false
	}

	xi, xf := x.Modf()
	return xf == 0 && int32(xi)&1 == 1
}
