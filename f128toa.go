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
	case 'x', 'X':
		return a.appendHex(buf, fmt, prec)
	case 'f', 'e', 'E', 'g', 'G':
		return a.append(buf, fmt, prec)
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

// %x: -0x1.yyyyyyyyp±ddd or -0x0p+0. (y is hex digit, d is decimal digit)
func (a Float128) appendHex(buf []byte, fmt byte, prec int) []byte {
	sign, exp, frac := a.split()

	// sign, 0x, leading digit
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
	buf = append(buf, '1')

	hex := lowerHex
	if fmt == 'X' {
		hex = upperHex
	}

	// Shift digits so leading 1 (if any) is at bit 1<<124.
	frac = frac.Lsh(124 - shift128)

	// Round if requested.
	if prec >= 0 && prec < 31 {
		one := ints.Uint128{0, 1}
		shift := uint(prec * 4)
		extra := frac.Lsh(shift).And(one.Lsh(124).Sub(one))
		frac = frac.Rsh(124 - shift)
		if extra.Or(frac.And(one)).Cmp(one.Lsh(123)) > 0 {
			frac = frac.Add(one)
		}
		frac = frac.Lsh(124 - shift)
		if frac.Cmp(one.Lsh(125)) >= 0 {
			// rounded up, e.g., 0x1.ffff... + 0x0.000...1 = 0x2.000...
			frac = frac.Rsh(1)
			exp++
		}
	}

	// .fraction
	frac = frac.Lsh(4) // remove leading 1
	if prec < 0 && !frac.IsZero() {
		buf = append(buf, '.')
		for !frac.IsZero() {
			buf = append(buf, hex[frac.Rsh(124).Uint64()&0xf])
			frac = frac.Lsh(4)
		}
	} else if prec > 0 {
		buf = append(buf, '.')
		for range prec {
			buf = append(buf, hex[frac.Rsh(124).Uint64()&0xf])
			frac = frac.Lsh(4)
		}
	}

	// p±
	buf = append(buf, fmt-('x'-'p')) // 'p' or 'P'
	if exp >= 0 {
		buf = append(buf, '+')
	} else {
		buf = append(buf, '-')
		exp = -exp
	}
	if exp < 10 {
		buf = append(buf, '0')
	}
	buf = strconv.AppendInt(buf, int64(exp), 10)
	return buf
}

func (a Float128) append(dst []byte, fmt byte, prec int) []byte {
	sign, exp, frac := a.split()
	d := new(decimal)
	d.AssignUint128(frac)
	d.Shift(exp - shift128)
	shortest := prec < 0
	if shortest {
		// TODO: implement roundShortest128
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
