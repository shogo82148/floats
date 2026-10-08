package floats

// Asin returns the arcsine, in radians, of a.
//
// Special cases are:
//
//	±0.Asin() = ±0
//	x.Asin() = NaN if x < -1 or x > 1
func (a Float256) Asin() Float256 {
	one := Float256(uvone256)
	ax := a.Abs()
	switch {
	case a.IsZero():
		return a
	case a.IsNaN() || ax.Gt(one):
		return NewFloat256NaN()
	case ax == one:
		return atan2Pi2Hi256.Copysign(a)
	}

	// asin(x) = atan2(x, c), where c = sqrt(1-x**2) = sqrt((1-x)(1+x)).
	// c is computed in double-Float256 (ch + cl) to keep the relative accuracy.
	xh, xl := twoSum256(one, ax.Neg()) // 1-x
	yh, yl := twoSum256(one, ax)       // 1+x
	mh, ml := twoProduct256(xh, yh)
	ml = ml.Add(xh.Mul(yl).Add(xl.Mul(yh))) // 1-x**2 = mh + ml
	ch := mh.Sqrt()
	cl := FMA256(ch.Neg(), ch, mh).Add(ml).Quo(ch.Add(ch))

	// atan2(x, c) = atan(p/q) or pi/2 - atan(p/q), where p = min(x, c), q = max(x, c).
	var r Float256
	if ax.Gt(ch) {
		th, tl := atan256RatioDD(ch, cl, ax, Float256{})
		th, tl = th.Neg(), tl.Neg()
		sh := atan2Pi2Hi256.Add(th)
		se := th.Sub(sh.Sub(atan2Pi2Hi256)) // pi/2 + th = sh + se
		r = sh.Add(se.Add(atan2Pi2Lo256.Add(tl)))
	} else {
		th, tl := atan256RatioDD(ax, Float256{}, ch, cl)
		r = th.Add(tl)
	}
	return r.Copysign(a)
}

// Acos returns the arccosine, in radians, of a.
//
// Special case is:
//
//	x.Acos() = NaN if x < -1 or x > 1
func (a Float256) Acos() Float256 {
	// acos(x) = pi/2 - asin(x)
	var Pi2 = Float256{
		0x3fff_f921_fb54_442d, 0x1846_9898_cc51_701b,
		0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
	}
	return Pi2.Sub(a.Asin())
}

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float256) Atan() Float256 {
	// special cases
	switch {
	case a.IsZero():
		return a
	case a.IsNaN():
		return NewFloat256NaN()
	case a.IsInf(0):
		return atan2Pi2Hi256.Copysign(a)
	}
	// atan(a) = atan2(a, 1)
	return atan2Finite256(a, Float256(uvone256))
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
func (a Float256) Atan2(b Float256) Float256 {
	var (
		Zero = Float256{}
		// Pi = Pi
		Pi = Float256{
			0x4000_0921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
		// Pi2 = Pi/2
		Pi2 = Float256{
			0x3fff_f921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
		// Pi4 = Pi/4
		Pi4 = Float256{
			0x3fff_e921_fb54_442d, 0x1846_9898_cc51_701b,
			0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
		}
		// Pi34 = 3Pi/4
		Pi34 = Float256{
			0x4000_02d9_7c7f_3321, 0xd234_f272_993d_1414,
			0xa2b3_9bd8_3750_ccf9, 0xbb2a_e031_19df_958a,
		}
	)

	// special cases
	switch {
	case a.IsNaN() || b.IsNaN():
		return NewFloat256NaN()
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

	return atan2Finite256(a, b)
}

// twoProduct256 returns hi, lo such that hi+lo = a*b exactly (as real numbers),
// with hi = a*b rounded to the nearest Float256.
func twoProduct256(a, b Float256) (hi, lo Float256) {
	hi = a.Mul(b)
	lo = FMA256(a, b, hi.Neg())
	return
}

// atan2Finite256 returns the arc tangent of y/x for finite and nonzero y and x.
func atan2Finite256(y, x Float256) Float256 {
	// atan2(y, x) = c + s*atan(p/q), where p = min(|y|, |x|), q = max(|y|, |x|).
	p, q := y.Abs(), x.Abs()
	swapped := p.Gt(q)
	if swapped {
		p, q = q, p
	}

	// t = th + tl = atan(p/q), where 0 < p/q <= 1 and |tl| is about ulp(th) or less.
	th, tl := atan256Ratio(p, q)

	var r Float256
	if !swapped && !x.Signbit() {
		// atan2 = t
		r = th.Add(tl)
	} else {
		// atan2 = c + s*t, where |c| >= |t|.
		var chi, clo Float256
		switch {
		case swapped && !x.Signbit():
			// atan2 = pi/2 - t
			chi, clo = atan2Pi2Hi256, atan2Pi2Lo256
			th, tl = th.Neg(), tl.Neg()
		case !swapped:
			// atan2 = pi - t
			chi, clo = atan2Pi2Hi256.Add(atan2Pi2Hi256), atan2Pi2Lo256.Add(atan2Pi2Lo256)
			th, tl = th.Neg(), tl.Neg()
		default:
			// atan2 = pi/2 + t
			chi, clo = atan2Pi2Hi256, atan2Pi2Lo256
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

// atan256Ratio returns atan(p/q) as a pair (hi, lo) such that atan(p/q) = hi + lo
// and |lo| is about ulp(hi) or less, for finite p and q with 0 < p <= q.
func atan256Ratio(p, q Float256) (hi, lo Float256) {
	r := p.Quo(q)

	// the table index: the nearest i/256 to r.
	i := int(r.Float64().BuiltIn()*256 + 0.5)
	if i == 0 {
		// atan(r) = r - r**3/3 + ..., where r < 1/512.
		ul := FMA256(r.Neg(), q, p).Quo(q) // p/q = r + ul
		return r, atan256Tail(r, ul)
	}

	// scale p and q so that 1 <= q < 2, to avoid overflow and underflow.
	e := q.Ilogb()
	p, q = p.Ldexp(-e), q.Ldexp(-e)

	// u = (p - c*q) / (q + c*p), where c = i/256.
	// |u| <= 1/512, and atan(p/q) = atan(c) + atan(u).
	c := NewFloat256(float64(i) / 256)
	cqh, cql := twoProduct256(c, q)
	cph, cpl := twoProduct256(c, p)
	nh, nl := twoSum256(p.Sub(cqh), cql.Neg()) // p - cqh is exact
	dh := q.Add(cph)
	dl := cph.Sub(dh.Sub(q)).Add(cpl) // q + cph = dh + (cph - (dh - q)), because q >= cph

	// u = uh + ul = n/d.
	uh := nh.Quo(dh)
	rem := FMA256(uh.Neg(), dh, nh).Add(nl).Sub(uh.Mul(dl))
	ul := rem.Quo(dh)

	// atan(c) + u + tail, where atan(c) > |u|.
	th := atan2Table256Hi[i]
	hi = th.Add(uh)
	lo = uh.Sub(hi.Sub(th)).Add(atan2Table256Lo[i].Add(atan256Tail(uh, ul)))
	return hi, lo
}

// atan256RatioDD returns atan(p/q) as a pair (hi, lo) such that atan(p/q) = hi + lo
// and |lo| is about ulp(hi) or less, for p = ph + pl and q = qh + ql with 0 < p <= q.
// The arguments are not scaled, so they must be in the range where the products
// of them do not overflow or underflow, such as [2**-1000, 1].
func atan256RatioDD(ph, pl, qh, ql Float256) (hi, lo Float256) {
	r := ph.Quo(qh)

	// the table index: the nearest i/256 to r.
	i := int(r.Float64().BuiltIn()*256 + 0.5)
	if i == 0 {
		// atan(r) = r - r**3/3 + ..., where r < 1/512.
		rem := FMA256(r.Neg(), qh, ph).Add(pl).Sub(r.Mul(ql)) // p - r*q
		return r, atan256Tail(r, rem.Quo(qh))
	}

	// u = (p - c*q) / (q + c*p), where c = i/256.
	// |u| <= 1/512, and atan(p/q) = atan(c) + atan(u).
	c := NewFloat256(float64(i) / 256)
	cqh, cql := twoProduct256(c, qh)
	cph, cpl := twoProduct256(c, ph)
	nh, nl := twoSum256(ph.Sub(cqh), pl.Sub(cql).Sub(c.Mul(ql))) // ph - cqh is exact
	dh := qh.Add(cph)
	dl := cph.Sub(dh.Sub(qh)).Add(ql.Add(cpl).Add(c.Mul(pl))) // qh + cph = dh + (cph - (dh - qh)), because qh >= cph

	// u = uh + ul = n/d.
	uh := nh.Quo(dh)
	rem := FMA256(uh.Neg(), dh, nh).Add(nl).Sub(uh.Mul(dl))
	ul := rem.Quo(dh)

	// atan(c) + u + tail, where atan(c) > |u|.
	th := atan2Table256Hi[i]
	hi = th.Add(uh)
	lo = uh.Sub(hi.Sub(th)).Add(atan2Table256Lo[i].Add(atan256Tail(uh, ul)))
	return hi, lo
}

// atan256Tail returns ul + (atan(u) - u) for u = uh + ul, |u| <= 1/511.
func atan256Tail(uh, ul Float256) Float256 {
	z := uh.Mul(uh)
	// (atan(u) - u)/u**3 = c[0] + c[1]*z + ... + c[13]*z**13
	n := len(atan2Coeffs256)
	s := atan2Coeffs256[n-1]
	for k := n - 2; k >= 0; k-- {
		s = FMA256(s, z, atan2Coeffs256[k])
	}
	return ul.Add(uh.Mul(z).Mul(s))
}
