package floats

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
