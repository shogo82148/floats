package floats

import (
	"strconv"

	"github.com/shogo82148/ints"
)

const fnParseFloat256 = "ParseFloat256"

func atof256(s string) (f Float256, n int, err error) {
	if val, n, ok := special(s); ok {
		return NewFloat256(val), n, nil
	}

	mantissa, exp, neg, trunc, hex, n, ok := readFloat256(s)
	if !ok {
		return Float256{}, n, syntaxError(fnParseFloat256, s)
	}

	if hex {
		f, err := atof256Hex(s[:n], mantissa, exp, neg, trunc)
		return f, n, err
	}

	var d decimal
	if !d.set(s[:n]) {
		return Float256{}, n, syntaxError(fnParseFloat256, s)
	}
	f, ovf := d.float256()
	if ovf {
		err = rangeError(fnParseFloat256, s)
	}
	return f, n, err
}

func readFloat256(s string) (mantissa ints.Uint256, exp int, neg, trunc, hex bool, i int, ok bool) {
	underscores := false

	// optional sign
	if i >= len(s) {
		return
	}
	switch s[i] {
	case '+':
		i++
	case '-':
		i++
		neg = true
	}

	// digits
	base := ints.Uint256{0, 0, 0, 10} // 10
	maxMantDigits := 77               // 10^77 fits in ints.Uint256
	expChar := byte('e')
	if i+2 < len(s) && s[i] == '0' && lower(s[i+1]) == 'x' {
		base = ints.Uint256{0, 0, 0, 16} // 16
		maxMantDigits = 64               // 16^64 fits in uint256
		i += 2
		expChar = 'p'
		hex = true
	}
	sawdot := false
	sawdigits := false
	nd := 0
	ndMant := 0
	dp := 0

loop:
	for ; i < len(s); i++ {
		switch c := s[i]; true {
		case c == '_':
			underscores = true
			continue

		case c == '.':
			if sawdot {
				break loop
			}
			sawdot = true
			dp = nd
			continue

		case '0' <= c && c <= '9':
			sawdigits = true
			if c == '0' && nd == 0 { // ignore leading zeros
				dp--
				continue
			}
			nd++
			if ndMant < maxMantDigits {
				mantissa = mantissa.Mul(base)
				mantissa = mantissa.Add(ints.Uint256{0, 0, 0, uint64(c - '0')})
				ndMant++
			} else if c != '0' {
				trunc = true
			}
			continue

		case hex && 'a' <= lower(c) && lower(c) <= 'f':
			sawdigits = true
			nd++
			if ndMant < maxMantDigits {
				mantissa = mantissa.Mul(base)
				mantissa = mantissa.Add(ints.Uint256{0, 0, 0, uint64(lower(c) - 'a' + 10)})
				ndMant++
			} else {
				trunc = true
			}
			continue
		}
		break
	}
	if !sawdigits {
		return
	}
	if !sawdot {
		dp = nd
	}

	if hex {
		dp *= 4
		ndMant *= 4
	}

	// optional exponent moves decimal point.
	// if we read a very large, very long number,
	// just be sure to move the decimal point by
	// a lot (say, 1000000).  it doesn't matter if it's
	// not the exact number.
	if i < len(s) && lower(s[i]) == expChar {
		i++
		if i >= len(s) {
			return
		}
		esign := 1
		switch s[i] {
		case '+':
			i++
		case '-':
			i++
			esign = -1
		}
		if i >= len(s) || s[i] < '0' || s[i] > '9' {
			return
		}
		e := 0
		for ; i < len(s) && ('0' <= s[i] && s[i] <= '9' || s[i] == '_'); i++ {
			if s[i] == '_' {
				underscores = true
				continue
			}
			if e < 1000000 {
				e = e*10 + int(s[i]) - '0'
			}
		}
		dp += e * esign
	} else if hex {
		// Must have exponent.
		return
	}

	if !mantissa.IsZero() {
		exp = dp - ndMant
	}

	if underscores && !underscoreOK(s[:i]) {
		return
	}

	ok = true
	return
}

// atof256Hex converts the hex floating-point string s
// to a rounded float256 value and returns it as a float256.
// The string s has already been parsed into a mantissa, exponent, and sign (neg==true for negative).
// If trunc is true, trailing non-zero bits have been omitted from the mantissa.
func atof256Hex(s string, mantissa ints.Uint256, exp int, neg, trunc bool) (Float256, error) {
	one := ints.Uint256{0, 0, 0, 1}
	const maxExp = mask256 - bias256 - 1
	const minExp = -bias256 + 1
	exp += shift256 // mantissa now implicitly divided by 2^shift256.

	// Shift mantissa and exponent to bring representation into float range.
	// Eventually we want a mantissa with a leading 1-bit followed by mantbits other bits.
	// For rounding, we need two more, where the bottom bit represents
	// whether that bit or any later bit was non-zero.
	// (If the mantissa has already lost non-zero bits, trunc is true,
	// and we OR in a 1 below after shifting left appropriately.)
	for !mantissa.IsZero() && mantissa.Rsh(shift256+2).IsZero() {
		mantissa = mantissa.Lsh(1)
		exp--
	}
	if trunc {
		mantissa[3] |= 1
	}
	for !mantissa.Rsh(1 + shift256 + 2).IsZero() {
		mantissa = mantissa.Rsh(1).Or(mantissa.And(one))
		exp++
	}

	// If exponent is too negative,
	// denormalize in hopes of making it representable.
	// (The -2 is for the rounding bits.)
	for mantissa.Cmp(one) > 0 && exp < minExp-2 {
		mantissa = mantissa.Rsh(1).Or(mantissa.And(one))
		exp++
	}

	// Round using two bottom bits.
	round := mantissa[3] & 3
	mantissa = mantissa.Rsh(2)
	round |= mantissa[3] & 1 // round to even (round up if mantissa is odd)
	exp += 2
	if round == 3 {
		mantissa = mantissa.Add(one)
		if mantissa.Cmp(ints.Uint256{1 << (1 + shift256 - 192), 0, 0, 0}) == 0 {
			mantissa = mantissa.Rsh(1)
			exp++
		}
	}

	if mantissa.Rsh(shift256).IsZero() { // Denormal or zero.
		exp = -bias256
	}
	var err error
	if exp > maxExp { // infinity and range error
		mantissa = ints.Uint256{1 << (1 + shift256 - 192), 0, 0, 0}
		exp = maxExp + 1
		err = rangeError(fnParseFloat256, s)
	}

	bits := mantissa.And(fracMask256)
	bits[0] |= uint64((exp+bias256)&mask256) << (shift256 - 192)
	if neg {
		bits = bits.Or(signMask256)
	}

	return Float256(bits), err
}

func (d *decimal) float256() (f Float256, overflow bool) {
	var exp int
	var mant ints.Uint256

	// Zero is always a special case.
	if d.nd == 0 {
		mant = ints.Uint256{0, 0, 0, 0}
		exp = -bias256
		goto out
	}

	// Obvious overflow/underflow.
	if d.dp > 78914 {
		goto overflow
	}
	if d.dp < -78985 {
		// underflow to zero
		mant = ints.Uint256{0, 0, 0, 0}
		exp = -bias256
		goto out
	}

	// Scale by powers of two until in range [0.5, 1.0)
	exp = 0
	for d.dp > 0 {
		var n int
		if d.dp >= len(powtab) {
			n = 27
		} else {
			n = powtab[d.dp]
		}
		d.Shift(-n)
		exp += n
	}
	for d.dp < 0 || d.dp == 0 && d.d[0] < '5' {
		var n int
		if -d.dp >= len(powtab) {
			n = 27
		} else {
			n = powtab[-d.dp]
		}
		d.Shift(n)
		exp -= n
	}

	// Our range is [0.5,1) but floating point range is [1,2).
	exp--

	// Minimum representable exponent is -bias256+1.
	// If the exponent is smaller, move it up and
	// adjust d accordingly.
	if exp < -bias256+1 {
		n := (-bias256 + 1) - exp
		d.Shift(-n)
		exp += n
	}

	// Check for overflow.
	if exp >= mask256-bias256 {
		goto overflow
	}

	// Extract 1+shift256 bits of mantissa.
	d.Shift(1 + shift256)
	mant = d.RoundedUint256()

	// Denormalized?
	if mant[0]&(1<<(shift256-192)) == 0 {
		exp = -bias256
	}
	goto out

overflow:
	// ±Inf
	mant = ints.Uint256{0, 0, 0, 0}
	exp = mask256 - bias256
	overflow = true

out:
	// Assemble bits.
	bits := mant.And(fracMask256)
	bits[0] |= uint64((exp+bias256)&mask256) << (shift256 - 192)
	if d.neg {
		bits = bits.Or(signMask256)
	}
	return Float256(bits), overflow
}

func (d *decimal) RoundedUint256() ints.Uint256 {
	if d.dp > 78 {
		return ints.Uint256{0xffff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}
	}
	var i int
	var n ints.Uint256
	ten := ints.Uint256{0, 0, 0, 10}
	for i = 0; i < d.dp && i < d.nd; i++ {
		n = n.Mul(ten).Add(ints.Uint256{0, 0, 0, uint64(d.d[i] - '0')})
	}
	for ; i < d.dp; i++ {
		n = n.Mul(ten)
	}
	if shouldRoundUp(d, d.dp) {
		n = n.Add(ints.Uint256{0, 0, 0, 1})
	}
	return n
}

// ParseFloat256 parses s as a Float256.
func ParseFloat256(s string) (Float256, error) {
	f, n, err := atof256(s)
	if n != len(s) && (err == nil || err.(*strconv.NumError).Err != strconv.ErrSyntax) {
		return NewFloat256(0), syntaxError(fnParseFloat256, s)
	}
	return f, err
}
