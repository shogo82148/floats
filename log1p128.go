package floats

// Log1p returns the natural logarithm of 1 plus its argument a.
// It is more accurate than [Log](1 + a) when a is near zero.
//
// Special cases are:
//
//	+Inf.Log1p() = +Inf
//	±0.Log1p() = ±0
//	-1.Log1p() = -Inf
//	(a < -1).Log1p() = NaN
//	NaN.Log1p() = NaN
func (a Float128) Log1p() Float128 {
	// special cases
	switch {
	case a.Lt(Float128(uvone128).Neg()) || a.IsNaN(): // includes -Inf
		return NewFloat128NaN()
	case a.Eq(Float128(uvone128).Neg()):
		return NewFloat128Inf(-1)
	case a.IsInf(1):
		return NewFloat128Inf(1)
	case a.IsZero():
		return a
	}

	// a = ±m × 2**(exp-112), where m = (m1:m0) is a 113-bit integer in [2**112, 2**113).
	sign, exp, frac := a.normalize()
	if exp < -115 {
		// log(1+a) = a × (1 - a/2 + ...) rounds to a when |a| < 2**-115.
		return a
	}

	if exp < -8 {
		// |a| < 2**-8. log(1+a) = a × G(w) = a - a² × H(w), where w = -a,
		// G(w) = 1 + w/2 + w²/3 + ... = 1 + w × H(w), and H(w) = 1/2 + w/3 + w²/4 + ...
		// H is computed in fixed point, and the product a² × H as a double-Float128,
		// so that the result is accurate even if a² × H is near a half ulp of a.
		// |a| = rnorm × 2**(exp-191), where rnorm = m × 2**79 is in [2**191, 2**192).
		// |w|, at scale 2**-192, is rnorm shifted right by (-1-exp).
		m1, m0 := frac[0], frac[1]
		w2, w1, w0 := rsh192(m1<<15|m0>>49, m0<<15, 0, uint(-1-exp))
		var h [3]uint64
		for _, c := range logGCoeffs128[:15] { // 1/16, ..., 1/2
			q := mul192x192(h, [3]uint64{w2, w1, w0})
			qTop := [3]uint64{q[0], q[1], q[2]}
			if sign == 0 {
				// w < 0
				h = sub192(c, qTop)
			} else {
				h = add192(qTop, c)
			}
		}

		// h is in fixed point with 191 fractional bits. split it exactly into two Float128.
		hHiBits := [3]uint64{h[0], h[1] &^ (1<<16 - 1), 0} // the top 112 bits
		hLoBits := sub192(h, hHiBits)
		hHi := fixToFloat128(0, hHiBits[0], hHiBits[1], hHiBits[2], false, -191)
		var hLo Float128
		if hLoBits != ([3]uint64{}) {
			lz, hLoNorm := logNormalizeTo192([4]uint64{hLoBits[0], hLoBits[1], hLoBits[2], 0})
			hLo = fixToFloat128(0, hLoNorm[0], hLoNorm[1], hLoNorm[2], false, -191-lz)
		}

		// c = a² × H
		s, e := twoProduct128(a, a)
		cHi, cLo := twoProduct128(s, hHi)
		cLo = cLo.Add(s.Mul(hLo)).Add(e.Mul(hHi))
		rHi, rLo := ddAdd128(a, Float128{}, cHi.Neg(), cLo.Neg())
		return rHi.Add(rLo)
	}

	// |a| >= 2**-8, so the result is not small. log(1+a) = log(hi) + log1p(lo/hi),
	// where hi + lo = 1 + a exactly, and |lo/hi| < 2**-112 so that log1p(lo/hi) ~ lo/hi.
	hi, lo := twoSum128(Float128(uvone128), a)
	logHi, logLo := hi.logDD()
	if lo.IsZero() {
		return logHi.Add(logLo)
	}
	logHi, logLo = ddAdd128(logHi, logLo, lo.Quo(hi), Float128{})
	return logHi.Add(logLo)
}
