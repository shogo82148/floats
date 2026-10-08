package floats

import "github.com/shogo82148/ints"

// Hypot128 returns [Sqrt](p*p + q*q), taking care to avoid
// unnecessary overflow and underflow.
//
// Special cases are:
//
//	Hypot128(±Inf, q) = +Inf
//	Hypot128(p, ±Inf) = +Inf
//	Hypot128(NaN, q) = NaN
//	Hypot128(p, NaN) = NaN
func Hypot128(p, q Float128) Float128 {
	p = p.Abs()
	q = q.Abs()

	// special cases
	switch {
	case p.IsInf(1) || q.IsInf(1):
		return NewFloat128Inf(1)
	case p.IsNaN() || q.IsNaN():
		return NewFloat128NaN()
	}

	if p.Lt(q) {
		p, q = q, p
	}
	if q.IsZero() {
		return p
	}

	// p = fracP * 2^(expP-112), q = fracQ * 2^(expQ-112), expP >= expQ.
	// Compute S = fracP^2 + fracQ^2 * 2^(-2d) with d = expP - expQ,
	// then hypot = sqrt(S) * 2^(expP-112), rounded once.
	_, expP, fracP := p.normalize()
	_, expQ, fracQ := q.normalize()
	sum := fracP.Mul256(fracP)
	sq := fracQ.Mul256(fracQ)
	sticky := false
	if d := uint(2 * (expP - expQ)); d >= 256 {
		sticky = true
		sq = ints.Uint256{}
	} else if d > 0 {
		sticky = !sq.Lsh(256 - d).IsZero()
		sq = sq.Rsh(d)
	}
	sum = sum.Add(sq) // sum is in [2^224, 2^227)

	// n = sum * 2^(2m) is in [2^254, 2^256), so that its root has 128 bits.
	// The sticky bit is the lowest bit of n; n is then odd, and the root is
	// reported as inexact even if n is a perfect square.
	m := (256 - uint(sum.BitLen())) / 2
	n := sum.Lsh(2 * m)
	if sticky {
		n[3] |= 1
	}
	root, rem := sqrtRem256(n)

	// hypot = root * 2^(expP-112-m), and root is in [2^127, 2^128).
	// Its leading bit has the exponent e.
	if !rem.IsZero() {
		root[1] |= 1
	}
	e := expP + 127 - 112 - int(m)
	exp := max(e, 1-bias128) // the result may be subnormal
	if exp >= mask128-bias128 {
		return Float128(uvinf128)
	}

	// keep shift128+1 bits, or fewer if the result is subnormal.
	shift := uint(128-(shift128+1)) + uint(exp-e)
	frac := roundToNearestEven128(root, shift)

	// The hidden bit of frac is added to the exponent.
	// It also handles carry-out caused by rounding, and subnormal results.
	frac[0] += uint64(exp-1+bias128) << (shift128 - 64)
	return Float128(frac)
}
