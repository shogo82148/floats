package floats

// Cbrt returns the cube root of a.
//
// Special cases are:
//
//	±0.Cbrt() = ±0
//	±Inf.Cbrt() = ±Inf
//	NaN.Cbrt() = NaN
func (a Float32) Cbrt() Float32 {
	const (
		B1             = 709958130 // B1 = (127-127.0/3-0.03306235651)*2**23
		B2             = 642849266 // B2 = (127-127.0/3-24/3-0.03306235651)*2**23
		SmallestNormal = 0x1p-126  // smallest positive normal number
	)

	// special cases
	switch {
	case a.IsZero() || a.IsInf(0) || a.IsNaN():
		return a
	}

	sign := false
	if a.Signbit() {
		a = -a
		sign = true
	}
	a64 := float64(a)

	t := float64(NewFloat32FromBits(a.Bits()/3 + B1))
	if a < SmallestNormal {
		// subnormal number
		t = a64 * 0x1p24
		t = float64(NewFloat32FromBits(Float32(t).Bits()/3 + B2))
	}

	// First step Newton iteration (solving t*t-x/t == 0) to 16 bits.  In
	// double precision so that its terms can be arranged for efficiency
	// without causing overflow or underflow.
	r := t * t * t
	t = t * (a64 + a64 + r) / (a64 + r + r)

	// Second step Newton iteration to 47 bits.  In double precision for
	// efficiency and accuracy.
	r = t * t * t
	t = t * (a64 + a64 + r) / (a64 + r + r)

	// restore the sign bit
	if sign {
		t = -t
	}

	// rounding to 24 bits is perfect in round-to-nearest mode
	return Float32(t)
}
