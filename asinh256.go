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

	// asinh(a) = log(u), where u = |a| + sqrt(a² + 1) = 2**k × v and v is in [1, 2).
	// vf is v × 2**319.
	var k int
	var vf ints.Uint512
	switch {
	case exp >= 128:
		// u = 2|a| × (1 + 1/(4a²) + ...) and the relative error of the approximation u ~ 2|a| is less than 2**-258.
		k = exp + 1
		vf = m.Uint512().Lsh(83)
	case exp >= 0:
		// sqrt(a² + 1) × 2**g = sqrt(m² × 2**(2exp-472+2f) + 2**(2f)), where f = 254 - exp.
		// the integer arithmetic is exact, and the root has 255 or 256 bits.
		n := m.Mul512(m).Lsh(36).Add(ints.Uint512{7: 1}.Lsh(uint(508 - 2*exp)))
		s := uint(n.LeadingZeros() &^ 1) // sqrtRem512 requires n[0] >= 1<<62.
		root, _ := sqrtRem512(n.Lsh(s))
		g := 254 - exp + int(s/2)
		u := m.Uint512().Lsh(uint(18 + s/2)).Add(root.Uint512()) // u × 2**g
		bl := u.BitLen()
		k = bl - 1 - g
		vf = u.Lsh(uint(320 - bl))
	default:
		// u = 1 + t, where t = |a| + a²/(1 + sqrt(a² + 1)) in fixed point with 492 fractional bits.
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

		if t.BitLen() <= 484 {
			// t < 2**-8. log(1+t) is computed without the reduction.
			q, lz := log256Log1p(false, t)
			return fixToFloat256(sign, q, false, -491-lz)
		}
		u := t.Add(ints.Uint512{0: 1 << 44}) // u × 2**492
		bl := u.BitLen()
		k = bl - 493
		vf = u.Rsh(uint(bl - 320))
	}

	idx, rneg, rmag := log256ReduceWide(vf)
	_, v, e := log256Combine(k, idx, rneg, rmag)
	return fixToFloat256(sign, v, false, e)
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

// Acosh returns the inverse hyperbolic cosine of a.
//
// Special cases are:
//
//	+Inf.Acosh() = +Inf
//	x.Acosh() = NaN if x < 1
//	NaN.Acosh() = NaN
func (a Float256) Acosh() Float256 {
	var (
		// Ln2 = ln(2)
		Ln2 = Float256{
			0x3fff_e62e_42fe_fa39, 0xef35_793c_7673_007e,
			0x5ed5_e81e_6864_ce53, 0x16c5_b141_a2eb_7175,
		}
		// Zero = 0.0
		Zero = Float256{}
		// One = 1.0
		One = Float256(uvone256)
		// Two = 2.0
		Two = Float256{
			0x4000_0000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
		// Large = 2**170
		Large = Float256{
			0x400a_9000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)
	switch {
	case a.Lt(One) || a.IsNaN():
		return NewFloat256NaN()
	case a.Eq(One):
		return Zero
	case a.Ge(Large):
		return a.Log().Add(Ln2) // a >= 2**170
	case a.Gt(Two):
		return (a.Add((a.Mul(a).Sub(One)).Sqrt())).Log() // 2**170 > a > 2.0
	}
	t := a.Sub(One)
	return (t.Add((t.Mul(t).Add(Two.Mul(t))).Sqrt())).Log1p() // 2 >= a > 1
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
	var (
		// Zero = 0.0
		Zero = Float256{}

		// Half = 0.5
		Half = Float256{
			0x3fff_e000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// One = 1.0
		One = Float256(uvone256)

		// NearZero = 2**-170
		NearZero = Float256{
			0x3ff5_5000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)

	// special cases
	switch {
	case a.Lt(One.Neg()) || a.Gt(One) || a.IsNaN():
		return NewFloat256NaN()
	case a.Eq(One):
		return NewFloat256Inf(1)
	case a.Eq(One.Neg()):
		return NewFloat256Inf(-1)
	}
	sign := false
	if a.Lt(Zero) {
		a = a.Neg()
		sign = true
	}
	var temp Float256
	switch {
	case a.Lt(NearZero):
		temp = a
	case a.Lt(Half):
		temp = a.Add(a)
		temp = Half.Mul(temp.Add(temp.Mul(a).Quo(One.Sub(a))).Log1p())
	default:
		temp = Half.Mul((a.Add(a).Quo(One.Sub(a))).Log1p())
	}
	if sign {
		temp = temp.Neg()
	}
	return temp
}
