package floats

// Expm1 returns e**a - 1, the base-e exponential of a minus 1.
// It is more accurate than Exp(a) - 1 when a is near zero.
//
// Special cases are:
//
//	+Inf.Expm1() = +Inf
//	-Inf.Expm1() = -1
//	NaN.Expm1() = NaN
//
// Very large values overflow to -1 or +Inf.
func (a Float32) Expm1() Float32 {
	// https://github.com/chewxy/math32/blob/912ef0b2e4151df0148d7645c92a7b5e22f887f5/expm1f.go#L131-L241
	const (
		Othreshold = 89.415985                   // 0x42b2d4fc
		Ln2X27     = 1.871497344970703125e+01    // 0x4195b844
		Ln2HalfX3  = 1.0397207736968994140625    // 0x3F851592
		Ln2Half    = 3.465735912322998046875e-01 // 0x3eb17218
		Ln2Hi      = 6.9313812256e-01            // 0x3f317180
		Ln2Lo      = 9.0580006145e-06            // 0x3717f7d1
		InvLn2     = 1.4426950216e+00            // 0x3fb8aa3b
		Tiny       = 1.0 / (1 << 24)             // 2**-24 = 0x3c900000

		/* scaled coefficients related to expm1 */
		Q1 = -3.3333335072e-02 /* 0xbd088889 */
		Q2 = 1.5873016091e-03  /* 0x3ad00d01 */
		Q3 = -7.9365076090e-05 /* 0xb8a670cd */
		Q4 = 4.0082177293e-06  /* 0x36867e54 */
		Q5 = -2.0109921195e-07 /* 0xb457edbb */
	)

	// special cases
	switch {
	case a.IsInf(1) || a.IsNaN():
		return a
	case a.IsInf(-1):
		return -1
	}

	absx := a
	sign := false
	if a < 0 {
		absx = -absx
		sign = true
	}

	// filter out huge argument
	if absx >= Ln2X27 { // if |x| >= 27 * ln2
		if sign {
			return -1 // x < -56*ln2, return -1
		}
		if absx >= Othreshold { // if |x| >= 89.415985...
			return NewFloat32Inf(1)
		}
	}

	// argument reduction
	var c Float32
	var k int
	if absx > Ln2Half { // if  |x| > 0.5 * ln2
		var hi, lo Float32
		if absx < Ln2HalfX3 { // and |x| < 1.5 * ln2
			if !sign {
				hi = a - Ln2Hi
				lo = Ln2Lo
				k = 1
			} else {
				hi = a + Ln2Hi
				lo = -Ln2Lo
				k = -1
			}
		} else {
			if !sign {
				k = int(InvLn2*a + 0.5)
			} else {
				k = int(InvLn2*a - 0.5)
			}
			t := Float32(k)
			hi = a - t*Ln2Hi // t * Ln2Hi is exact here
			lo = t * Ln2Lo
		}
		a = hi - lo
		c = (hi - a) - lo
	} else if absx < Tiny { // when |a| < 2**-24, return a
		return a
	} else {
		k = 0
	}

	// a is now in primary range
	hfx := 0.5 * a
	hxs := a * hfx
	r1 := 1 + hxs*(Q1+hxs*(Q2+hxs*(Q3+hxs*(Q4+hxs*Q5))))
	t := 3 - r1*hfx
	e := hxs * ((r1 - t) / (6.0 - a*t))
	if k != 0 {
		e = (a*(e-c) - c)
		e -= hxs
		switch {
		case k == -1:
			return 0.5*(a-e) - 0.5
		case k == 1:
			if a < -0.25 {
				return -2 * (e - (a + 0.5))
			}
			return 1 + 2*(a-e)
		case k <= -2 || k > 56: // suffice to return exp(a)-1
			y := 1 - (e - a)
			y = NewFloat32FromBits(y.Bits() + uint32(k)<<23) // add k to y's exponent
			return y - 1
		}
		if k < 20 {
			t := NewFloat32FromBits(0x3f800000 - (0x1000000 >> uint(k))) // t=1-2**-k
			y := t - (e - a)
			y = NewFloat32FromBits(y.Bits() + uint32(k)<<23) // add k to y's exponent
			return y
		}
		t := NewFloat32FromBits(uint32(0x7f-k) << 23) // 2**-k
		y := a - (e + t)
		y += 1
		y = NewFloat32FromBits(y.Bits() + uint32(k)<<23) // add k to y's exponent
		return y
	}
	return a - (a*e - hxs) // c is 0
}
