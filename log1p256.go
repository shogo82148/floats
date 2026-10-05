package floats

import "github.com/shogo82148/ints"

// Log1p returns the natural logarithm of 1 plus its argument a.
// It is more accurate than [Log](1 + a) when a is near zero.
//
// Special cases are:
//
//	+Inf.Log1p() = +Inf
//	±0.Log1p() = ±0
//	-1.Log1p() = -Inf
//	(a < -1).Log1p() = NaN
//	NaN.Log1p() = NaN
func (a Float256) Log1p() Float256 {
	// special cases
	switch {
	case a.Lt(Float256(uvone256).Neg()) || a.IsNaN(): // includes -Inf
		return NewFloat256NaN()
	case a.Eq(Float256(uvone256).Neg()):
		return NewFloat256Inf(-1)
	case a.IsInf(1):
		return NewFloat256Inf(1)
	case a.IsZero():
		return a
	}

	// a = ±m × 2**(exp-236), where m is a 237-bit integer in [2**236, 2**237).
	sign, exp, m := a.normalize()
	if exp < -239 {
		// log(1+a) = a × (1 - a/2 + ...) rounds to a when |a| < 2**-239.
		return a
	}

	if exp < -8 {
		// |a| < 2**-8. log(1+a) = a × G(w) = a - a² × H(w), where w = -a,
		// G(w) = 1 + w/2 + w²/3 + ... = 1 + w × H(w), and H(w) = 1/2 + w/3 + w²/4 + ...
		// a² × H is computed with the relative precision and subtracted from (or, if a < 0, added to) a
		// without rounding, so that the result is correct even if a² × H is near a half ulp of a.

		// |w| in fixed point with 264 fractional bits.
		var w ints.Uint256
		if s := exp + 28; s >= 0 {
			w = m.Lsh(uint(s))
		} else {
			w = m.Rsh(uint(-s))
		}

		// H in fixed point with 255 fractional bits.
		h := ints.Uint256(log256Coeffs[0])
		for _, c := range log256Coeffs[1 : len(log256Coeffs)-1] { // 1/32, ..., 1/2
			t := shr512to256(h.Mul512(w), 264)
			if sign == 0 {
				// w < 0
				h = ints.Uint256(c).Sub(t)
			} else {
				h = ints.Uint256(c).Add(t)
			}
		}

		// a² × H = m² × H × 2**(2exp-472-255). Keep the top 256 bits of m²,
		// which are accurate enough since H has only 255 bits.
		m2 := shr512to256(m.Mul512(m), 218)
		c := m2.Mul512(h) // a² × H = c × 2**(2exp-509)

		// |a| = m × 2**(exp-236) = (m × 2**274) × 2**(exp-510)
		// a² × H = c × 2**(2exp-509) = (c >> (-1-exp)) × 2**(exp-510)
		r := ints.Uint512{4: m[0], 5: m[1], 6: m[2], 7: m[3]}.Lsh(274)
		if s := uint(-1 - exp); s < 512 {
			c = c.Rsh(s)
		} else {
			c = ints.Uint512{}
		}
		if sign == 0 {
			r = r.Sub(c)
		} else {
			r = r.Add(c)
		}
		return fixToFloat256(sign, r, false, exp-510)
	}

	// |a| >= 2**-8, so the result is not small. log(1+a) = log(hi) + lo/hi,
	// where hi + lo = 1 + a exactly, and |lo/hi| < 2**-235 so that log1p(lo/hi) ~ lo/hi.
	hi, lo := twoSum256(Float256(uvone256), a)
	lsign, v, vexp := log256Fix(hi)
	if lo.IsZero() {
		return fixToFloat256(lsign, v, false, vexp)
	}

	// log(hi) is a fixed-point value with 320 fractional bits, since hi is not close to 1.
	// add lo/hi to it in the same fixed-point.
	dsign, dexp, dm := lo.Quo(hi).normalize()
	var d ints.Uint512 // |lo/hi| × 2**320
	if s := dexp - 236 + 320; s >= 0 {
		d = ints.Uint512{4: dm[0], 5: dm[1], 6: dm[2], 7: dm[3]}.Lsh(uint(s))
	} else if s > -256 {
		d = ints.Uint512{4: dm[0], 5: dm[1], 6: dm[2], 7: dm[3]}.Rsh(uint(-s))
	}
	if lsign != 0 {
		v = v.Neg()
	}
	if dsign != 0 {
		v = v.Sub(d)
	} else {
		v = v.Add(d)
	}
	var sign2 uint64
	if v[0]>>63 != 0 {
		sign2 = signMask256[0]
		v = v.Neg()
	}
	return fixToFloat256(sign2, v, false, -320)
}

// twoSum256 returns hi, lo such that hi+lo = a+b exactly (as real numbers),
// with hi = a+b rounded to the nearest Float256.
func twoSum256(a, b Float256) (hi, lo Float256) {
	hi = a.Add(b)
	v := hi.Sub(a)
	lo = a.Sub(hi.Sub(v)).Add(b.Sub(v))
	return
}
