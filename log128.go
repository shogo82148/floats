package floats

import "math/bits"

// logGCoeffs128[i] is 1/(16-i) × 2**191 for i = 0, ..., 15, the coefficients of
// G(w) = log(1+r)/r in terms of w = -r, i.e. G = 1 + w/2 + w²/3 + ... + w¹⁵/16,
// in Horner's method order (highest power first), in fixed point with 191 fractional bits.
var logGCoeffs128 = [16][3]uint64{
	{0x0800000000000000, 0x0000000000000000, 0x0000000000000000}, // 1/16
	{0x0888888888888888, 0x8888888888888888, 0x8888888888888889}, // 1/15
	{0x0924924924924924, 0x9249249249249249, 0x2492492492492492}, // 1/14
	{0x09d89d89d89d89d8, 0x9d89d89d89d89d89, 0xd89d89d89d89d89e}, // 1/13
	{0x0aaaaaaaaaaaaaaa, 0xaaaaaaaaaaaaaaaa, 0xaaaaaaaaaaaaaaab}, // 1/12
	{0x0ba2e8ba2e8ba2e8, 0xba2e8ba2e8ba2e8b, 0xa2e8ba2e8ba2e8ba}, // 1/11
	{0x0ccccccccccccccc, 0xcccccccccccccccc, 0xcccccccccccccccd}, // 1/10
	{0x0e38e38e38e38e38, 0xe38e38e38e38e38e, 0x38e38e38e38e38e4}, // 1/9
	{0x1000000000000000, 0x0000000000000000, 0x0000000000000000}, // 1/8
	{0x1249249249249249, 0x2492492492492492, 0x4924924924924925}, // 1/7
	{0x1555555555555555, 0x5555555555555555, 0x5555555555555555}, // 1/6
	{0x1999999999999999, 0x9999999999999999, 0x999999999999999a}, // 1/5
	{0x2000000000000000, 0x0000000000000000, 0x0000000000000000}, // 1/4
	{0x2aaaaaaaaaaaaaaa, 0xaaaaaaaaaaaaaaaa, 0xaaaaaaaaaaaaaaab}, // 1/3
	{0x4000000000000000, 0x0000000000000000, 0x0000000000000000}, // 1/2
	{0x8000000000000000, 0x0000000000000000, 0x0000000000000000}, // 1/1
}

// add192 returns a+b modulo 2**192.
func add192(a, b [3]uint64) [3]uint64 {
	x2, c := bits.Add64(a[2], b[2], 0)
	x1, c := bits.Add64(a[1], b[1], c)
	x0, _ := bits.Add64(a[0], b[0], c)
	return [3]uint64{x0, x1, x2}
}

// sub192 returns a-b modulo 2**192.
func sub192(a, b [3]uint64) [3]uint64 {
	x2, c := bits.Sub64(a[2], b[2], 0)
	x1, c := bits.Sub64(a[1], b[1], c)
	x0, _ := bits.Sub64(a[0], b[0], c)
	return [3]uint64{x0, x1, x2}
}

// sub256 returns a-b as a signed magnitude: (neg, |a-b|), for 256-bit unsigned a, b.
func sub256(a, b [4]uint64) (neg bool, mag [4]uint64) {
	x3, borrow := bits.Sub64(a[3], b[3], 0)
	x2, borrow := bits.Sub64(a[2], b[2], borrow)
	x1, borrow := bits.Sub64(a[1], b[1], borrow)
	x0, borrow := bits.Sub64(a[0], b[0], borrow)
	if borrow == 0 {
		return false, [4]uint64{x0, x1, x2, x3}
	}
	y3, borrow := bits.Sub64(b[3], a[3], 0)
	y2, borrow := bits.Sub64(b[2], a[2], borrow)
	y1, borrow := bits.Sub64(b[1], a[1], borrow)
	y0, _ := bits.Sub64(b[0], a[0], borrow)
	return true, [4]uint64{y0, y1, y2, y3}
}

// logNormalizeTo192 finds the leading 1 bit of the 256-bit value x, shifts x
// left so that bit 255 (the MSB of the result) is that leading 1, and returns
// the shift amount lz and the top 192 bits of the shifted value.
// If x is zero, it returns lz = 256 and a zero result.
func logNormalizeTo192(x [4]uint64) (lz int, out [3]uint64) {
	idx := 0
	for idx < 4 && x[idx] == 0 {
		idx++
	}
	if idx == 4 {
		return 256, [3]uint64{}
	}
	wlz := bits.LeadingZeros64(x[idx])
	lz = idx*64 + wlz

	get := func(i int) uint64 {
		if i < 0 || i > 3 {
			return 0
		}
		return x[i]
	}
	s := uint(lz)
	wordsShift, bitShift := s/64, s%64
	word := func(i int) uint64 {
		hi, lo := get(i+int(wordsShift)), get(i+int(wordsShift)+1)
		if bitShift == 0 {
			return hi
		}
		return hi<<bitShift | lo>>(64-bitShift)
	}
	return lz, [3]uint64{word(0), word(1), word(2)}
}

// twoSum128 returns hi, lo such that hi+lo = a+b exactly (as real numbers),
// with hi = a+b rounded to the nearest Float128.
func twoSum128(a, b Float128) (hi, lo Float128) {
	hi = a.Add(b)
	v := hi.Sub(a)
	lo = a.Sub(hi.Sub(v)).Add(b.Sub(v))
	return
}

// twoProduct128 returns hi, lo such that hi+lo = a*b exactly (as real numbers),
// with hi = a*b rounded to the nearest Float128.
func twoProduct128(a, b Float128) (hi, lo Float128) {
	hi = a.Mul(b)
	lo = FMA128(a, b, hi.Neg())
	return
}

// ddAdd128 adds two double-Float128 numbers (aHi+aLo) and (bHi+bLo),
// returning a double-Float128 result (hi+lo) that is accurate even when
// the two numbers are of very different magnitude or nearly cancel.
// Every step uses the unconditional twoSum128 (rather than a quick-two-sum
// that requires knowing which operand is larger), since that relation isn't
// known in general for the terms this is applied to in Log.
func ddAdd128(aHi, aLo, bHi, bLo Float128) (hi, lo Float128) {
	s1, s2 := twoSum128(aHi, bHi)
	t1, t2 := twoSum128(aLo, bLo)
	s2 = s2.Add(t1)
	s1, s2 = twoSum128(s1, s2)
	s2 = s2.Add(t2)
	hi, lo = twoSum128(s1, s2)
	return
}

// Log returns the natural logarithm of a.
//
// Special cases are:
//
//	+Inf.Log() = +Inf
//	0.Log() = -Inf
//	(x < 0).Log() = NaN
//	NaN.Log() = NaN
func (a Float128) Log() Float128 {
	// special cases
	switch {
	case a.IsNaN() || a.IsInf(1):
		return a
	case a.Lt(Float128{}): // a < 0
		return NewFloat128NaN()
	case a.IsZero():
		return NewFloat128Inf(-1)
	}

	var (
		// Ln2Hi = ln(2) ~ 0.6931471805599453094172321214581765
		// Ln2Lo = ln(2) - Ln2Hi ~ 8.928835774481220748938623512047474e-35
		Ln2Hi = Float128{0x3ffe_62e4_2fef_a39e, 0xf357_93c7_6730_07e5}
		Ln2Lo = Float128{0x3f8d_dabd_03cd_0c99, 0xca62_d8b6_2834_5d6e}
	)

	// a = m × 2**(exp-112), where m = (m1:m0) is a 113-bit integer in [2**112, 2**113).
	// normalize (rather than reading the exponent field directly) so that a
	// subnormal a, whose fraction has no implicit leading bit, is handled too.
	_, exp, frac := a.normalize()
	m1, m0 := frac[0], frac[1]

	// table-driven reduction: idx picks a breakpoint c ≈ m/2**112 in [1,2),
	// for which 1/c and log(c) are precomputed to avoid a run-time division.
	// Buckets 0 and 255 border a power of two (c = 1 and c = 2) exactly rather
	// than their usual midpoint, and fold that power of two into k instead of
	// using log(c): every other bucket's log(c) is bounded away from 0 by the
	// bucket width, but these two border the one point (a a power of two) where
	// log(a) itself can be arbitrarily small, so they must have no rounding
	// error of their own, or that error would dominate log(a) there.
	idx := (m1 >> 40) & 0xff
	k := NewFloat128(float64(exp))
	var logCHi, logCLo Float128
	var rneg bool
	var rmag [4]uint64
	switch idx {
	case 0:
		// r = m/2**112 - 1, read directly from m's fraction bits: exact, and never negative.
		rmag = [4]uint64{m1 &^ (1 << 48), m0, 0, 0}
	case 255:
		// r = m/2**113 - 1 = -(2**113-m)/2**113, i.e. c = 2 = 2**(exp+1); 2**113-m
		// is read directly from m's bits via a two's-complement-style negation.
		k = NewFloat128(float64(exp + 1))
		lo64, borrow := bits.Sub64(0, m0, 0)
		hi64, _ := bits.Sub64(1<<49, m1, borrow)
		rneg = true
		// |r| = (2**113-m)/2**113, so rmag = (2**113-m) × 2**127 (2**240/2**113)
		// is (hi64:lo64) shifted left by 127, i.e. by 64 and then 63 more bits.
		rmag = [4]uint64{hi64 >> 1, hi64<<63 | lo64>>1, lo64 << 63, 0}
	default:
		logCHi, logCLo = logCHi128[idx], logCLo128[idx]

		// p = m × invC × 2**304, i.e. (m/2**112) × (1/c) in fixed point with 304 fractional bits.
		p := mul192x192([3]uint64{0, m1, m0}, logInvC128[idx])

		// r = m×invC/2**112 - 1 ~ m/(c×2**112) - 1, computed from the middle 256 bits of p
		// (the top word is always 0 and the bottom word is negligible at our precision).
		phi4 := [4]uint64{p[1], p[2], p[3], p[4]}
		const252 := [4]uint64{1 << 48, 0, 0, 0} // 2**240 within this 256-bit window
		rneg, rmag = sub256(phi4, const252)
	}

	// e×ln2, as a double-Float128, with the exact error of k×Ln2Hi recovered via
	// FMA128 so that this term keeps full precision even when it nearly cancels
	// with log(c) below (e.g. a just under a power of 2, so c is close to 2).
	eHi, eLo := twoProduct128(k, Ln2Hi)
	eLo = eLo.Add(k.Mul(Ln2Lo))
	hi, lo := ddAdd128(eHi, eLo, logCHi, logCLo)

	if rmag == ([4]uint64{}) {
		// m is exactly c×2**112; log(1+r) = 0.
		return hi.Add(lo)
	}

	lz, rnorm := logNormalizeTo192(rmag) // rmag is never all zero here, so lz <= 231.

	// w = -r, so G(w) = log(1+r)/r = 1 + w/2 + w²/3 + ... has only positive terms,
	// with w's sign handled by adding when w > 0 (r < 0) and subtracting when w < 0 (r > 0).
	// |w|, at scale 2**-192 (so that mul192x192(g, w) truncated to its top 192 bits
	// lands back at g's own scale of 2**-191), is rnorm shifted right by (lz-16).
	w2, w1, w0 := rsh192(rnorm[0], rnorm[1], rnorm[2], uint(lz-16))
	wneg := !rneg
	var g [3]uint64
	for _, c := range logGCoeffs128 {
		q := mul192x192(g, [3]uint64{w2, w1, w0})
		qTop := [3]uint64{q[0], q[1], q[2]}
		if wneg {
			g = sub192(c, qTop)
		} else {
			g = add192(qTop, c)
		}
	}

	// log(1+r) = r × G(w), computed directly in fixed point from the normalized
	// mantissa rnorm (in [2**191, 2**192)) and g (also in [2**191, 2**192)),
	// so their product's top 192 bits always sit in q[0]:q[1]:q[2]. log1r is kept
	// as a double-Float128 (log1rHi+log1rLo): when a is close to a power of c
	// (e.g. a itself close to 1), log1rHi alone is only Float128-accurate relative
	// to log(1+r)'s own size, which isn't enough once it nearly cancels with hi+lo.
	q := mul192x192(rnorm, g)
	finalExp := -175 - lz
	log1rHi := fixToFloat128(0, q[0], q[1], q[2], q[3] != 0 || q[4] != 0 || q[5] != 0, finalExp)

	// log1rLo is the exact residual q - log1rHi (both expressed as 384-bit integers
	// at the common scale finalExp), re-rounded to Float128: this accounts correctly
	// for log1rHi's own rounding, unlike simply slicing the next bits of q.
	hExp := int((log1rHi[0]>>(shift128-64))&mask128) - bias128
	hm1 := log1rHi[0]&fracMask128[0] | 1<<(shift128-64)
	hm0 := log1rHi[1]
	// q is a full 384-bit integer, while fixToFloat128 above only used its top 192
	// bits at scale finalExp, so q's own scale (for this 384-bit comparison) is
	// finalExp-192; hiFrame must be placed at that same scale for the subtraction
	// below to be a true bit-for-bit comparison.
	hiFrame := place128In384(hExp-(finalExp-192)-112, hm1, hm0)
	loNeg, loMag := sub384(q, hiFrame)
	var log1rLo Float128
	if loMag != ([6]uint64{}) {
		loLz, loNorm := logNormalizeTo192From384(loMag)
		log1rLo = fixToFloat128(0, loNorm[0], loNorm[1], loNorm[2], false, finalExp-loLz)
		if loNeg {
			log1rLo = log1rLo.Neg()
		}
	}
	if rneg {
		log1rHi, log1rLo = log1rHi.Neg(), log1rLo.Neg()
	}

	hi, lo = ddAdd128(hi, lo, log1rHi, log1rLo)
	return hi.Add(lo)
}

// place128In384 returns (m1:m0) positioned in a 384-bit (6-word) value such that
// m0's bit 0 sits at global bit position lsbPos (bit 0 = LSB of word 5, bit 383 =
// MSB of word 0); bits that would fall outside [0,384) are dropped.
func place128In384(lsbPos int, m1, m0 uint64) [6]uint64 {
	var out [6]uint64
	place := func(val uint64, lowBit int) {
		for b := range 64 {
			if val&(1<<uint(b)) == 0 {
				continue
			}
			gbit := lowBit + b
			if gbit < 0 || gbit >= 384 {
				continue
			}
			out[5-gbit/64] |= 1 << uint(gbit%64)
		}
	}
	place(m0, lsbPos)
	place(m1, lsbPos+64)
	return out
}

// sub384 returns a-b as a signed magnitude: (neg, |a-b|), for 384-bit unsigned a, b.
func sub384(a, b [6]uint64) (neg bool, mag [6]uint64) {
	var borrow uint64
	var x [6]uint64
	for i := 5; i >= 0; i-- {
		x[i], borrow = bits.Sub64(a[i], b[i], borrow)
	}
	if borrow == 0 {
		return false, x
	}
	var y [6]uint64
	borrow = 0
	for i := 5; i >= 0; i-- {
		y[i], borrow = bits.Sub64(b[i], a[i], borrow)
	}
	return true, y
}

// logNormalizeTo192From384 finds the leading 1 bit of the 384-bit value x, shifts x
// left so that bit 383 (the MSB of the result) is that leading 1, and returns the
// shift amount lz and the top 192 bits of the shifted value.
func logNormalizeTo192From384(x [6]uint64) (lz int, out [3]uint64) {
	idx := 0
	for idx < 6 && x[idx] == 0 {
		idx++
	}
	wlz := bits.LeadingZeros64(x[idx])
	lz = idx*64 + wlz

	get := func(i int) uint64 {
		if i < 0 || i > 5 {
			return 0
		}
		return x[i]
	}
	s := uint(lz)
	wordsShift, bitShift := s/64, s%64
	word := func(i int) uint64 {
		hi, lo := get(i+int(wordsShift)), get(i+int(wordsShift)+1)
		if bitShift == 0 {
			return hi
		}
		return hi<<bitShift | lo>>(64-bitShift)
	}
	return lz, [3]uint64{word(0), word(1), word(2)}
}

// Log10 returns the decimal logarithm of a.
// The special cases are the same as for [Log].
func (a Float128) Log10() Float128 {
	// // 1/ln(10) ~ 0.4342944819032518276511289189166051
	var Ln10Inv = Float128{0x3ffd_bcb7_b152_6e50, 0xe32a_6ab7_555f_5a68}
	return a.Log().Mul(Ln10Inv)
}

// Log2 returns the binary logarithm of a.
// The special cases are the same as for [Log].
func (a Float128) Log2() Float128 {
	var (
		// Half = 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}

		// Ln2Inv = 1/ln(2)
		// ~ 1.442695040888963407359924681001892
		Ln2Inv = Float128{0x3fff_7154_7652_b82f, 0xe177_7d0f_fda0_d23a}
	)

	frac, exp := a.Frexp()
	// Make sure exact powers of two give an exact answer.
	// Don't depend on Log(0.5)*(1/Ln2)+exp being exactly exp-1.
	if frac.Eq(Half) {
		return NewFloat128(float64(exp - 1))
	}
	return frac.Log().Mul(Ln2Inv).Add(NewFloat128(float64(exp)))
}
