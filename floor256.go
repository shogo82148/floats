package floats

import "github.com/shogo82148/ints"

// Floor returns the greatest integer value less than or equal to x.
//
// Special cases are:
//
//	Floor(±0) = ±0
//	Floor(±Inf) = ±Inf
//	Floor(NaN) = NaN
func (a Float256) Floor() Float256 {
	// Handle special cases
	if a.IsZero() || a.IsNaN() || a.IsInf(0) {
		return a
	}

	if a.Signbit() {
		d, fract := a.Neg().Modf()
		if !fract.IsZero() {
			d = d.Add(Float256(uvone256))
		}
		return d.Neg()
	}
	d, _ := a.Modf()
	return d
}

// Ceil returns the least integer value greater than or equal to a.
//
// Special cases are:
//
//	±0.Ceil() = ±0
//	±Inf.Ceil() = ±Inf
//	NaN.Ceil() = NaN
func (a Float256) Ceil() Float256 {
	return a.Neg().Floor().Neg()
}

// Trunc returns the integer value of a.
//
// Special cases are:
//
//	±0.Trunc() = ±0
//	±Inf.Trunc() = ±Inf
//	NaN.Trunc() = NaN
func (a Float256) Trunc() Float256 {
	if a.IsZero() || a.IsNaN() || a.IsInf(0) {
		return a
	}
	d, _ := a.Modf()
	return d
}

// Round returns the nearest integer, rounding half away from zero.
//
// Special cases are:
//
//	±0.Round() = ±0
//	±Inf.Round() = ±Inf
//	NaN.Round() = NaN
func (a Float256) Round() Float256 {
	bits := a.Bits()
	e := uint((a[0] >> (shift256 - 192)) & mask256)
	if e < bias256 {
		// Round abs(x) < 1 including denormals.
		bits = bits.And(signMask256) // ±0
		if e == bias256-1 {
			// abs(x) >= 0.5
			bits = bits.Or(uvone256) // ±1
		}
	} else if e < bias256+shift256 {
		// Round any abs(x) >= 1 containing a fractional component [0,1).
		//
		// Numbers with larger exponents are returned unchanged since they
		// must be either an integer, infinity, or NaN.
		half := ints.Uint256{1 << (shift256 - 192 - 1), 0, 0, 0}
		e -= bias256
		bits = bits.Add(half.Rsh(e))
		bits = bits.AndNot(fracMask256.Rsh(e))
	}
	return NewFloat256FromBits(bits)
}

// RoundToEven returns the nearest integer, rounding ties to even.
//
// Special cases are:
//
//	±0.RoundToEven() = ±0
//	±Inf.RoundToEven() = ±Inf
//	NaN.RoundToEven() = NaN
func (a Float256) RoundToEven() Float256 {
	bits := a.Bits()
	e := uint((a[0] >> (shift256 - 192)) & mask256)
	if e >= bias256 {
		// Round abs(x) >= 1.
		// - Large numbers without fractional components, infinity, and NaN are unchanged.
		// - Add 0.499.. or 0.5 before truncating depending on whether the truncated
		//   number is even or odd (respectively).
		halfMinusULP := ints.Uint256{1 << (shift256 - 192 - 1), 0, 0, 0}.Sub(ints.Uint256{0, 0, 0, 1})
		e -= bias256
		bits = bits.Add(halfMinusULP.Add(bits.Rsh(shift256 - e).And(ints.Uint256{0, 0, 0, 1})).Rsh(e))
		bits = bits.AndNot(fracMask256.Rsh(e))
	} else if e == bias256-1 && !bits.And(fracMask256).IsZero() {
		// Round 0.5 < abs(x) < 1.
		bits = bits.And(signMask256).Or(uvone256) // ±1
	} else {
		// Round abs(x) <= 0.5 including denormals.
		bits = bits.And(signMask256) // ±0
	}
	return NewFloat256FromBits(bits)
}
