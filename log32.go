package floats

import "math"

// Log returns the natural logarithm of a.
//
// Special cases are:
//
//	+Inf.Log() = +Inf
//	0.Log() = -Inf
//	(x < 0).Log() = NaN
//	NaN.Log() = NaN
func (a Float32) Log() Float32 {
	// https://github.com/chewxy/math32/blob/912ef0b2e4151df0148d7645c92a7b5e22f887f5/log.go#L85-L124
	const (
		Ln2Hi = 6.9313812256e-01 /* 0x3f317180 */
		Ln2Lo = 9.0580006145e-06 /* 0x3717f7d1 */
		L1    = 6.6666668653e-01 /* 0x3f2aaaab */
		L2    = 4.0000000596e-01 /* 0x3ecccccd */
		L3    = 2.8571429849e-01 /* 0x3e924925 */
		L4    = 2.2222198546e-01 /* 0x3e638e29 */
		L5    = 1.8183572590e-01 /* 0x3e3a3325 */
		L6    = 1.5313838422e-01 /* 0x3e1cd04f */
		L7    = 1.4798198640e-01 /* 0x3e178897 */
	)

	// special cases
	switch {
	case a.IsNaN() || a.IsInf(1):
		return a
	case a < 0:
		return NewFloat32NaN()
	case a == 0:
		return NewFloat32Inf(-1)
	}

	// reduce
	f1, ki := a.Frexp()
	if f1 < math.Sqrt2/2 {
		f1 *= 2
		ki--
	}
	f := f1 - 1
	k := Float32(ki)

	// compute
	s := f / (2 + f)
	s2 := s * s
	s4 := s2 * s2
	t1 := s2 * (L1 + s4*(L3+s4*(L5+s4*L7)))
	t2 := s4 * (L2 + s4*(L4+s4*L6))
	R := t1 + t2
	hfsq := 0.5 * f * f
	return k*Ln2Hi - ((hfsq - (s*(hfsq+R) + k*Ln2Lo)) - f)
}

// Log10 returns the decimal logarithm of a.
// The special cases are the same as for [Log].
func (a Float32) Log10() Float32 {
	return a.Log() * (1 / math.Ln10)
}

// Log2 returns the binary logarithm of a.
// The special cases are the same as for [Log].
func (a Float32) Log2() Float32 {
	frac, exp := a.Frexp()
	// Make sure exact powers of two give an exact answer.
	// Don't depend on Log(0.5)*(1/Ln2)+exp being exactly exp-1.
	if frac == 0.5 {
		return Float32(exp - 1)
	}
	return frac.Log()*(1/math.Ln2) + Float32(exp)
}
