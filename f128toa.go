package floats

import (
	"fmt"
	"strconv"

	"github.com/shogo82148/ints"
)

var _ fmt.Stringer = Float128{}

// String returns the string representation of a.
func (a Float128) String() string {
	return a.Text('g', -1)
}

// Text returns the string representation of a in the given format and precision.
func (a Float128) Text(fmt byte, prec int) string {
	return string(a.Append(make([]byte, 0, 32), fmt, prec))
}

// Append appends the string representation of a in the given format and precision to buf and returns the extended buffer.
func (a Float128) Append(buf []byte, fmt byte, prec int) []byte {
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
		// case 'x', 'X':
		// 	return a.appendHex(buf, fmt, prec)
		// case 'f', 'e', 'E', 'g', 'G':
		// 	return a.append(buf, fmt, prec)
	}

	// unknown format
	return append(buf, '%', fmt)
}

func (a Float128) appendBin(buf []byte) []byte {
	if a.Signbit() {
		buf = append(buf, '-')
	}
	b := ints.Uint128(a)
	exp := int((b[0]>>(shift128-64))&mask128) - bias128
	frac := b.And(fracMask128)
	if exp == -bias128 {
		exp++
	} else {
		frac[0] = frac[0] | (1 << (shift128 - 64))
	}
	exp -= shift128

	buf = frac.Append(buf, 10)
	buf = append(buf, 'p')
	if exp >= 0 {
		buf = append(buf, '+')
	} else {
		buf = append(buf, '-')
		exp = -exp
	}
	buf = strconv.AppendInt(buf, int64(exp), 10)

	return buf
}
