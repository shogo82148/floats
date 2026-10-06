package floats

import (
	"github.com/shogo82148/ints"
)

// Lgamma returns the natural logarithm and sign (-1 or +1) of Gamma(a).
//
// Special cases are:
//
//	Lgamma(+Inf) = +Inf
//	Lgamma(0) = +Inf
//	Lgamma(-integer) = +Inf
//	Lgamma(-Inf) = -Inf
//	Lgamma(NaN) = NaN
func (a Float128) Lgamma() (Float128, int) {
	switch {
	case a.IsNaN() || a.IsInf(0):
		return a, 1
	case a.IsZero() || isNegInt128(a):
		return NewFloat128Inf(1), 1
	}

	// |a| = m × 2**(exp-112), where m is a 113-bit integer in [2**112, 2**113).
	sign, exp, m := a.normalize()
	neg := sign != 0
	switch {
	case exp >= 15:
		return lgamma128Large(a, neg)
	case exp >= 11:
		// 2048 <= |a| < 2**15
		if !neg {
			return lgamma128FromFix256(false, gammaLogGamma256(lgamma128Fix256(exp, m))), 1
		}
		sign, v := lgamma320Neg(exp, m)
		return lgamma128FromFix256(v.neg, v.v), sign
	case !neg && (a == Float128(uvone128) || a == Float128{0x4000_0000_0000_0000, 0}):
		return Float128{}, 1 // the zeros of Lgamma
	}

	// the fixed point calculation with 192 fractional bits, which is used unless the result is close to zero.
	sgn := 1
	var t lgammaFix
	switch {
	case exp < -40:
		t = lgamma128Tiny(neg, exp, m)
		if neg {
			sgn = -1
		}
	case !neg:
		t = lgamma128Pos(m.Uint256().Lsh(uint(80 + exp)))
	default:
		sgn, t = lgamma128Neg(exp, m)
	}
	if t.v[0] != 0 || t.v[1] >= 1<<58 {
		return lgamma128FromFix(t), sgn
	}

	// |Lgamma(a)| < 2**-6. The relative error of the calculation above may be too large, so use 320 fractional bits.
	if !neg {
		v := lgamma320Pos(lgamma128Fix256(exp, m))
		return lgamma128FromFix256(v.neg, v.v), 1
	}
	sgn, v := lgamma320Neg(exp, m)
	return lgamma128FromFix256(v.neg, v.v), sgn
}

// lgammaFix is a signed number in fixed point with 192 fractional bits in ints.Uint256,
// whose first word is the integer part.
type lgammaFix struct {
	neg bool
	v   ints.Uint256
}

func (a lgammaFix) add(b lgammaFix) lgammaFix {
	if a.neg == b.neg {
		return lgammaFix{a.neg, a.v.Add(b.v)}
	}
	if a.v.Cmp(b.v) >= 0 {
		return lgammaFix{a.neg, a.v.Sub(b.v)}
	}
	return lgammaFix{b.neg, b.v.Sub(a.v)}
}

func (a lgammaFix) negate() lgammaFix {
	return lgammaFix{!a.neg, a.v}
}

// lgamma128FromFix returns a as Float128. a must not be zero.
func lgamma128FromFix(a lgammaFix) Float128 {
	var sign uint64
	if a.neg {
		sign = signMask128[0]
	}
	lz := a.v.LeadingZeros()
	v := a.v.Lsh(uint(lz))
	return fixToFloat128(sign, v[0], v[1], v[2], v[3] != 0, -128-int(lz))
}

// lgammaMulLn2 returns k × log(2) in fixed point with 192 fractional bits.
func lgammaMulLn2(k int) ints.Uint256 {
	return ints.Uint256{3: uint64(k)}.Mul(ints.Uint256{0, asinhLn2Fix128[0], asinhLn2Fix128[1], asinhLn2Fix128[2]})
}

// lgammaLogMant returns log(v/2**192) for v in [2**192, 2**193) in fixed point with 192 fractional bits.
func lgammaLogMant(v ints.Uint256) ints.Uint256 {
	if v != gammaOne && gammaLogBucket0(v) {
		// logKernel128 is not accurate for 1 <= v/2**192 < 1 + 2**-8. log(v) = log(3 v/2) - log(3/2).
		return logKernel128(0, v.Add(v.Rsh(1))).Sub(ints.Uint256(lgamma128Ln1p5))
	}
	return logKernel128(0, v)
}

// lgammaLogPair returns log(P) for P = pm × 2**(pe-192), where pm is in [2**192, 2**193).
func lgammaLogPair(pm ints.Uint256, pe int) lgammaFix {
	t := lgammaFix{false, lgammaLogMant(pm)}
	if pe >= 0 {
		return t.add(lgammaFix{false, lgammaMulLn2(pe)})
	}
	return t.add(lgammaFix{true, lgammaMulLn2(-pe)})
}

// lgamma128Pos returns Lgamma(y) for y = Y × 2**-192, 2**-40 <= y < 2049.
// The absolute error is less than 2**-145.
func lgamma128Pos(y ints.Uint256) lgammaFix {
	// Lgamma(y) = Lgamma(z) - log(y (y+1) ... (z-1)), where z = y + n >= 24.
	// log(z) is not accurate for 1 <= z/2**k < 1 + 2**-8, so z is moved out of the range.
	z, pm, pe := gammaRecur128(y)
	for gammaLogBucket0(z) {
		pm, pe = gammaNormalize(gammaMul(pm, z), pe)
		z = z.Add(gammaOne)
	}
	return lgammaFix{false, gammaLogGamma192(z)}.add(lgammaLogPair(pm, pe).negate())
}

// lgamma128Tiny returns Lgamma(x) for x = ±m × 2**(exp-112), where m is a 113-bit integer, and exp < -40.
func lgamma128Tiny(neg bool, exp int, m ints.Uint128) lgammaFix {
	// |x| in fixed point with 192 fractional bits.
	var x ints.Uint256
	if s := exp + 80; s >= 0 {
		x = m.Uint256().Lsh(uint(s))
	} else if s > -256 {
		x = m.Uint256().Rsh(uint(-s))
	}

	// Lgamma(x) = -log|x| + Lgamma(1+x), and -log|x| = -exp log(2) - log(m/2**112) is positive.
	r := lgammaMulLn2(-exp).Sub(lgammaLogMant(m.Uint256().Lsh(80)))

	// Lgamma(1+x) = -c1 x + c2 x**2 - c3 x**3, whose terms are positive for x < 0. The next term is less than 2**-160.
	c := &lgamma128Taylor
	p := ints.Uint256(c[2])
	if neg {
		for k := 1; k >= 0; k-- {
			p = ints.Uint256(c[k]).Add(gammaMul(x, p))
		}
		return lgammaFix{false, r.Add(gammaMul(x, p))}
	}
	for k := 1; k >= 0; k-- {
		p = ints.Uint256(c[k]).Sub(gammaMul(x, p))
	}
	return lgammaFix{false, r.Sub(gammaMul(x, p))}
}

// lgamma128Neg returns the sign of Gamma(x) and Lgamma(x) for x = -m × 2**(exp-112), where m is a 113-bit integer,
// and -40 <= exp < 11. x must not be an integer.
func lgamma128Neg(exp int, m ints.Uint128) (int, lgammaFix) {
	// |x| in fixed point with 192 fractional bits. It is exact.
	x := m.Uint256().Lsh(uint(80 + exp))

	// |x| = n + r, where n is the nearest integer and -1/2 <= r < 1/2. The reflection formula is
	// |Gamma(x)| = 1 / (|r| sinc(r) Gamma(1+|x|)), where sinc(r) = sin(pi r)/(pi r).
	half := ints.Uint256{0, 1 << 63, 0, 0}
	n := x.Add(half)
	n[1], n[2], n[3] = 0, 0, 0 // the integer part
	var r ints.Uint256
	rneg := x.Cmp(n) < 0
	if rneg {
		r = n.Sub(x)
	} else {
		r = x.Sub(n)
	}
	// sign of Gamma(x) = -(-1)**n sgn(r)
	sign := 1
	if (n[0]&1 != 0) == rneg {
		sign = -1
	}

	// Lgamma(x) = -log(|r| sinc(r)) - Lgamma(1+|x|).
	// |r| sinc(r) = sin(pi |r|)/pi is less than 1/3, so the logarithm is not close to zero.
	rm, re := gammaNormalize(r, 0)
	dm, de := gammaNormalize(gammaMul(rm, gammaSinc192(r)), re)
	t := lgammaLogPair(dm, de).negate().add(lgamma128Pos(x.Add(gammaOne)).negate())
	return sign, t
}

// lgammaFix256 is a signed number in fixed point with 320 fractional bits.
type lgammaFix256 struct {
	neg bool
	v   gammaFix256
}

func (a lgammaFix256) add(b lgammaFix256) lgammaFix256 {
	if a.neg == b.neg {
		return lgammaFix256{a.neg, a.v.add(b.v)}
	}
	if a.v.cmp(b.v) >= 0 {
		return lgammaFix256{a.neg, a.v.sub(b.v)}
	}
	return lgammaFix256{b.neg, b.v.sub(a.v)}
}

func (a lgammaFix256) negate() lgammaFix256 {
	return lgammaFix256{!a.neg, a.v}
}

// lgamma128Fix256 returns |x| for x = ±m × 2**(exp-112) in fixed point with 320 fractional bits.
func lgamma128Fix256(exp int, m ints.Uint128) gammaFix256 {
	return gammaFix256FromUint256(m.Uint256(), uint(exp+208))
}

// lgamma128FromFix256 returns ±v as Float128. v must not be zero.
func lgamma128FromFix256(neg bool, v gammaFix256) Float128 {
	bl := v.bitLen()
	var sign uint64
	if neg {
		sign = signMask128[0]
	}
	lz := 384 - bl
	l := v.shl(uint(lz))
	return fixToFloat128(sign, l[0], l[1], l[2], l[3]|l[4]|l[5] != 0, -128-lz)
}

// lgamma320LogPair returns log(P) for P = pm × 2**pe, where pm is in [1, 2).
func lgamma320LogPair(pm gammaFix256, pe int) lgammaFix256 {
	t := lgammaFix256{false, gammaLog256(pm)}
	if pe >= 0 {
		return t.add(lgammaFix256{false, gammaMulLn2(&gamma256Ln2By256, uint64(pe)*256)})
	}
	return t.add(lgammaFix256{true, gammaMulLn2(&gamma256Ln2By256, uint64(-pe)*256)})
}

// lgamma320Pos returns Lgamma(y) for 2**-40 <= y < 2**15 in fixed point with 320 fractional bits.
// The absolute error is less than 2**-285.
func lgamma320Pos(y gammaFix256) lgammaFix256 {
	// Lgamma(y) = Lgamma(z) - log(y (y+1) ... (z-1)), where z = y + n >= 48.
	z, pm, pe := gammaRecur256(y)
	return lgammaFix256{false, gammaLogGamma256(z)}.add(lgamma320LogPair(pm, pe).negate())
}

// lgamma320Neg returns the sign of Gamma(x) and Lgamma(x) for x = -m × 2**(exp-112), where m is a 113-bit integer,
// and -40 <= exp < 15. x must not be an integer.
func lgamma320Neg(exp int, m ints.Uint128) (int, lgammaFix256) {
	return lgamma320NegFix(lgamma128Fix256(exp, m))
}

// lgamma320NegFix returns the sign of Gamma(x) and Lgamma(x) for -2**15 < x < 0, where |x| is in fixed point with 320 fractional bits.
// x must not be an integer.
func lgamma320NegFix(x gammaFix256) (int, lgammaFix256) {
	// |x| = n + r, where n is the nearest integer and -1/2 <= r < 1/2.
	n := x.add(gammaFix256{0, 1 << 63})
	n = gammaFix256{n[0]} // the integer part
	var r gammaFix256
	rneg := x.cmp(n) < 0
	if rneg {
		r = n.sub(x)
	} else {
		r = x.sub(n)
	}
	// sign of Gamma(x) = -(-1)**n sgn(r)
	sign := 1
	if (n[0]&1 != 0) == rneg {
		sign = -1
	}

	// Lgamma(x) = -log(|r| sinc(r)) - Lgamma(1+|x|).
	rm, re := gammaNormalize256(r, 0)
	dm, de := gammaNormalize256(gammaMul6(rm, gammaSinc256(r)), re)
	return sign, lgamma320LogPair(dm, de).negate().add(lgamma320Pos(x.add(gammaOne256)).negate())
}

// lgamma128Large returns Lgamma(a) for |a| >= 2**15 by the calculation in Float256.
func lgamma128Large(a Float128, neg bool) (Float128, int) {
	q := a.Abs()
	if !neg {
		return lgamma256Stirling(q.Float256()).Float128(), 1
	}

	// Lgamma(x) = log(pi / |sin(pi x)|) - Lgamma(1-x), where |sin(pi x)| = sin(pi |r|) for x = n + r, |r| <= 1/2.
	// n and r are calculated exactly, since a is not an integer here.
	half := Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}
	one := Float128(uvone128)
	n := q.Floor()
	r := q.Sub(n)
	if r.Gt(half) {
		n = n.Add(one)
		r = q.Sub(n)
	}
	// The sign of Gamma(x) is negative if floor(x) = -ceil(|x|) is odd. ceil(|x|) is n+1 if r > 0, and n otherwise.
	sign := 1
	if isOddInt128(n) != !r.Signbit() {
		sign = -1
	}
	s := lgamma128Pi256.Mul(r.Abs().Float256()).Sin()
	y := lgamma128LnPi256.Sub(s.Log()).Sub(lgamma256Stirling(q.Float256().Add(Float256(uvone256))))
	return y.Float128(), sign
}

// lgamma256Stirling returns Lgamma(x) for x >= 2**15 by Stirling's series.
func lgamma256Stirling(x Float256) Float256 {
	half := Float256{0x3fff_e000_0000_0000, 0, 0, 0}
	// log(Gamma(x)) = (x-1/2) log(x) - x + log(2 pi)/2 + sum (-1)**(k+1) B(2k) / (2k (2k-1) x**(2k-1)).
	// The error of the sum of the first four terms is less than 2**-135 of the result for x >= 2**15.
	c := &lgamma128Stirling256
	w := Float256(uvone256).Quo(x.Mul(x))
	s := c[3]
	s = c[2].Sub(w.Mul(s))
	s = c[1].Sub(w.Mul(s))
	s = c[0].Sub(w.Mul(s))
	return x.Sub(half).Mul(x.Log()).Sub(x).Add(lgamma128HalfLn2Pi256).Add(s.Quo(x))
}
