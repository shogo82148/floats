package floats

import (
	"encoding"
	"encoding/json"
	"strconv"
)

const fnParseBFloat16 = "ParseBFloat16"

func atofbf16(s string) (f BFloat16, n int, err error) {
	if val, n, ok := special(s); ok {
		return NewBFloat16(val), n, nil
	}

	mantissa, exp, neg, trunc, hex, n, ok := readFloat(s)
	if !ok {
		return 0, n, syntaxError(fnParseBFloat16, s)
	}

	if hex {
		f, err := atofbf16Hex(s[:n], mantissa, exp, neg, trunc)
		return f, n, err
	}

	var buf [decimalDigitsBF16]byte
	d := decimal{d: buf[:]}
	if !d.set(s[:n]) {
		return 0, n, syntaxError(fnParseBFloat16, s)
	}
	f, ovf := d.bfloat16()
	if ovf {
		err = rangeError(fnParseBFloat16, s)
	}
	return f, n, err
}

// atofHex converts the hex floating-point string s
// to a rounded bfloat16 value and returns it as a bfloat16.
// The string s has already been parsed into a mantissa, exponent, and sign (neg==true for negative).
// If trunc is true, trailing non-zero bits have been omitted from the mantissa.
func atofbf16Hex(s string, mantissa uint64, exp int, neg, trunc bool) (BFloat16, error) {
	const maxExp = maskbf16 - biasbf16 - 1
	const minExp = -biasbf16 + 1
	exp += shiftbf16 // mantissa now implicitly divided by 2^shiftbf16.

	// Shift mantissa and exponent to bring representation into float range.
	// Eventually we want a mantissa with a leading 1-bit followed by mantbits other bits.
	// For rounding, we need two more, where the bottom bit represents
	// whether that bit or any later bit was non-zero.
	// (If the mantissa has already lost non-zero bits, trunc is true,
	// and we OR in a 1 below after shifting left appropriately.)
	for mantissa != 0 && mantissa>>(shiftbf16+2) == 0 {
		mantissa <<= 1
		exp--
	}
	if false {
		mantissa |= 1
	}
	for mantissa>>(1+shiftbf16+2) != 0 {
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
		if mantissa == 1<<(1+shiftbf16) {
			mantissa >>= 1
			exp++
		}
	}

	if mantissa>>shiftbf16 == 0 { // Denormal or zero.
		exp = -biasbf16
	}
	var err error
	if exp > maxExp { // infinity and range error
		mantissa = 1 << shiftbf16
		exp = maxExp + 1
		err = rangeError(fnParseBFloat16, s)
	}

	bits := mantissa & fracMaskbf16
	bits |= uint64((exp+biasbf16)&maskbf16) << shiftbf16
	if neg {
		bits |= signMaskbf16
	}

	return BFloat16(bits), err
}

func (d *decimal) bfloat16() (f BFloat16, overflow bool) {
	var exp int
	var mant uint16

	// Zero is always a special case.
	if d.nd == 0 {
		mant = 0
		exp = -biasbf16
		goto out
	}

	// Obvious overflow/underflow.
	if d.dp > 40 {
		goto overflow
	}
	if d.dp < -42 {
		// underflow to zero
		mant = 0
		exp = -biasbf16
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

	// Minimum representable exponent is -biasbf16+1.
	// If the exponent is smaller, move it up and
	// adjust d accordingly.
	if exp < -biasbf16+1 {
		n := (-biasbf16 + 1) - exp
		d.Shift(-n)
		exp += n
	}

	// Check for overflow.
	if exp >= maskbf16-biasbf16 {
		goto overflow
	}

	// Extract 1+shiftbf16 bits of mantissa.
	d.Shift(1 + shiftbf16)
	mant = d.RoundedUint16()

	// Rounding might have added a bit; shift down.
	if mant == 2<<shiftbf16 {
		mant >>= 1
		exp++
		if exp >= maskbf16-biasbf16 {
			goto overflow
		}
	}

	// Denormalized?
	if mant&(1<<shiftbf16) == 0 {
		exp = -biasbf16
	}
	goto out

overflow:
	// ±Inf
	mant = 0
	exp = maskbf16 - biasbf16
	overflow = true

out:
	// Assemble bits.
	bits := mant & fracMaskbf16
	bits |= uint16((exp+biasbf16)&maskbf16) << shiftbf16
	if d.neg {
		bits |= signMaskbf16
	}
	return BFloat16(bits), overflow
}

// ParseBFloat16 parses s as a BFloat16.
func ParseBFloat16(s string) (BFloat16, error) {
	f, n, err := atofbf16(s)
	if n != len(s) && (err == nil || err.(*strconv.NumError).Err != strconv.ErrSyntax) {
		return NewBFloat16(0), syntaxError(fnParseBFloat16, s)
	}
	return f, err
}

var _ json.Unmarshaler = (*BFloat16)(nil)

// UnmarshalJSON implements [json.Unmarshaler].
func (a *BFloat16) UnmarshalJSON(data []byte) error {
	ret, err := ParseBFloat16(string(data))
	if err != nil {
		return err
	}
	*a = ret
	return nil
}

var _ encoding.TextUnmarshaler = (*BFloat16)(nil)

// UnmarshalText implements [encoding.TextUnmarshaler].
func (a *BFloat16) UnmarshalText(data []byte) error {
	ret, err := ParseBFloat16(string(data))
	if err != nil {
		return err
	}
	*a = ret
	return nil
}
