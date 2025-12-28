package floats

import (
	"fmt"
	"strconv"

	"github.com/shogo82148/ints"
)

var _ fmt.Stringer = Float256{}

// String returns the string representation of a.
func (a Float256) String() string {
	return a.Text('g', -1)
}

// Text returns the string representation of a in the given format and precision.
func (a Float256) Text(fmt byte, prec int) string {
	return string(a.Append(make([]byte, 0, 64), fmt, prec))
}

// Append appends the string representation of a in the given format and precision to buf and returns the extended buffer.
func (a Float256) Append(buf []byte, fmt byte, prec int) []byte {
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
	case 'f', 'e', 'E', 'g', 'G':
		return a.append(buf, fmt, prec)
	}

	// unknown format
	return append(buf, '%', fmt)
}

func (a Float256) appendBin(dst []byte) []byte {
	if a.Signbit() {
		dst = append(dst, '-')
	}
	b := ints.Uint256(a)
	exp := int((b[0]>>(shift256-192))&mask256) - bias256
	frac := b.And(fracMask256)
	if exp == -bias256 {
		exp++
	} else {
		frac[0] = frac[0] | (1 << (shift256 - 192))
	}
	exp -= shift256

	dst = frac.Append(dst, 10)
	dst = append(dst, 'p')
	if exp >= 0 {
		dst = append(dst, '+')
	} else {
		dst = append(dst, '-')
		exp = -exp
	}
	dst = strconv.AppendInt(dst, int64(exp), 10)
	return dst
}

// %x: -0x1.yyyyyyyyp±ddd or -0x0p+0. (y is hex digit, d is decimal digit)
func (a Float256) appendHex(dst []byte, fmt byte, prec int) []byte {
	return dst
}

func (a Float256) append(dst []byte, fmt byte, prec int) []byte {
	return dst
}
