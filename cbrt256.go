package floats

import "github.com/shogo82148/ints"

// Cbrt returns the cube root of a.
//
// Special cases are:
//
//	±0.Cbrt() = ±0
//	±Inf.Cbrt() = ±Inf
//	NaN.Cbrt() = NaN
func (a Float256) Cbrt() Float256 {
	// special cases
	switch {
	case a.IsZero() || a.IsInf(0) || a.IsNaN():
		return a
	}

	// |a| = m × 2**(exp-236), where m is a 237-bit integer in [2**236, 2**237).
	sign, exp, m := a.normalize()

	// |a| = 2**(3q) × X, where X = mx × 2**-236 is in [1, 8), and exp = 3q + r with r = 0, 1, 2.
	// cbrt(a) = 2**q × cbrt(X), and cbrt(X) is in [1, 2).
	// exp + 262380 is positive and divisible by 3 for the shift, so that the quotient is floored.
	t := exp + 262380
	q, r := t/3-87460, uint(t%3)

	// the estimate T × 2**-112 of cbrt(X) with an error of about 2**-112,
	// from the cube root of the top 113 bits of X in Float128.
	x128 := Float128{
		uint64(bias128+int(r))<<(shift128-64) | (m[0]<<4|m[1]>>60)&fracMask128[0], // the upper 112 bits of the fraction
		m[1]<<4 | m[2]>>60,
	}
	c := x128.Cbrt()
	est := ints.Uint128{1 << (shift128 - 64) << 1, 0} // 2 if the cube root is rounded up to 2
	if int(c[0]>>(shift128-64)) == bias128 {
		est = ints.Uint128{c[0]&fracMask128[0] | 1<<(shift128-64), c[1]}
	}

	// Halley's iteration: t' = t - t × (t³ - X) / (2t³ + X) converges cubically,
	// so that one step makes the error much smaller than 2**-256.
	// t³ = T³ × 2**-336 and X = xs × 2**-336, where xs = m × 2**(100+r), are computed exactly.
	t3 := est.Mul256(est).Mul512(est.Uint256())
	xs := m.Uint512().Lsh(100 + r)
	over := t3.Cmp(xs) >= 0
	var d ints.Uint512
	if over {
		d = t3.Sub(xs)
	} else {
		d = xs.Sub(t3)
	}
	den := t3.Add(t3).Add(xs)

	// the correction T × d / den × 2**144 is about 2**145, rounded down.
	corr := d.Uint256().Mul512(est.Uint256()).Lsh(144).Quo(den)

	// y = cbrt(X) × 2**256 in [2**256, 2**257), with an error less than 2.
	y := est.Uint512().Lsh(144)
	if over {
		y = y.Sub(corr)
	} else {
		y = y.Add(corr)
	}

	// round y to 237 bits: the result significand is in [2**236, 2**237].
	// If the 20 bits below are close to the half, the exact comparison is needed.
	sig := y.Rsh(20)
	switch low := y[7] & 0xfffff; {
	case low > 0x80000+4:
		sig = sig.Add(ints.Uint512{7: 1})
	case low >= 0x80000-4:
		// round up if (sig + 1/2)³ < X × 2**708, i.e., (2 sig + 1)³ < m × 2**(475+r) (no tie is possible).
		u := sig.Lsh(1).Add(ints.Uint512{7: 1})
		cube := u.Mul1024(u).Mul(u.Uint1024())
		if m.Uint512().Uint1024().Lsh(475+r).Cmp(cube) > 0 {
			sig = sig.Add(ints.Uint512{7: 1})
		}
	}

	// sign, exponent, and fraction. If the significand is rounded up to 2**237, it carries to the exponent.
	hi := sign | uint64(q+bias256)<<(shift256-192)
	hi += sig[4] - 1<<(shift256-192)
	return Float256{hi, sig[5], sig[6], sig[7]}
}
