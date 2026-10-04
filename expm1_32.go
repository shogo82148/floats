package floats

// Expm1 returns e**a - 1, the base-e exponential of a minus 1.
// It is more accurate than Exp(a) - 1 when a is near zero.
//
// Special cases are:
//
//	+Inf.Expm1() = +Inf
//	-Inf.Expm1() = -1
//	NaN.Expm1() = NaN
//
// Very large values overflow to -1 or +Inf.
func (a Float32) Expm1() Float32 {
	ix := a.Bits() &^ signMask32
	if ix < 0x33800000 { // |a| < 2**-24
		// expm1(a) = a + a**2/2 + ... rounds to a.
		return a
	}

	x := float64(a)
	if !(x >= -104 && x <= 89) {
		switch {
		case x != x:
			// expm1(NaN) = NaN
			return NewFloat32NaN()
		case x > 0:
			// expm1(x) overflows.
			return NewFloat32Inf(1)
		default:
			// expm1(x) rounds to -1.
			return -1
		}
	}

	// The result may overflow in float32,
	// but it is correctly handled by the conversion from float64.
	return Float32(expm1Small(x))
}
