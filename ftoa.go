package floats

import "fmt"

type floatN interface {
	IsNaN() bool
	Signbit() bool
	Append(dst []byte, fmt byte, prec int) []byte
}

// format implements [fmt.Formatter] in the same way as the floating-point numbers of the fmt package.
func format(x floatN, s fmt.State, verb rune) {
	// the default precision
	var prec int
	var f byte
	switch verb {
	case 'v':
		f, prec = 'g', -1
	case 'g', 'G', 'x', 'X', 'b':
		f, prec = byte(verb), -1
	case 'e', 'E':
		f, prec = byte(verb), 6
	case 'f', 'F':
		f, prec = 'f', 6
	default:
		_, _ = fmt.Fprintf(s, "%%!%c(%T=%s)", verb, x, x.Append(nil, 'g', -1))
		return
	}
	if p, ok := s.Precision(); ok {
		prec = p
	}

	// For %+v, the "+" flag of the Formatter means the "plusV" flag of fmt, that is not for the sign.
	plus := s.Flag('+') && verb != 'v'
	space := s.Flag(' ')
	// For %#v, the "#" flag of the Formatter means the "sharpV" flag of fmt, that is for the Go syntax.
	sharp := s.Flag('#') && verb != 'v'
	minus := s.Flag('-')
	zero := s.Flag('0') && !minus
	width, hasWidth := s.Width()

	// num has the sign at the first byte.
	num := x.Append(make([]byte, 1, 16), f, prec)
	if num[1] == '-' || num[1] == '+' {
		num = num[1:]
	} else {
		num[0] = '+'
	}
	// space means to add a leading space instead of a "+" sign unless plus is used.
	if space && num[0] == '+' && !plus {
		num[0] = ' '
	}

	// Infinities and NaN don't look like a number so shouldn't be padded with zeros.
	if num[1] == 'I' || num[1] == 'N' {
		if num[1] == 'N' && !space && !plus {
			num = num[1:]
		}
		writePadding(s, num, width, hasWidth, false, minus)
		return
	}

	if sharp && f != 'b' {
		num = formatSharp(num, f, prec)
	}

	// We want a sign if asked for and if the sign is not positive.
	if plus || num[0] != '+' {
		// If we're zero padding to the left we want the sign before the leading zeros.
		if zero && hasWidth && width > len(num) {
			_, _ = s.Write(num[:1])
			writePadding(s, num[1:], width-1, true, true, false)
			return
		}
		writePadding(s, num, width, hasWidth, false, minus)
		return
	}

	// No sign to show and the number is positive; just print the unsigned number.
	writePadding(s, num[1:], width, hasWidth, zero, minus)
}

// formatSharp changes num, which has the sign at the first byte, for the "#" flag:
// it always has a decimal point, and %g and %x keep the trailing zeros.
func formatSharp(num []byte, verb byte, prec int) []byte {
	digits := 0
	switch verb {
	case 'g', 'G', 'x':
		digits = prec
		// If no precision is set explicitly use a precision of 6.
		if digits == -1 {
			digits = 6
		}
	}

	// Buffer pre-allocated with enough room for exponent notations of the form "e+123" or "p-1023".
	var tailBuf [6]byte
	tail := tailBuf[:0]

	hasDecimalPoint := false
	sawNonzeroDigit := false
	// Starting from i = 1 to skip sign at num[0].
loop:
	for i := 1; i < len(num); i++ {
		switch num[i] {
		case '.':
			hasDecimalPoint = true
		case 'p', 'P':
			tail = append(tail, num[i:]...)
			num = num[:i]
			break loop
		case 'e', 'E':
			if verb != 'x' && verb != 'X' {
				tail = append(tail, num[i:]...)
				num = num[:i]
				break loop
			}
			fallthrough
		default:
			// Count significant digits after the first non-zero digit.
			if num[i] != '0' {
				sawNonzeroDigit = true
			}
			if sawNonzeroDigit {
				digits--
			}
		}
	}
	if !hasDecimalPoint {
		// Leading digit 0 should contribute once to digits.
		if len(num) == 2 && num[1] == '0' {
			digits--
		}
		num = append(num, '.')
	}
	for ; digits > 0; digits-- {
		num = append(num, '0')
	}
	return append(num, tail...)
}

// writePadding writes data to s with padding to the width.
// The padding is on the left side with spaces or zeros, or on the right side with spaces if minus is true.
func writePadding(s fmt.State, data []byte, width int, hasWidth, zero, minus bool) {
	pad := 0
	if hasWidth && width > len(data) {
		pad = width - len(data)
	}
	var buf [1]byte
	if minus {
		_, _ = s.Write(data)
	}
	buf[0] = ' '
	if zero {
		buf[0] = '0'
	}
	for range pad {
		_, _ = s.Write(buf[:])
	}
	if !minus {
		_, _ = s.Write(data)
	}
}

func formatDigits(dst []byte, neg bool, d *decimal, shortest bool, prec int, fmt byte) []byte {
	switch fmt {
	case 'e', 'E':
		return fmtE(dst, neg, d, prec, fmt)
	case 'f':
		return fmtF(dst, neg, d, prec)
	case 'g', 'G':
		// trailing fractional zeros in 'e' form will be trimmed.
		eprec := prec
		if eprec > d.nd && d.nd >= d.dp {
			eprec = d.nd
		}
		// %e is used if the exponent from the conversion
		// is less than -4 or greater than or equal to the precision.
		// if precision was the shortest possible, use precision 6 for this decision.
		if shortest {
			eprec = 6
		}
		exp := d.dp - 1
		if exp < -4 || exp >= eprec {
			if prec > d.nd {
				prec = d.nd
			}
			return fmtE(dst, neg, d, prec-1, fmt+'e'-'g')
		}
		if prec > d.dp {
			prec = d.nd
		}
		return fmtF(dst, neg, d, max(prec-d.dp, 0))
	}

	// unknown format
	return append(dst, '%', fmt)
}

// %e: -d.ddddde±dd
func fmtE(dst []byte, neg bool, d *decimal, prec int, fmt byte) []byte {
	// sign
	if neg {
		dst = append(dst, '-')
	}

	// first digit
	ch := byte('0')
	if d.nd != 0 {
		ch = d.d[0]
	}
	dst = append(dst, ch)

	// .moredigits
	if prec > 0 {
		dst = append(dst, '.')
		i := 1
		m := min(d.nd, prec+1)
		if i < m {
			dst = append(dst, d.d[i:m]...)
			i = m
		}
		for ; i <= prec; i++ {
			dst = append(dst, '0')
		}
	}

	// e±
	dst = append(dst, fmt)
	exp := d.dp - 1
	if d.nd == 0 { // special case: 0 has exponent 0
		exp = 0
	}
	if exp < 0 {
		ch = '-'
		exp = -exp
	} else {
		ch = '+'
	}
	dst = append(dst, ch)

	// dd or ddd
	switch {
	case exp < 10:
		dst = append(dst, '0', byte(exp)+'0')
	case exp < 100:
		dst = append(dst, byte(exp/10)+'0', byte(exp%10)+'0')
	case exp < 1000:
		dst = append(dst, byte(exp/100)+'0', byte(exp/10%10)+'0', byte(exp%10)+'0')
	case exp < 10000:
		dst = append(dst, byte(exp/1000)+'0', byte(exp/100%10)+'0', byte(exp/10%10)+'0', byte(exp%10)+'0')
	default:
		dst = append(dst, byte(exp/10000)+'0', byte(exp/1000%10)+'0', byte(exp/100%10)+'0', byte(exp/10%10)+'0', byte(exp%10)+'0')
	}

	return dst
}

// %f: -ddddddd.ddddd
func fmtF(dst []byte, neg bool, d *decimal, prec int) []byte {
	// sign
	if neg {
		dst = append(dst, '-')
	}

	// integer, padded with zeros as needed.
	if d.dp > 0 {
		m := min(d.nd, d.dp)
		dst = append(dst, d.d[:m]...)
		for ; m < d.dp; m++ {
			dst = append(dst, '0')
		}
	} else {
		dst = append(dst, '0')
	}

	// fraction
	if prec > 0 {
		dst = append(dst, '.')
		for i := range prec {
			ch := byte('0')
			if j := d.dp + i; 0 <= j && j < d.nd {
				ch = d.d[j]
			}
			dst = append(dst, ch)
		}
	}

	return dst
}
