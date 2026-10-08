package floats

import "github.com/shogo82148/ints"

// Hypot256 returns [Sqrt](p*p + q*q), taking care to avoid
// unnecessary overflow and underflow.
//
// Special cases are:
//
//	Hypot256(±Inf, q) = +Inf
//	Hypot256(p, ±Inf) = +Inf
//	Hypot256(NaN, q) = NaN
//	Hypot256(p, NaN) = NaN
func Hypot256(p, q Float256) Float256 {
	p = p.Abs()
	q = q.Abs()

	// special cases
	switch {
	case p.IsInf(1) || q.IsInf(1):
		return NewFloat256Inf(1)
	case p.IsNaN() || q.IsNaN():
		return NewFloat256NaN()
	}

	if p.Lt(q) {
		p, q = q, p
	}
	if q.IsZero() {
		return p
	}

	// p = fracP * 2^(expP-236), q = fracQ * 2^(expQ-236), expP >= expQ.
	// Compute S = fracP^2 + fracQ^2 * 2^(-2d) with d = expP - expQ,
	// then hypot = sqrt(S) * 2^(expP-236), rounded once.
	_, expP, fracP := p.normalize()
	_, expQ, fracQ := q.normalize()
	sum := fracP.Mul512(fracP)
	sq := fracQ.Mul512(fracQ)
	sticky := false
	if d := uint(2 * (expP - expQ)); d >= 512 {
		sticky = true
		sq = ints.Uint512{}
	} else if d > 0 {
		sticky = !sq.Lsh(512 - d).IsZero()
		sq = sq.Rsh(d)
	}
	sum = sum.Add(sq) // sum is in [2^472, 2^475)

	// n = sum * 2^(2m) is in [2^510, 2^512), so that its root has 256 bits.
	// The sticky bit is the lowest bit of n; n is then odd, and the root is
	// reported as inexact even if n is a perfect square.
	m := (512 - uint(sum.BitLen())) / 2
	n := sum.Lsh(2 * m)
	if sticky {
		n[7] |= 1
	}
	root, inexact := sqrtRem512(n)

	// hypot = root * 2^(expP-236-m), and root is in [2^255, 2^256).
	// Its leading bit has the exponent e.
	if inexact {
		root[3] |= 1
	}
	e := expP + 255 - 236 - int(m)
	exp := max(e, 1-bias256) // the result may be subnormal
	if exp >= mask256-bias256 {
		return Float256(uvinf256)
	}

	// keep shift256+1 bits, or fewer if the result is subnormal.
	shift := uint(256-(shift256+1)) + uint(exp-e)
	frac := roundToNearestEven256(root, shift)

	// The hidden bit of frac is added to the exponent.
	// It also handles carry-out caused by rounding, and subnormal results.
	frac[0] += uint64(exp-1+bias256) << (shift256 - 192)
	return Float256(frac)
}
