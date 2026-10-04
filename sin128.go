package floats

// reduce128 reduces the non-negative finite argument a by Pi/4.
// It returns the octant j (0 <= j < 8) and the reduced argument z = hi + lo
// such that a = j * Pi/4 + z (mod 2*Pi) and |z| <= Pi/4.
// lo is a correction term with |lo| <= ulp(hi).
//
// a is exactly representable in Float256, and reduce256 computes
// the reduced argument with more precision than Float128 has.
func reduce128(a Float128) (j uint64, hi, lo Float128) {
	j, hi256, lo256 := reduce256(a.Float256())
	hi = hi256.Float128()
	lo = hi256.Sub(hi.Float256()).Add(lo256).Float128()
	return j, hi, lo
}

// sinCoeffs128 are the coefficients of the Taylor series of sin(x)/x - 1 in x^2,
// in the order of Horner's method: (-1)^n / (2n+1)! for n = 15, 14, ..., 1.
var sinCoeffs128 = [...]Float128{
	{0xbf8e_434d_2e78_3f5b, 0xc42e_1ee4_6fa6_bfc4}, // -1/31!
	{0x3f98_259f_98b4_358a, 0xd7ab_e30e_7766_f129}, // +1/29!
	{0xbfa1_d1ab_1c2d_ccea, 0x320a_9a18_f15d_4277}, // -1/27!
	{0x3fab_3f3c_cdd1_65fa, 0x8d4e_44a4_1977_6f11}, // +1/25!
	{0xbfb4_761b_4131_6381, 0x9d97_b870_4dd7_f628}, // -1/23!
	{0x3fbd_71b8_ef6d_cf57, 0x18be_f146_fcee_6e45}, // +1/21!
	{0xbfc6_2f49_b468_1415, 0x724c_a1ec_3b7b_9675}, // -1/19!
	{0x3fce_952c_7703_0ad4, 0xa6b2_6051_9777_1b00}, // +1/17!
	{0xbfd6_ae7f_3e73_3b81, 0xf11d_8656_b0ee_8cb0}, // -1/15!
	{0x3fde_6124_613a_86d0, 0x97ca_3833_1d23_af68}, // +1/13!
	{0xbfe5_ae64_567f_544e, 0x38fe_747e_4b83_7dc7}, // -1/11!
	{0x3fec_71de_3a55_6c73, 0x38fa_ac1c_88e5_0017}, // +1/9!
	{0xbff2_a01a_01a0_1a01, 0xa01a_01a0_1a01_a01a}, // -1/7!
	{0x3ff8_1111_1111_1111, 0x1111_1111_1111_1111}, // +1/5!
	{0xbffc_5555_5555_5555, 0x5555_5555_5555_5555}, // -1/3!
}

// cosCoeffs128 are the coefficients of the Taylor series of (cos(x) - 1 + x^2/2) / x^4 in x^2,
// in the order of Horner's method: (-1)^n / (2n)! for n = 15, 14, ..., 2.
var cosCoeffs128 = [...]Float128{
	{0xbf93_3932_c504_7d60, 0xe60c_aded_4c29_89c5}, // -1/30!
	{0x3f9d_0a18_a263_5085, 0xd373_c5c5_1c35_4a8d}, // +1/28!
	{0xbfa6_88e8_5fc6_a4e5, 0x9a38_f205_0ba6_b015}, // -1/26!
	{0x3faf_f2cf_0197_2f57, 0x7cca_4b40_67ca_9d8a}, // +1/24!
	{0xbfb9_0ce3_96db_7f85, 0x2945_0c90_b7f3_38ec}, // -1/22!
	{0x3fc1_e542_ba40_2022, 0x507a_9cad_2bf8_f0bb}, // +1/20!
	{0xbfca_6827_863b_97d9, 0x77bb_0048_86a2_c2ab}, // -1/18!
	{0x3fd2_ae7f_3e73_3b81, 0xf11d_8656_b0ee_8cb0}, // +1/16!
	{0xbfda_9397_4a8c_07c9, 0xd20b_adf1_45df_a3e5}, // -1/14!
	{0x3fe2_1eed_8eff_8d89, 0x7b54_4da9_87ac_fe85}, // +1/12!
	{0xbfe9_27e4_fb77_89f5, 0xc72e_f016_d3ea_6679}, // -1/10!
	{0x3fef_a01a_01a0_1a01, 0xa01a_01a0_1a01_a01a}, // +1/8!
	{0xbff5_6c16_c16c_16c1, 0x6c16_c16c_16c1_6c17}, // -1/6!
	{0x3ffa_5555_5555_5555, 0x5555_5555_5555_5555}, // +1/4!
}

// sinPoly128 returns (sin(x)/x - 1) / x^2 for z = x^2.
func sinPoly128(z Float128) Float128 {
	r := sinCoeffs128[0]
	for _, c := range sinCoeffs128[1:] {
		r = FMA128(r, z, c)
	}
	return r
}

// cosPoly128 returns (cos(x) - 1 + x^2/2) / x^4 for z = x^2.
func cosPoly128(z Float128) Float128 {
	r := cosCoeffs128[0]
	for _, c := range cosCoeffs128[1:] {
		r = FMA128(r, z, c)
	}
	return r
}

// kernelSin128 returns sin(x + y) for |x| <= Pi/4 and |y| <= ulp(x).
func kernelSin128(x, y Float128) Float128 {
	// -1/2
	var NegHalf = Float128{0xbffe_0000_0000_0000, 0}

	z := x.Mul(x)
	v := z.Mul(x)
	r := sinPoly128(z)

	// sin(x + y) ≈ sin(x) + cos(x)*y ≈ x + x^3*r + (y - y*x^2/2)
	t := FMA128(y.Mul(z), NegHalf, y)
	return x.Add(FMA128(v, r, t))
}

// kernelSinExt128 is the same as kernelSin128, but returns the result as hi + lo
// with more precision. lo is a correction term with |lo| <= ulp(hi).
func kernelSinExt128(x, y Float128) (hi, lo Float128) {
	// -1/2
	var NegHalf = Float128{0xbffe_0000_0000_0000, 0}

	z := x.Mul(x)
	zl := FMA128(x, x, z.Neg()) // x*x = z + zl exactly
	v := x.Mul(z)
	vl := FMA128(x, zl, FMA128(x, z, v.Neg())) // x^3 ≈ v + vl
	r := sinPoly128(z)

	// sin(x + y) ≈ sin(x) + cos(x)*y ≈ x + x^3*r + (y - y*x^2/2)
	t := FMA128(y.Mul(z), NegHalf, y)
	t = FMA128(v, r, FMA128(vl, r, t))
	hi = x.Add(t)
	lo = x.Sub(hi).Add(t) // |x| >= |t|, so x - hi is exact
	return
}

// kernelCos128 returns cos(x + y) for |x| <= Pi/4 and |y| <= ulp(x).
func kernelCos128(x, y Float128) Float128 {
	w, t := kernelCosParts128(x, y)
	return w.Add(t)
}

// kernelCosExt128 is the same as kernelCos128, but returns the result as hi + lo
// with more precision. lo is a correction term with |lo| <= ulp(hi).
func kernelCosExt128(x, y Float128) (hi, lo Float128) {
	w, t := kernelCosParts128(x, y)
	hi = w.Add(t)
	lo = w.Sub(hi).Add(t) // |w| >= |t|, so w - hi is exact
	return
}

// kernelCosParts128 returns cos(x + y) as w + t, where |w| >= |t|.
func kernelCosParts128(x, y Float128) (w, t Float128) {
	var (
		One     = Float128(uvone128)
		Half    = Float128{0x3ffe_0000_0000_0000, 0}
		NegHalf = Float128{0xbffe_0000_0000_0000, 0}
	)

	z := x.Mul(x)
	zl := FMA128(x, x, z.Neg()) // x*x = z + zl exactly
	r := cosPoly128(z)

	// cos(x + y) ≈ cos(x) - sin(x)*y ≈ 1 - (z + zl)/2 + z^2*r - x*y
	hz := z.Mul(Half)
	w = One.Sub(hz)
	c := One.Sub(w).Sub(hz) // the rounding error of w, computed exactly
	t = FMA128(x, y.Neg(), zl.Mul(NegHalf))
	t = FMA128(z.Mul(z), r, t)
	t = c.Add(t)
	return
}

// Sin returns the sine of the radian argument a.
//
// Special cases are:
//
//	±0.Sin() = ±0
//	±Inf.Sin() = NaN
//	NaN.Sin() = NaN
func (a Float128) Sin() Float128 {
	// special cases
	switch {
	case a.IsZero():
		return a
	case a.IsNaN() || a.IsInf(0):
		return NewFloat128NaN()
	}

	// make argument positive but save the sign
	sign := a.Signbit()
	a = a.Abs()

	j, hi, lo := reduce128(a)

	// reflect in x axis
	if j > 3 {
		sign = !sign
		j -= 4
	}

	var y Float128
	if j == 1 || j == 2 {
		y = kernelCos128(hi, lo)
	} else {
		y = kernelSin128(hi, lo)
	}
	if sign {
		y = y.Neg()
	}
	return y
}

// Cos returns the cosine of the radian argument a.
//
// Special cases are:
//
//	±Inf.Cos() = NaN
//	NaN.Cos() = NaN
func (a Float128) Cos() Float128 {
	// special cases
	switch {
	case a.IsNaN() || a.IsInf(0):
		return NewFloat128NaN()
	}

	// make argument positive
	sign := false
	a = a.Abs()

	j, hi, lo := reduce128(a)

	if j > 3 {
		j -= 4
		sign = !sign
	}
	if j > 1 {
		sign = !sign
	}

	var y Float128
	if j == 1 || j == 2 {
		y = kernelSin128(hi, lo)
	} else {
		y = kernelCos128(hi, lo)
	}
	if sign {
		y = y.Neg()
	}
	return y
}

// Sincos returns Sin(a), Cos(a).
//
// Special cases are:
//
//	±0.Sincos() = ±0, 1
//	±Inf.Sincos() = NaN, NaN
//	NaN.Sincos() = NaN, NaN
func (a Float128) Sincos() (sin, cos Float128) {
	var One = Float128(uvone128)

	// special cases
	switch {
	case a.IsZero():
		return a, One // return ±0.0, 1.0
	case a.IsNaN() || a.IsInf(0):
		return NewFloat128NaN(), NewFloat128NaN()
	}

	// make argument positive
	sinSign, cosSign := a.Signbit(), false
	a = a.Abs()

	j, hi, lo := reduce128(a)

	if j > 3 { // reflect in x axis
		j -= 4
		sinSign, cosSign = !sinSign, !cosSign
	}
	if j > 1 {
		cosSign = !cosSign
	}

	sin = kernelSin128(hi, lo)
	cos = kernelCos128(hi, lo)
	if j == 1 || j == 2 {
		sin, cos = cos, sin
	}
	if cosSign {
		cos = cos.Neg()
	}
	if sinSign {
		sin = sin.Neg()
	}
	return
}

// Tan returns the tangent of the radian argument a.
//
// Special cases are:
//
//	±0.Tan() = ±0
//	±Inf.Tan() = NaN
//	NaN.Tan() = NaN
func (a Float128) Tan() Float128 {
	// special cases
	switch {
	case a.IsZero():
		return a
	case a.IsNaN() || a.IsInf(0):
		return NewFloat128NaN()
	}

	// make argument positive but save the sign
	sign := a.Signbit()
	a = a.Abs()

	j, hi, lo := reduce128(a)

	// tan(x) = sin(x) / cos(x) for j = 0 (mod 4),
	// tan(x) = -cos(x) / sin(x) for j = 2 (mod 4).
	sh, sl := kernelSinExt128(hi, lo)
	ch, cl := kernelCosExt128(hi, lo)
	nh, nl, dh, dl := sh, sl, ch, cl
	if j&2 != 0 {
		nh, nl, dh, dl = ch, cl, sh, sl
		sign = !sign
	}

	// (nh + nl) / (dh + dl) ≈ q + (nh - q*dh + nl - q*dl) / dh
	q := nh.Quo(dh)
	r := FMA128(q.Neg(), dh, nh) // the remainder nh - q*dh, computed exactly
	r = FMA128(q.Neg(), dl, r.Add(nl))
	y := q.Add(r.Quo(dh))
	if sign {
		y = y.Neg()
	}
	return y
}
