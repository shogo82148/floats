package floats

import "strconv"

const fnParseFloat16 = "ParseFloat16"

func atof16(s string) (Float16, int, error) {
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
	return NewFloat16(0), len(s), nil
}

// atofHex converts the hex floating-point string s
// to a rounded float16 value and returns it as a float16.
// The string s has already been parsed into a mantissa, exponent, and sign (neg==true for negative).
// If trunc is true, trailing non-zero bits have been omitted from the mantissa.
func atof16Hex(s string, mantissa uint64, exp int, neg, trunc bool) (Float16, error) {
	maxExp := mask16 + bias16 - 2
	minExp := bias16 + 1
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
		exp = bias16
	}
	var err error
	if exp > maxExp { // infinity and range error
		mantissa = 1 << shift16
		exp = maxExp + 1
		err = rangeError(fnParseFloat16, s)
	}

	bits := mantissa & fracMask16
	bits |= uint64((exp-bias16)&mask16) << shift16
	if neg {
		bits |= signMask16
	}

	return Float16(bits), err
}

// ParseFloat16 parses s as a Float16.
func ParseFloat16(s string) (Float16, error) {
	f, n, err := atof16(s)
	if n != len(s) && (err == nil || err.(*strconv.NumError).Err != strconv.ErrSyntax) {
		return NewFloat16(0), syntaxError(fnParseFloat16, s)
	}
	return f, err
}
