package floats

import (
	"math"
	"math/bits"

	"github.com/shogo82148/ints"
)

// Hypot64 returns [Sqrt](p*p + q*q), taking care to avoid
// unnecessary overflow and underflow.
//
// Special cases are:
//
//	Hypot64(±Inf, q) = +Inf
//	Hypot64(p, ±Inf) = +Inf
//	Hypot64(NaN, q) = NaN
//	Hypot64(p, NaN) = NaN
func Hypot64(p, q Float64) Float64 {
	p = p.Abs()
	q = q.Abs()

	// special cases
	switch {
	case p.IsInf(1) || q.IsInf(1):
		return NewFloat64Inf(1)
	case p.IsNaN() || q.IsNaN():
		return NewFloat64NaN()
	}

	if p.Lt(q) {
		p, q = q, p
	}
	if q.IsZero() {
		return p
	}

	if r, ok := hypot64Fast(p.BuiltIn(), q.BuiltIn()); ok {
		return Float64(r)
	}

	// p = fracP * 2^(expP-52), q = fracQ * 2^(expQ-52), expP >= expQ.
	// Compute S = fracP^2 + fracQ^2 * 2^(-2d) with d = expP - expQ,
	// then hypot = sqrt(S) * 2^(expP-52), rounded once.
	_, expP, fracP := p.normalize()
	_, expQ, fracQ := q.normalize()
	hi, lo := bits.Mul64(fracP, fracP)
	sum := ints.Uint128{hi, lo}
	hi, lo = bits.Mul64(fracQ, fracQ)
	sq := ints.Uint128{hi, lo}
	sticky := false
	if d := uint(2 * (expP - expQ)); d >= 128 {
		sticky = true
		sq = ints.Uint128{}
	} else if d > 0 {
		sticky = !sq.Lsh(128 - d).IsZero()
		sq = sq.Rsh(d)
	}
	sum = sum.Add(sq) // sum is in [2^104, 2^107)

	// n = sum * 2^(2m) is in [2^126, 2^128), so that its root has 64 bits.
	// The sticky bit is the lowest bit of n; n is then odd, and the root is
	// reported as inexact even if n is a perfect square.
	m := (128 - uint(sum.BitLen())) / 2
	n := sum.Lsh(2 * m)
	if sticky {
		n[1] |= 1
	}
	root := isqrt128(n[0], n[1])
	if h, l := bits.Mul64(root, root); h != n[0] || l != n[1] {
		root |= 1
	}

	// hypot = root * 2^(expP-52-m), and root is in [2^63, 2^64).
	// Its leading bit has the exponent e.
	e := expP + 63 - 52 - int(m)
	exp := max(e, 1-bias64) // the result may be subnormal
	if exp >= mask64-bias64 {
		return NewFloat64Inf(1)
	}

	// keep shift64+1 bits, or fewer if the result is subnormal.
	shift := uint(64-(shift64+1)) + uint(exp-e)
	frac := root >> shift
	r := root << (64 - shift)
	const half = 1 << 63
	if r > half || (r == half && frac&1 != 0) {
		frac++
	}

	// The hidden bit of frac is added to the exponent.
	// It also handles carry-out caused by rounding, and subnormal results.
	frac += uint64(exp-1+bias64) << shift64
	return Float64(math.Float64frombits(frac))
}

// hypot64Fast computes hypot(x, y) for x >= y > 0 with floating-point
// arithmetic only. It reports false if it cannot prove that the result is
// correctly rounded, in which case the caller must use the exact algorithm.
func hypot64Fast(x, y float64) (float64, bool) {
	// make sure that x*x and y*y neither overflow nor lose their low bits to underflow.
	if x > 0x1p500 || y < 0x1p-450 {
		return 0, false
	}

	// x*x + y*y = hi + lo, where the relative error of lo is about 2^-105.
	// float64() prevents the compiler from fusing the products into the sums.
	xx := float64(x * x)
	xxl := math.FMA(x, x, -xx)
	yy := float64(y * y)
	yyl := math.FMA(y, y, -yy)
	hi := xx + yy
	bv := hi - xx
	t := (xx - (hi - bv)) + (yy - bv)
	lo := t + (xxl + yyl)

	// r = sqrt(hi) is correctly rounded, so hi - r*r is computed exactly.
	// The exact root is r + d/(2r) within about 2^-50 ulp.
	r := math.Sqrt(hi)
	d := math.FMA(-r, r, hi) + lo

	// ulp of r. Give up for powers of two, whose lower neighbour has a smaller ulp.
	rb := math.Float64bits(r)
	if rb&fracMask64 == 0 {
		return 0, false
	}
	ulp := math.Float64frombits(rb&(mask64<<shift64)) * 0x1p-52
	k := d / (2 * r * ulp)
	const tol = 0x1p-30
	switch {
	case k > -0.5+tol && k < 0.5-tol:
		return r, true
	case k > 0.5+tol && k < 1.5-tol:
		return r + ulp, true
	case k < -0.5-tol && k > -1.5+tol:
		return r - ulp, true
	}
	return 0, false
}
