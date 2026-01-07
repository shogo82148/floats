package floats

import (
	"fmt"
	"io"
)

type floatN interface {
	IsNaN() bool
	Signbit() bool
	Append(dst []byte, fmt byte, prec int) []byte
}

func format(x floatN, s fmt.State, verb rune) {
	if x.IsNaN() {
		_, _ = io.WriteString(s, "NaN")
		return
	}

	var prefix []byte
	var data []byte

	// sign
	if !x.Signbit() {
		if s.Flag('+') {
			prefix = append(prefix, '+')
		} else if s.Flag(' ') {
			prefix = append(prefix, ' ')
		}
	}

	switch verb {
	case 'b':
		data = x.Append(data, 'b', -1)
	case 'f', 'e', 'E', 'g', 'G', 'x', 'X':
		if prec, ok := s.Precision(); ok {
			data = x.Append(data, byte(verb), prec)
		} else {
			data = x.Append(data, byte(verb), -1)
		}
	case 'v':
		data = x.Append(data, 'g', -1)
	}

	if w, ok := s.Width(); ok {
		var buf [1]byte
		if s.Flag('-') {
			_, _ = s.Write(prefix)
			_, _ = s.Write(data)
			buf[0] = ' '
			for i := len(data); i < w; i++ {
				_, _ = s.Write(buf[:1])
			}
		} else {
			buf[0] = ' '
			for i := len(data); i < w; i++ {
				_, _ = s.Write(buf[:1])
			}
			_, _ = s.Write(prefix)
			_, _ = s.Write(data)
		}
		return
	}

	if len(prefix) > 0 {
		_, _ = s.Write(prefix)
	}
	_, _ = s.Write(data)
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
