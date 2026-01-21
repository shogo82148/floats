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
func (a Float32) Pow(b Float32) Float32 {
	// https://github.com/chewxy/math32/blob/912ef0b2e4151df0148d7645c92a7b5e22f887f5/pow.go#L37-L136
	switch {
	case b == 0 || a == 1:
		return 1
	case b == 1:
		return a
	case b == 0.5:
		return a.Sqrt()
	case b == -0.5:
		return 1 / a.Sqrt()
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

	absy := b
	flip := false
	if absy < 0 {
		absy = -absy
		flip = true
	}
	yi, yf := absy.Modf()
	if yf != 0 && a < 0 {
		return NewFloat32NaN()
	}
	if yi >= 1<<31 {
		return (b * a.Log()).Exp()
	}

	// ans = a1 * 2**ae (= 1 for now).
	a1 := Float32(1.0)
	ae := 0

	// ans *= x**yf
	if yf != 0 {
		if yf > 0.5 {
			yf--
			yi++
		}
		a1 = (yf * a.Log()).Exp()
	}

	// ans *= x**yi
	// by multiplying in successive squarings
	// of x according to bits of yi.
	// accumulate powers of two into exp.
	x1, xe := a.Frexp()
	for i := int32(yi); i != 0; i >>= 1 {
		if i&1 == 1 {
			a1 *= x1
			ae += xe
		}
		x1 *= x1
		xe <<= 1
		if x1 < .5 {
			x1 += x1
			xe--
		}
	}

	// ans = a1*2**ae
	// if flip { ans = 1 / ans }
	// but in the opposite order
	if flip {
		a1 = 1 / a1
		ae = -ae
	}
	return a1.Ldexp(ae)
}

func isOddInt32(x Float32) bool {
	if x >= 1<<24 {
		// 1 << 24 is the largest exact integer in the float32 format.
		// Any number outside this range will be truncated before the decimal point and therefore will always be
		// an even integer.
		return false
	}

	xi, xf := x.Modf()
	return xf == 0 && int32(xi)&1 == 1
}
