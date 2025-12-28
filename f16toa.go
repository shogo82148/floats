// convert float16 to string

package floats

import "fmt"

var _ fmt.Stringer = Float16(0)

// String returns the string representation of a.
func (a Float16) String() string {
	return a.Text('g', -1)
}

// Text returns the string representation of a in the given format and precision.
func (a Float16) Text(fmt byte, prec int) string {
	return string(a.Append(make([]byte, 0, 8), fmt, prec))
}

// Append appends the string representation of a in the given format and precision to buf and returns the extended buffer.
func (a Float16) Append(buf []byte, fmt byte, prec int) []byte {
	// special numbers
	switch {
	case a.IsNaN():
		return append(buf, "NaN"...)
	case a.IsInf(1):
		return append(buf, "+Inf"...)
	case a.IsInf(-1):
		return append(buf, "-Inf"...)
	}

	switch fmt {
	case 'b':
		return a.appendBin(buf)
	case 'x', 'X':
		return a.appendHex(buf, fmt, prec)
	case 'f', 'F':
	case 'e', 'E':
	case 'g', 'G':
	}

	// unknown format
	return append(buf, '%', fmt)
}

func (a Float16) appendBin(buf []byte) []byte {
	if a&signMask16 != 0 {
		buf = append(buf, '-')
	}
	exp := int(a>>shift16&mask16) - bias16
	frac := a & fracMask16
	if exp == -bias16 {
		exp++
	} else {
		frac |= 1 << shift16
	}
	exp -= shift16

	switch {
	case frac >= 1000:
		buf = append(buf, byte((frac/1000)%10)+'0')
		fallthrough
	case frac >= 100:
		buf = append(buf, byte((frac/100)%10)+'0')
		fallthrough
	case frac >= 10:
		buf = append(buf, byte((frac/10)%10)+'0')
		fallthrough
	default:
		buf = append(buf, byte(frac%10)+'0')
	}

	buf = append(buf, 'p')
	if exp >= 0 {
		buf = append(buf, '+')
	} else {
		buf = append(buf, '-')
		exp = -exp
	}

	switch {
	case exp >= 10:
		buf = append(buf, byte(exp/10)+'0')
		fallthrough
	default:
		buf = append(buf, byte(exp%10)+'0')
	}
	return buf
}

func (a Float16) appendHex(buf []byte, fmt byte, prec int) []byte {
	sign, exp, frac := a.split()
	if sign != 0 {
		buf = append(buf, '-')
	}
	buf = append(buf, '0', fmt) // 0x or 0X
	if a.IsZero() {
		buf = append(buf, '0')
		if prec >= 1 {
			buf = append(buf, '.')
			for range prec {
				buf = append(buf, '0')
			}
		}
		buf = append(buf, fmt-('x'-'p')) // 'p' or 'P'
		return append(buf, "+00"...)
	}

	switch {
	case prec < 0:
		if frac&0x3ff == 0 {
			buf = append(buf, '1')
		} else if frac&0x3f == 0 {
			buf = append(buf, '1', '.')
			buf = append(buf, nibble(fmt, byte(frac>>6)))
		} else if frac&0x3 == 0 {
			buf = append(buf, '1', '.')
			buf = append(buf, nibble(fmt, byte(frac>>6)))
			buf = append(buf, nibble(fmt, byte(frac>>2)))
		} else {
			buf = append(buf, '1', '.')
			buf = append(buf, nibble(fmt, byte(frac>>6)))
			buf = append(buf, nibble(fmt, byte(frac>>2)))
			buf = append(buf, nibble(fmt, byte(frac<<2)))
		}
	case prec == 0:
		// round to nearest even
		frac += 1 << (shift16 - 1)
		if frac >= 1<<(shift16+1) {
			exp++
		}
		buf = append(buf, '1')
	case prec == 1:
		// round to nearest even
		frac += 0x1f + (frac>>6)&1
		if frac >= 1<<(shift16+1) {
			exp++
			frac >>= 1
		}

		buf = append(buf, '1', '.')
		buf = append(buf, nibble(fmt, byte(frac>>6)))
	case prec == 2:
		// round to nearest even
		frac += 1 + (frac>>2)&1
		if frac >= 1<<(shift16+1) {
			exp++
			frac >>= 1
		}

		buf = append(buf, '1', '.')
		buf = append(buf, nibble(fmt, byte(frac>>6)))
		buf = append(buf, nibble(fmt, byte(frac>>2)))
	default:
		buf = append(buf, '1', '.')
		buf = append(buf, nibble(fmt, byte(frac>>6)))
		buf = append(buf, nibble(fmt, byte(frac>>2)))
		buf = append(buf, nibble(fmt, byte(frac<<2)))
		for i := 3; i < prec; i++ {
			buf = append(buf, '0')
		}
	}

	buf = append(buf, fmt-('x'-'p'))
	if exp >= 0 {
		buf = append(buf, '+')
	} else {
		buf = append(buf, '-')
		exp = -exp
	}
	buf = append(buf, byte(exp/10)+'0', byte(exp%10)+'0')
	return buf
}
