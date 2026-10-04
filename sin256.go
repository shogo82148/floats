package floats

import (
	"math/bits"

	"github.com/shogo82148/ints"
)

// pi4Bits returns 64 bits of 4/pi.
// The most significant bit of the result is the bit of 4/pi whose weight is 2^-k.
func pi4Bits(k int) uint64 {
	// g is the position of the bit in the bit stream of mPi4Long.
	g := k + 63
	if g < 0 {
		if g <= -64 {
			return 0
		}
		return mPi4Long[0] >> uint(-g)
	}
	i, s := g/64, uint(g%64)
	w := mPi4Long[i] << s
	if s != 0 {
		w |= mPi4Long[i+1] >> (64 - s)
	}
	return w
}

// reduce256 reduces the non-negative finite argument a by Pi/4.
// It returns the octant j (0 <= j < 8) and the reduced argument z = hi + lo
// such that a = j * Pi/4 + z (mod 2*Pi) and |z| <= Pi/4.
// lo is a correction term with |lo| <= ulp(hi).
//
// It uses the Payne-Hanek algorithm, so the result is accurate
// even if a is extremely large.
func reduce256(a Float256) (j uint64, hi, lo Float256) {
	// PI4 = Pi/4
	var PI4 = Float256{
		0x3fff_e921_fb54_442d, 0x1846_9898_cc51_701b,
		0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
	}
	if a.Lt(PI4) {
		return 0, a, Float256{}
	}

	// a = m * 2^e, where m is an integer.
	_, exp, m := a.normalize()
	e := exp - shift256

	// Compute a * 4/pi mod 8 in fixed point arithmetic.
	// The bits of 4/pi whose weight is 2^-k for k < e-2 are skipped,
	// because m * 2^(e-k) is a multiple of 8 for such k.
	// The window w holds the bits of 4/pi from 2^-(e-2) to 2^-(e-2+windowBits-1),
	// and p = m * w represents a * 4/pi mod 8 with fixed point at bit fracBits.
	// The low bits of p are inaccurate because the window is truncated,
	// but the error is less than m < 2^(shift256+1).
	const (
		windowWords = 13
		windowBits  = 64 * windowWords
		fracBits    = windowBits - 3
	)
	var w [windowWords]uint64 // little endian
	for i := range windowWords {
		w[windowWords-1-i] = pi4Bits(e - 2 + 64*i)
	}
	mw := [4]uint64{m[3], m[2], m[1], m[0]} // little endian
	var p [windowWords]uint64               // little endian, the bits above windowBits are discarded
	mulWords(p[:], mw[:], w[:])

	// the integer part of a * 4/pi mod 8
	j = p[windowWords-1] >> (64 - 3)
	p[windowWords-1] &= 1<<(64-3) - 1

	// map zeros to origin
	var sign uint64
	if j&1 == 1 {
		j = (j + 1) & 7
		sign = signMask256[0]

		// p = 1 - p
		var borrow uint64
		for i := range windowWords {
			p[i], borrow = bits.Sub64(0, p[i], borrow)
		}
		p[windowWords-1] &= 1<<(64-3) - 1
	}

	// normalize the fraction part: f = frac * 2^(l - 64*fracWords - fracBits)
	const fracWords = 6
	l := bitLenWords(p[:])
	if l == 0 {
		return j, Float256{}, Float256{}
	}
	var frac [fracWords]uint64 // little endian
	for i := range fracWords {
		frac[i] = bitsAt(p[:], l-64*(fracWords-i))
	}

	// z = f * Pi/4 = q * 2^qexp
	// pi4 = Pi/4 * 2^384
	pi4 := [fracWords]uint64{
		0xef95_19b3_cd3a_431b, 0x514a_0879_8e34_04dd,
		0x020b_bea6_3b13_9b22, 0x2902_4e08_8a67_cc74,
		0xc4c6_628b_80dc_1cd1, 0xc90f_daa2_2168_c234,
	}
	var q [2 * fracWords]uint64 // little endian
	mulWords(q[:], frac[:], pi4[:])
	qexp := l - 64*fracWords - fracBits - 64*fracWords

	// hi = the top bits of q rounded to nearest (ties away from zero).
	// lo = q - hi.
	ql := bitLenWords(q[:])
	s := ql - (shift256 + 1)
	mhi := bitsAt256(q[:], s)
	hexp := s + qexp + shift256
	roundUp := bitsAt(q[:], s-1)&1 != 0
	clearWordsFrom(q[:], s)
	losign := sign
	if roundUp {
		mhi = mhi.Add(ints.Uint256{0, 0, 0, 1})
		if mhi.BitLen() > shift256+1 {
			mhi = mhi.Rsh(1)
			hexp++
		}
		// q = 2^s - q
		var borrow uint64
		for i := range q {
			q[i], borrow = bits.Sub64(0, q[i], borrow)
		}
		clearWordsFrom(q[:], s)
		losign ^= signMask256[0]
	}
	hi = makeFloat256(sign, hexp, mhi)

	if ll := bitLenWords(q[:]); ll > 0 {
		t := ll - (shift256 + 1)
		lo = makeFloat256(losign, t+qexp+shift256, bitsAt256(q[:], t))
	}
	return j, hi, lo
}

// mulWords adds x * y to dst.
// dst, x, and y are little endian multi-word integers.
// The bits above len(dst) words are discarded.
func mulWords(dst, x, y []uint64) {
	for i, xi := range x {
		if xi == 0 {
			continue
		}
		var carry uint64
		for k := 0; i+k < len(dst) && k < len(y); k++ {
			hi, lo := bits.Mul64(xi, y[k])
			var c uint64
			lo, c = bits.Add64(lo, carry, 0)
			hi += c
			dst[i+k], c = bits.Add64(dst[i+k], lo, 0)
			carry = hi + c
		}
		for k := i + len(y); k < len(dst) && carry != 0; k++ {
			dst[k], carry = bits.Add64(dst[k], carry, 0)
		}
	}
}

// bitLenWords returns the bit length of the little endian multi-word integer x.
func bitLenWords(x []uint64) int {
	for i := len(x) - 1; i >= 0; i-- {
		if x[i] != 0 {
			return 64*i + bits.Len64(x[i])
		}
	}
	return 0
}

// bitsAt returns the 64 bits of the little endian multi-word integer x
// from bit position pos. The bits outside of x are treated as zero.
func bitsAt(x []uint64, pos int) uint64 {
	if pos <= -64 || pos >= 64*len(x) {
		return 0
	}
	if pos < 0 {
		return x[0] << uint(-pos)
	}
	k, s := pos/64, uint(pos%64)
	v := x[k] >> s
	if s != 0 && k+1 < len(x) {
		v |= x[k+1] << (64 - s)
	}
	return v
}

// bitsAt256 returns the bits of the little endian multi-word integer x
// from bit position pos to pos+shift256.
func bitsAt256(x []uint64, pos int) ints.Uint256 {
	v := ints.Uint256{bitsAt(x, pos+192), bitsAt(x, pos+128), bitsAt(x, pos+64), bitsAt(x, pos)}
	return v.And(ints.Uint256{1<<(shift256+1-192) - 1, ^uint64(0), ^uint64(0), ^uint64(0)})
}

// clearWordsFrom clears the bits of x at position pos and above.
func clearWordsFrom(x []uint64, pos int) {
	if pos < 0 {
		pos = 0
	}
	for i := range x {
		switch {
		case 64*i >= pos:
			x[i] = 0
		case 64*(i+1) > pos:
			x[i] &= 1<<uint(pos-64*i) - 1
		}
	}
}

// makeFloat256 returns sign * mant * 2^(exp - shift256).
// mant must be normalized: 2^shift256 <= mant < 2^(shift256+1).
// The result must be a normal number.
func makeFloat256(sign uint64, exp int, mant ints.Uint256) Float256 {
	mant = mant.And(fracMask256)
	mant[0] |= sign | uint64(exp+bias256)<<(shift256-192)
	return Float256(mant)
}

// sinCoeffs256 are the coefficients of the Taylor series of sin(x)/x - 1 in x^2,
// in the order of Horner's method: (-1)^n / (2n+1)! for n = 26, 25, ..., 1.
var sinCoeffs256 = [...]Float256{
	{0x3ff1_79d4_f105_8674, 0xdf40_f206_da45_3565, 0x15f5_4b34_3c3e_4ccf, 0x98ff_ed69_39f5_5aec}, // +1/53!
	{0xbff2_3161_872b_f7b8, 0x2337_72d9_1c5d_112d, 0x4406_4bd9_e588_eb2e, 0xaeb2_337e_0d3f_566d}, // -1/51!
	{0x3ff2_e5a4_2f0d_feb0, 0x85d9_4a3f_d410_e123, 0x1c72_d6f1_8e8c_3ad4, 0x6004_619d_32fe_9cdd}, // +1/49!
	{0xbff3_98da_8e0a_127e, 0xb9b7_8b45_4d8b_628e, 0x52ab_e2d9_69b5_0b8f, 0xe645_081a_8890_6825}, // -1/47!
	{0x3ff4_4a3c_b872_2206, 0x478e_02c5_e91c_64ca, 0xbec5_f43a_03d7_5f74, 0xa8d6_5fce_02aa_71f1}, // +1/45!
	{0xbff4_f95d_b452_57e5, 0x122d_cbae_56de_f372, 0x0370_619e_16b6_b8c9, 0x493b_419f_ab93_c929}, // -1/43!
	{0x3ff5_a65e_61c3_9d02, 0x40c7_e25c_fd1b_1b2d, 0xca88_5a15_2887_a173, 0x8054_011e_8d8d_92a4}, // +1/41!
	{0xbff6_51e9_9449_a4ba, 0xcde0_1044_76ae_b4c3, 0xab2f_3022_f174_9c49, 0x7dc3_44e5_775a_5e6d}, // -1/39!
	{0x3ff6_f9ec_8d1c_94e8, 0x5af4_c78b_15c3_d89d, 0x2f3f_cb2a_9273_4430, 0x5c83_1b36_193c_49a9}, // +1/37!
	{0xbff7_a0dc_59c7_16d9, 0x1f28_33c7_f5a7_e062, 0x3b3a_fda3_303f_f7d9, 0x742b_4532_af69_b5e8}, // -1/35!
	{0x3ff8_4398_1254_dd0d, 0x51b5_382c_dffa_9742, 0x27d5_0dc1_2492_5687, 0x3480_48ea_66d9_58e6}, // +1/33!
	{0xbff8_e434_d2e7_83f5, 0xbc42_e1ee_46fa_6bfc, 0x3913_b62f_2db6_e93b, 0x6e24_4b31_ba10_23ad}, // -1/31!
	{0x3ff9_8259_f98b_4358, 0xad7a_be30_e776_6f12, 0x91d6_66f5_d904_9ed2, 0x7987_f64a_a97b_a866}, // +1/29!
	{0xbffa_1d1a_b1c2_dcce, 0xa320_a9a1_8f15_d427, 0x734a_0749_e62d_53e1, 0xccbd_a09a_68ca_1d12}, // -1/27!
	{0x3ffa_b3f3_ccdd_165f, 0xa8d4_e44a_4197_76f1, 0x0b89_3fff_294c_1301, 0x4bdb_ff99_dad6_8eee}, // +1/25!
	{0xbffb_4761_b413_1638, 0x19d9_7b87_04dd_7f62, 0x7984_d6ff_0465_2645, 0x84e5_cf88_4c73_6f7f}, // -1/23!
	{0x3ffb_d71b_8ef6_dcf5, 0x718b_ef14_6fce_e6e4, 0x5218_487a_0757_f6d2, 0xb457_1e19_b38e_1531}, // +1/21!
	{0xbffc_62f4_9b46_8141, 0x5724_ca1e_c3b7_b967, 0x4b57_eb74_1a06_2878, 0xd7ef_76b1_154a_8d62}, // -1/19!
	{0x3ffc_e952_c770_30ad, 0x4a6b_2605_1977_71af, 0xfea7_748d_1ac4_3a11, 0x7079_e890_9271_98e1}, // +1/17!
	{0xbffd_6ae7_f3e7_33b8, 0x1f11_d865_6b0e_e8ca, 0xfe91_ebd5_ec70_7db2, 0x8781_8719_9b98_b26f}, // -1/15!
	{0x3ffd_e612_4613_a86d, 0x097c_a383_31d2_3af6, 0x84d3_b375_7bf4_471c, 0x7328_40d3_01a3_425f}, // +1/13!
	{0xbffe_5ae6_4567_f544, 0xe38f_e747_e4b8_37dc, 0x71e2_02b7_2f11_b6aa, 0xac59_0f01_29fe_f8e4}, // -1/11!
	{0x3ffe_c71d_e3a5_56c7, 0x338f_aac1_c88e_5001, 0x71de_3a55_6c73_38fa, 0xac1c_88e5_0017_1de4}, // +1/9!
	{0xbfff_2a01_a01a_01a0, 0x1a01_a01a_01a0_1a01, 0xa01a_01a0_1a01_a01a, 0x01a0_1a01_a01a_01a0}, // -1/7!
	{0x3fff_8111_1111_1111, 0x1111_1111_1111_1111, 0x1111_1111_1111_1111, 0x1111_1111_1111_1111}, // +1/5!
	{0xbfff_c555_5555_5555, 0x5555_5555_5555_5555, 0x5555_5555_5555_5555, 0x5555_5555_5555_5555}, // -1/3!
}

// cosCoeffs256 are the coefficients of the Taylor series of (cos(x) - 1 + x^2/2) / x^4 in x^2,
// in the order of Horner's method: (-1)^n / (2n)! for n = 26, 25, ..., 2.
var cosCoeffs256 = [...]Float256{
	{0x3ff1_d564_5798_9358, 0xc8e1_c86d_acc1_5037, 0xb62f_2247_41e3_979b, 0xeab3_f09b_23ff_2f4b}, // +1/52!
	{0xbff2_8bb3_6f6e_12cd, 0x7820_5f0a_0534_5360, 0x246a_08e3_45d2_36d2, 0x666c_0210_e51c_f1be}, // -1/50!
	{0x3ff3_4091_b406_b6ff, 0x267a_5cd8_de5c_ec5e, 0xe1c7_ec90_f123_5d0a, 0x9983_5abc_5b0a_f019}, // +1/48!
	{0xbff3_f240_804f_6595, 0x1062_ca46_e4f2_5c60, 0x84b6_3a97_a9a0_f47d, 0xad1a_b1f3_7c4a_0c7b}, // -1/46!
	{0x3ff4_a272_b1b0_3fec, 0x6a4f_d9f3_27e7_f6de, 0x8e23_2fb8_cab3_6f1e, 0x06b6_bb5c_d9df_d81e}, // +1/44!
	{0xbff5_510a_f527_530d, 0xe836_c4d9_225d_cb90, 0x9a4f_8196_3742_c427, 0x3d33_d017_4747_4b28}, // -1/42!
	{0x3ff5_fca8_ed42_a12a, 0xe300_1a07_244a_bad2, 0xab7e_b36b_1bed_c6db, 0xfc6b_a16f_255d_63e2}, // +1/40!
	{0xbff6_a5d4_acb9_c0c3, 0xaae9_13d3_70a4_ec4e, 0x78a1_82aa_9646_1e79, 0x9145_fbf7_a976_2315}, // -1/38!
	{0x3ff7_4df9_8329_0c2c, 0xa92b_06b8_d12a_7275, 0xbea1_c2e9_3955_46d7, 0xeaf7_9776_8d2d_b52b}, // +1/36!
	{0xbff7_f271_0231_c0fd, 0x7a13_f8a2_b4af_9d6b, 0x70c8_856a_7cc5_f715, 0xd70f_53af_6fdb_9ef6}, // -1/34!
	{0x3ff8_9434_d2e7_83f5, 0xbc42_e1ee_46fa_6bfc, 0x3913_b62f_2db6_e93b, 0x6e24_4b31_ba10_23ad}, // +1/32!
	{0xbff9_3393_2c50_47d6, 0x0e60_cade_d4c2_989c, 0x574b_187d_b449_31f1, 0x92b3_28d8_2c3f_a28f}, // -1/30!
	{0x3ff9_d0a1_8a26_3508, 0x5d37_3c5c_51c3_54a8, 0xd42a_4d4e_ccac_2fee, 0xbe23_3733_a998_109d}, // +1/28!
	{0xbffa_688e_85fc_6a4e, 0x59a3_8f20_50ba_6b01, 0x4946_7626_5a36_3ec6, 0x84bf_ff82_486a_8888}, // -1/26!
	{0x3ffa_ff2c_f019_72f5, 0x77cc_a4b4_067c_a9d8, 0xa206_73fe_b086_ddb2, 0x0687_bf60_65ef_3f54}, // +1/24!
	{0xbffb_90ce_396d_b7f8, 0x5294_50c9_0b7f_338e, 0xc757_7a87_4b28_b381, 0xf785_2d29_f6f2_f823}, // -1/22!
	{0x3ffc_1e54_2ba4_0202, 0x2507_a9ca_d2bf_8f0b, 0xabbf_df20_29a3_73f4, 0x8cb2_5781_bbaa_7bd0}, // +1/20!
	{0xbffc_a682_7863_b97d, 0x977b_b004_886a_2c2a, 0xa978_6799_dee7_500f, 0x806c_5cf2_4948_87e4}, // -1/18!
	{0x3ffd_2ae7_f3e7_33b8, 0x1f11_d865_6b0e_e8ca, 0xfe91_ebd5_ec70_7db2, 0x8781_8719_9b98_b26f}, // +1/16!
	{0xbffd_a939_74a8_c07c, 0x9d20_badf_145d_fa3e, 0x4ea8_cd18_8da9_75d7, 0x5f09_6ea8_01df_2748}, // -1/14!
	{0x3ffe_21ee_d8ef_f8d8, 0x97b5_44da_987a_cfe8, 0x4bec_01cf_74b6_79c7, 0x1d90_b4ab_7154_a5ed}, // +1/12!
	{0xbffe_927e_4fb7_789f, 0x5c72_ef01_6d3e_a667, 0x8e4b_61dd_f05c_2d95, 0x567d_3a50_ccdf_4b1d}, // -1/10!
	{0x3ffe_fa01_a01a_01a0, 0x1a01_a01a_01a0_1a01, 0xa01a_01a0_1a01_a01a, 0x01a0_1a01_a01a_01a0}, // +1/8!
	{0xbfff_56c1_6c16_c16c, 0x16c1_6c16_c16c_16c1, 0x6c16_c16c_16c1_6c16, 0xc16c_16c1_6c16_c16c}, // -1/6!
	{0x3fff_a555_5555_5555, 0x5555_5555_5555_5555, 0x5555_5555_5555_5555, 0x5555_5555_5555_5555}, // +1/4!
}

// kernelSin256 returns sin(x + y) for |x| <= Pi/4 and |y| <= ulp(x).
func kernelSin256(x, y Float256) Float256 {
	// -1/2
	var NegHalf = Float256{0xbfff_e000_0000_0000, 0, 0, 0}

	z := x.Mul(x)
	v := z.Mul(x)
	r := sinCoeffs256[0]
	for _, c := range sinCoeffs256[1:] {
		r = FMA256(r, z, c)
	}

	// sin(x + y) ≈ sin(x) + cos(x)*y ≈ x + x^3*r + (y - y*x^2/2)
	t := FMA256(y.Mul(z), NegHalf, y)
	return x.Add(FMA256(v, r, t))
}

// kernelCos256 returns cos(x + y) for |x| <= Pi/4 and |y| <= ulp(x).
func kernelCos256(x, y Float256) Float256 {
	var (
		One     = Float256(uvone256)
		Half    = Float256{0x3fff_e000_0000_0000, 0, 0, 0}
		NegHalf = Float256{0xbfff_e000_0000_0000, 0, 0, 0}
	)

	z := x.Mul(x)
	zl := FMA256(x, x, z.Neg()) // x*x = z + zl exactly
	r := cosCoeffs256[0]
	for _, c := range cosCoeffs256[1:] {
		r = FMA256(r, z, c)
	}

	// cos(x + y) ≈ cos(x) - sin(x)*y ≈ 1 - (z + zl)/2 + z^2*r - x*y
	hz := z.Mul(Half)
	w := One.Sub(hz)
	c := One.Sub(w).Sub(hz) // the rounding error of w, computed exactly
	t := FMA256(x, y.Neg(), zl.Mul(NegHalf))
	t = FMA256(z.Mul(z), r, t)
	return w.Add(c.Add(t))
}

// Sin returns the sine of the radian argument a.
//
// Special cases are:
//
//	±0.Sin() = ±0
//	±Inf.Sin() = NaN
//	NaN.Sin() = NaN
func (a Float256) Sin() Float256 {
	// special cases
	switch {
	case a.IsZero():
		return a
	case a.IsNaN() || a.IsInf(0):
		return NewFloat256NaN()
	}

	// make argument positive but save the sign
	sign := a.Signbit()
	a = a.Abs()

	j, hi, lo := reduce256(a)

	// reflect in x axis
	if j > 3 {
		sign = !sign
		j -= 4
	}

	var y Float256
	if j == 1 || j == 2 {
		y = kernelCos256(hi, lo)
	} else {
		y = kernelSin256(hi, lo)
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
func (a Float256) Cos() Float256 {
	// special cases
	switch {
	case a.IsNaN() || a.IsInf(0):
		return NewFloat256NaN()
	}

	// make argument positive
	sign := false
	a = a.Abs()

	j, hi, lo := reduce256(a)

	if j > 3 {
		j -= 4
		sign = !sign
	}
	if j > 1 {
		sign = !sign
	}

	var y Float256
	if j == 1 || j == 2 {
		y = kernelSin256(hi, lo)
	} else {
		y = kernelCos256(hi, lo)
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
func (a Float256) Sincos() (sin, cos Float256) {
	var One = Float256(uvone256)

	// special cases
	switch {
	case a.IsZero():
		return a, One // return ±0.0, 1.0
	case a.IsNaN() || a.IsInf(0):
		return NewFloat256NaN(), NewFloat256NaN()
	}

	// make argument positive
	sinSign, cosSign := a.Signbit(), false
	a = a.Abs()

	j, hi, lo := reduce256(a)

	if j > 3 { // reflect in x axis
		j -= 4
		sinSign, cosSign = !sinSign, !cosSign
	}
	if j > 1 {
		cosSign = !cosSign
	}

	sin = kernelSin256(hi, lo)
	cos = kernelCos256(hi, lo)
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
func (a Float256) Tan() Float256 {
	sin, cos := a.Sincos()
	return sin.Quo(cos)
}
