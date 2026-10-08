package floats

// Asin returns the arcsine, in radians, of a.
//
// Special cases are:
//
//	±0.Asin() = ±0
//	x.Asin() = NaN if x < -1 or x > 1
func (a Float128) Asin() Float128 {
	switch {
	case a.IsZero():
		return a
	case a.IsNaN():
		return NewFloat128NaN()
	}

	sign := a.Signbit()
	if sign {
		a = a.Neg()
	}

	var Dot7 = Float128{0x3ffe_6666_6666_6666, 0x6666_6666_6666_6666} // 0.7
	temp := Float128(uvone128).Sub(a.Mul(a)).Sqrt()
	if a.Gt(Dot7) {
		// asin(x) = pi/2 - atan(sqrt(1-x²)/x)
		var Pi2 = Float128{0x3fff_921f_b544_42d1, 0x8469_898c_c517_01b8}
		temp = Pi2.Sub(satan128(temp.Quo(a)))
	} else {
		// asin(x) = atan(x/sqrt(1-x²))
		temp = satan128(a.Quo(temp))
	}

	if sign {
		temp = temp.Neg()
	}
	return temp
}

// Acos returns the arccosine, in radians, of a.
//
// Special case is:
//
//	x.Acos() = NaN if x < -1 or x > 1
func (a Float128) Acos() Float128 {
	// acos(x) = pi/2 - asin(x)
	var Pi2 = Float128{0x3fff_921f_b544_42d1, 0x8469_898c_c517_01b8}
	return Pi2.Sub(a.Asin())
}

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float128) Atan() Float128 {
	// special cases
	switch {
	case a.IsZero():
		return a
	case a.IsNaN():
		return NewFloat128NaN()
	}
	if a.Signbit() {
		return satan128(a.Neg()).Neg()
	}
	return satan128(a)
}

// satan128 reduces its argument (known to be positive)
// to the range [0, 0.66] and calls xatan.
func satan128(x Float128) Float128 {
	var (
		One = Float128(uvone128)

		// Dot66 = 0.66
		Dot66 = Float128{0x3ffe_51eb_851e_b851, 0xeb85_1eb8_51eb_851f}

		// Tan3pio8 = tan(3*pi/8) = 1 + sqrt(2)
		Tan3pio8 = Float128{0x4000_3504_f333_f9de, 0x6484_597d_89b3_754b}

		// Pi/2 split into two parts
		Pi2Hi = Float128{0x3fff_921f_b544_42d1, 0x8469_898c_c517_01b8}
		Pi2Lo = Float128{0x3f8c_cd12_9024_e088, 0xa67c_c740_20bb_ea64}

		// Pi/4 split into two parts
		Pi4Hi = Float128{0x3ffe_921f_b544_42d1, 0x8469_898c_c517_01b8}
		Pi4Lo = Float128{0x3f8b_cd12_9024_e088, 0xa67c_c740_20bb_ea64}
	)

	switch {
	case x.Le(Dot66):
		return xatan128(x)
	case x.Gt(Tan3pio8):
		// atan(x) = pi/2 - atan(1/x)
		return Pi2Hi.Sub(xatan128(One.Quo(x))).Add(Pi2Lo)
	default:
		// atan(x) = pi/4 + atan((x-1)/(x+1))
		return Pi4Hi.Add(xatan128((x.Sub(One)).Quo(x.Add(One)))).Add(Pi4Lo)
	}
}

// xatan128 returns the arctangent.
// it is valid in the range [0, 0.66].
func xatan128(x Float128) Float128 {
	var y Float128
	for n := 60; n >= 0; n-- {
		term := power128(x, 2*n+1).Quo(NewFloat128(float64(2*n + 1)))
		if n%2 != 0 {
			term = term.Neg()
		}
		y = y.Add(term)
	}
	return y
}

// Atan2 returns the arc tangent of a/b, using
// the signs of the two to determine the quadrant
// of the return value.
//
// Special cases are (in order):
//
//	y.Atan2(NaN) = NaN
//	NaN.Atan2(x) = NaN
//	+0.Atan2(x>=0) = +0
//	-0.Atan2(x>=0) = -0
//	+0.Atan2(x<=-0) = +Pi
//	-0.Atan2(x<=-0) = -Pi
//	y>0.Atan2(0) = +Pi/2
//	y<0.Atan2(0) = -Pi/2
//	+Inf.Atan2(+Inf) = +Pi/4
//	-Inf.Atan2(+Inf) = -Pi/4
//	+Inf.Atan2(-Inf) = 3Pi/4
//	-Inf.Atan2(-Inf) = -3Pi/4
//	y.Atan2(+Inf) = 0
//	(y>0).Atan2(-Inf) = +Pi
//	(y<0).Atan2(-Inf) = -Pi
//	+Inf.Atan2(x) = +Pi/2
//	-Inf.Atan2(x) = -Pi/2
func (a Float128) Atan2(b Float128) Float128 {
	var (
		Zero = Float128{}
		// Pi = Pi
		Pi = Float128{0x4000_921f_b544_42d1, 0x8469_898c_c517_01b8}
		// Pi2 = Pi/2
		Pi2 = Float128{0x3fff_921f_b544_42d1, 0x8469_898c_c517_01b8}
		// Pi4 = Pi/4
		Pi4 = Float128{0x3ffe_921f_b544_42d1, 0x8469_898c_c517_01b8}
		// Pi34 = 3Pi/4
		Pi34 = Float128{0x4000_2d97_c7f3_321d, 0x234f_2729_93d1_414a}
	)

	// special cases
	switch {
	case a.IsNaN() || b.IsNaN():
		return NewFloat128NaN()
	case a.IsZero():
		if b.Ge(Zero) && !b.Signbit() {
			return Zero.Copysign(a)
		}
		return Pi.Copysign(a)
	case b.IsZero():
		return Pi2.Copysign(a)
	case b.IsInf(0):
		if b.IsInf(1) {
			switch {
			case a.IsInf(0):
				return Pi4.Copysign(a)
			default:
				return Zero.Copysign(a)
			}
		}
		switch {
		case a.IsInf(0):
			return Pi34.Copysign(a)
		default:
			return Pi.Copysign(a)
		}
	case a.IsInf(0):
		return Pi2.Copysign(a)
	}

	return atan2Finite128(a, b)
}

// atan2Finite128 returns the arc tangent of y/x for finite and nonzero y and x.
func atan2Finite128(y, x Float128) Float128 {
	// atan2(y, x) = c + s*atan(p/q), where p = min(|y|, |x|), q = max(|y|, |x|).
	p, q := y.Abs(), x.Abs()
	swapped := p.Gt(q)
	if swapped {
		p, q = q, p
	}

	// t = th + tl = atan(p/q), where 0 < p/q <= 1 and |tl| is about ulp(th) or less.
	th, tl := atan128Ratio(p, q)

	var r Float128
	if !swapped && !x.Signbit() {
		// atan2 = t
		r = th.Add(tl)
	} else {
		// atan2 = c + s*t, where |c| >= |t|.
		var chi, clo Float128
		switch {
		case swapped && !x.Signbit():
			// atan2 = pi/2 - t
			chi, clo = atan2Pi2Hi128, atan2Pi2Lo128
			th, tl = th.Neg(), tl.Neg()
		case !swapped:
			// atan2 = pi - t
			chi, clo = atan2Pi2Hi128.Add(atan2Pi2Hi128), atan2Pi2Lo128.Add(atan2Pi2Lo128)
			th, tl = th.Neg(), tl.Neg()
		default:
			// atan2 = pi/2 + t
			chi, clo = atan2Pi2Hi128, atan2Pi2Lo128
		}
		sh := chi.Add(th)
		se := th.Sub(sh.Sub(chi)) // chi + th = sh + se
		r = sh.Add(se.Add(clo.Add(tl)))
	}
	if y.Signbit() {
		return r.Neg()
	}
	return r
}

// atan128Ratio returns atan(p/q) as a pair (hi, lo) such that atan(p/q) = hi + lo
// and |lo| is about ulp(hi) or less, for finite p and q with 0 < p <= q.
func atan128Ratio(p, q Float128) (hi, lo Float128) {
	r := p.Quo(q)

	// the table index: the nearest i/64 to r.
	i := int(r.Float64().BuiltIn()*64 + 0.5)
	if i == 0 {
		// atan(r) = r - r**3/3 + ..., where r < 1/128.
		ul := FMA128(r.Neg(), q, p).Quo(q) // p/q = r + ul
		return r, atan128Tail(r, ul)
	}

	// scale p and q so that 1 <= q < 2, to avoid overflow and underflow.
	e := q.Ilogb()
	p, q = p.Ldexp(-e), q.Ldexp(-e)

	// u = (p - c*q) / (q + c*p), where c = i/64.
	// |u| <= 1/128, and atan(p/q) = atan(c) + atan(u).
	c := NewFloat128(float64(i) / 64)
	cqh, cql := twoProduct128(c, q)
	cph, cpl := twoProduct128(c, p)
	nh, nl := twoSum128(p.Sub(cqh), cql.Neg()) // p - cqh is exact
	dh := q.Add(cph)
	dl := cph.Sub(dh.Sub(q)).Add(cpl) // q + cph = dh + (cph - (dh - q)), because q >= cph

	// u = uh + ul = n/d.
	uh := nh.Quo(dh)
	rem := FMA128(uh.Neg(), dh, nh).Add(nl).Sub(uh.Mul(dl))
	ul := rem.Quo(dh)

	// atan(c) + u + tail, where atan(c) > |u|.
	th := atan2TableHi128[i]
	hi = th.Add(uh)
	lo = uh.Sub(hi.Sub(th)).Add(atan2TableLo128[i].Add(atan128Tail(uh, ul)))
	return hi, lo
}

// atan128Tail returns ul + (atan(u) - u) for u = uh + ul, |u| <= 1/127.
func atan128Tail(uh, ul Float128) Float128 {
	z := uh.Mul(uh)
	// (atan(u) - u)/u**3 = c[0] + c[1]*z + ... + c[8]*z**8
	s := atan2Coeffs128[8]
	for k := 7; k >= 0; k-- {
		s = FMA128(s, z, atan2Coeffs128[k])
	}
	return ul.Add(uh.Mul(z).Mul(s))
}
