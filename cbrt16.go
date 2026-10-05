package floats

import "math/bits"

// Cbrt returns the cube root of a.
//
// Special cases are:
//
//	±0.Cbrt() = ±0
//	±Inf.Cbrt() = ±Inf
//	NaN.Cbrt() = NaN
func (a Float16) Cbrt() Float16 {
	b := uint16(a)
	sign := b & uint16(signMask16)
	exp := int(b>>10) & 0x1f
	frac := b & 0x3ff

	// a = ±2**e × (1 + f/1024)
	var e int
	var f uint16
	switch {
	case exp == 0x1f:
		if frac != 0 {
			return Float16(sign | 0x7e00) // NaN
		}
		return a // ±Inf
	case exp == 0:
		if frac == 0 {
			return a // ±0
		}
		// subnormal: a = frac × 2**-24
		l := bits.Len16(frac)
		e = l - 25
		f = frac << (11 - l) & 0x3ff
	default:
		e = exp - 15
		f = frac
	}

	// cbrt(a) = 2**q × cbrt(2**r × (1 + f/1024)), where e = 3q + r and r = 0, 1, 2.
	// The result is always normal, and a carry of the significand is added to the exponent.
	q, r := (e+27)/3-9, (e+27)%3
	return Float16(sign | (uint16((q+15)<<10) + cbrt16Table[r<<10|int(f)]))
}
