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
func (a Float256) Append(dst []byte, fmt byte, prec int) []byte {
	// special numbers
	switch {
	case a.IsNaN():
		return append(dst, "NaN"...)
	case a.IsInf(1):
		return append(dst, "+Inf"...)
	case a.IsInf(-1):
		return append(dst, "-Inf"...)
	}

	switch fmt {
	case 'b':
		return a.appendBin(dst)
	case 'x', 'X':
		return a.appendHex(dst, fmt, prec)
	case 'f', 'e', 'E', 'g', 'G':
		return a.append(dst, fmt, prec)
	}

	// unknown format
	return append(dst, '%', fmt)
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
	sign, exp, frac := a.normalize()

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
	b := ints.Uint256(a)
	exp := int((b[0]>>(shift256-192))&mask256) - bias256
	sign := b[0] & signMask256[0]
	frac := b.And(fracMask256)
	if exp > -bias256 {
		frac[0] = frac[0] | (1 << (shift256 - 192))
	}
	d := new(decimal)
	d.AssignUint256(frac)
	d.Shift(exp - shift256)
	shortest := prec < 0
	if shortest {
		roundShortest256(d, frac, exp)
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

func roundShortest256(d *decimal, frac ints.Uint256, exp int) {
	// If mantissa is zero, the number is zero; stop now.
	if frac.IsZero() {
		d.nd = 0
		return
	}

	one := ints.Uint256{0, 0, 0, 1}
	minexp := -bias256 + 1 // minimum possible exponent

	// d = frac << (exp - shift16)
	// Next highest floating point number is frac+1 << exp-shift16.
	// Our upper bound is halfway between, frac*2+1 << exp-shift16-1.
	upper := new(decimal)
	upper.AssignUint256(frac.Lsh(1).Add(one))
	upper.Shift(exp - shift256 - 1)
	// d = frac << (exp - shift16)
	// Next lowest floating point number is frac-1 << exp-shift16,
	// unless frac-1 drops the significant bit and exp is not the minimum exp,
	// in which case the next lowest is frac*2-1 << exp-shift16-1.
	// Either way, call it fraclo << explo-shift16.
	// Our lower bound is halfway between, fraclo*2+1 << explo-shift16-1.
	var fraclo ints.Uint256
	var explo int
	if frac.Cmp(one.Lsh(shift256)) > 0 || exp == minexp {
		fraclo = frac.Sub(one)
		explo = exp
	} else {
		fraclo = frac.Lsh(1).Sub(one)
		explo = exp - 1
	}
	lower := new(decimal)
	lower.AssignUint256(fraclo.Lsh(1).Add(one))
	lower.Shift(explo - shift256 - 1)

	// The upper and lower bounds are possible outputs only if
	// the original mantissa is even, so that IEEE round-to-even
	// would round to the original mantissa and not the neighbors.
	inclusive := frac.And(one).IsZero()

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
