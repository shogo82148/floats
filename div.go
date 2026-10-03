package floats

import (
	"math"
	"math/bits"

	"github.com/shogo82148/ints"
)

// div3by2 divides (u2, u1, u0) by (d1, d0) using Knuth's Algorithm D.
// It returns the quotient q and the remainder (r1, r0).
// d1 must be normalized (its most significant bit is set), and (u2, u1) must be less than (d1, d0).
func div3by2(u2, u1, u0, d1, d0 uint64) (q, r1, r0 uint64) {
	// estimate the quotient
	q = math.MaxUint64
	if u2 < d1 {
		var rhat uint64
		q, rhat = bits.Div64(u2, u1, d1)
		for {
			ph, pl := bits.Mul64(q, d0)
			if ph < rhat || (ph == rhat && pl <= u0) {
				break
			}
			q--
			rhat += d1
			if rhat < d1 {
				// rhat overflowed
				break
			}
		}
	}

	// r = u - q*d
	p0h, p0l := bits.Mul64(q, d0)
	p1h, p1l := bits.Mul64(q, d1)
	w1, c := bits.Add64(p0h, p1l, 0)
	w2 := p1h + c
	var borrow uint64
	r0, borrow = bits.Sub64(u0, p0l, 0)
	r1, borrow = bits.Sub64(u1, w1, borrow)
	r2, borrow := bits.Sub64(u2, w2, borrow)

	// q is too large; add back.
	for borrow != 0 {
		r0, c = bits.Add64(r0, d0, 0)
		r1, c = bits.Add64(r1, d1, c)
		r2, c = bits.Add64(r2, 0, c)
		q--
		borrow ^= c
	}
	return
}

// div5by4 divides (r, u4) by d using Knuth's Algorithm D.
// r and d are big-endian; r[0] is the most significant word.
// It returns the quotient q, and stores the remainder into r.
// d[0] must be normalized (its most significant bit is set), and r must be less than d.
func div5by4(r *ints.Uint256, u4 uint64, d ints.Uint256) (q uint64) {
	u0, u1, u2, u3 := r[0], r[1], r[2], r[3]

	// estimate the quotient
	q = math.MaxUint64
	if u0 < d[0] {
		var rhat uint64
		q, rhat = bits.Div64(u0, u1, d[0])
		for {
			ph, pl := bits.Mul64(q, d[1])
			if ph < rhat || (ph == rhat && pl <= u2) {
				break
			}
			q--
			rhat += d[0]
			if rhat < d[0] {
				// rhat overflowed
				break
			}
		}
	}

	// u = u - q*d
	var c, borrow uint64
	p3h, p3l := bits.Mul64(q, d[3])
	p2h, p2l := bits.Mul64(q, d[2])
	p1h, p1l := bits.Mul64(q, d[1])
	p0h, p0l := bits.Mul64(q, d[0])
	w3, c := bits.Add64(p2l, p3h, 0)
	w2, c := bits.Add64(p1l, p2h, c)
	w1, c := bits.Add64(p0l, p1h, c)
	w0 := p0h + c
	u4, borrow = bits.Sub64(u4, p3l, 0)
	u3, borrow = bits.Sub64(u3, w3, borrow)
	u2, borrow = bits.Sub64(u2, w2, borrow)
	u1, borrow = bits.Sub64(u1, w1, borrow)
	u0, borrow = bits.Sub64(u0, w0, borrow)

	// q is too large; add back.
	for borrow != 0 {
		u4, c = bits.Add64(u4, d[3], 0)
		u3, c = bits.Add64(u3, d[2], c)
		u2, c = bits.Add64(u2, d[1], c)
		u1, c = bits.Add64(u1, d[0], c)
		u0, c = bits.Add64(u0, 0, c)
		q--
		borrow ^= c
	}
	*r = ints.Uint256{u1, u2, u3, u4}
	return q
}

// quo256by128 returns q = floor((n << 128) / d) and whether the remainder is not zero.
// d must be normalized (its most significant bit is set), and n must be less than d.
func quo256by128(n ints.Uint128, d ints.Uint128) (q ints.Uint128, inexact bool) {
	q1, r1, r0 := div3by2(n[0], n[1], 0, d[0], d[1])
	q0, r1, r0 := div3by2(r1, r0, 0, d[0], d[1])
	return ints.Uint128{q1, q0}, r1|r0 != 0
}

// quo512by256 returns q = floor((n << 256) / d) and whether the remainder is not zero.
// d must be normalized (its most significant bit is set), and n must be less than d.
func quo512by256(n ints.Uint256, d ints.Uint256) (q ints.Uint256, inexact bool) {
	q[0] = div5by4(&n, 0, d)
	q[1] = div5by4(&n, 0, d)
	q[2] = div5by4(&n, 0, d)
	q[3] = div5by4(&n, 0, d)
	return q, !n.IsZero()
}
