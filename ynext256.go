package floats

import (
	"math/bits"

	"github.com/shogo82148/ints"
)

// ynExt256 is a floating-point number with the 256-bit significand, which is used by the recurrence of Yn.
// The rounding errors of Float256, whose significand is 237 bits, are accumulated in the recurrence;
// the extended precision makes them negligible.
//
// The value is (-1)**neg * m * 2**(e-255), where m is normalized, that is, m >= 2**255, or m = 0.
// The exponent e is not limited.
type ynExt256 struct {
	m   ints.Uint256
	e   int
	neg bool
}

// ynExt256FromFloat256 converts a finite Float256 to ynExt256.
func ynExt256FromFloat256(a Float256) ynExt256 {
	if a.IsZero() {
		return ynExt256{}
	}
	sign, e, frac := a.normalize()
	return ynExt256{m: frac.Lsh(256 - 1 - shift256), e: e, neg: sign != 0}
}

// ynExt256FromUint converts a positive integer to ynExt256.
func ynExt256FromUint(v uint64) ynExt256 {
	l := uint(bits.LeadingZeros64(v))
	return ynExt256{m: ints.Uint256{v << l, 0, 0, 0}, e: 63 - int(l)}
}

// Float256 converts x to Float256 rounded to nearest even.
func (x ynExt256) Float256() Float256 {
	if x.m.IsZero() {
		return Float256{}
	}
	frac := roundToNearestEven256(x.m, 256-1-shift256)
	e := x.e
	if frac[0]>>(shift256-192+1) != 0 {
		// carry-out
		frac = frac.Rsh(1)
		e++
	}
	// f is frac * 2**(-shift256), and it is in [1, 2).
	frac[0] &^= 1 << (shift256 - 192)
	frac[0] |= uint64(bias256) << (shift256 - 192)
	if x.neg {
		frac[0] |= signMask256[0]
	}
	// Ldexp handles overflow and underflow.
	return Float256(frac).Ldexp(e)
}

// ynExt256Inv returns 1/x for x != 0.
func ynExt256Inv(x ynExt256) ynExt256 {
	if x.m == (ints.Uint256{1 << 63, 0, 0, 0}) {
		// x is a power of two.
		return ynExt256{m: x.m, e: -x.e, neg: x.neg}
	}
	// q = floor(2**511 / m) is in (2**255, 2**256).
	q, r := ints.Uint512{1 << 63}.DivMod(ints.Uint512{4: x.m[0], 5: x.m[1], 6: x.m[2], 7: x.m[3]})
	m := ints.Uint256{q[4], q[5], q[6], q[7]}
	e := -x.e - 1
	// round to nearest. r < m, so 2r does not overflow in 512 bits.
	if r.Lsh(1).Cmp(ints.Uint512{4: x.m[0], 5: x.m[1], 6: x.m[2], 7: x.m[3]}) >= 0 {
		m = m.Add(ints.Uint256{0, 0, 0, 1})
		if m.IsZero() {
			m = ints.Uint256{1 << 63, 0, 0, 0}
			e++
		}
	}
	return ynExt256{m: m, e: e, neg: x.neg}
}

// ynExt256Mul returns a * b rounded to nearest even.
func ynExt256Mul(a, b ynExt256) ynExt256 {
	if a.m.IsZero() || b.m.IsZero() {
		return ynExt256{}
	}
	// p is in [2**510, 2**512).
	p := a.m.Mul512(b.m)
	e := a.e + b.e + 1
	if p[0]>>63 == 0 {
		p = ints.Uint512{
			p[0]<<1 | p[1]>>63, p[1]<<1 | p[2]>>63, p[2]<<1 | p[3]>>63, p[3]<<1 | p[4]>>63,
			p[4]<<1 | p[5]>>63, p[5]<<1 | p[6]>>63, p[6]<<1 | p[7]>>63, p[7] << 1,
		}
		e--
	}
	m := ints.Uint256{p[0], p[1], p[2], p[3]}
	// round to nearest even by the lower half.
	const half = 1 << 63
	if p[4] > half || (p[4] == half && (p[5]|p[6]|p[7] != 0 || m[3]&1 != 0)) {
		m = m.Add(ints.Uint256{0, 0, 0, 1})
		if m.IsZero() {
			m = ints.Uint256{1 << 63, 0, 0, 0}
			e++
		}
	}
	return ynExt256{m: m, e: e, neg: a.neg != b.neg}
}

// ynExt256MulUint returns a * v rounded to nearest even for v != 0.
func ynExt256MulUint(a ynExt256, v uint64) ynExt256 {
	if a.m.IsZero() {
		return ynExt256{}
	}
	// p = a.m * v is in 320 bits: p4 is the high word, and the rest are the low 256 bits.
	var p [5]uint64
	var c uint64
	for i := 3; i >= 0; i-- {
		hi, lo := bits.Mul64(a.m[i], v)
		var carry uint64
		p[i+1], carry = bits.Add64(lo, c, 0)
		c = hi + carry
	}
	p[0] = c
	// normalize p so that its most significant bit is set.
	l := uint(bits.LeadingZeros64(p[0]))
	e := a.e + 64 - int(l)
	if l != 0 {
		for i := range 4 {
			p[i] = p[i]<<l | p[i+1]>>(64-l)
		}
		p[4] <<= l
	}
	m := ints.Uint256{p[0], p[1], p[2], p[3]}
	// the bits shifted out are in p[4], aligned to the most significant bit.
	const half = 1 << 63
	if p[4] > half || (p[4] == half && m[3]&1 != 0) {
		m = m.Add(ints.Uint256{0, 0, 0, 1})
		if m.IsZero() {
			m = ints.Uint256{1 << 63, 0, 0, 0}
			e++
		}
	}
	return ynExt256{m: m, e: e, neg: a.neg}
}

// ynExt256Add returns a + b rounded to nearest even.
func ynExt256Add(a, b ynExt256) ynExt256 {
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
	if diff > 258 {
		// |b| is less than the quarter of the ulp of a.
		return a
	}
	d := uint(diff)

	// A and B are the significands in 512 bits with the 254 bits of headroom for the shift of B.
	// The bits of B shifted out are the sticky bit.
	// m << 254 = (m << 62) << 192, and the low 192 bits are zero.
	A := lsh512(ints.Uint512{4: a.m[0], 5: a.m[1], 6: a.m[2], 7: a.m[3]}, 254)
	B := lsh512(ints.Uint512{4: b.m[0], 5: b.m[1], 6: b.m[2], 7: b.m[3]}, 254)
	sticky := false
	if d > 0 {
		sticky = lsh512(B, 512-d) != (ints.Uint512{})
		B = rsh512(B, d)
	}
	var R ints.Uint512
	if a.neg == b.neg {
		R = A.Add(B)
	} else {
		R = A.Sub(B)
		if sticky {
			// the exact B is greater than B by the fraction.
			R = R.Sub(ints.Uint512{7: 1})
		}
	}
	if R.IsZero() && !sticky {
		return ynExt256{}
	}

	// normalize R. R * 2**(a.e-509) is the value.
	var lz uint
	for lz = 0; lz < 512 && R[lz/64] == 0; lz += 64 {
	}
	if lz < 512 {
		lz += uint(bits.LeadingZeros64(R[lz/64]))
	}
	bl := 512 - int(lz)
	R = lsh512(R, lz)
	e := a.e + bl - 510
	m := ints.Uint256{R[0], R[1], R[2], R[3]}
	const half = 1 << 63
	if R[4] > half || (R[4] == half && (sticky || R[5]|R[6]|R[7] != 0 || m[3]&1 != 0)) {
		m = m.Add(ints.Uint256{0, 0, 0, 1})
		if m.IsZero() {
			m = ints.Uint256{1 << 63, 0, 0, 0}
			e++
		}
	}
	return ynExt256{m: m, e: e, neg: a.neg}
}
