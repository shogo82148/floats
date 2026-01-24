package floats

// Cbrt returns the cube root of a.
//
// Special cases are:
//
//	±0.Cbrt() = ±0
//	±Inf.Cbrt() = ±Inf
//	NaN.Cbrt() = NaN
func (a Float128) Cbrt() Float128 {
	const (
		B1 = 709958130 // B1 = (127-127.0/3-0.03306235651)*2**23
	)

	// special cases
	switch {
	case a.IsZero() || a.IsInf(0) || a.IsNaN():
		return a
	}

	sign := false
	if a.Signbit() {
		a = a.Neg()
		sign = true
	}

	_, exp, frac := a.normalize()
	a = Float128{bias128<<(shift128-64) | frac[0]&fracMask128[0], frac[1] & fracMask128[1]}
	switch exp % 3 {
	case 1, -2:
		a = a.Add(a)
		exp--
	case 2, -1:
		a = a.Add(a)
		a = a.Add(a)
		exp -= 2
	}
	f := Float128{(uint64(exp/3 + bias128)) << (shift128 - 64), 0}

	// ~5-bit estimate
	t32 := NewFloat32FromBits(a.Float32().Bits()/3 + B1)

	// ~16-bit estimate:
	x64 := a.Float64()
	t64 := t32.Float64()
	r64 := t64 * t64 * t64
	t64 = t64 * (x64 + x64 + r64) / (x64 + r64 + r64)

	// ~47-bit estimate:
	r64 = t64 * t64 * t64
	t64 = t64 * (x64 + x64 + r64) / (x64 + r64 + r64)

	// Round t64 away from zero to 47 bits.
	t64 = NewFloat64FromBits(t64.Bits()&0xffff_ffff_ffff_ffe0 + 0x20)
	t := t64.Float128()

	// Final step Newton iteration to 113 bits.
	s := t.Mul(t)
	r := a.Quo(s)
	w := t.Add(t)
	r = r.Sub(t).Quo(w.Add(r))
	t = t.Add(t.Mul(r))

	// Adjust exponent
	t = t.Mul(f)

	// restore the sign bit
	if sign {
		t = t.Neg()
	}
	return t
}
