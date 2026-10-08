package floats

import (
	"encoding/binary"
	"math"
	"math/big"
)

// Pow returns a**b, the base-a exponential of b.
//
// Special cases are (in order):
//
//	a.Pow(±0) = 1 for any a
//	1.Pow(b) = 1 for any b
//	a.Pow(1) = a for any a
//	NaN.Pow(b) = NaN
//	a.Pow(NaN) = NaN
//	±0.Pow(b) = ±Inf for b an odd integer < 0
//	±0.Pow(-Inf) = +Inf
//	±0.Pow(+Inf) = +0
//	±0.Pow(b) = +Inf for finite b < 0 and not an odd integer
//	±0.Pow(b) = ±0 for b an odd integer > 0
//	±0.Pow(b) = +0 for finite b > 0 and not an odd integer
//	-1.Pow(±Inf) = 1
//	a.Pow(+Inf) = +Inf for |a| > 1
//	a.Pow(-Inf) = +0 for |a| > 1
//	a.Pow(+Inf) = +0 for |a| < 1
//	a.Pow(-Inf) = +Inf for |a| < 1
//	+Inf.Pow(b) = +Inf for b > 0
//	+Inf.Pow(b) = +0 for b < 0
//	-Inf.Pow(b) = (-0).Pow(-b)
//	a.Pow(b) = NaN for finite a < 0 and finite non-integer b
func (a Float256) Pow(b Float256) Float256 {
	var (
		// Zero = 0
		Zero = Float256{}

		// One = 1.0
		One = Float256(uvone256)

		// Half = 0.5
		Half = Float256{
			0x3fff_e000_0000_0000, 0x0000_0000_0000_0000,
			0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
		}
	)

	switch {
	case b.IsZero() || a.Eq(One):
		return One
	case b.Eq(One):
		return a
	case a.IsNaN() || b.IsNaN():
		return NewFloat256NaN()
	case a.IsZero():
		switch {
		case b.Lt(Zero):
			if isOddInt256(b) {
				return NewFloat256Inf(1).Copysign(a)
			}
			return NewFloat256Inf(1)
		case b.Gt(Zero):
			if isOddInt256(b) {
				return a
			}
			return Zero
		}
	case b.IsInf(0):
		switch {
		case a.Eq(One.Neg()):
			return One
		case (a.Abs().Lt(One)) == b.IsInf(1):
			return Zero
		default:
			return NewFloat256Inf(1)
		}
	case a.IsInf(0):
		if a.IsInf(-1) {
			return (Zero.Neg()).Pow(b.Neg()) // Pow(-0, -b)
		}
		switch {
		case b.Lt(Zero):
			return Zero
		case b.Gt(Zero):
			return NewFloat256Inf(1)
		}
	case b.Eq(Half):
		return a.Sqrt()
	}

	bi, bf := b.Abs().Modf()
	if !bf.IsZero() && a.Lt(Zero) {
		return NewFloat256NaN()
	}
	neg := a.Signbit() && isOddInt256(b)
	r := pow256Abs(a.Abs(), b, bi, bf.IsZero())
	if neg {
		return r.Neg()
	}
	return r
}

// pow256Abs returns x**b for x > 0 and x != 1. bi is |b| rounded toward zero, and isInt reports whether b is an integer.
func pow256Abs(x, b, bi Float256, isInt bool) Float256 {
	var (
		half = Float256{0x3fff_e000_0000_0000}
	)

	if isInt {
		if x1, xe := x.Frexp(); x1.Eq(half) {
			// x is a power of two, so that the result is exactly a power of two, and it may be the midpoint of
			// the subnormal numbers.
			return pow256Two(xe-1, bi, b.Lt(Float256{}))
		}

		// The exact product of the integers has at least (s-1) |b| + 1 bits, where s is the number of the significant
		// bits of x. It can be the midpoint of two adjacent Float256 values only if it has 238 bits. In that case the
		// partial products of successive squarings are exact, and only the last multiplication rounds, so that
		// it is correctly rounded. Otherwise the result can not be the midpoint, and exp and log below are accurate enough.
		_, _, m := x.normalize()
		s := 237 - m.TrailingZeros()
		if bi.Le(NewFloat256(238)) && (s-1)*int(bi.Int64())+1 <= 238 {
			return powInt256(x, int(bi.Int64()), b.Lt(Float256{}))
		}
	}

	// x**b = exp(b ln(x)). b ln(x) is calculated in fixed point with 320 fractional bits,
	// whose absolute error is about 2**-290 if the result does not overflow or underflow. The relative error of ln(x)
	// is about 2**-254 if x is close to 1 (|x-1| < 2**-8), and much smaller otherwise.
	lsign, v, e := log256Fix(x)

	// the estimate of b ln(x) to detect the overflow and underflow
	zf := b.Float64().BuiltIn() * math.Ldexp(float64(v.Rsh(uint(max(v.BitLen()-64, 0)))[7]), e+max(v.BitLen()-64, 0))
	if lsign != 0 {
		zf = -zf
	}
	switch {
	case zf > 182500: // ln(2**262144) ~ 181704
		return NewFloat256Inf(1)
	case zf < -182500:
		return Float256{}
	}

	// z = |b ln(x)| = mb × v × 2**(eb-236+e)
	_, eb, mb := b.normalize()
	prod := new(big.Int).Mul(bigFromWords(mb[:]), bigFromWords(v[:]))
	// The shift is negative: it is eb - 236 if ln(x) is not tiny (e = -320), where eb <= 26 because |ln(x)| >= 2**-8
	// and |b ln(x)| < 2**18, and it is eb - 236 + e + 320 with e <= -746 otherwise.
	prod.Rsh(prod, uint(236-eb-e-320))
	// prod < 2**338 because |b ln(x)| < 2**18, so that it fits z. big.Word may be 32 bits, so that the 64-bit limbs are
	// extracted through the big-endian bytes.
	var buf [48]byte
	prod.FillBytes(buf[:])
	var z gammaFix256
	for i := range z {
		z[i] = binary.BigEndian.Uint64(buf[8*i:])
	}

	var mant gammaFix256
	var k int
	if zneg := (lsign != 0) != b.Signbit(); zneg {
		mant, k = gammaExpNeg256(z)
	} else {
		mant, k = gammaExp256(z)
	}
	return lgamma256FromPair(false, mant, k)
}

// bigFromWords returns the integer whose big-endian 64-bit words are w.
func bigFromWords(w []uint64) *big.Int {
	r := new(big.Int)
	for _, v := range w {
		r.Lsh(r, 64)
		r.Or(r, new(big.Int).SetUint64(v))
	}
	return r
}

// pow256Two returns (2**e)**n for e != 0 and the non-negative integer n, or its reciprocal if flip is true.
// |e| >= 1, so that the result overflows or underflows if n > 2**20 (|e n| > 1000000).
func pow256Two(e int, n Float256, flip bool) Float256 {
	if n.Gt(Float256{0x4001_3000_0000_0000}) { // n > 2**20
		n = Float256{0x4001_3000_0000_0000}
	}
	// |e n| <= 262378 × 2**20 fits in int64.
	ae := int64(e) * n.Int64()
	if flip {
		ae = -ae
	}
	return Float256(uvone256).Ldexp(int(max(min(ae, 1000000), -1000000)))
}

func isOddInt256(x Float256) bool {
	// MaxSafeInteger = 2**237
	var MaxSafeInteger = Float256{
		0x400e_c000_0000_0000, 0x0000_0000_0000_0000,
		0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
	}
	if x.Abs().Ge(MaxSafeInteger) {
		// 1 << 237 is the largest exact integer in the float256 format.
		// Any number outside this range will be truncated before the decimal point and therefore will always be
		// an even integer.
		return false
	}

	xi, xf := x.Modf()
	return xf.IsZero() && xi.Int256()[3]&1 == 1
}
