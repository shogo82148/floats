package floats

import (
	"math/bits"

	"github.com/shogo82148/ints"
)

// Asinh returns the inverse hyperbolic sine of a.
//
// Special cases are:
//
//	±0.Asinh() = ±0
//	±Inf.Asinh() = ±Inf
//	NaN.Asinh() = NaN
func (a Float128) Asinh() Float128 {
	// a = m × 2**(exp-112), where m = (m1:m0) is a 113-bit integer in [2**112, 2**113).
	exp := int((a[0]>>(shift128-64))&mask128) - bias128
	if exp == mask128-bias128 {
		return a // NaN and ±Inf
	}
	if exp < -56 {
		// |a| < 2**-56, and |asinh(a) - a| ~ |a|³/6 is less than the half ulp of a.
		// This also handles ±0 and subnormal numbers.
		return a
	}
	sign := a[0] & signMask128[0]
	m1 := a[0]&fracMask128[0] | 1<<(shift128-64)
	return logHyp128(sign, exp, m1, a[1], false)
}

// logHyp128 returns ±log(a + sqrt(a² ± 1)), i.e. ±asinh(a) if acosh is false and ±acosh(a) otherwise,
// for a = m × 2**(exp-112), where m = (m1:m0) is a 113-bit integer in [2**112, 2**113).
// sign is the sign bit of the result. a must be greater than 1 if acosh is true, and the exp must be at least -56.
func logHyp128(sign uint64, exp int, m1, m0 uint64, acosh bool) Float128 {
	// u = a + sqrt(a² ± 1) = 2**k × v, where v is in [1, 2) and in fixed point with 192 fractional bits.
	var k int
	var v ints.Uint256
	if exp >= 60 {
		// u = 2a × (1 ± 1/(4a²) + ...) and the relative error of the approximation u ~ 2a is less than 2**-122.
		k = exp + 1
		v = ints.Uint256{2: m1, 3: m0}.Lsh(80)
	} else {
		// sqrt(a² ± 1) × 2**192 = sqrt(m² × 2**(2exp+160) ± 2**384).
		// the integer arithmetic is exact even if a is close to 1 and a² - 1 cancels.
		p3, p2, p1, p0 := mul128x128(m1, m0, m1, m0)
		n := ints.Uint512{4: p3, 5: p2, 6: p1, 7: p0}.Lsh(uint(2*exp + 160))
		if acosh {
			n = n.Sub(ints.Uint512{1: 1})
		} else {
			n = n.Add(ints.Uint512{1: 1})
		}
		s := uint(n.LeadingZeros() &^ 1) // sqrtRem512 requires n[0] >= 1<<62.
		root, _ := sqrtRem512(n.Lsh(s))
		u := ints.Uint256{2: m1, 3: m0}.Lsh(uint(exp + 80)).Add(root.Rsh(s / 2)) // u × 2**192
		k = u.BitLen() - 193
		v = u.Rsh(uint(k))
	}
	return logFix128(sign, k, v, 0)
}

// logFix128 returns ±log(u) × 2**e, where u = 2**k × v, and v is in [1, 2) and in fixed point with 192 fractional bits.
// u must be greater than 1 and its logarithm must not be less than 2**-58.
// sign is the sign bit of the result.
func logFix128(sign uint64, k int, v ints.Uint256, e int) Float128 {
	t := logKernel128(k, v)
	if t[0] != 0 {
		return fixToFloat128(sign, t[0], t[1], t[2], t[3] != 0, e-128)
	}
	return fixToFloat128(sign, t[1], t[2], t[3], false, e-192)
}

// logKernel128 returns log(u) × 2**192 for u = 2**k × v, where v is in [1, 2) and in fixed point with 192 fractional bits.
// u must be greater than 1 and its logarithm must not be less than 2**-58.
// The absolute error is about 2**-150 + k × 2**-192.
func logKernel128(k int, v ints.Uint256) ints.Uint256 {
	// v = 1 + f × 2**-192. the table-driven reduction picks a breakpoint c ≈ v in [1, 2)
	// for which 1/c and log(c) are precomputed, and log(v) = log(c) + log(1+r) with r = v/c - 1.
	// Bucket 0 borders c = 1, where log(v) itself is tiny, and uses c = 1 exactly.
	f := [3]uint64{v[1], v[2], v[3]}
	idx := f[0] >> 56
	var rneg bool
	var rmag [3]uint64 // |r| × 2**192
	if idx == 0 {
		rmag = f
	} else {
		// p = v × (1/c) in fixed point with 192 fractional bits.
		q := mul192x192(f, logInvC128[idx])
		var c uint64
		p2, c := bits.Add64(q[2], logInvC128[idx][2], 0)
		p1, c := bits.Add64(q[1], logInvC128[idx][1], c)
		p0, c := bits.Add64(q[0], logInvC128[idx][0], c)
		rmag = [3]uint64{p0, p1, p2}
		if c == 0 {
			rneg = true
			rmag = sub192([3]uint64{}, rmag)
		}
	}

	// log(1+r) = r × G(w), where w = -r and G(w) = 1 + w/2 + w²/3 + ... + w¹⁵/16.
	//
	// |r| < 2**-8, so an error in the partial sum made i steps before the end of Horner's method is reduced
	// by 2**(-8i). The earlier steps use fewer bits (192 bits for the last 2 steps, 128 bits for the 8 steps before them,
	// and 64 bits for the first 5 steps), and the error is less than 2**-150 in the result.
	g := logGCoeffs128[0]
	for i, c := range logGCoeffs128[1:] {
		var top [3]uint64
		switch {
		case i < 5:
			hi, lo := bits.Mul64(g[0], rmag[0])
			top = [3]uint64{hi, lo, 0}
		case i < 13:
			p3, p2, p1, _ := mul128x128(g[0], g[1], rmag[0], rmag[1])
			top = [3]uint64{p3, p2, p1}
		default:
			q := mul192x192(g, rmag)
			top = [3]uint64{q[0], q[1], q[2]}
		}
		if rneg {
			g = add192(top, c)
		} else {
			g = sub192(c, top)
		}
	}
	q := mul192x192(rmag, g)
	log1r := ints.Uint256{1: q[0]<<1 | q[1]>>63, 2: q[1]<<1 | q[2]>>63, 3: q[2]<<1 | q[3]>>63} // |log(1+r)| × 2**192

	// k × ln(2) + log(c) ± log(1+r) in fixed point with 192 fractional bits.
	var t ints.Uint256
	kk := uint64(k)
	h2, l2 := bits.Mul64(asinhLn2Fix128[2], kk)
	h1, l1 := bits.Mul64(asinhLn2Fix128[1], kk)
	h0, l0 := bits.Mul64(asinhLn2Fix128[0], kk)
	var c uint64
	t[3] = l2
	t[2], c = bits.Add64(l1, h2, 0)
	t[1], c = bits.Add64(l0, h1, c)
	t[0] = h0 + c
	if idx != 0 {
		lc := &asinhLogC128[idx]
		t = t.Add(ints.Uint256{1: lc[0], 2: lc[1], 3: lc[2]})
	}
	if rneg {
		t = t.Sub(log1r)
	} else {
		t = t.Add(log1r)
	}
	return t
}

// Acosh returns the inverse hyperbolic cosine of a.
//
// Special cases are:
//
//	+Inf.Acosh() = +Inf
//	x.Acosh() = NaN if x < 1
//	NaN.Acosh() = NaN
func (a Float128) Acosh() Float128 {
	// a = m × 2**(exp-112), where m = (m1:m0) is a 113-bit integer in [2**112, 2**113).
	exp := int((a[0]>>(shift128-64))&mask128) - bias128
	switch {
	case a.IsNaN() || a[0]&signMask128[0] != 0 || exp < 0: // NaN or a < 1
		return NewFloat128NaN()
	case exp == mask128-bias128:
		return a // +Inf
	case a == Float128(uvone128):
		return Float128{}
	}
	return logHyp128(0, exp, a[0]&fracMask128[0]|1<<(shift128-64), a[1], true)
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
func (a Float128) Atanh() Float128 {
	// a = m × 2**(exp-112), where m = (m1:m0) is a 113-bit integer in [2**112, 2**113).
	exp := int((a[0]>>(shift128-64))&mask128) - bias128
	sign := a[0] & signMask128[0]
	switch {
	case exp > 0 || (exp == 0 && (a[0]&fracMask128[0] != 0 || a[1] != 0)):
		return NewFloat128NaN() // NaN or |a| > 1
	case exp == 0:
		return Float128{sign | uvinf128[0], uvinf128[1]} // ±1
	case exp < -58:
		// |a| < 2**-58, and |atanh(a) - a| ~ |a|³/3 is less than the half ulp of a.
		// This also handles ±0 and subnormal numbers.
		return a
	}

	// atanh(a) = log(u)/2, where u = (1+|a|)/(1-|a|) = (2**f + m)/(2**f - m) with f = 112 - exp.
	m1 := a[0]&fracMask128[0] | 1<<(shift128-64)
	f := uint(112 - exp)
	one := ints.Uint512{7: 1}.Lsh(f)
	m := ints.Uint512{6: m1, 7: a[1]}
	q := one.Add(m).Lsh(192).Quo(one.Sub(m)) // u × 2**192
	k := q.BitLen() - 193
	q = q.Rsh(uint(k))
	return logFix128(sign, k, ints.Uint256{q[4], q[5], q[6], q[7]}, -1)
}
