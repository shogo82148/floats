package floats

import (
	"math/bits"

	"github.com/shogo82148/ints"
)

// ynExt128 is a floating-point number with the 128-bit significand, which is used by the recurrence of Yn.
// The rounding errors of Float128, whose significand is 113 bits, are accumulated in the recurrence;
// the extended precision makes them negligible.
//
// The value is (-1)**neg * m * 2**(e-127), where m is normalized, that is, m >= 2**127, or m = 0.
// The exponent e is not limited.
type ynExt128 struct {
	m   ints.Uint128
	e   int
	neg bool
}

// ynExt128FromFloat128 converts a finite Float128 to ynExt128.
func ynExt128FromFloat128(a Float128) ynExt128 {
	if a.IsZero() {
		return ynExt128{}
	}
	sign, e, frac := a.normalize()
	return ynExt128{m: frac.Lsh(128 - 1 - shift128), e: e, neg: sign != 0}
}

// ynExt128FromUint converts a positive integer to ynExt128.
func ynExt128FromUint(v uint64) ynExt128 {
	l := uint(bits.LeadingZeros64(v))
	return ynExt128{m: ints.Uint128{v << l, 0}, e: 63 - int(l)}
}

// Float128 converts x to Float128 rounded to nearest even.
func (x ynExt128) Float128() Float128 {
	if x.m.IsZero() {
		return Float128{}
	}
	frac := roundToNearestEven128(x.m, 128-1-shift128)
	e := x.e
	if frac[0]>>(shift128-64+1) != 0 {
		// carry-out
		frac = frac.Rsh(1)
		e++
	}
	// f is frac * 2**(-shift128), and it is in [1, 2).
	frac[0] &^= 1 << (shift128 - 64)
	frac[0] |= uint64(bias128) << (shift128 - 64)
	if x.neg {
		frac[0] |= signMask128[0]
	}
	// Ldexp handles overflow and underflow.
	return Float128(frac).Ldexp(e)
}

// ynExt128Inv returns 1/x for x != 0.
func ynExt128Inv(x ynExt128) ynExt128 {
	if x.m == (ints.Uint128{1 << 63, 0}) {
		// x is a power of two.
		return ynExt128{m: x.m, e: -x.e, neg: x.neg}
	}
	// q = floor(2**255 / m) is in (2**127, 2**128).
	d := ints.Uint256{2: x.m[0], 3: x.m[1]}
	q, r := ints.Uint256{1 << 63}.DivMod(d)
	m := ints.Uint128{q[2], q[3]}
	e := -x.e - 1
	// round to nearest. r < m, so 2r does not overflow in 256 bits.
	if r.Lsh(1).Cmp(d) >= 0 {
		m = m.Add(ints.Uint128{0, 1})
		if m.IsZero() {
			m = ints.Uint128{1 << 63, 0}
			e++
		}
	}
	return ynExt128{m: m, e: e, neg: x.neg}
}

// ynExt128Mul returns a * b rounded to nearest even.
func ynExt128Mul(a, b ynExt128) ynExt128 {
	if a.m.IsZero() || b.m.IsZero() {
		return ynExt128{}
	}
	// p is in [2**254, 2**256).
	p := a.m.Mul256(b.m)
	e := a.e + b.e + 1
	if p[0]>>63 == 0 {
		p = p.Lsh(1)
		e--
	}
	m := ints.Uint128{p[0], p[1]}
	// round to nearest even by the lower half.
	const half = 1 << 63
	if p[2] > half || (p[2] == half && (p[3] != 0 || m[1]&1 != 0)) {
		m = m.Add(ints.Uint128{0, 1})
		if m.IsZero() {
			m = ints.Uint128{1 << 63, 0}
			e++
		}
	}
	return ynExt128{m: m, e: e, neg: a.neg != b.neg}
}

// ynExt128Add returns a + b rounded to nearest even.
func ynExt128Add(a, b ynExt128) ynExt128 {
	if b.m.IsZero() {
		return a
	}
	if a.m.IsZero() {
		return b
	}
	// make |a| >= |b|.
	if a.e < b.e || (a.e == b.e && a.m.Cmp(b.m) < 0) {
		a, b = b, a
	}
	diff := a.e - b.e
	if diff > 130 {
		// |b| is less than the quarter of the ulp of a.
		return a
	}
	d := uint(diff)

	// A and B are the significands in 256 bits with the 126 bits of headroom for the shift of B.
	// The bits of B shifted out are the sticky bit.
	A := ints.Uint256{2: a.m[0], 3: a.m[1]}.Lsh(126)
	B := ints.Uint256{2: b.m[0], 3: b.m[1]}.Lsh(126)
	sticky := false
	if d > 0 {
		sticky = !B.Lsh(256 - d).IsZero()
		B = B.Rsh(d)
	}
	var R ints.Uint256
	if a.neg == b.neg {
		R = A.Add(B)
	} else {
		R = A.Sub(B)
		if sticky {
			// the exact B is greater than B by the fraction.
			R = R.Sub(ints.Uint256{3: 1})
		}
	}
	if R.IsZero() && !sticky {
		return ynExt128{}
	}

	// normalize R. R * 2**(a.e-253) is the value.
	bl := R.BitLen()
	R = R.Lsh(uint(256 - bl))
	e := a.e + bl - 254
	m := ints.Uint128{R[0], R[1]}
	const half = 1 << 63
	if R[2] > half || (R[2] == half && (sticky || R[3] != 0 || m[1]&1 != 0)) {
		m = m.Add(ints.Uint128{0, 1})
		if m.IsZero() {
			m = ints.Uint128{1 << 63, 0}
			e++
		}
	}
	return ynExt128{m: m, e: e, neg: a.neg}
}
