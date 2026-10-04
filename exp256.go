package floats

import "github.com/shogo82148/ints"

// Exp returns e**x, the base-e exponential of a.
//
// Special cases are:
//
//	+Inf.Exp() = +Inf
//	NaN.Exp() = NaN
//
// Very large values overflow to 0 or +Inf.
// Very small values underflow to 1.
func (a Float256) Exp() Float256 {
	var (
		// ln(max float256 + 0.5ulp) = ln(2²⁶²¹⁴³×(2-2⁻²³⁷))
		// ~ 181704.374500706303191870897247532238261583907221734753336211540408636177
		Overflow = Float256{
			0x4001_062e_42fe_fa39, 0xef35_793c_7673_007e,
			0x5ed5_e81e_6864_ce53, 0x16c5_b141_a2eb_7175,
		}

		// ln(min float256 - 0.5ulp) = ln(2⁻²⁶²³⁷⁹)
		// ~ 181867.264088137890339583946796074909755081649753309413320929900210867126
		Underflow = Float256{
			0xc001_0633_5a1c_da3d, 0xdc01_1ec2_0913_2f62,
			0xbbd6_2bb5_6d5f_4375, 0x7057_2b20_10bb_94fe,
		}
	)

	sign := a[0] & signMask256[0]
	exp := int((a[0]>>(shift256-192))&mask256) - bias256

	// special cases
	switch {
	case exp < -238:
		// e**a rounds to 1 when |a| < 2**-238, including ±0 and subnormal values.
		return Float256(uvone256)
	case a.IsNaN():
		return a
	case a.Gt(Overflow):
		return NewFloat256Inf(1)
	case a.Lt(Underflow):
		return Float256{} // 0
	}

	// a = ±m × 2**(exp-236), where m is a 237-bit integer.
	m := ints.Uint256{a[0]&fracMask256[0] | 1<<(shift256-192), a[1], a[2], a[3]}

	n := expN256(exp, m)
	if n == 0 {
		// |a| < ln(2)/128. e**a = 1 + a × (e**a - 1)/a.
		// a × (e**a - 1)/a is computed with the relative precision,
		// so that 1 + a is rounded correctly even when |a| is tiny.
		x := expFix256(exp, m)
		g := expm1Poly256(sign != 0, x)
		return expOnePlus256(sign, exp, m, g)
	}

	k, v := expKernel256(sign, exp, m, n)
	return fixToFloat256(0, v, false, k-383)
}

// Exp2 returns 2**x, the base-2 exponential of x.
//
// Special cases are the same as [Exp].
func (a Float256) Exp2() Float256 {
	var (
		// log2(max float256 + 0.5ulp) = log2(2²⁶²¹⁴³×(2-2⁻²³⁷)) ~ 262144
		Overflow = Float256{
			0x4001_1000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}

		// log2(min float256 - 0.5ulp) = log2(2⁻²⁶²³⁷⁹) = -262379
		Underflow = Float256{
			0xc001_1003_ac00_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)

	sign := a[0] & signMask256[0]
	exp := int((a[0]>>(shift256-192))&mask256) - bias256

	// special cases
	switch {
	case exp < -238:
		// 2**a rounds to 1 when |a| < 2**-238, including ±0 and subnormal values.
		return Float256(uvone256)
	case a.IsNaN():
		return a
	case a.Gt(Overflow):
		return NewFloat256Inf(1)
	case a.Lt(Underflow):
		return Float256{} // 0
	}

	// a = ±m × 2**(exp-236), where m is a 237-bit integer.
	m := ints.Uint256{a[0]&fracMask256[0] | 1<<(shift256-192), a[1], a[2], a[3]}

	// |a|×2**262 modulo 2**256, that is, the fraction part of |a|×64 in fixed point
	// with 256 fractional bits. It is exact for exp >= -26.
	f := expFix256(exp, m)
	l := ints.Uint256(expm1Ln2Fix256)

	if exp < -7 {
		// |a| < 1/128. 2**a = 1 + a × ln(2) × (e**b - 1)/b, where b = a × ln(2).
		b := shr512to256(f.Mul512(l), 256) // |b|×2**262
		g := expm1Poly256(sign != 0, b)
		h := shr512to256(g.Mul512(l), 256) // ln(2) × g × 2**255
		return expOnePlus256(sign, exp, m, h)
	}

	// reduce: |a|×64 = n + f, where n is an integer and |f| <= 1/2.
	// n = round(|a|×64). |a|×64 = m × 2**(exp-230), and -7 <= exp <= 18.
	n := m[0] >> uint(38-exp)
	n += f[0] >> 63 // round up if f >= 1/2; then f is negative in two's complement.

	// r = f × ln(2)/64 in fixed point with 262 fractional bits.
	fneg := f[0]>>63 != 0
	if fneg {
		f = f.Neg()
	}
	r := shr512to256(f.Mul512(l), 256)
	if fneg {
		r = r.Neg()
	}

	k, v := expScale256(sign, n, r)
	return fixToFloat256(0, v, false, k-383)
}
