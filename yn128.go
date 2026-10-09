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
func (a Float128) Yn(n int) Float128 {
	switch {
	case a.IsNaN():
		return a
	case a.IsInf(1):
		return Float128{}
	}

	neg := false
	if n < 0 {
		n = -n
		neg = !neg
	}

	switch {
	case a.IsZero():
		if neg && n%2 == 1 {
			return NewFloat128Inf(1)
		}
		return NewFloat128Inf(-1)
	case a.Signbit():
		return NewFloat128NaN()
	}

	var y Float128
	switch n {
	case 0:
		y = a.Y0()
	case 1:
		y = a.Y1()
	default:
		var ok bool
		if y, ok = ynAsymptotic128(n, a); !ok {
			y = ynForward128(n, a)
		}
	}
	if neg && n%2 == 1 {
		y = y.Neg()
	}
	return y
}

// ynAsymptotic128 returns Yn(x) for n >= 2, x > 0 using Hankel's asymptotic expansion
//
//	Yn(x) ~ sqrt(2/(pi*x)) * (P(x)*sin(x-n*pi/2-pi/4) + Q(x)*cos(x-n*pi/2-pi/4))
//	      = sqrt(1/(pi*x)) * (P(x)*(s-c) + Q(x)*(s+c)),  s = sin(x-n*pi/2), c = cos(x-n*pi/2)
//	P(x) = sum (-1)**k a(2k)/x**(2k),  Q(x) = sum (-1)**k a(2k+1)/x**(2k+1)
//	a(k) = prod (4n**2-(2j-1)**2, j = 1, ..., k) / (k! 8**k)
//
// The series is asymptotic, and it is accurate enough only for x >= 64 and x >= n**2/2.
// ok is false if it does not converge enough, and then the forward recurrence should be used.
func ynAsymptotic128(n int, x Float128) (y Float128, ok bool) {
	if !x.Ge(NewFloat128(64)) || !x.Mul(NewFloat128(2)).Ge(NewFloat128(float64(n)).Mul(NewFloat128(float64(n)))) {
		return Float128{}, false
	}
	var (
		One = Float128(uvone128)
		Pi  = Float128{0x4000_921f_b544_42d1, 0x8469_898c_c517_01b8}
	)

	// t is a(k)/x**k.
	fn := NewFloat128(float64(n))
	mu := fn.Mul(fn).Mul(NewFloat128(4))
	inv8x := One.Quo(x.Mul(NewFloat128(8)))
	t := One
	p, q := One, Float128{}
	converged := false
	// 2**-125
	const small = uint64(bias128-125) << (shift128 - 64)
	for k := 1; k <= 500; k++ {
		odd := NewFloat128(float64((2*k - 1) * (2*k - 1)))
		next := t.Mul(mu.Sub(odd)).Mul(inv8x).Quo(NewFloat128(float64(k)))
		if next.Abs()[0] < small {
			converged = true
			break
		}
		if next.Abs().Gt(t.Abs()) && next.Abs()[0] > uint64(bias128-100)<<(shift128-64) {
			// The terms start increasing before they are small enough.
			return Float128{}, false
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
		return Float128{}, false
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

// ynForward128 returns Yn(x) for n >= 2, x > 0 by the forward recurrence
//
//	Y(k+1)(x) = 2k/x * Y(k)(x) - Y(k-1)(x)
//
// from Y0(x) and Y1(x). It is stable, but the rounding errors of each step are accumulated, so it is calculated
// in the extended precision; see ynExt128.
func ynForward128(n int, x Float128) Float128 {
	y0, y1 := y0y1_128(x)
	if y0.IsInf(0) || y1.IsInf(0) {
		// x is so small that Y1(x) overflows. |Yn(x)| >= |Y1(x)|, and it is negative.
		return NewFloat128Inf(-1)
	}
	inv := ynExt128Inv(ynExt128FromFloat128(x))
	prev, cur := ynExt128FromFloat128(y0), ynExt128FromFloat128(y1)
	for k := 1; k < n; k++ {
		c := ynExt128Mul(ynExt128FromUint(2*uint64(k)), inv)
		prev, cur = cur, ynExt128Add(ynExt128Mul(c, cur), ynExt128{m: prev.m, e: prev.e, neg: !prev.neg})
		if cur.e > bias128+1 {
			// The result overflows. |Yk(x)| grows monotonically for k > x/2 once it exceeds 1, and it is negative.
			return NewFloat128Inf(-1)
		}
	}
	return cur.Float128()
}

// y0y1_128 returns Y0(x) and Y1(x) for x > 0, sharing the calculation of them.
func y0y1_128(x Float128) (y0, y1 Float128) {
	var (
		Two         = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
		Threshold64 = Float128{0x4005_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	switch {
	case x.Lt(Two):
		return y0y1Series128(x)
	case x.Lt(Threshold64):
		// Y1(x) = -Y0'(x)
		y0, dy := y0Taylor128(x, true)
		return y0, dy.Neg()
	default:
		return y0y1Asymptotic128(x)
	}
}

// y0y1Series128 returns Y0(x) and Y1(x) for 0 < x < 2; see y0Series128 and y1Series128.
func y0y1Series128(x Float128) (y0, y1 Float128) {
	var (
		Two     = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}
		Quarter = Float128{0x3ffd_0000_0000_0000, 0x0000_0000_0000_0000}
		Euler   = Float128{0x3ffe_2788_cfc6_fb61, 0x8f49_a37c_7f02_02a6}
	)

	z := x.Mul(x).Mul(Quarter)
	j0 := y0128SeriesJ[len(y0128SeriesJ)-1]
	s0 := y0128SeriesS[len(y0128SeriesS)-1]
	j1 := y1128SeriesJ[len(y1128SeriesJ)-1]
	s1 := y1128SeriesS[len(y1128SeriesS)-1]
	for k := len(y0128SeriesJ) - 2; k >= 0; k-- {
		j0 = FMA128(j0, z, y0128SeriesJ[k])
		s0 = FMA128(s0, z, y0128SeriesS[k])
	}
	for k := len(y1128SeriesJ) - 2; k >= 0; k-- {
		j1 = FMA128(j1, z, y1128SeriesJ[k])
		s1 = FMA128(s1, z, y1128SeriesS[k])
	}
	l := x.Quo(Two).Log().Add(Euler)
	y0 = FMA128(l, j0, s0).Mul(y0128TwoOverPi)
	y1 = FMA128(FMA128(l, j1, s1), x.Mul(y1128OneOverPi), y1128TwoOverPi.Quo(x).Neg())
	return y0, y1
}

// y0y1Asymptotic128 returns Y0(x) and Y1(x) for x >= 64; see y0Asymptotic128 and y1Asymptotic128.
func y0y1Asymptotic128(x Float128) (y0, y1 Float128) {
	var (
		One = Float128(uvone128)
		Pi  = Float128{0x4000_921f_b544_42d1, 0x8469_898c_c517_01b8}
	)

	w := One.Quo(x.Mul(x))

	terms := len(y0128HankelP) // 64 <= x
	switch xf := x.Float64().BuiltIn(); {
	case xf >= 4096:
		terms = 6
	case xf >= 512:
		terms = 9
	case xf >= 128:
		terms = 13
	}

	p0 := y0128HankelP[terms-1]
	q0 := y0128HankelQ[terms-1]
	p1 := y1128HankelP[terms-1]
	q1 := y1128HankelQ[terms-1]
	for k := terms - 2; k >= 0; k-- {
		p0 = FMA128(p0, w, y0128HankelP[k])
		q0 = FMA128(q0, w, y0128HankelQ[k])
		p1 = FMA128(p1, w, y1128HankelP[k])
		q1 = FMA128(q1, w, y1128HankelQ[k])
	}
	q0 = q0.Quo(x)
	q1 = q1.Quo(x)

	sin, cos := x.Sincos()
	amp := One.Quo(Pi.Mul(x)).Sqrt()
	y0 = amp.Mul(p0.Add(q0).Mul(sin).Add(q0.Sub(p0).Mul(cos)))
	y1 = amp.Mul(q1.Sub(p1).Mul(sin).Sub(p1.Add(q1).Mul(cos)))
	return y0, y1
}
