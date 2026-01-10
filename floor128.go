package floats

import "github.com/shogo82148/ints"

// Floor returns the greatest integer value less than or equal to a.
//
// Special cases are:
//
//	±0.Floor() = ±0
//	±Inf.Floor() = ±Inf
//	NaN.Floor() = NaN
func (a Float128) Floor() Float128 {
	// Handle special cases
	if a.IsZero() || a.IsNaN() || a.IsInf(0) {
		return a
	}

	if a.Signbit() {
		d, fract := a.Neg().Modf()
		if !fract.IsZero() {
			d = d.Add(Float128(uvone128))
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
func (a Float128) Ceil() Float128 {
	return a.Neg().Floor().Neg()
}

// Trunc returns the integer value of a.
//
// Special cases are:
//
//	±0.Trunc() = ±0
//	±Inf.Trunc() = ±Inf
//	NaN.Trunc() = NaN
func (a Float128) Trunc() Float128 {
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
func (a Float128) Round() Float128 {
	bits := a.Bits()
	e := uint((a[0] >> (shift128 - 64)) & mask128)
	if e < bias128 {
		// Round abs(x) < 1 including denormals.
		bits = bits.And(signMask128) // ±0
		if e == bias128-1 {
			// abs(x) >= 0.5
			bits = bits.Or(uvone128) // ±1
		}
	} else if e < bias128+shift128 {
		// Round any abs(x) >= 1 containing a fractional component [0,1).
		//
		// Numbers with larger exponents are returned unchanged since they
		// must be either an integer, infinity, or NaN.
		half := ints.Uint128{1 << (shift128 - 64 - 1), 0}
		e -= bias128
		bits = bits.Add(half.Rsh(e))
		bits = bits.AndNot(fracMask128.Rsh(e))
	}
	return NewFloat128FromBits(bits)
}

// RoundToEven returns the nearest integer, rounding ties to even.
//
// Special cases are:
//
//	±0.RoundToEven() = ±0
//	±Inf.RoundToEven() = ±Inf
//	NaN.RoundToEven() = NaN
func (a Float128) RoundToEven() Float128 {
	bits := a.Bits()
	e := uint((a[0] >> (shift128 - 64)) & mask128)
	if e >= bias128 {
		// Round abs(x) >= 1.
		// - Large numbers without fractional components, infinity, and NaN are unchanged.
		// - Add 0.499.. or 0.5 before truncating depending on whether the truncated
		//   number is even or odd (respectively).
		halfMinusULP := ints.Uint128{1 << (shift128 - 64 - 1), 0}.Sub(ints.Uint128{0, 1})
		e -= bias128
		bits = bits.Add(halfMinusULP.Add(bits.Rsh(shift128 - e).And(ints.Uint128{0, 1})).Rsh(e))
		bits = bits.AndNot(fracMask128.Rsh(e))
	} else if e == bias128-1 && !bits.And(fracMask128).IsZero() {
		// Round 0.5 < abs(x) < 1.
		bits = bits.And(signMask128).Or(uvone128) // ±1
	} else {
		// Round abs(x) <= 0.5 including denormals.
		bits = bits.And(signMask128) // ±0
	}
	return NewFloat128FromBits(bits)
}
