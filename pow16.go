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
func (a Float16) Pow(b Float16) Float16 {
	if a-1 < uvinf16-1 && (b&^signMask16)-1 < uvinf16-1 {
		// fast path: a is positive finite, and b is finite non-zero.
		return NewFloat16(pow16(a, finite16ToFloat64(b)))
	}
	if a > signMask16 && a < uvneginf16 && (b&^signMask16)-1 < uvinf16-1 {
		// a is negative finite, and b is finite non-zero.
		y := finite16ToFloat64(b)
		if y != math.Trunc(y) {
			return NewFloat16NaN()
		}
		r := pow16(a&^signMask16, y)
		if math.Mod(y, 2) != 0 {
			r = -r
		}
		return NewFloat16(r)
	}
	// special cases
	return NewFloat16(math.Pow(a.Float64().BuiltIn(), b.Float64().BuiltIn()))
}

// pow16 returns x**y, where x is a positive finite Float16 value.
// The relative error is less than 2**-40, and exact results on the midpoint of
// two adjacent Float16 values are computed exactly, so the result rounded to Float16 is correctly rounded.
func pow16(x Float16, y float64) float64 {
	// x**y = e**(y*log(x))
	k, l := logKernel64(positive16ToFloat64(x))
	t := y * (k*math.Ln2 + l)
	if !(t >= -20 && t <= 12) {
		if t > 0 {
			// overflow
			return math.Inf(1)
		}
		// underflow
		return 0
	}
	scale, p := expReduce(t)
	r := scale + scale*p

	// An exact result may be the midpoint of two adjacent Float16 values,
	// and it must be rounded to even. Such results are close to the midpoint,
	// so check them again if r is within 2**-38 of the midpoint, or in the subnormal range.
	const mask = 1<<(shift64-shift16) - 1
	const half = 1 << (shift64 - shift16 - 1)
	if d := math.Float64bits(r)&mask - half; d+1<<14 < 1<<15 || r < 0x1p-14 {
		return powExact16(x, y, k, l, r)
	}
	return r
}

// powExact16 returns x**y exactly if it is easily computed, and r otherwise.
// x is a positive finite Float16 value, and log(x) = k*ln(2) + l.
func powExact16(x Float16, y, k, l, r float64) float64 {
	if l == 0 {
		// x is a power of two.
		if e := k * y; e == math.Trunc(e) {
			return math.Ldexp(1, int(e))
		}
		return r
	}

	// If y = n/2**q, x**y = (x**(1/2**q))**n.
	// Take square roots while they are exact.
	f := positive16ToFloat64(x)
	for i := 0; y != math.Trunc(y); i++ {
		s := math.Sqrt(f)
		if i == 4 || s*s != f {
			return r
		}
		f, y = s, y*2
	}
	if y > 0 && y <= 64 {
		// If x**y is the midpoint of two adjacent Float16 values,
		// it has 12 significant bits, and so does f**i for i <= y.
		// So the following multiplications are exact.
		ret := 1.0
		for n := int(y); n > 0; n >>= 1 {
			if n&1 != 0 {
				ret *= f
			}
			f *= f
		}
		return ret
	}
	return r
}

// finite16ToFloat64 converts a finite Float16 value a to float64.
func finite16ToFloat64(a Float16) float64 {
	return math.Copysign(positive16ToFloat64(a&^signMask16), float64(int16(a)))
}
