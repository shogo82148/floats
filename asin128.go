package floats

// Asin returns the arcsine, in radians, of a.
//
// Special cases are:
//
//	±0.Asin() = ±0
//	x.Asin() = NaN if x < -1 or x > 1
func (a Float128) Asin() Float128 {
	one := Float128(uvone128)
	ax := a.Abs()
	switch {
	case a.IsZero():
		return a
	case a.IsNaN() || ax.Gt(one):
		return NewFloat128NaN()
	case ax == one:
		return atan2Pi2Hi128.Copysign(a)
	}

	// asin(x) = atan2(x, c), where c = sqrt(1-x**2) = sqrt((1-x)(1+x)).
	ch, cl := sqrt1mx2128(ax)

	// atan2(x, c) = atan(p/q) or pi/2 - atan(p/q), where p = min(x, c), q = max(x, c).
	var r Float128
	if ax.Gt(ch) {
		th, tl := atan128RatioDD(ch, cl, ax, Float128{})
		th, tl = th.Neg(), tl.Neg()
		sh := atan2Pi2Hi128.Add(th)
		se := th.Sub(sh.Sub(atan2Pi2Hi128)) // pi/2 + th = sh + se
		r = sh.Add(se.Add(atan2Pi2Lo128.Add(tl)))
	} else {
		th, tl := atan128RatioDD(ax, Float128{}, ch, cl)
		r = th.Add(tl)
	}
	return r.Copysign(a)
}

// sqrt1mx2128 returns sqrt(1-x**2) as a double-Float128 (hi + lo) for 0 <= x < 1.
// It keeps the relative accuracy near 0 and near 1.
func sqrt1mx2128(x Float128) (hi, lo Float128) {
	one := Float128(uvone128)
	xh, xl := twoSum128(one, x.Neg()) // 1-x
	yh, yl := twoSum128(one, x)       // 1+x
	mh, ml := twoProduct128(xh, yh)
	ml = ml.Add(xh.Mul(yl).Add(xl.Mul(yh))) // 1-x**2 = mh + ml
	hi = mh.Sqrt()
	lo = FMA128(hi.Neg(), hi, mh).Add(ml).Quo(hi.Add(hi))
	return
}

// Acos returns the arccosine, in radians, of a.
//
// Special cases are:
//
//	1.Acos() = +0
//	x.Acos() = NaN if x < -1 or x > 1
func (a Float128) Acos() Float128 {
	one := Float128(uvone128)
	ax := a.Abs()
	switch {
	case a.IsNaN() || ax.Gt(one):
		return NewFloat128NaN()
	case ax == one:
		if a.Signbit() {
			return atan2Pi2Hi128.Add(atan2Pi2Hi128) // pi
		}
		return Float128{}
	case a.IsZero():
		return atan2Pi2Hi128
	}

	// acos(x) = atan2(c, x), where c = sqrt(1-x**2) is computed in double-Float128 (ch + cl).
	ch, cl := sqrt1mx2128(ax)

	// acos(|x|) = atan(c/|x|) if c <= |x|, otherwise pi/2 - atan(|x|/c).
	// acos(x) = pi - acos(-x) for x < 0.
	var th, tl Float128
	small := !ch.Gt(ax)
	if small {
		th, tl = atan128RatioDD(ch, cl, ax, Float128{})
	} else {
		th, tl = atan128RatioDD(ax, Float128{}, ch, cl)
	}

	// acos(x) = c + s*t, where c is 0, pi/2 or pi.
	var chi, clo Float128
	switch {
	case small && !a.Signbit():
		return th.Add(tl)
	case small:
		// pi - t
		chi, clo = atan2Pi2Hi128.Add(atan2Pi2Hi128), atan2Pi2Lo128.Add(atan2Pi2Lo128)
		th, tl = th.Neg(), tl.Neg()
	case !a.Signbit():
		// pi/2 - t
		chi, clo = atan2Pi2Hi128, atan2Pi2Lo128
		th, tl = th.Neg(), tl.Neg()
	default:
		// pi/2 + t
		chi, clo = atan2Pi2Hi128, atan2Pi2Lo128
	}
	sh := chi.Add(th)
	se := th.Sub(sh.Sub(chi)) // chi + th = sh + se
	return sh.Add(se.Add(clo.Add(tl)))
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
	case a.IsInf(0):
		return atan2Pi2Hi128.Copysign(a)
	}
	// atan(a) = atan2(a, 1)
	return atan2Finite128(a, Float128(uvone128))
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

// atan128RatioDD returns atan(p/q) as a pair (hi, lo) such that atan(p/q) = hi + lo
// and |lo| is about ulp(hi) or less, for p = ph + pl and q = qh + ql with 0 < p <= q.
// The arguments are not scaled, so they must be in the range where the products
// of them do not overflow or underflow, such as [2**-1000, 1].
func atan128RatioDD(ph, pl, qh, ql Float128) (hi, lo Float128) {
	r := ph.Quo(qh)

	// the table index: the nearest i/64 to r.
	i := int(r.Float64().BuiltIn()*64 + 0.5)
	if i == 0 {
		// atan(r) = r - r**3/3 + ..., where r < 1/128.
		rem := FMA128(r.Neg(), qh, ph).Add(pl).Sub(r.Mul(ql)) // p - r*q
		return r, atan128Tail(r, rem.Quo(qh))
	}

	// u = (p - c*q) / (q + c*p), where c = i/64.
	// |u| <= 1/128, and atan(p/q) = atan(c) + atan(u).
	c := NewFloat128(float64(i) / 64)
	cqh, cql := twoProduct128(c, qh)
	cph, cpl := twoProduct128(c, ph)
	nh, nl := twoSum128(ph.Sub(cqh), pl.Sub(cql).Sub(c.Mul(ql))) // ph - cqh is exact
	dh := qh.Add(cph)
	dl := cph.Sub(dh.Sub(qh)).Add(ql.Add(cpl).Add(c.Mul(pl))) // qh + cph = dh + (cph - (dh - qh)), because qh >= cph

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
