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
	m0 := a[1]

	// asinh(|a|) = log(u) = k × ln(2) + log(v), where u = |a| + sqrt(1+a²) = 2**k × v and v is in [1, 2).
	// v is in fixed point with 192 fractional bits.
	var k int
	var v ints.Uint256
	if exp >= 60 {
		// u = 2|a| × (1 + 1/(4a²) + ...) and the relative error of the approximation u ~ 2|a| is less than 2**-122.
		k = exp + 1
		v = ints.Uint256{2: m1, 3: m0}.Lsh(80)
	} else {
		// sqrt(1+a²) × 2**192 = sqrt(2**384 + m² × 2**(2exp+160)).
		p3, p2, p1, p0 := mul128x128(m1, m0, m1, m0)
		n := ints.Uint512{4: p3, 5: p2, 6: p1, 7: p0}.Lsh(uint(2*exp + 160)).Add(ints.Uint512{1: 1})
		s := uint(n.LeadingZeros() &^ 1) // sqrtRem512 requires n[0] >= 1<<62.
		root, _ := sqrtRem512(n.Lsh(s))
		u := ints.Uint256{2: m1, 3: m0}.Lsh(uint(exp + 80)).Add(root.Rsh(s / 2)) // u × 2**192
		k = u.BitLen() - 193
		v = u.Rsh(uint(k))
	}

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
	g := logGCoeffs128[0]
	for _, c := range logGCoeffs128[1:] {
		q := mul192x192(g, rmag)
		top := [3]uint64{q[0], q[1], q[2]}
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

	if t[0] != 0 {
		return fixToFloat128(sign, t[0], t[1], t[2], t[3] != 0, -128)
	}
	return fixToFloat128(sign, t[1], t[2], t[3], false, -192)
}

// Acosh returns the inverse hyperbolic cosine of a.
//
// Special cases are:
//
//	+Inf.Acosh() = +Inf
//	x.Acosh() = NaN if x < 1
//	NaN.Acosh() = NaN
func (a Float128) Acosh() Float128 {
	var (
		Ln2   = Float128{0x3ffe_62e4_2fef_a39e, 0xf357_93c7_6730_07e6} // 6.93147180559945286227e-01
		One   = Float128(uvone128)                                     // 1.0
		Two   = Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000} // 2.0
		Large = Float128{0x4039_0000_0000_0000, 0x0000_0000_0000_0000} // 2**58
	)
	switch {
	case a.Lt(One) || a.IsNaN():
		return NewFloat128NaN()
	case a.Eq(One):
		return Float128{}
	case a.Ge(Large):
		return a.Log().Add(Ln2) // a > 2**58
	case a.Gt(Two):
		return (a.Add((a.Mul(a).Sub(One)).Sqrt())).Log() // 2**58 > a > 2.0
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
func (a Float128) Atanh() Float128 {
	var (
		// Zero = 0.0
		Zero = Float128{}

		// Half = 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}

		// One = 1.0
		One = Float128(uvone128)

		// NearZero = 2**-58
		NearZero = Float128{0x3fc5_0000_0000_0000, 0x0000_0000_0000_0000}
	)

	// special cases
	switch {
	case a.Lt(One.Neg()) || a.Gt(One) || a.IsNaN():
		return NewFloat128NaN()
	case a.Eq(One):
		return NewFloat128Inf(1)
	case a.Eq(One.Neg()):
		return NewFloat128Inf(-1)
	}
	sign := false
	if a.Lt(Zero) {
		a = a.Neg()
		sign = true
	}
	var temp Float128
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
