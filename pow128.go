package floats

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
func (a Float128) Pow(b Float128) Float128 {
	var (
		// Zero = 0
		Zero = Float128{}

		// One = 1.0
		One = Float128(uvone128)

		// Half = 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	switch {
	case b.IsZero() || a.Eq(One):
		return One
	case b.Eq(One):
		return a
	case a.IsNaN() || b.IsNaN():
		return NewFloat128NaN()
	case a.IsZero():
		switch {
		case b.Lt(Zero):
			if isOddInt128(b) {
				return NewFloat128Inf(1).Copysign(a)
			}
			return NewFloat128Inf(1)
		case b.Gt(Zero):
			if isOddInt128(b) {
				return a
			}
			return Zero
		}
	case b.IsInf(0):
		switch {
		case a.Eq(One.Neg()):
			return One
		case (a.Abs().Lt(One)) == b.IsInf(1):
			return Zero
		default:
			return NewFloat128Inf(1)
		}
	case a.IsInf(0):
		if a.IsInf(-1) {
			return (Zero.Neg()).Pow(b.Neg()) // Pow(-0, -b)
		}
		switch {
		case b.Lt(Zero):
			return Zero
		case b.Gt(Zero):
			return NewFloat128Inf(1)
		}
	case b.Eq(Half):
		return a.Sqrt()
	}

	// a**b is calculated in Float256, whose precision is enough to round correctly except for the cases which are
	// extremely close to the midpoint of two adjacent Float128 values.
	bi, bf := b.Abs().Modf()
	if !bf.IsZero() && a.Lt(Zero) {
		return NewFloat128NaN()
	}
	neg := a.Signbit() && isOddInt128(b)
	x := a.Abs().Float256()

	var r Float256
	if x1, xe := x.Frexp(); bf.IsZero() && x1.Eq(Float256{0x3fff_e000_0000_0000}) {
		// a is a power of two, so that the result is exactly a power of two, and it may be the midpoint of
		// the subnormal numbers.
		r = powTwo256(xe-1, bi, b.Lt(Zero))
	} else if bf.IsZero() && bi.Le(Float128{0x4006_0000_0000_0000, 0}) { // |b| <= 2**7
		// The product of the integers is exact if it has at most 237 bits, so that it is correctly rounded even if
		// it is the midpoint of two adjacent Float128 values. The result of a**b has more than 114 bits, so that it
		// can not be the midpoint, if |b| > 71 (a is not a power of two). Otherwise it is calculated by exp and log.
		r = powInt256(x, int(bi.Int64()), b.Lt(Zero))
	} else {
		// a**b = exp(b ln(a)). The absolute error of b ln(a) is about |b ln(a)| 2**-236, that is less than 2**-220
		// if the result does not overflow or underflow.
		z := b.Float256().Mul(x.Log())
		switch {
		case z.Gt(Float256{0x4000_d388_0000_0000}): // 20000
			r = NewFloat256Inf(1) // overflow
		case z.Lt(Float256{0xc000_d388_0000_0000}):
			r = Float256{} // underflow
		default:
			r = z.Exp()
		}
	}

	y := r.Float128()
	if neg {
		return y.Neg()
	}
	return y
}

// powTwo256 returns (2**e)**n for e != 0 and the non-negative integer n, or its reciprocal if flip is true.
func powTwo256(e int, n Float128, flip bool) Float256 {
	// the exponent of the result is e n, which overflows or underflows if |e n| > 20000. n is limited not to overflow int.
	if n.Gt(Float128{0x4010_0000_0000_0000, 0}) { // n > 2**17
		n = Float128{0x4010_0000_0000_0000, 0}
	}
	ae := e * int(n.Int64())
	if flip {
		ae = -ae
	}
	return Float256(uvone256).Ldexp(ae)
}

// powInt256 returns x**n for x > 0 and n >= 0 by multiplying in successive squarings of x according to bits of n.
// The exponent of the result is accumulated separately so that it does not overflow in the intermediate steps.
// If flip is true, it returns x**-n.
func powInt256(x Float256, n int, flip bool) Float256 {
	var (
		one  = Float256(uvone256)
		half = Float256{0x3fff_e000_0000_0000}
	)

	// ans = a1 * 2**ae (= 1 for now).
	a1 := one
	ae := 0

	x1, xe := x.Frexp()
	for i := n; i != 0; i >>= 1 {
		if i&1 != 0 {
			a1 = a1.Mul(x1)
			ae += xe
		}
		x1 = x1.Mul(x1)
		xe <<= 1
		if x1.Lt(half) {
			x1 = x1.Add(x1)
			xe--
		}
	}

	// ans = a1*2**ae
	// if flip { ans = 1 / ans }
	// but in the opposite order
	if flip {
		a1 = one.Quo(a1)
		ae = -ae
	}
	return a1.Ldexp(ae)
}

func isOddInt128(x Float128) bool {
	// MaxSafeInteger = 2**113
	var MaxSafeInteger = Float128{0x4070000000000000, 0x0000000000000000}
	if x.Abs().Ge(MaxSafeInteger) {
		// 1 << 113 is the largest exact integer in the float128 format.
		// Any number outside this range will be truncated before the decimal point and therefore will always be
		// an even integer.
		return false
	}

	xi, xf := x.Modf()
	return xf.IsZero() && xi.Int128()[1]&1 == 1
}
