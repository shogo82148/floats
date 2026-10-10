package floats

import (
	"encoding"
	"encoding/json"
	"fmt"
	"strconv"
)

// float8Params describes the layout of an 8-bit format.
type float8Params struct {
	fn    string // name of the parse function, for the errors
	mbits int    // number of the bits of the fraction
	bias  int    // bias of the exponent
}

var (
	paramsE4M3 = float8Params{"ParseFloat8E4M3", 3, 7}
	paramsE5M2 = float8Params{"ParseFloat8E5M2", 2, 15}
)

// decimalDigits8 is the number of the decimal digits that is enough
// to format and parse the 8-bit formats.
const decimalDigits8 = 64

// float8Overflow is the encoding that is returned by the parsers
// when the value is too large; it is larger than all of the finite encodings.
const float8Overflow = 0x3ff

// split splits the magnitude and the sign of b.
// frac contains the implicit bit if b is normal.
func (p float8Params) split(b uint8) (sign uint8, exp int, frac uint8) {
	sign = b & 0x80
	exp = int(b>>p.mbits&(1<<(7-p.mbits)-1)) - p.bias
	frac = b & (1<<p.mbits - 1)
	if exp == -p.bias {
		// subnormal
		exp++
	} else {
		frac |= 1 << p.mbits
	}
	return
}

// append appends the representation of the finite number b.
func (p float8Params) append(dst []byte, b uint8, fmt byte, prec int, hex func([]byte, byte, int) []byte) []byte {
	switch fmt {
	case 'b':
		sign, exp, frac := p.split(b)
		exp -= p.mbits
		if sign != 0 {
			dst = append(dst, '-')
		}
		dst = strconv.AppendUint(dst, uint64(frac), 10)
		dst = append(dst, 'p')
		if exp >= 0 {
			dst = append(dst, '+')
		}
		return strconv.AppendInt(dst, int64(exp), 10)
	case 'x', 'X':
		// The hexadecimal representation depends only on the value,
		// and the value is exactly representable in Float32.
		return hex(dst, fmt, prec)
	case 'f', 'e', 'E', 'g', 'G':
		return p.appendDecimal(dst, b, fmt, prec)
	}

	// unknown format
	return append(dst, '%', fmt)
}

func (p float8Params) appendDecimal(dst []byte, b uint8, fmt byte, prec int) []byte {
	sign, exp, frac := p.split(b)

	var buf [decimalDigits8]byte
	d := &decimal{d: buf[:]}
	d.AssignUint64(uint64(frac))
	d.Shift(exp - p.mbits)
	shortest := prec < 0
	if shortest {
		p.roundShortest(d, frac, exp)
		// Precision for shortest representation mode.
		switch fmt {
		case 'e', 'E':
			prec = d.nd - 1
		case 'f':
			prec = max(d.nd-d.dp, 0)
		case 'g', 'G':
			prec = d.nd
		}
	} else {
		// Round appropriately.
		switch fmt {
		case 'e', 'E':
			d.Round(prec + 1)
		case 'f':
			d.Round(d.dp + prec)
		case 'g', 'G':
			if prec == 0 {
				prec = 1
			}
			d.Round(prec)
		}
	}
	return formatDigits(dst, sign != 0, d, shortest, prec, fmt)
}

func (p float8Params) roundShortest(d *decimal, frac uint8, exp int) {
	// If mantissa is zero, the number is zero; stop now.
	if frac == 0 {
		d.nd = 0
		return
	}

	minexp := -p.bias + 1 // minimum possible exponent

	// d = frac << (exp - mbits)
	// Next highest floating point number is frac+1 << exp-mbits.
	// Our upper bound is halfway between, frac*2+1 << exp-mbits-1.
	var upperBuf [decimalDigits8]byte
	upper := &decimal{d: upperBuf[:]}
	upper.AssignUint64(uint64(frac)*2 + 1)
	upper.Shift(exp - p.mbits - 1)

	// d = frac << (exp - mbits)
	// Next lowest floating point number is frac-1 << exp-mbits,
	// unless frac-1 drops the significant bit and exp is not the minimum exp,
	// in which case the next lowest is frac*2-1 << exp-mbits-1.
	// Either way, call it fraclo << explo-mbits.
	// Our lower bound is halfway between, fraclo*2+1 << explo-mbits-1.
	var fraclo uint8
	var explo int
	if frac > 1<<p.mbits || exp == minexp {
		fraclo = frac - 1
		explo = exp
	} else {
		fraclo = frac*2 - 1
		explo = exp - 1
	}
	var lowerBuf [decimalDigits8]byte
	lower := &decimal{d: lowerBuf[:]}
	lower.AssignUint64(uint64(fraclo)*2 + 1)
	lower.Shift(explo - p.mbits - 1)

	// The upper and lower bounds are possible outputs only if
	// the original mantissa is even, so that IEEE round-to-even
	// would round to the original mantissa and not the neighbors.
	inclusive := frac%2 == 0

	// As we walk the digits we want to know whether rounding up would fall
	// within the upper bound. This is tracked by upperdelta:
	//
	// If upperdelta == 0, the digits of d and upper are the same so far.
	//
	// If upperdelta == 1, we saw a difference of 1 between d and upper on a
	// previous digit and subsequently only 9s for d and 0s for upper.
	// (Thus rounding up may fall outside the bound, if it is exclusive.)
	//
	// If upperdelta == 2, then the difference is greater than 1
	// and we know that rounding up falls within the bound.
	var upperdelta uint8

	// Now we can figure out the minimum number of digits required.
	// Walk along until d has distinguished itself from upper and lower.
	for ui := 0; ; ui++ {
		// lower, d, and upper may have the decimal points at different
		// places. In this case upper is the longest, so we iterate from
		// ui==0 and start li and mi at (possibly) -1.
		mi := ui - upper.dp + d.dp
		if mi >= d.nd {
			break
		}
		li := ui - upper.dp + lower.dp
		l := byte('0') // lower digit
		if li >= 0 && li < lower.nd {
			l = lower.d[li]
		}
		m := byte('0') // middle digit
		if mi >= 0 {
			m = d.d[mi]
		}
		u := byte('0') // upper digit
		if ui < upper.nd {
			u = upper.d[ui]
		}

		// Okay to round down (truncate) if lower has a different digit
		// or if lower is inclusive and is exactly the result of rounding
		// down (i.e., and we have reached the final digit of lower).
		okdown := l != m || inclusive && li+1 == lower.nd

		switch {
		case upperdelta == 0 && m+1 < u:
			// Example:
			// m = 12345xxx
			// u = 12347xxx
			upperdelta = 2
		case upperdelta == 0 && m != u:
			// Example:
			// m = 12345xxx
			// u = 12346xxx
			upperdelta = 1
		case upperdelta == 1 && (m != '9' || u != '0'):
			// Example:
			// m = 1234598x
			// u = 1234600x
			upperdelta = 2
		}
		// Okay to round up if upper has a different digit and either upper
		// is inclusive or upper is bigger than the result of rounding up.
		okup := upperdelta > 0 && (inclusive || upperdelta > 1 || ui+1 < upper.nd)

		// If it's okay to do either, then round to the nearest one.
		// If it's okay to do only one, do it.
		switch {
		case okdown && okup:
			d.Round(mi + 1)
			return
		case okdown:
			d.RoundDown(mi + 1)
			return
		case okup:
			d.RoundUp(mi + 1)
			return
		}
	}
}

// Text returns the string representation of a in the given format and precision.
func (a Float8E4M3) Text(fmt byte, prec int) string {
	return string(a.Append(make([]byte, 0, 8), fmt, prec))
}

// Text returns the string representation of a in the given format and precision.
func (a Float8E5M2) Text(fmt byte, prec int) string {
	return string(a.Append(make([]byte, 0, 8), fmt, prec))
}

// Append appends the string representation of a in the given format and precision to dst and returns the extended buffer.
func (a Float8E4M3) Append(dst []byte, fmt byte, prec int) []byte {
	if a.IsNaN() {
		return append(dst, "NaN"...)
	}
	return paramsE4M3.append(dst, uint8(a), fmt, prec, func(dst []byte, fmt byte, prec int) []byte {
		return a.Float32().Append(dst, fmt, prec)
	})
}

// Append appends the string representation of a in the given format and precision to dst and returns the extended buffer.
func (a Float8E5M2) Append(dst []byte, fmt byte, prec int) []byte {
	switch {
	case a.IsNaN():
		return append(dst, "NaN"...)
	case a.IsInf(1):
		return append(dst, "+Inf"...)
	case a.IsInf(-1):
		return append(dst, "-Inf"...)
	}
	return paramsE5M2.append(dst, uint8(a), fmt, prec, func(dst []byte, fmt byte, prec int) []byte {
		return a.Float32().Append(dst, fmt, prec)
	})
}

var _ fmt.Formatter = Float8E4M3(0)

// Format implements [fmt.Formatter].
func (a Float8E4M3) Format(s fmt.State, verb rune) {
	format(a, s, verb)
}

var _ fmt.Formatter = Float8E5M2(0)

// Format implements [fmt.Formatter].
func (a Float8E5M2) Format(s fmt.State, verb rune) {
	format(a, s, verb)
}

var _ fmt.Stringer = Float8E4M3(0)

// String returns the string representation of a.
func (a Float8E4M3) String() string {
	return a.Text('g', -1)
}

var _ fmt.Stringer = Float8E5M2(0)

// String returns the string representation of a.
func (a Float8E5M2) String() string {
	return a.Text('g', -1)
}

var _ json.Marshaler = Float8E4M3(0)

// MarshalJSON implements [json.Marshaler].
func (a Float8E4M3) MarshalJSON() ([]byte, error) {
	// JSON does not support NaN values.
	if a.IsNaN() {
		return nil, fmt.Errorf("floats: cannot marshal %v to JSON", a)
	}
	return a.Append(nil, 'g', -1), nil
}

var _ json.Marshaler = Float8E5M2(0)

// MarshalJSON implements [json.Marshaler].
func (a Float8E5M2) MarshalJSON() ([]byte, error) {
	// JSON does not support NaN and Inf values.
	if a.IsNaN() || a.IsInf(0) {
		return nil, fmt.Errorf("floats: cannot marshal %v to JSON", a)
	}
	return a.Append(nil, 'g', -1), nil
}

var _ encoding.TextMarshaler = Float8E4M3(0)

// MarshalText implements [encoding.TextMarshaler].
func (a Float8E4M3) MarshalText() ([]byte, error) {
	return a.Append(nil, 'g', -1), nil
}

var _ encoding.TextMarshaler = Float8E5M2(0)

// MarshalText implements [encoding.TextMarshaler].
func (a Float8E5M2) MarshalText() ([]byte, error) {
	return a.Append(nil, 'g', -1), nil
}

// atof8 parses the prefix of s as a floating-point number except the special values.
// It returns the sign and the encoding of the magnitude rounded to nearest even.
// The encoding is not checked whether it is finite:
// it is larger than the finite encodings if the value is too large.
func atof8(s string, p float8Params) (neg bool, enc uint32, n int, err error) {
	mantissa, exp, neg, trunc, hex, n, ok := readFloat(s)
	if !ok {
		return false, 0, n, syntaxError(p.fn, s)
	}
	if hex {
		return neg, atof8Hex(mantissa, exp, trunc, p), n, nil
	}

	var buf [decimalDigits8]byte
	d := decimal{d: buf[:]}
	if !d.set(s[:n]) {
		return false, 0, n, syntaxError(p.fn, s)
	}
	return neg, d.float8(p), n, nil
}

// atof8Hex rounds the hex floating-point number, which has been parsed into a mantissa and an exponent.
// If trunc is true, trailing non-zero bits have been omitted from the mantissa.
func atof8Hex(mantissa uint64, exp int, trunc bool, p float8Params) uint32 {
	minExp := -p.bias + 1
	exp += p.mbits // mantissa now implicitly divided by 2^mbits.

	// Shift mantissa and exponent to bring representation into float range.
	// Eventually we want a mantissa with a leading 1-bit followed by mbits other bits.
	// For rounding, we need two more, where the bottom bit represents
	// whether that bit or any later bit was non-zero.
	for mantissa != 0 && mantissa>>(p.mbits+2) == 0 {
		mantissa <<= 1
		exp--
	}
	if trunc {
		mantissa |= 1
	}
	for mantissa>>(1+p.mbits+2) != 0 {
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
		if mantissa == 1<<(1+p.mbits) {
			mantissa >>= 1
			exp++
		}
	}

	field := 0
	if mantissa>>p.mbits != 0 { // not denormal nor zero
		field = exp + p.bias
	}
	if field > 63 {
		return float8Overflow
	}
	return uint32(field)<<p.mbits | uint32(mantissa)&(1<<p.mbits-1)
}

// float8 rounds d to the 8-bit format.
// See [atof8] for the result.
func (d *decimal) float8(p float8Params) uint32 {
	// Zero is always a special case.
	if d.nd == 0 {
		return 0
	}

	// Obvious overflow/underflow.
	if d.dp > 40 {
		return float8Overflow
	}
	if d.dp < -6 {
		// underflow to zero
		return 0
	}

	// Scale by powers of two until in range [0.5, 1.0)
	exp := 0
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

	// Minimum representable exponent is -bias+1.
	// If the exponent is smaller, move it up and
	// adjust d accordingly.
	if exp < -p.bias+1 {
		n := (-p.bias + 1) - exp
		d.Shift(-n)
		exp += n
	}

	// Extract 1+mbits bits of mantissa.
	d.Shift(1 + p.mbits)
	mant := uint32(d.RoundedUint16())

	// Rounding might have added a bit; shift down.
	if mant == 2<<p.mbits {
		mant >>= 1
		exp++
	}

	field := 0
	if mant&(1<<p.mbits) != 0 { // not denormal
		field = exp + p.bias
	}
	if field > 63 {
		return float8Overflow
	}
	return uint32(field)<<p.mbits | mant&(1<<p.mbits-1)
}

func atofE4M3(s string) (Float8E4M3, int, error) {
	if val, n, ok := special(s); ok {
		return NewFloat8E4M3(val), n, nil
	}
	neg, enc, n, err := atof8(s, paramsE4M3)
	if err != nil {
		return 0, n, err
	}
	var sign uint8
	if neg {
		sign = 0x80
	}
	if enc > 0x7e {
		// Float8E4M3 has no infinity.
		return Float8E4M3(sign | uvnane4m3), n, rangeError(paramsE4M3.fn, s)
	}
	return Float8E4M3(sign | uint8(enc)), n, nil
}

func atofE5M2(s string) (Float8E5M2, int, error) {
	if val, n, ok := special(s); ok {
		return NewFloat8E5M2(val), n, nil
	}
	neg, enc, n, err := atof8(s, paramsE5M2)
	if err != nil {
		return 0, n, err
	}
	var sign uint8
	if neg {
		sign = 0x80
	}
	if enc >= uvinfe5m2 {
		return Float8E5M2(sign | uvinfe5m2), n, rangeError(paramsE5M2.fn, s)
	}
	return Float8E5M2(sign | uint8(enc)), n, nil
}

// ParseFloat8E4M3 parses s as a Float8E4M3.
// If s is out of range, it returns NaN and an error whose Err is [strconv.ErrRange].
func ParseFloat8E4M3(s string) (Float8E4M3, error) {
	f, n, err := atofE4M3(s)
	if n != len(s) && (err == nil || err.(*strconv.NumError).Err != strconv.ErrSyntax) {
		return 0, syntaxError(paramsE4M3.fn, s)
	}
	return f, err
}

// ParseFloat8E5M2 parses s as a Float8E5M2.
// If s is out of range, it returns ±Inf and an error whose Err is [strconv.ErrRange].
func ParseFloat8E5M2(s string) (Float8E5M2, error) {
	f, n, err := atofE5M2(s)
	if n != len(s) && (err == nil || err.(*strconv.NumError).Err != strconv.ErrSyntax) {
		return 0, syntaxError(paramsE5M2.fn, s)
	}
	return f, err
}

var _ json.Unmarshaler = (*Float8E4M3)(nil)

// UnmarshalJSON implements [json.Unmarshaler].
func (a *Float8E4M3) UnmarshalJSON(data []byte) error {
	ret, err := ParseFloat8E4M3(string(data))
	if err != nil {
		return err
	}
	*a = ret
	return nil
}

var _ json.Unmarshaler = (*Float8E5M2)(nil)

// UnmarshalJSON implements [json.Unmarshaler].
func (a *Float8E5M2) UnmarshalJSON(data []byte) error {
	ret, err := ParseFloat8E5M2(string(data))
	if err != nil {
		return err
	}
	*a = ret
	return nil
}

var _ encoding.TextUnmarshaler = (*Float8E4M3)(nil)

// UnmarshalText implements [encoding.TextUnmarshaler].
func (a *Float8E4M3) UnmarshalText(data []byte) error {
	ret, err := ParseFloat8E4M3(string(data))
	if err != nil {
		return err
	}
	*a = ret
	return nil
}

var _ encoding.TextUnmarshaler = (*Float8E5M2)(nil)

// UnmarshalText implements [encoding.TextUnmarshaler].
func (a *Float8E5M2) UnmarshalText(data []byte) error {
	ret, err := ParseFloat8E5M2(string(data))
	if err != nil {
		return err
	}
	*a = ret
	return nil
}

var _ encoding.TextAppender = Float8E4M3(0)

// AppendText implements [encoding.TextAppender].
func (a Float8E4M3) AppendText(dst []byte) ([]byte, error) {
	return a.Append(dst, 'g', -1), nil
}

var _ encoding.TextAppender = Float8E5M2(0)

// AppendText implements [encoding.TextAppender].
func (a Float8E5M2) AppendText(dst []byte) ([]byte, error) {
	return a.Append(dst, 'g', -1), nil
}
