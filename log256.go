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

	exp, idx, rneg, rmag := log256Reduce(a)

	// log(a) = exp × ln(2) + log(c) + log(1+r).
	// If a is close to a power of two, log(a) ~ log(1+r) is much smaller than the other terms,
	// and it must be computed with the relative precision.
	tiny := (idx == 0 && exp == 0) || (idx == 255 && exp == -1)
	if rmag.IsZero() && tiny {
		return Float256{}
	}

	var log1r ints.Uint512 // |log(1+r)| × 2**320
	if !rmag.IsZero() {
		q, lz := log256Log1p(rneg, rmag) // |log(1+r)| = q × 2**(-491-lz)
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

// mul320x320 returns the 640-bit product of a and b.
// The limbs are in the big-endian order.
func mul320x320(a, b [5]uint64) (p [10]uint64) {
	for i := 4; i >= 0; i-- {
		var carry uint64
		for j := 4; j >= 0; j-- {
			hi, lo := bits.Mul64(a[j], b[i])
			var c1, c2 uint64
			lo, c1 = bits.Add64(lo, carry, 0)
			p[i+j+1], c2 = bits.Add64(p[i+j+1], lo, 0)
			carry = hi + c1 + c2
		}
		p[i] = carry
	}
	return
}

// log256Reduce reduces a > 0 to a = 2**exp × c × (1+r),
// where c is the breakpoint of the bucket idx, and returns r = ±rmag × 2**-492, |r| < 2**-8.
// The reduction is exact.
func log256Reduce(a Float256) (exp int, idx uint64, rneg bool, rmag ints.Uint512) {
	// a = m × 2**(exp-236), where m is a 237-bit integer in [2**236, 2**237).
	// normalize (rather than reading the exponent field directly) so that a
	// subnormal a, whose fraction has no implicit leading bit, is handled too.
	_, exp, m := a.normalize()

	// table-driven reduction: idx picks a breakpoint c ≈ m/2**236 in [1,2),
	// for which 1/c and log(c) are precomputed to avoid a run-time division.
	// log256InvC is rounded, and log256LogC is the logarithm of the rounded value,
	// so that the reduction is exact.
	// Bucket 0 borders c = 1 and bucket 255 uses c = 2, so that r is computed exactly
	// even when a is close to a power of two, where log(a) itself is tiny.
	idx = (m[0] >> 36) & 0xff

	// p = m/2**236 × (1/c) in fixed point with 492 fractional bits. It is exact.
	var p ints.Uint512
	if idx == 0 {
		p = ints.Uint512{m[0], m[1], m[2], m[3]} // m × 2**256
	} else {
		p = m.Mul512(ints.Uint256(log256InvC[idx]))
	}

	// r = p - 1
	one := ints.Uint512{1 << 44} // 2**492
	rneg = p.Cmp(one) < 0
	if rneg {
		rmag = one.Sub(p)
	} else {
		rmag = p.Sub(one)
	}
	return
}

// log256Log1p returns |log(1+r)| = q × 2**(-491-lz) for r = ±rmag × 2**-492 ≠ 0, |r| < 2**-8.
// q is in [2**510, 2**512), so it keeps the relative precision even if r is tiny.
func log256Log1p(rneg bool, rmag ints.Uint512) (q ints.Uint512, lz int) {
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
	lz = rmag.LeadingZeros()
	s := rmag.Lsh(uint(lz))
	rn := ints.Uint256{s[0], s[1], s[2], s[3]} // |r| = rn × 2**(-236-lz)
	return rn.Mul512(g), lz
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
	// special cases
	switch {
	case a.IsNaN() || a.IsInf(1):
		return a
	case a.Lt(Float256{}): // a < 0
		return NewFloat256NaN()
	case a.IsZero():
		return NewFloat256Inf(-1)
	}

	// log2(a) = k + (log(c) + log(1+r))/ln(2).
	// Bucket 255 is c = 2, so it is folded into k to make exact powers of two exact.
	exp, idx, rneg, rmag := log256Reduce(a)
	k := exp
	var logC [5]uint64
	if idx == 255 {
		k++
	} else {
		logC = log256LogC[idx]
	}

	// If a is close to a power of two, log2(a) is much smaller than k and log(c) is zero,
	// and it must be computed with the relative precision.
	tiny := k == 0 && (idx == 0 || idx == 255)
	if rmag.IsZero() && tiny {
		return Float256{}
	}

	var log1r ints.Uint512 // |log(1+r)| × 2**320
	if !rmag.IsZero() {
		q, lz := log256Log1p(rneg, rmag) // |log(1+r)| = q × 2**(-491-lz)
		if tiny {
			// |log2(1+r)| = q × 2**(-491-lz) / ln(2)
			var sign uint64
			if rneg {
				sign = signMask256[0]
			}
			top := ints.Uint256{q[0], q[1], q[2], q[3]} // q × 2**-256
			p := top.Mul512(ints.Uint256(log256Ln2Inv)) // × (1/ln(2)) × 2**255
			return fixToFloat256(sign, p, false, -490-lz)
		}
		if shift := uint(171 + lz); shift < 512 {
			log1r = q.Rsh(shift)
		}
	}

	// t = log(c) + log(1+r) in fixed point with 320 fractional bits.
	t := ints.Uint512{3: logC[0], 4: logC[1], 5: logC[2], 6: logC[3], 7: logC[4]}
	if rneg {
		t = t.Sub(log1r)
	} else {
		t = t.Add(log1r)
	}
	tneg := t[0]>>63 != 0
	if tneg {
		t = t.Neg()
	}

	// t/ln(2) in fixed point with 320 fractional bits.
	// t < 2**320 and 1/ln(2) × 2**319 < 2**320.
	p := mul320x320([5]uint64{t[3], t[4], t[5], t[6], t[7]}, log256Ln2Inv319)
	var v ints.Uint512 // (t × log256Ln2Inv319) >> 319
	v[7] = p[4]<<1 | p[5]>>63
	v[6] = p[3]<<1 | p[4]>>63
	v[5] = p[2]<<1 | p[3]>>63
	v[4] = p[1]<<1 | p[2]>>63
	v[3] = p[0]<<1 | p[1]>>63
	if tneg {
		v = v.Neg()
	}

	// k in fixed point with 320 fractional bits, in two's complement.
	kv := ints.Uint512{2: uint64(int64(k))}
	if k < 0 {
		kv[0], kv[1] = ^uint64(0), ^uint64(0)
	}
	v = v.Add(kv)

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
