package floats

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

// Trunc returns the integer value of x.
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
