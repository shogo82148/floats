package floats

func formatDigits(dst []byte, neg bool, d *decimal, shortest bool, prec int, fmt byte) []byte {
	switch fmt {
	case 'e', 'E':
	case 'f':
		return fmtF(dst, neg, d, prec)
	case 'g', 'G':
	}

	// unknown format
	return append(dst, '%', fmt)
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
