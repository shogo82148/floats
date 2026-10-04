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
// It returns the octant j (0 <= j < 8) and the reduced argument z
// such that a = j * Pi/4 + z (mod 2*Pi) and |z| <= Pi/4.
//
// It uses the Payne-Hanek algorithm, so the result is accurate
// even if a is extremely large.
func reduce256(a Float256) (j uint64, z Float256) {
	// PI4 = Pi/4
	var PI4 = Float256{
		0x3fff_e921_fb54_442d, 0x1846_9898_cc51_701b,
		0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
	}
	if a.Lt(PI4) {
		return 0, a
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
	var p [windowWords]uint64 // little endian, the bits above windowBits are discarded
	for i := range len(m) {
		mi := m[len(m)-1-i]
		if mi == 0 {
			continue
		}
		var carry uint64
		for k := 0; i+k < windowWords; k++ {
			hi, lo := bits.Mul64(mi, w[k])
			var c uint64
			lo, c = bits.Add64(lo, carry, 0)
			hi += c
			p[i+k], c = bits.Add64(p[i+k], lo, 0)
			carry = hi + c
		}
	}

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

	// normalize the fraction part: f = frac * 2^(l - 256 - fracBits)
	var l int // bit length of p
	for i := windowWords - 1; i >= 0; i-- {
		if p[i] != 0 {
			l = 64*i + bits.Len64(p[i])
			break
		}
	}
	if l == 0 {
		return j, Float256{}
	}
	var frac ints.Uint256 // big endian
	for i := range 4 {
		// the 64 bits from bit position l-64*(i+1)
		pos := l - 64*(i+1)
		var v uint64
		if pos >= 0 {
			k, s := pos/64, uint(pos%64)
			v = p[k] >> s
			if s != 0 && k+1 < windowWords {
				v |= p[k+1] << (64 - s)
			}
		} else if pos > -64 {
			v = p[0] << uint(-pos)
		}
		frac[i] = v
	}

	// z = f * Pi/4
	// pi4 = Pi/4 * 2^256
	pi4 := ints.Uint256{
		0xc90f_daa2_2168_c234, 0xc4c6_628b_80dc_1cd1,
		0x2902_4e08_8a67_cc74, 0x020b_bea6_3b13_9b22,
	}
	q := frac.Mul512(pi4)
	r := ints.Uint256{q[0], q[1], q[2], q[3]}
	if q[4]|q[5]|q[6]|q[7] != 0 {
		r[3] |= 1 // sticky bit
	}
	shift := r.BitLen() - (shift256 + 1)
	mant := roundToNearestEven256(r, uint(shift))
	zexp := shift + l - fracBits - 256 + shift256
	if mant.BitLen() > shift256+1 {
		mant = mant.Rsh(1)
		zexp++
	}
	mant = mant.And(fracMask256)
	mant[0] |= sign | uint64(zexp+bias256)<<(shift256-192)
	return j, Float256(mant)
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

	var Zero = Float256{}

	// make argument positive but save the sign
	sign := false
	if a.Lt(Zero) {
		a = a.Neg()
		sign = true
	}

	j, z := reduce256(a)
	var y Float256

	// reflect in x axis
	if j > 3 {
		sign = !sign
		j -= 4
	}

	if j == 1 || j == 2 {
		// taylor series expansion of cos around 0
		y = Zero
		for n := 50; n >= 0; n-- {
			term := power256(z, 2*n).Quo(factorial256(2 * n))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
	} else {
		// taylor series expansion of sin around 0
		y = Zero
		for n := 50; n >= 0; n-- {
			term := power256(z, 2*n+1).Quo(factorial256(2*n + 1))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
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

	var Zero = Float256{}

	// make argument positive but save the sign
	sign := false
	a = a.Abs()

	j, z := reduce256(a)
	var y Float256

	if j > 3 {
		j -= 4
		sign = !sign
	}
	if j > 1 {
		sign = !sign
	}

	if j == 1 || j == 2 {
		// taylor series expansion of sin around 0
		y = Zero
		for n := 50; n >= 0; n-- {
			term := power256(z, 2*n+1).Quo(factorial256(2*n + 1))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
	} else {
		// taylor series expansion of cos around 0
		y = Zero
		for n := 50; n >= 0; n-- {
			term := power256(z, 2*n).Quo(factorial256(2 * n))
			if n%2 != 0 {
				term = term.Neg()
			}
			y = y.Add(term)
		}
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
	var (
		Zero = Float256{}
		One  = Float256(uvone256)
	)

	// special cases
	switch {
	case a.IsZero():
		return a, One // return ±0.0, 1.0
	case a.IsNaN() || a.IsInf(0):
		return NewFloat256NaN(), NewFloat256NaN()
	}

	// make argument positive
	sinSign, cosSign := false, false
	if a.Lt(Zero) {
		a = a.Neg()
		sinSign = true
	}

	j, z := reduce256(a)

	if j > 3 { // reflect in x axis
		j -= 4
		sinSign, cosSign = !sinSign, !cosSign
	}
	if j > 1 {
		cosSign = !cosSign
	}

	// taylor series expansion of sin around 0
	sin = Zero
	for n := 50; n >= 0; n-- {
		term := power256(z, 2*n+1).Quo(factorial256(2*n + 1))
		if n%2 != 0 {
			term = term.Neg()
		}
		sin = sin.Add(term)
	}

	// taylor series expansion of cos around 0
	cos = Zero
	for n := 50; n >= 0; n-- {
		term := power256(z, 2*n).Quo(factorial256(2 * n))
		if n%2 != 0 {
			term = term.Neg()
		}
		cos = cos.Add(term)
	}

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
