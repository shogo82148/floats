package floats

import (
	"math"
)

// Sin returns the sine of the radian argument a.
//
// Special cases are:
//
//	±0.Sin() = ±0
//	±Inf.Sin() = NaN
//	NaN.Sin() = NaN
func (a Float32) Sin() Float32 {
	ix := a.Bits() &^ signMask32
	if ix < 0x3f490fdb { // |a| < Pi/4
		if ix < 0x39800000 { // |a| < 2**-12
			// sin(a) = a - a**3/6 + ... rounds to a.
			return a
		}
		x := float64(a)
		return Float32(sinKernel32(x, x*x))
	}
	if ix >= uvinf32 {
		// sin(±Inf) = NaN, sin(NaN) = NaN
		return NewFloat32NaN()
	}

	n, r := trigReduce32(ix)
	z := r * r
	var s float64
	if n&1 == 0 {
		s = sinKernel32(r, z)
	} else {
		s = cosKernel32(z)
	}
	if n&2 != 0 {
		s = -s
	}
	if a < 0 {
		s = -s
	}
	return Float32(s)
}

// Cos returns the cosine of the radian argument a.
//
// Special cases are:
//
//	±Inf.Cos() = NaN
//	NaN.Cos() = NaN
func (a Float32) Cos() Float32 {
	ix := a.Bits() &^ signMask32
	if ix < 0x3f490fdb { // |a| < Pi/4
		if ix < 0x39800000 { // |a| < 2**-12
			// cos(a) = 1 - a**2/2 + ... rounds to 1.
			return 1
		}
		x := float64(a)
		return Float32(cosKernel32(x * x))
	}
	if ix >= uvinf32 {
		// cos(±Inf) = NaN, cos(NaN) = NaN
		return NewFloat32NaN()
	}

	n, r := trigReduce32(ix)
	z := r * r
	var c float64
	if n&1 == 0 {
		c = cosKernel32(z)
	} else {
		c = sinKernel32(r, z)
	}
	if (n+1)&2 != 0 {
		c = -c
	}
	return Float32(c)
}

// Sincos returns Sin(a), Cos(a).
//
// Special cases are:
//
//	±0.Sincos() = ±0, 1
//	±Inf.Sincos() = NaN, NaN
//	NaN.Sincos() = NaN, NaN
func (a Float32) Sincos() (sin, cos Float32) {
	ix := a.Bits() &^ signMask32
	if ix < 0x3f490fdb { // |a| < Pi/4
		if ix < 0x39800000 { // |a| < 2**-12
			return a, 1
		}
		x := float64(a)
		z := x * x
		return Float32(sinKernel32(x, z)), Float32(cosKernel32(z))
	}
	if ix >= uvinf32 {
		nan := NewFloat32NaN()
		return nan, nan
	}

	n, r := trigReduce32(ix)
	z := r * r
	s, c := sinKernel32(r, z), cosKernel32(z)
	if n&1 != 0 {
		s, c = c, s
	}
	if n&2 != 0 {
		s = -s
	}
	if (n+1)&2 != 0 {
		c = -c
	}
	if a < 0 {
		s = -s
	}
	return Float32(s), Float32(c)
}

// Tan returns the tangent of the radian argument a.
//
// Special cases are:
//
//	±0.Tan() = ±0
//	±Inf.Tan() = NaN
//	NaN.Tan() = NaN
func (a Float32) Tan() Float32 {
	ix := a.Bits() &^ signMask32
	if ix < 0x3f490fdb { // |a| < Pi/4
		if ix < 0x39800000 { // |a| < 2**-12
			// tan(a) = a + a**3/3 + ... rounds to a.
			return a
		}
		x := float64(a)
		p, q := tanKernel32(x, x*x)
		return Float32(p / q)
	}
	if ix >= uvinf32 {
		// tan(±Inf) = NaN, tan(NaN) = NaN
		return NewFloat32NaN()
	}

	n, r := trigReduce32(ix)
	p, q := tanKernel32(r, r*r)
	var t float64
	if n&1 == 0 {
		t = p / q
	} else {
		t = -q / p
	}
	if a < 0 {
		t = -t
	}
	return Float32(t)
}

// sinKernel32 returns sin(x) for |x| <= Pi/4, where z = x*x.
// The relative error is less than 2**-45.
func sinKernel32(x, z float64) float64 {
	const (
		s1 = -0x1.555555555516bp-3
		s2 = 0x1.1111110fd3d43p-7
		s3 = -0x1.a019fd9b35ee5p-13
		s4 = 0x1.71d9a9f41c5a9p-19
		s5 = -0x1.aa285788aaa42p-26
	)
	return x + x*z*(s1+z*(s2+z*(s3+z*(s4+z*s5))))
}

// cosKernel32 returns cos(x) for |x| <= Pi/4, where z = x*x.
// The absolute error is less than 2**-50.
func cosKernel32(z float64) float64 {
	const (
		c2 = 0x1.5555555555437p-5
		c3 = -0x1.6c16c16b614fcp-10
		c4 = 0x1.a019ff53a6a1cp-16
		c5 = -0x1.27e25f4bb4e6fp-22
		c6 = 0x1.1c81c3531fff2p-29
	)
	return 1 - 0.5*z + z*z*(c2+z*(c3+z*(c4+z*(c5+z*c6))))
}

// tanKernel32 returns p and q such that tan(x) = p/q for |x| <= Pi/4, where z = x*x.
// The relative error of p/q is less than 2**-54.
func tanKernel32(x, z float64) (p, q float64) {
	const (
		p1 = -0x1.06b97bdbd0256p-3
		p2 = 0x1.6fc6fd9814ca2p-9
		p3 = -0x1.f637dbd78d4e5p-18
		q1 = -0x1.d8b213433d64dp-2
		q2 = 0x1.7e7b689111f03p-6
		q3 = -0x1.b525afbf1a0cep-13
	)
	p = x * (1 + z*(p1+z*(p2+z*p3)))
	q = 1 + z*(q1+z*(q2+z*q3))
	return
}

// trigReduce32 reduces |x| >= Pi/4 to r in [-Pi/4, Pi/4] such that
// |x| = n*Pi/2 + r (mod 2*Pi). ix is the bits of |x|, which must be finite.
func trigReduce32(ix uint32) (n uint64, r float64) {
	if ix < 0x49800000 { // |x| < 2**20
		// Cody-Waite reduction.
		// pio2_1 has 31 significant bits, so n*pio2_1 and x-n*pio2_1 are exact for n < 2**20
		// even if the multiplication is not fused.
		const (
			twoOverPi = 0x1.45f306dc9c883p-1
			pio2_1    = 0x1.921fb544p+0            // first 31 bits of Pi/2
			pio2_1t   = 6.07710050650619224932e-11 // Pi/2 - pio2_1
		)
		x := float64(math.Float32frombits(ix))
		fn := float64(int64(x*twoOverPi + 0.5))
		return uint64(fn), (x - fn*pio2_1) - fn*pio2_1t
	}

	// Payne-Hanek reduction.
	// |x| = m * 2**e where m is a 24-bit integer.
	e := int(ix>>shift32) - bias32 - shift32
	m := uint64(ix&fracMask32 | 1<<shift32)

	// Take 96 bits of 4/Pi starting from the bit with weight 2**(e-2).
	// The bits with larger weights contribute multiples of 8 to x*4/Pi,
	// i.e. multiples of 2*Pi to x.
	g := uint(61 + e)
	w, s := g/64, g%64
	hi := mPi4Long[w]<<s | mPi4Long[w+1]>>(64-s)
	lo := (mPi4Long[w+1]<<s | mPi4Long[w+2]>>(64-s)) >> 32

	// p = x*4/Pi mod 8 in 3.61 fixed point.
	p := m*hi + m*lo>>32

	// round x*2/Pi to the nearest integer n,
	// and the remaining d = x*4/Pi - 2n in [-1, 1) as a signed 2.61 fixed point.
	p += 1 << 61
	n = p >> 62
	d := int64(p&(1<<62-1)) - 1<<61
	return n, float64(d) * (math.Pi / 4 * 0x1p-61)
}
