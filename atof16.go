package floats

import "strconv"

const fnParseFloat16 = "ParseFloat16"

func atof16(s string) (f Float16, n int, err error) {
	if val, n, ok := special(s); ok {
		return NewFloat16(val), n, nil
	}

	mantissa, exp, neg, trunc, hex, n, ok := readFloat(s)
	if !ok {
		return 0, n, syntaxError(fnParseFloat16, s)
	}

	if hex {
		f, err := atof16Hex(s[:n], mantissa, exp, neg, trunc)
		return f, n, err
	}

	var d decimal
	if !d.set(s[:n]) {
		return 0, n, syntaxError(fnParseFloat16, s)
	}
	f, ovf := d.float16()
	if ovf {
		err = rangeError(fnParseFloat16, s)
	}
	return f, n, err
}

// atofHex converts the hex floating-point string s
// to a rounded float16 value and returns it as a float16.
// The string s has already been parsed into a mantissa, exponent, and sign (neg==true for negative).
// If trunc is true, trailing non-zero bits have been omitted from the mantissa.
func atof16Hex(s string, mantissa uint64, exp int, neg, trunc bool) (Float16, error) {
	const maxExp = mask16 - bias16 - 1
	const minExp = -bias16 + 1
	exp += shift16 // mantissa now implicitly divided by 2^shift16.

	// Shift mantissa and exponent to bring representation into float range.
	// Eventually we want a mantissa with a leading 1-bit followed by mantbits other bits.
	// For rounding, we need two more, where the bottom bit represents
	// whether that bit or any later bit was non-zero.
	// (If the mantissa has already lost non-zero bits, trunc is true,
	// and we OR in a 1 below after shifting left appropriately.)
	for mantissa != 0 && mantissa>>(shift16+2) == 0 {
		mantissa <<= 1
		exp--
	}
	if trunc {
		mantissa |= 1
	}
	for mantissa>>(1+shift16+2) != 0 {
		mantissa = mantissa>>1 | mantissa&1
		exp++
	}

	// If exponent is too negative,
	// denormalize in hopes of making it representable.
	// (The -2 is for the rounding bits.)
	for mantissa > 1 && exp < minExp-2 {
		mantissa = mantissa>>1 | mantissa&1
		exp++
	}

	// Round using two bottom bits.
	round := mantissa & 3
	mantissa >>= 2
	round |= mantissa & 1 // round to even (round up if mantissa is odd)
	exp += 2
	if round == 3 {
		mantissa++
		if mantissa == 1<<(1+shift16) {
			mantissa >>= 1
			exp++
		}
	}

	if mantissa>>shift16 == 0 { // Denormal or zero.
		exp = -bias16
	}
	var err error
	if exp > maxExp { // infinity and range error
		mantissa = 1 << shift16
		exp = maxExp + 1
		err = rangeError(fnParseFloat16, s)
	}

	bits := mantissa & fracMask16
	bits |= uint64((exp+bias16)&mask16) << shift16
	if neg {
		bits |= signMask16
	}

	return Float16(bits), err
}

func (d *decimal) float16() (f Float16, overflow bool) {
	var exp int
	var mant uint16

	// Zero is always a special case.
	if d.nd == 0 {
		mant = 0
		exp = -bias16
		goto out
	}

	// Obvious overflow/underflow.
	if d.dp > 5 {
		goto overflow
	}
	if d.dp < -8 {
		// underflow to zero
		mant = 0
		exp = -bias16
		goto out
	}

	// Scale by powers of two until in range [0.5, 1.0)
	exp = 0
	for d.dp > 0 {
		var n int
		if d.dp >= len(powtab) {
			n = 27
		} else {
			n = powtab[d.dp]
		}
		d.Shift(-n)
		exp += n
	}
	for d.dp < 0 || d.dp == 0 && d.d[0] < '5' {
		var n int
		if -d.dp >= len(powtab) {
			n = 27
		} else {
			n = powtab[-d.dp]
		}
		d.Shift(n)
		exp -= n
	}

	// Our range is [0.5,1) but floating point range is [1,2).
	exp--

	// Minimum representable exponent is -bias16+1.
	// If the exponent is smaller, move it up and
	// adjust d accordingly.
	if exp < -bias16+1 {
		n := (-bias16 + 1) - exp
		d.Shift(-n)
		exp += n
	}

	// Check for overflow.
	if exp >= mask16-bias16 {
		goto overflow
	}

	// Extract 1+shift16 bits of mantissa.
	d.Shift(1 + shift16)
	mant = d.RoundedUint16()

	// Denormalized?
	if mant&(1<<shift16) == 0 {
		exp = -bias16
	}
	goto out

overflow:
	// ±Inf
	mant = 0
	exp = mask16 - bias16
	overflow = true

out:
	// Assemble bits.
	bits := mant & fracMask16
	bits |= uint16((exp+bias16)&mask16) << shift16
	if d.neg {
		bits |= signMask16
	}
	return Float16(bits), overflow
}

func (d *decimal) RoundedUint16() uint16 {
	if d.dp > 5 {
		return 0xffff
	}
	var i int
	var n uint16
	for i = 0; i < d.dp && i < d.nd; i++ {
		n = n*10 + uint16(d.d[i]-'0')
	}
	for ; i < d.dp; i++ {
		n *= 10
	}
	if shouldRoundUp(d, d.dp) {
		n++
	}
	return n
}

// ParseFloat16 parses s as a Float16.
func ParseFloat16(s string) (Float16, error) {
	f, n, err := atof16(s)
	if n != len(s) && (err == nil || err.(*strconv.NumError).Err != strconv.ErrSyntax) {
		return NewFloat16(0), syntaxError(fnParseFloat16, s)
	}
	return f, err
}
