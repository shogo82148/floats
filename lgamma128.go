package floats

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
	var (
		// Half is 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}

		// One is 1
		One = Float128(uvone128)

		// Pi is π
		Pi = Float128{0x4000_921f_b544_42d1, 0x8469_898c_c517_01b8}

		// LnPi is ln(π)
		LnPi = Float128{0x3fff_250d_048e_7a1b, 0xd0bd_5f95_6c6a_843f}
	)

	switch {
	case a.IsNaN() || a.IsInf(0):
		return a, 1
	case a.IsZero() || isNegInt128(a):
		return NewFloat128Inf(1), 1
	}

	if g := a.Gamma(); !g.IsInf(0) && !g.IsZero() {
		// Gamma(a) didn't overflow (or underflow) here, so just take its log directly.
		sign := 1
		if g.Signbit() {
			sign = -1
		}
		return g.Abs().Log(), sign
	}

	// Gamma(a) over/underflowed, so work in log space instead. Note that
	// |a| need not exceed MaxStirling for this to happen: Gamma is
	// increasing quickly enough near its overflow point that some
	// non-integer a below MaxStirling already overflow Gamma(a).
	q := a.Abs()
	if !a.Signbit() {
		return lnStirling128(a), 1
	}

	// Reflection formula in log space, mirroring the one used by Gamma:
	//
	//	|Gamma(a)| = pi / (|q * sin(pi*frac)| * Gamma(q))
	p := q.Floor()
	z := q.Sub(p)
	if z.Gt(Half) {
		p = p.Add(One)
		z = q.Sub(p)
	}
	s := q.Mul(Pi.Mul(z).Sin())
	sign := -1
	if isOddInt128(p) {
		sign = 1
	}
	if s.IsZero() {
		return NewFloat128Inf(1), sign
	}
	return LnPi.Sub(s.Abs().Log()).Sub(lnStirling128(q)), sign
}

// lnStirling128 returns ln(Gamma(x)) for x large enough that Gamma(x)
// itself would overflow. It uses the same Stirling asymptotic series as
// stirling128, but evaluated directly in log space so that it stays finite
// well beyond the point where Gamma(x) would overflow.
func lnStirling128(x Float128) Float128 {
	var (
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}
		One  = Float128(uvone128)

		// HalfLn2Pi is ln(sqrt(2*pi))
		HalfLn2Pi = Float128{0x3ffe_d67f_1c86_4beb, 0x4a69_2979_2002_8832}
	)

	// Coefficients of the asymptotic series
	//
	//	ln Gamma(n+1) = 0.5*ln(2*pi*n) + n*ln(n) - n + M1/n + M2/n**3 + M3/n**5 + ...
	//
	// i.e. the Bernoulli-derived log-domain Stirling series (OEIS A046968/A046969).
	var (
		M1 = Float128{0x3ffb_5555_5555_5555, 0x5555_5555_5555_5555}
		M2 = Float128{0xbff6_6c16_c16c_16c1, 0x6c16_c16c_16c1_6c17}
		M3 = Float128{0x3ff4_a01a_01a0_1a01, 0xa01a_01a0_1a01_a01a}
		M4 = Float128{0xbff4_3813_8138_1381, 0x3813_8138_1381_3814}
		M5 = Float128{0x3ff4_b951_e2b1_8ff2, 0x3570_ea73_806e_5479}
		M6 = Float128{0xbff5_f6ab_0d99_93c7, 0xc81f_6ab0_d999_3c7d}
	)

	n := x.Sub(One)
	w := One.Quo(n)
	w2 := w.Mul(w)

	mu := M6
	mu = FMA128(mu, w2, M5)
	mu = FMA128(mu, w2, M4)
	mu = FMA128(mu, w2, M3)
	mu = FMA128(mu, w2, M2)
	mu = FMA128(mu, w2, M1)
	mu = mu.Mul(w)

	return x.Sub(Half).Mul(n.Log()).Sub(n).Add(HalfLn2Pi).Add(mu)
}
