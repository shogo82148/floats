package floats

// Log1p returns the natural logarithm of 1 plus its argument x.
// It is more accurate than [Log](1 + x) when x is near zero.
//
// Special cases are:
//
//	+Inf.Log1p() = +Inf
//	±0.Log1p() = ±0
//	-1.Log1p() = -Inf
//	(x < -1).Log1p() = NaN
//	NaN.Log1p() = NaN
func (a Float32) Log1p() Float32 {
	const (
		Sqrt2M1     = 4.142135623730950488017e-01  // Sqrt(2)-1 = 0x3fda827999fcef34
		Sqrt2HalfM1 = -2.928932188134524755992e-01 // Sqrt(2)/2-1 = 0xbfd2bec333018866
		Small       = 0x1p-24                      // 2**-24
		Ln2Hi       = 6.9313812256e-01             // 0x3f317180
		Ln2Lo       = 9.0580006145e-06             // 0x3717f7d1
		/* |(log(1+s)-log(1-s))/s - Lg(s)| < 2**-34.24 (~[-4.95e-11, 4.97e-11]). */
		Lg1 = 0xaaaaaa.0p-24 // 0.66666662693
		Lg2 = 0xccce13.0p-25 // 0.40000972152
		Lg3 = 0x91e9ee.0p-25 // 0.28498786688
		Lg4 = 0xf89e26.0p-26 // 0.24279078841
	)

	// special cases
	switch {
	case a < -1 || a.IsNaN(): // includes -Inf
		return NewFloat32NaN()
	case a == -1:
		return NewFloat32Inf(-1)
	case a.IsInf(1):
		return NewFloat32Inf(1)
	}

	absx := a.Abs()
	var f Float32
	var c Float32
	k := 1
	if absx < Sqrt2M1 { // |x| < Sqrt(2)-1
		if absx < Small {
			return a
		}
		if a > Sqrt2HalfM1 { // Sqrt(2)/2-1 < x
			// (Sqrt(2)/2-1) < x < (Sqrt(2)-1)
			k = 0
			c = 0
			f = a
		}
	}
	if k != 0 {
		g := 1 + a
		iu := g.Bits()
		iu += 0x3f800000 - 0x3f3504f3
		k = int(iu>>shift32) - bias32
		// correction term ~ log(1+x)-log(u), avoid underflow in c/u
		if k < 25 {
			if k >= 2 {
				c = 1 - (g - a)
			} else {
				c = a - (g - 1)
			}
			c /= g
		} else {
			c = 0
		}
		// reduce u into [sqrt(2)/2, sqrt(2)]
		iu = (iu & 0x007fffff) + 0x3f3504f3
		f = NewFloat32FromBits(iu) - 1
	}

	s := f / (2.0 + f)
	z := s * s
	w := z * z
	t1 := w * (Lg2 + w*Lg4)
	t2 := z * (Lg1 + w*Lg3)
	R := t2 + t1
	hfsq := 0.5 * f * f
	dk := Float32(k)
	return s*(hfsq+R) + (dk*Ln2Lo + c) - hfsq + f + dk*Ln2Hi
}
