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

		// MaxInt64 = 2^63
		MaxInt64 = Float128{0x403e_0000_0000_0000, 0x0000_0000_0000_0000}
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
	case b.Eq(Half.Neg()):
		return One.Quo(a.Sqrt())
	}

	absy := b
	flip := false
	if absy.Lt(Zero) {
		absy = absy.Neg()
		flip = true
	}
	yi, yf := absy.Modf()
	if !yf.IsZero() && a.Lt(Zero) {
		return NewFloat128NaN()
	}
	if yi.Ge(MaxInt64) {
		// yi is a large even int that will lead to overflow (or underflow to 0)
		// for all x except -1 (x == 1 was handled earlier)
		switch {
		case a.Eq(One.Neg()):
			return One
		case a.Abs().Lt(One) == (b.Gt(Zero)):
			return Zero
		default:
			return NewFloat128Inf(1)
		}
	}

	// ans = a1 * 2**ae (= 1 for now).
	a1 := One
	ae := 0

	// ans *= x**yf
	if !yf.IsZero() {
		if yf.Gt(Half) {
			yf = yf.Sub(One)
			yi = yi.Add(One)
		}
		a1 = (yf.Mul(a.Log())).Exp()
	}

	// ans *= x**yi
	// by multiplying in successive squarings
	// of x according to bits of yi.
	// accumulate powers of two into exp.
	x1, xe := a.Frexp()
	for i := yi.Int64(); i != 0; i >>= 1 {
		if i&1 != 0 {
			a1 = a1.Mul(x1)
			ae += xe
		}
		x1 = x1.Mul(x1)
		xe <<= 1
		if x1.Lt(Half) {
			x1 = x1.Add(x1)
			xe--
		}
	}

	// ans = a1*2**ae
	// if flip { ans = 1 / ans }
	// but in the opposite order
	if flip {
		a1 = One.Quo(a1)
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
