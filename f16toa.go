// convert float16 to string

package floats

import (
	"fmt"
)

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
	case 'f', 'e', 'E', 'g', 'G':
		return a.append(buf, fmt, prec)
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

	hex := lowerHex
	if fmt == 'X' {
		hex = upperHex
	}

	switch {
	case prec < 0:
		if frac&0x3ff == 0 {
			buf = append(buf, '1')
		} else if frac&0x3f == 0 {
			buf = append(buf, '1', '.')
			buf = append(buf, hex[(frac>>6)&0xF])
		} else if frac&0x3 == 0 {
			buf = append(buf, '1', '.')
			buf = append(buf, hex[(frac>>6)&0xF])
			buf = append(buf, hex[(frac>>2)&0xF])
		} else {
			buf = append(buf, '1', '.')
			buf = append(buf, hex[(frac>>6)&0xF])
			buf = append(buf, hex[(frac>>2)&0xF])
			buf = append(buf, hex[(frac<<2)&0xF])
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
		buf = append(buf, hex[(frac>>6)&0xF])
	case prec == 2:
		// round to nearest even
		frac += 1 + (frac>>2)&1
		if frac >= 1<<(shift16+1) {
			exp++
			frac >>= 1
		}

		buf = append(buf, '1', '.')
		buf = append(buf, hex[(frac>>6)&0xF])
		buf = append(buf, hex[(frac>>2)&0xF])
	default:
		buf = append(buf, '1', '.')
		buf = append(buf, hex[(frac>>6)&0xF])
		buf = append(buf, hex[(frac>>2)&0xF])
		buf = append(buf, hex[(frac<<2)&0xF])
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

func (a Float16) append(buf []byte, fmt byte, prec int) []byte {
	sign := uint16(a & signMask16)
	exp := int((a>>shift16)&mask16) - bias16
	frac := uint16(a & fracMask16)
	if exp == -bias16 {
		exp++
	} else {
		frac |= 1 << shift16
	}

	d := new(decimal)
	d.AssignUint64(uint64(frac))
	d.Shift(exp - shift16)
	shortest := prec < 0
	if shortest {
		roundShortest16(d, frac, exp)
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
	return formatDigits(buf, sign != 0, d, shortest, prec, fmt)
}

func roundShortest16(d *decimal, frac uint16, exp int) {
	// If mantissa is zero, the number is zero; stop now.
	if frac == 0 {
		d.nd = 0
		return
	}

	minexp := -bias16 + 1 // minimum possible exponent

	// d = frac << (exp - shift16)
	// Next highest floating point number is frac+1 << exp-shift16.
	// Our upper bound is halfway between, frac*2+1 << exp-shift16-1.
	upper := new(decimal)
	upper.AssignUint64(uint64(frac*2 + 1))
	upper.Shift(exp - shift16 - 1)

	// d = frac << (exp - shift16)
	// Next lowest floating point number is frac-1 << exp-shift16,
	// unless frac-1 drops the significant bit and exp is not the minimum exp,
	// in which case the next lowest is frac*2-1 << exp-shift16-1.
	// Either way, call it fraclo << explo-shift16.
	// Our lower bound is halfway between, fraclo*2+1 << explo-shift16-1.
	var fraclo uint16
	var explo int
	if frac > 1<<shift16 || exp == minexp {
		fraclo = frac - 1
		explo = exp
	} else {
		fraclo = frac*2 - 1
		explo = exp - 1
	}
	lower := new(decimal)
	lower.AssignUint64(uint64(fraclo*2 + 1))
	lower.Shift(explo - shift16 - 1)

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
