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

	sign, v, exp := log256Fix(a)
	if v.IsZero() {
		return Float256{}
	}
	return fixToFloat256(sign, v, false, exp)
}

// log256Fix returns log(a) = ±v × 2**exp for a finite a > 0 as the sign bit and v, with v = 0 if a = 1.
// The relative error is about 2**-254 even if a is close to 1.
func log256Fix(a Float256) (sign uint64, v ints.Uint512, exp int) {
	k, idx, rneg, rmag := log256Reduce(a)
	return log256Combine(k, idx, rneg, rmag)
}

// log256Combine returns log(2**k × c × (1+r)) = ±v × 2**exp, where c is the breakpoint of the bucket idx
// and r = ±rmag × 2**-492, |r| < 2**-8, as the sign bit and v. v = 0 if the logarithm is exactly zero.
func log256Combine(k int, idx uint64, rneg bool, rmag ints.Uint512) (sign uint64, v ints.Uint512, exp int) {
	// log(a) = k × ln(2) + log(c) + log(1+r).
	// If a is close to a power of two, log(a) ~ log(1+r) is much smaller than the other terms,
	// and it must be computed with the relative precision.
	tiny := (idx == 0 && k == 0) || (idx == 255 && k == -1)
	if rmag.IsZero() && tiny {
		return 0, ints.Uint512{}, 0
	}

	var log1r ints.Uint512 // |log(1+r)| × 2**320
	if !rmag.IsZero() {
		q, lz := log256Log1p(rneg, rmag) // |log(1+r)| = q × 2**(-491-lz)
		if tiny {
			if rneg {
				sign = signMask256[0]
			}
			return sign, q, -491 - lz
		}
		if shift := uint(171 + lz); shift < 512 {
			log1r = q.Rsh(shift)
		}
	}

	// k × ln(2) in fixed point with 320 fractional bits.
	kabs := uint64(k)
	if k < 0 {
		kabs = -kabs
	}
	var carry uint64
	for i := 4; i >= 0; i-- {
		hi, lo := bits.Mul64(log256Ln2[i], kabs)
		var c uint64
		v[i+3], c = bits.Add64(lo, carry, 0)
		carry = hi + c
	}
	v[2] = carry
	if k < 0 {
		v = v.Neg()
	}

	c := &log256LogC[idx]
	v = v.Add(ints.Uint512{3: c[0], 4: c[1], 5: c[2], 6: c[3], 7: c[4]})
	if rneg {
		v = v.Sub(log1r)
	} else {
		v = v.Add(log1r)
	}

	if v[0]>>63 != 0 {
		sign = signMask256[0]
		v = v.Neg()
	}
	return sign, v, -320
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

// log256G returns G(w) = 1 + w/2 + w²/3 + ... + w**32/33 in fixed point with 255 fractional bits,
// where w is in fixed point with 264 fractional bits, and 0 <= |w| < 2**-8. w is negative if wneg is true.
func log256G(w ints.Uint256, wneg bool) ints.Uint256 {
	// |w| < 2**-8, so an error in the partial sum made i steps before the end of Horner's method
	// is reduced by 2**(-8i). The earlier steps need fewer bits, as in sinhPoly256:
	// the step p = 1, 2, ..., 32 uses at least 8p + 16 bits, and its error is less than 2**-269 in the result.
	c := &log256Coeffs
	const n = len(log256Coeffs)

	// 128 bits (p <= 14): g in fixed point with 127 fractional bits, w with 136 fractional bits.
	g1, g0 := c[0][0], c[0][1]
	for i := 1; i < 15; i++ {
		x3, x2, _, _ := mul128x128(g1, g0, w[0], w[1])
		t1, t0 := x3>>8, x3<<56|x2>>8
		var b uint64
		if wneg {
			g0, b = bits.Sub64(c[i][1], t0, 0)
			g1, _ = bits.Sub64(c[i][0], t1, b)
		} else {
			g0, b = bits.Add64(c[i][1], t0, 0)
			g1, _ = bits.Add64(c[i][0], t1, b)
		}
	}

	// 192 bits (p <= 22): g in fixed point with 191 fractional bits, w with 200 fractional bits.
	g := [3]uint64{g1, g0, 0}
	w192 := [3]uint64{w[0], w[1], w[2]}
	for i := 15; i < 23; i++ {
		x := mul192x192(g, w192)
		t := [3]uint64{x[0] >> 8, x[0]<<56 | x[1]>>8, x[1]<<56 | x[2]>>8}
		var b uint64
		if wneg {
			g[2], b = bits.Sub64(c[i][2], t[2], 0)
			g[1], b = bits.Sub64(c[i][1], t[1], b)
			g[0], _ = bits.Sub64(c[i][0], t[0], b)
		} else {
			g[2], b = bits.Add64(c[i][2], t[2], 0)
			g[1], b = bits.Add64(c[i][1], t[1], b)
			g[0], _ = bits.Add64(c[i][0], t[0], b)
		}
	}

	// 256 bits: g in fixed point with 255 fractional bits, w with 264 fractional bits.
	g256 := ints.Uint256{g[0], g[1], g[2], 0}
	for i := 23; i < n; i++ {
		h := mulHi256(g256, w) // about (g × w) >> 256
		t := ints.Uint256{h[0] >> 8, h[0]<<56 | h[1]>>8, h[1]<<56 | h[2]>>8, h[2]<<56 | h[3]>>8}
		if wneg {
			g256 = ints.Uint256(c[i]).Sub(t)
		} else {
			g256 = ints.Uint256(c[i]).Add(t)
		}
	}
	return g256
}

// log256Log1p returns |log(1+r)| = q × 2**(-491-lz) for r = ±rmag × 2**-492 ≠ 0, |r| < 2**-8.
// q is in [2**510, 2**512), so it keeps the relative precision even if r is tiny.
func log256Log1p(rneg bool, rmag ints.Uint512) (q ints.Uint512, lz int) {
	// log(1+r) = r × G(w), where w = -r and G(w) = 1 + w/2 + w²/3 + ... only has positive terms
	// if w > 0. G is in fixed point with 255 fractional bits, and w is with 264 fractional bits.
	w := shr512to256(rmag, 228)
	g := log256G(w, !rneg)

	// normalize r so that it keeps the relative precision.
	lz = rmag.LeadingZeros()
	s := rmag.Lsh(uint(lz))
	rn := ints.Uint256{s[0], s[1], s[2], s[3]} // |r| = rn × 2**(-236-lz)
	return rn.Mul512(g), lz
}

// Log10 returns the decimal logarithm of a.
// The special cases are the same as for [Log].
func (a Float256) Log10() Float256 {
	// special cases
	switch {
	case a.IsNaN() || a.IsInf(1):
		return a
	case a.Lt(Float256{}): // a < 0
		return NewFloat256NaN()
	case a.IsZero():
		return NewFloat256Inf(-1)
	}

	// log10(a) = log(a)/ln(10), where log(a) is computed without rounding.
	sign, v, exp := log256Fix(a)
	if v.IsZero() {
		return Float256{}
	}

	// v = vn × 2**(256-lz) for the top 256 bits vn of v.
	lz := v.LeadingZeros()
	s := v.Lsh(uint(lz))
	vn := ints.Uint256{s[0], s[1], s[2], s[3]}
	p := vn.Mul512(ints.Uint256(log256Ln10Inv)) // × (1/ln(10)) × 2**256
	return fixToFloat256(sign, p, false, exp-lz)
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
