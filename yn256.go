package floats

// Yn returns the order-n Bessel function of the second kind.
//
// Special cases are:
//
//	Yn(n, +Inf) = 0
//	Yn(n >= 0, 0) = -Inf
//	Yn(n < 0, 0) = +Inf if n is odd, -Inf if n is even
//	Yn(n, x < 0) = NaN
//	Yn(n, NaN) = NaN
func (a Float256) Yn(n int) Float256 {
	switch {
	case a.IsNaN():
		return a
	case a.IsInf(1):
		return Float256{}
	}

	neg := false
	if n < 0 {
		n = -n
		neg = !neg
	}

	switch {
	case a.IsZero():
		if neg && n%2 == 1 {
			return NewFloat256Inf(1)
		}
		return NewFloat256Inf(-1)
	case a.Signbit():
		return NewFloat256NaN()
	}

	var y Float256
	switch n {
	case 0:
		y = a.Y0()
	case 1:
		y = a.Y1()
	default:
		var ok bool
		if y, ok = ynAsymptotic256(n, a); !ok {
			y = ynForward256(n, a)
		}
	}
	if neg && n%2 == 1 {
		y = y.Neg()
	}
	return y
}

// ynAsymptotic256 returns Yn(x) for n >= 2, x > 0 using Hankel's asymptotic expansion
//
//	Yn(x) ~ sqrt(2/(pi*x)) * (P(x)*sin(x-n*pi/2-pi/4) + Q(x)*cos(x-n*pi/2-pi/4))
//	      = sqrt(1/(pi*x)) * (P(x)*(s-c) + Q(x)*(s+c)),  s = sin(x-n*pi/2), c = cos(x-n*pi/2)
//	P(x) = sum (-1)**k a(2k)/x**(2k),  Q(x) = sum (-1)**k a(2k+1)/x**(2k+1)
//	a(k) = prod (4n**2-(2j-1)**2, j = 1, ..., k) / (k! 8**k)
//
// The series is asymptotic, and it is accurate enough only for x >= 128 and x >= n**2/2.
// ok is false if it does not converge enough, and then the forward recurrence should be used.
func ynAsymptotic256(n int, x Float256) (y Float256, ok bool) {
	if !x.Ge(NewFloat256(128)) || !x.Mul(NewFloat256(2)).Ge(NewFloat256(float64(n)).Mul(NewFloat256(float64(n)))) {
		return Float256{}, false
	}
	var (
		One = Float256(uvone256)
		Pi  = Float256{
			0x4000_0921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
	)

	// t is a(k)/x**k.
	fn := NewFloat256(float64(n))
	mu := fn.Mul(fn).Mul(NewFloat256(4))
	inv8x := One.Quo(x.Mul(NewFloat256(8)))
	t := One
	p, q := One, Float256{}
	converged := false
	// 2**-250
	const small = uint64(bias256-250) << (shift256 - 192)
	for k := 1; k <= 500; k++ {
		odd := NewFloat256(float64((2*k - 1) * (2*k - 1)))
		next := t.Mul(mu.Sub(odd)).Mul(inv8x).Quo(NewFloat256(float64(k)))
		if next.Abs()[0] < small {
			converged = true
			break
		}
		if next.Abs().Gt(t.Abs()) && next.Abs()[0] > uint64(bias256-200)<<(shift256-192) {
			// The terms start increasing before they are small enough.
			return Float256{}, false
		}
		t = next
		// the sign is (-1)**(k/2).
		if k&2 != 0 {
			next = next.Neg()
		}
		if k&1 == 0 {
			p = p.Add(next)
		} else {
			q = q.Add(next)
		}
	}
	if !converged {
		return Float256{}, false
	}

	sin, cos := x.Sincos()
	switch n & 3 {
	case 1:
		sin, cos = cos.Neg(), sin
	case 2:
		sin, cos = sin.Neg(), cos.Neg()
	case 3:
		sin, cos = cos, sin.Neg()
	}
	amp := One.Quo(Pi.Mul(x)).Sqrt()
	return amp.Mul(p.Mul(sin.Sub(cos)).Add(q.Mul(sin.Add(cos)))), true
}

// ynForward256 returns Yn(x) for n >= 2, x > 0 by the forward recurrence
//
//	Y(k+1)(x) = 2k/x * Y(k)(x) - Y(k-1)(x)
//
// from Y0(x) and Y1(x). It is stable, but the rounding errors of each step are accumulated, so it is calculated
// in the extended precision; see ynExt256.
func ynForward256(n int, x Float256) Float256 {
	y0, y1 := y0y1_256(x)
	if y0.IsInf(0) || y1.IsInf(0) {
		// x is so small that Y1(x) overflows. |Yn(x)| >= |Y1(x)|, and it is negative.
		return NewFloat256Inf(-1)
	}
	inv := ynExt256Inv(ynExt256FromFloat256(x))
	prev, cur := ynExt256FromFloat256(y0), ynExt256FromFloat256(y1)
	for k := 1; k < n; k++ {
		c := ynExt256Mul(ynExt256FromUint(2*uint64(k)), inv)
		prev, cur = cur, ynExt256Add(ynExt256Mul(c, cur), ynExt256{m: prev.m, e: prev.e, neg: !prev.neg})
		if cur.e > bias256+1 {
			// The result overflows. |Yk(x)| grows monotonically for k > x/2 once it exceeds 1, and it is negative.
			return NewFloat256Inf(-1)
		}
	}
	return cur.Float256()
}

// y0y1_256 returns Y0(x) and Y1(x) for x > 0, sharing the calculation of them.
func y0y1_256(x Float256) (y0, y1 Float256) {
	var (
		Two = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Threshold128 = Float256{
			0x4000_6000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)

	switch {
	case x.Lt(Two):
		return y0y1Series256(x)
	case x.Lt(Threshold128):
		// Y1(x) = -Y0'(x)
		y0, dy := y0Taylor256(x, true)
		return y0, dy.Neg()
	default:
		return y0y1Asymptotic256(x)
	}
}

// y0y1Series256 returns Y0(x) and Y1(x) for 0 < x < 2; see y0Series256 and y1Series256.
func y0y1Series256(x Float256) (y0, y1 Float256) {
	var (
		Two = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Quarter = Float256{
			0x3fff_d000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		Euler = Float256{
			0x3fff_e278_8cfc_6fb6, 0x18f4_9a37_c7f0_202a,
			0x596a_d439_d987_5ecb, 0x9803_2180_7be6_8e13,
		}
	)

	z := x.Mul(x).Mul(Quarter)
	j0 := y0256SeriesJ[len(y0256SeriesJ)-1]
	s0 := y0256SeriesS[len(y0256SeriesS)-1]
	j1 := y1256SeriesJ[len(y1256SeriesJ)-1]
	s1 := y1256SeriesS[len(y1256SeriesS)-1]
	for k := len(y0256SeriesJ) - 2; k >= 0; k-- {
		j0 = FMA256(j0, z, y0256SeriesJ[k])
		s0 = FMA256(s0, z, y0256SeriesS[k])
	}
	for k := len(y1256SeriesJ) - 2; k >= 0; k-- {
		j1 = FMA256(j1, z, y1256SeriesJ[k])
		s1 = FMA256(s1, z, y1256SeriesS[k])
	}
	l := x.Quo(Two).Log().Add(Euler)
	y0 = FMA256(l, j0, s0).Mul(y0256TwoOverPi)
	y1 = FMA256(FMA256(l, j1, s1), x.Mul(y1256OneOverPi), y1256TwoOverPi.Quo(x).Neg())
	return y0, y1
}

// y0y1Asymptotic256 returns Y0(x) and Y1(x) for x >= 128; see y0Asymptotic256 and y1Asymptotic256.
func y0y1Asymptotic256(x Float256) (y0, y1 Float256) {
	var (
		One = Float256(uvone256)
		Pi  = Float256{
			0x4000_0921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
	)

	w := One.Quo(x.Mul(x))

	terms := len(y0256HankelP) // 128 <= x
	switch xf := x.Float64().BuiltIn(); {
	case xf >= 1e5:
		terms = 9
	case xf >= 4096:
		terms = 13
	case xf >= 1024:
		terms = 17
	case xf >= 512:
		terms = 20
	case xf >= 256:
		terms = 26
	}

	p0 := y0256HankelP[terms-1]
	q0 := y0256HankelQ[terms-1]
	p1 := y1256HankelP[terms-1]
	q1 := y1256HankelQ[terms-1]
	for k := terms - 2; k >= 0; k-- {
		p0 = FMA256(p0, w, y0256HankelP[k])
		q0 = FMA256(q0, w, y0256HankelQ[k])
		p1 = FMA256(p1, w, y1256HankelP[k])
		q1 = FMA256(q1, w, y1256HankelQ[k])
	}
	q0 = q0.Quo(x)
	q1 = q1.Quo(x)

	sin, cos := x.Sincos()
	amp := One.Quo(Pi.Mul(x)).Sqrt()
	y0 = amp.Mul(p0.Add(q0).Mul(sin).Add(q0.Sub(p0).Mul(cos)))
	y1 = amp.Mul(q1.Sub(p1).Mul(sin).Sub(p1.Add(q1).Mul(cos)))
	return y0, y1
}
