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
	sign, exp, frac := a.split()

	// sign, 0x, leading digit
	if sign != 0 {
		dst = append(dst, '-')
	}
	dst = append(dst, '0', fmt) // 0x or 0X
	if a.IsZero() {
		dst = append(dst, '0')
		if prec >= 1 {
			dst = append(dst, '.')
			for range prec {
				dst = append(dst, '0')
			}
		}
		dst = append(dst, fmt-('x'-'p')) // 'p' or 'P'
		return append(dst, "+00"...)
	}
	dst = append(dst, '1')

	hex := lowerHex
	if fmt == 'X' {
		hex = upperHex
	}

	// Shift digits so leading 1 (if any) is at bit 1<<252.
	frac = frac.Lsh(252 - shift256)

	// Round if requested.
	if prec >= 0 && prec < 63 {
		one := ints.Uint256{0, 0, 0, 1}
		shift := uint(prec * 4)
		extra := frac.Lsh(shift).And(one.Lsh(252).Sub(one))
		frac = frac.Rsh(252 - shift)
		if extra.Or(frac.And(one)).Cmp(one.Lsh(251)) > 0 {
			frac = frac.Add(one)
		}
		frac = frac.Lsh(252 - shift)
		if frac.Cmp(one.Lsh(253)) >= 0 {
			// rounded up, e.g., 0x1.ffff... + 0x0.000...1 = 0x2.000...
			frac = frac.Rsh(1)
			exp++
		}
	}

	// .fraction
	frac = frac.Lsh(4) // remove leading 1
	if prec < 0 && !frac.IsZero() {
		dst = append(dst, '.')
		for !frac.IsZero() {
			dst = append(dst, hex[frac.Rsh(252).Uint64()&0xf])
			frac = frac.Lsh(4)
		}
	} else if prec > 0 {
		dst = append(dst, '.')
		for range prec {
			dst = append(dst, hex[frac.Rsh(252).Uint64()&0xf])
			frac = frac.Lsh(4)
		}
	}

	// p±
	dst = append(dst, fmt-('x'-'p')) // 'p' or 'P'
	if exp >= 0 {
		dst = append(dst, '+')
	} else {
		dst = append(dst, '-')
		exp = -exp
	}
	if exp < 10 {
		dst = append(dst, '0')
	}
	dst = strconv.AppendInt(dst, int64(exp), 10)
	return dst
}

func (a Float256) append(dst []byte, fmt byte, prec int) []byte {
	return dst
}
