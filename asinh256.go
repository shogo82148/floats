package floats

import "github.com/shogo82148/ints"

// Asinh returns the inverse hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Asinh() = ±0
//	±Inf.Asinh() = ±Inf
//	NaN.Asinh() = NaN
func (a Float256) Asinh() Float256 {
	if a.IsNaN() || a.IsInf(0) {
		return a
	}

	// |a| = m × 2**(exp-236), where m is a 237-bit integer in [2**236, 2**237).
	sign, exp, m := a.normalize()
	if exp < -118 {
		// |a| < 2**-118, and |asinh(a) - a| ~ |a|³/6 is less than the half ulp of a.
		// This also handles ±0 and subnormal numbers.
		return a
	}
	if exp >= 0 {
		return logHyp256(sign, exp, m, false)
	}

	// asinh(a) = log(1+t), where t = |a| + a²/(1 + sqrt(a² + 1)) in fixed point with 492 fractional bits.
	// sqrt(a² + 1) × 2**255 = sqrt(m² × 2**(2exp+38) + 2**510) is in [2**255, 2**256).
	m2 := m.Mul512(m)
	var n ints.Uint512
	if sh := 2*exp + 38; sh >= 0 {
		n = m2.Lsh(uint(sh))
	} else {
		n = m2.Rsh(uint(-sh))
	}
	root, _ := sqrtRem512(n.Add(ints.Uint512{0: 1 << 62}))

	// a²/(1 + sqrt(a² + 1)) = M2 × 2**(2exp-254) / (d × 2**-255), where d = 2**255 + root, and M2 is the top 256 bits of m².
	d := ints.Uint512{4: 1 << 63}.Add(root.Uint512())
	q := shr512to256(m2, 218).Uint512().Lsh(255).Quo(d)
	t := m.Uint512().Lsh(uint(exp + 256)).Add(q.Lsh(uint(2*exp + 238)))
	return log1pFix256(sign, t)
}

// Acosh returns the inverse hyperbolic cosine of a.
//
// Special cases are:
//
//	+Inf.Acosh() = +Inf
//	x.Acosh() = NaN if x < 1
//	NaN.Acosh() = NaN
func (a Float256) Acosh() Float256 {
	if a.IsNaN() {
		return a
	}

	// a = m × 2**(exp-236), where m is a 237-bit integer in [2**236, 2**237).
	sign, exp, m := a.normalize()
	switch {
	case sign != 0 || exp < 0: // a < 1
		return NewFloat256NaN()
	case a.IsInf(1):
		return a
	case exp > 0:
		return logHyp256(0, exp, m, true)
	}

	// 1 <= a < 2. acosh(a) = log(1+t), where t = (a-1) + sqrt(a²-1) in fixed point with 492 fractional bits.
	// a-1 = d × 2**-236 and a²-1 = (a-1)(a+1) = d × (2**237 + d) × 2**-472 are exact integers.
	d := m.Sub(ints.Uint256{0: 1 << 44}) // m - 2**236
	if d.IsZero() {
		return Float256{}
	}
	n := d.Mul512(ints.Uint256{0: 1 << 45}.Add(d))
	lz := uint(n.LeadingZeros() &^ 1) // sqrtRem512 requires n[0] >= 1<<62.
	root, _ := sqrtRem512(n.Lsh(lz))
	// sqrt(a²-1) × 2**492 = root × 2**(256-lz/2)
	t := d.Uint512().Lsh(256).Add(root.Uint512().Lsh(256 - lz/2))
	return log1pFix256(0, t)
}

// logHyp256 returns ±log(a + sqrt(a² ± 1)), i.e. ±asinh(a) if acosh is false and ±acosh(a) otherwise,
// for a = m × 2**(exp-236), where m is a 237-bit integer in [2**236, 2**237).
// sign is the sign bit of the result. exp must be at least 0, and at least 1 if acosh is true.
func logHyp256(sign uint64, exp int, m ints.Uint256, acosh bool) Float256 {
	// u = a + sqrt(a² ± 1) = 2**k × v, where v is in [1, 2). vf is v × 2**319.
	var k int
	var vf ints.Uint512
	if exp >= 128 {
		// u = 2a × (1 ± 1/(4a²) + ...) and the relative error of the approximation u ~ 2a is less than 2**-258.
		k = exp + 1
		vf = m.Uint512().Lsh(83)
	} else {
		// sqrt(a² ± 1) × 2**g = sqrt(m² × 2**(2exp-472+2f) ± 2**(2f)), where f = 254 - exp.
		// the integer arithmetic is exact, and the root has 255 or 256 bits.
		n := m.Mul512(m).Lsh(36)
		one := ints.Uint512{7: 1}.Lsh(uint(508 - 2*exp))
		if acosh {
			n = n.Sub(one)
		} else {
			n = n.Add(one)
		}
		s := uint(n.LeadingZeros() &^ 1) // sqrtRem512 requires n[0] >= 1<<62.
		root, _ := sqrtRem512(n.Lsh(s))
		g := 254 - exp + int(s/2)
		u := m.Uint512().Lsh(uint(18 + s/2)).Add(root.Uint512()) // u × 2**g
		bl := u.BitLen()
		k = bl - 1 - g
		vf = u.Lsh(uint(320 - bl))
	}
	return log256Wide(sign, k, vf, 0)
}

// log1pFix256 returns ±log(1+t) for t = |t| × 2**-492 > 0 in fixed point, where |t| < 2**493.
// sign is the sign bit of the result.
func log1pFix256(sign uint64, t ints.Uint512) Float256 {
	if t.BitLen() <= 484 {
		// t < 2**-8. log(1+t) is computed without the reduction to keep the relative precision.
		q, lz := log256Log1p(false, t)
		return fixToFloat256(sign, q, false, -491-lz)
	}
	u := t.Add(ints.Uint512{0: 1 << 44}) // u × 2**492
	bl := u.BitLen()
	return log256Wide(sign, bl-493, u.Rsh(uint(bl-320)), 0)
}

// log256Wide returns ±log(u) × 2**e, where u = 2**k × v, v is in [1, 2), and vf is v × 2**319.
// u must not be close to 1: log(u) >= 2**-9. sign is the sign bit of the result.
func log256Wide(sign uint64, k int, vf ints.Uint512, e int) Float256 {
	idx, rneg, rmag := log256ReduceWide(vf)
	_, v, exp := log256Combine(k, idx, rneg, rmag)
	return fixToFloat256(sign, v, false, exp+e)
}

// log256ReduceWide reduces v = 1 + f to v = c × (1+r), where c is the breakpoint of the bucket idx,
// and returns r = ±rmag × 2**-492, |r| < 2**-8. vf is v × 2**319, where v is in [1, 2).
// Unlike log256Reduce, the 320 bits of v are used, but the reduction is truncated at 2**-492.
func log256ReduceWide(vf ints.Uint512) (idx uint64, rneg bool, rmag ints.Uint512) {
	idx = (vf[3] >> 55) & 0xff
	if idx == 0 {
		vf[3] &^= 1 << 63 // r = v - 1
		return idx, false, vf.Lsh(173)
	}

	// p = v × (1/c) in fixed point with 638 fractional bits.
	w := ints.Uint256(log256InvC[idx]).Uint512().Lsh(63)
	p := vf.Mul1024(w)
	one := ints.Uint1024{6: 1 << 62} // 2**638
	var d ints.Uint1024
	rneg = p.Cmp(one) < 0
	if rneg {
		d = one.Sub(p)
	} else {
		d = p.Sub(one)
	}
	d = d.Rsh(146)
	return idx, rneg, ints.Uint512(d[8:])
}

// Atanh returns the inverse hyperbolic tangent of a.
//
// Special cases are:
//
//	1.Atanh() = +Inf
//	±0.Atanh() = ±0
//	-1.Atanh() = -Inf
//	x.Atanh() = NaN if x < -1 or x > 1
//	NaN.Atanh() = NaN
func (a Float256) Atanh() Float256 {
	// |a| = m × 2**(exp-236), where m is a 237-bit integer in [2**236, 2**237).
	sign, exp, m := a.normalize()
	switch {
	case exp > 0 || (exp == 0 && m != ints.Uint256{0: 1 << 44}):
		return NewFloat256NaN() // NaN or |a| > 1
	case exp == 0:
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]} // ±1
	case exp < -119:
		// |a| < 2**-119, and |atanh(a) - a| ~ |a|³/3 is less than the half ulp of a.
		// This also handles ±0 and subnormal numbers.
		return a
	}

	if exp < -9 {
		// |a| < 2**-9. atanh(a) = a + a × z × H(z), where z = a² < 2**-18 and H(z) = 1/3 + z/5 + z²/7 + ...
		// a × z × H is computed with the relative precision and added to a without rounding.
		// z = M2 × 2**(2exp-254), where M2 is the top 256 bits of m².
		m2 := shr512to256(m.Mul512(m), 218)

		// z in fixed point with 264 fractional bits, and H in fixed point with 257 fractional bits.
		z := m2.Rsh(uint(-2*exp - 10))
		h := ints.Uint256(atanhCoeffs256[0])
		for _, c := range atanhCoeffs256[1:] {
			h = ints.Uint256(c).Add(shr512to256(z.Mul512(h), 264))
		}

		// z × H = p × 2**(2exp-511). keep the top 256 bits of p: z × H = p1 × 2**(2exp-255).
		p1 := shr512to256(m2.Mul512(h), 256)
		c := m.Mul512(p1) // a × z × H = c × 2**(3exp-491)

		// a = (m × 2**274) × 2**(exp-510) and a × z × H = (c >> (-2exp-19)) × 2**(exp-510)
		s := ints.Uint512{4: m[0], 5: m[1], 6: m[2], 7: m[3]}.Lsh(274).Add(c.Rsh(uint(-2*exp - 19)))
		return fixToFloat256(sign, s, false, exp-510)
	}

	// atanh(a) = log(u)/2, where u = (1+|a|)/(1-|a|) = (2**f + m)/(2**f - m) with f = 236 - exp.
	// u × 2**265 is computed by the integer division, and the relative error is less than 2**-265
	// even if a is close to 1, since u >= 1.
	one := ints.Uint512{7: 1}.Lsh(uint(236 - exp))
	mm := m.Uint512()
	q := one.Add(mm).Lsh(265).Quo(one.Sub(mm))
	bl := q.BitLen()
	var vf ints.Uint512
	if bl >= 320 {
		vf = q.Rsh(uint(bl - 320))
	} else {
		vf = q.Lsh(uint(320 - bl))
	}
	return log256Wide(sign, bl-266, vf, -1)
}
