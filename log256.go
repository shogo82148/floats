package floats

import (
	"math/bits"

	"github.com/shogo82148/ints"
)

// Log returns the natural logarithm of a.
//
// Special cases are:
//
//	+Inf.Log() = +Inf
//	0.Log() = -Inf
//	(x < 0).Log() = NaN
//	NaN.Log() = NaN
func (a Float256) Log() Float256 {
	// special cases
	switch {
	case a.IsNaN() || a.IsInf(1):
		return a
	case a.Lt(Float256{}): // a < 0
		return NewFloat256NaN()
	case a.IsZero():
		return NewFloat256Inf(-1)
	}

	// a = m × 2**(exp-236), where m is a 237-bit integer in [2**236, 2**237).
	// normalize (rather than reading the exponent field directly) so that a
	// subnormal a, whose fraction has no implicit leading bit, is handled too.
	_, exp, m := a.normalize()

	// table-driven reduction: idx picks a breakpoint c ≈ m/2**236 in [1,2),
	// for which 1/c and log(c) are precomputed to avoid a run-time division.
	// log256InvC is rounded, and log256LogC is the logarithm of the rounded value,
	// so that the reduction is exact: a = 2**exp × c × (1+r).
	// Bucket 0 borders c = 1 and bucket 255 uses c = 2, so that r is computed exactly
	// even when a is close to a power of two, where log(a) itself is tiny.
	idx := (m[0] >> 36) & 0xff

	// p = m/2**236 × (1/c) in fixed point with 492 fractional bits. It is exact.
	var p ints.Uint512
	if idx == 0 {
		p = ints.Uint512{m[0], m[1], m[2], m[3]} // m × 2**256
	} else {
		p = m.Mul512(ints.Uint256(log256InvC[idx]))
	}

	// r = p - 1, |r| < 2**-8.
	one := ints.Uint512{1 << 44} // 2**492
	rneg := p.Cmp(one) < 0
	var rmag ints.Uint512 // |r| × 2**492
	if rneg {
		rmag = one.Sub(p)
	} else {
		rmag = p.Sub(one)
	}

	// log(a) = exp × ln(2) + log(c) + log(1+r).
	// If a is close to a power of two, log(a) ~ log(1+r) is much smaller than the other terms,
	// and it must be computed with the relative precision.
	tiny := (idx == 0 && exp == 0) || (idx == 255 && exp == -1)
	if rmag.IsZero() && tiny {
		return Float256{}
	}

	var log1r ints.Uint512 // |log(1+r)| × 2**320
	if !rmag.IsZero() {
		// log(1+r) = r × G(w), where w = -r and G(w) = 1 + w/2 + w²/3 + ... only has positive terms
		// if w > 0. G is in fixed point with 255 fractional bits, and w is with 264 fractional bits.
		w := shr512to256(rmag, 228)
		wneg := !rneg
		g := ints.Uint256(log256Coeffs[0])
		for _, c := range log256Coeffs[1:] {
			t := shr512to256(g.Mul512(w), 264)
			if wneg {
				g = ints.Uint256(c).Sub(t)
			} else {
				g = ints.Uint256(c).Add(t)
			}
		}

		// normalize r so that it keeps the relative precision.
		lz := rmag.LeadingZeros()
		s := rmag.Lsh(uint(lz))
		rn := ints.Uint256{s[0], s[1], s[2], s[3]} // |r| = rn × 2**(-236-lz)
		q := rn.Mul512(g)                          // |log(1+r)| = q × 2**(-491-lz)
		if tiny {
			var sign uint64
			if rneg {
				sign = signMask256[0]
			}
			return fixToFloat256(sign, q, false, -491-lz)
		}
		if shift := uint(171 + lz); shift < 512 {
			log1r = q.Rsh(shift)
		}
	}

	// exp × ln(2) in fixed point with 320 fractional bits.
	kabs := uint64(exp)
	if exp < 0 {
		kabs = -kabs
	}
	var v ints.Uint512
	var carry uint64
	for i := 4; i >= 0; i-- {
		hi, lo := bits.Mul64(log256Ln2[i], kabs)
		var c uint64
		v[i+3], c = bits.Add64(lo, carry, 0)
		carry = hi + c
	}
	v[2] = carry
	if exp < 0 {
		v = v.Neg()
	}

	c := &log256LogC[idx]
	v = v.Add(ints.Uint512{3: c[0], 4: c[1], 5: c[2], 6: c[3], 7: c[4]})
	if rneg {
		v = v.Sub(log1r)
	} else {
		v = v.Add(log1r)
	}

	if v.IsZero() {
		return Float256{}
	}
	var sign uint64
	if v[0]>>63 != 0 {
		sign = signMask256[0]
		v = v.Neg()
	}
	return fixToFloat256(sign, v, false, -320)
}

// Log10 returns the decimal logarithm of a.
// The special cases are the same as for [Log].
func (a Float256) Log10() Float256 {
	// 1/ln(10) ~ 0.43429448190325182765112891891660508229439700580366656611445378316586465
	var Ln10Inv = Float256{
		0x3fff_dbcb_7b15_26e5, 0x0e32_a6ab_7555_f5a6,
		0x7b86_47dc_68c0_48b9, 0x3440_4747_e5a8_9ef2,
	}
	return a.Log().Mul(Ln10Inv)
}

// Log2 returns the binary logarithm of a.
// The special cases are the same as for [Log].
func (a Float256) Log2() Float256 {
	var (
		// Half = 0.5
		Half = Float256{
			0x3fff_e000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// Ln2Inv = 1/ln(2)
		// ~ 1.44269504088896340735992468100189213742664595415298593413544940693110922
		Ln2Inv = Float256{
			0x3fff_f715_4765_2b82, 0xfe17_77d0_ffda_0d23,
			0xa7d1_1d6a_ef55_1bad, 0x2b4b_1164_a2cd_9a34,
		}
	)

	frac, exp := a.Frexp()
	// Make sure exact powers of two give an exact answer.
	// Don't depend on Log(0.5)*(1/Ln2)+exp being exactly exp-1.
	if frac.Eq(Half) {
		return NewFloat256(float64(exp - 1))
	}
	return frac.Log().Mul(Ln2Inv).Add(NewFloat256(float64(exp)))
}
