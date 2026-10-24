package floats

import (
	"math/big"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"

	"github.com/shogo82148/ints"
)

// The inputs below are valid but have exponents far beyond the range that
// float256 scales by shifting the decimal. The expected values were produced
// by scaling with the decimal shifts, which is slow for them (seconds).
func TestParseFloat256_LargeExponent(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  Float256
		err   error
	}{
		{"1e78000", Float256{0x7f4254fc964320e1, 0xffea80e131dad665, 0x7f2d2b30e8ce741a, 0x5dbba421a932b59c}, nil},
		{"-1e-78000", Float256{0x80bd88657d25b18e, 0xc43a807de1bfca9a, 0x5e3cd3c664feb388, 0x8031a990def9b9ea}, nil},
		{"1e5000", Float256{0x440e08f1040081b2, 0x8879e2c8fef29b0e, 0xe30ccc3c4b3807a2, 0x28f0e42e70c8148a}, nil},
		{"1e-5000", Float256{0x3bf1d4872f1ab276, 0xd1e7e890bc4d29f3, 0xe53dde83d4a0ee69, 0x21deb58079fed8c2}, nil},
		{"1e1001", Float256{0x40cfc30711e78548, 0xd1706d7073c3a882, 0xbe4875b8a6f57d66, 0xfde6338ca0c0f5a3}, nil},
		{"1e-1001", Float256{0x3f301ae88381bb52, 0xd0e1699d6a600b13, 0x60b1a2a677b79bfe, 0x6c0a0bab21d66b79}, nil},
		{"1e78913", Float256{0x7fffe3dc032f91b9, 0xaab7e56573728587, 0x1e9441f174c69d38, 0xe0d5c9e773eccdd6}, nil},
		{"1e-78950", Float256{0, 0, 0x0000db5281b3f788, 0x4721d6fe46c4448e}, nil},
		{"1.2e-78984", Float256{0, 0, 0, 1}, nil}, // the smallest subnormal number
		{"1e-78984", exact256(0), nil},
		{"3.7e-78985", exact256(0), nil},
		{"2e78913", Float256{0x7ffff00000000000, 0, 0, 0}, strconv.ErrRange},
		{"-2e78913", Float256{0xfffff00000000000, 0, 0, 0}, strconv.ErrRange},
		{"1e78915", Float256{0x7ffff00000000000, 0, 0, 0}, strconv.ErrRange},
	}
	for _, tt := range tests {
		got, err := ParseFloat256(tt.input)
		if numErr, ok := err.(*strconv.NumError); ok {
			err = numErr.Err
		}
		if !eq256(got, tt.want) || err != tt.err {
			t.Errorf("ParseFloat256(%q) = (%#x, %v) want (%#x, %v)", tt.input, [4]uint64(got), err, [4]uint64(tt.want), tt.err)
		}
	}
}

// halfway256 returns the decimal digits and the exponent of the exact value
// (m + 1/2) * 2^(e-shift256), the midpoint of the float256 numbers m*2^(e-shift256)
// and (m+1)*2^(e-shift256).
func halfway256(m *big.Int, e int) (digits *big.Int, exp10 int) {
	digits = new(big.Int).Lsh(m, 1)
	digits.Add(digits, big.NewInt(1)) // (2m+1) * 2^(e-shift256-1)
	shift := e - shift256 - 1
	if shift >= 0 {
		return digits.Lsh(digits, uint(shift)), 0
	}
	// 2^shift = 5^-shift * 10^shift
	return digits.Mul(digits, new(big.Int).Exp(big.NewInt(5), big.NewInt(int64(-shift)), nil)), shift
}

// float256FromMantissa returns m * 2^(e-shift256), where m includes the hidden bit.
// If m is less than 2^shift256, the result is subnormal for e = 1-bias256.
func float256FromMantissa(m *big.Int, e int) Float256 {
	bits := big.NewInt(int64(e + bias256))
	bits.Lsh(bits, shift256)
	bits.Add(bits, m)
	bits.Sub(bits, new(big.Int).Lsh(big.NewInt(1), shift256))
	var buf [32]byte
	bits.FillBytes(buf[:])
	var f Float256
	for i := range f {
		for _, b := range buf[i*8 : i*8+8] {
			f[i] = f[i]<<8 | uint64(b)
		}
	}
	return f
}

func TestParseFloat256_LargeExponentRounding(t *testing.T) {
	t.Parallel()
	hidden := new(big.Int).Lsh(big.NewInt(1), shift256)
	maxMant := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), shift256+1), big.NewInt(1))
	tests := []struct {
		e int
		m *big.Int
	}{
		{-30000, new(big.Int).Add(hidden, big.NewInt(2))}, // even
		{-30000, new(big.Int).Add(hidden, big.NewInt(3))}, // odd
		{-30000, maxMant}, // carries into the exponent
		{-3500, new(big.Int).Add(hidden, big.NewInt(4))},
		{-3500, new(big.Int).Add(hidden, big.NewInt(5))},
		{3500, new(big.Int).Add(hidden, big.NewInt(6))},
		{3500, new(big.Int).Add(hidden, big.NewInt(7))},
		{100000, new(big.Int).Add(hidden, big.NewInt(8))},
		{100000, new(big.Int).Add(hidden, big.NewInt(9))},
		{262142, maxMant},
		{262143, new(big.Int).Sub(maxMant, big.NewInt(1))}, // the largest even mantissa
		{262143, maxMant},                                      // rounds up to infinity
		{1 - bias256, big.NewInt(0)},                           // subnormal numbers
		{1 - bias256, new(big.Int).Sub(hidden, big.NewInt(1))}, // carries to the smallest normal number
	}
	for _, tt := range tests {
		digits, exp10 := halfway256(tt.m, tt.e)
		tie := digits.String()
		below := new(big.Int).Sub(digits, big.NewInt(1)).String() + "9" // the midpoint minus 10^(exp10-1)
		above := tie + "1"                                              // the midpoint plus 10^(exp10-1)
		// The midpoint followed by more digits than a decimal can hold.
		// The decimal drops them, but they still make the value greater.
		truncated := "." + tie + strings.Repeat("0", decimalDigits256) + "1"
		lo := float256FromMantissa(tt.m, tt.e)
		hi := float256FromMantissa(new(big.Int).Add(tt.m, big.NewInt(1)), tt.e)
		even := lo
		if tt.m.Bit(0) != 0 {
			even = hi
		}

		for _, sign := range []string{"", "-"} {
			for _, c := range []struct {
				digits string
				exp10  int
				want   Float256
			}{
				{tie, exp10, even},
				{below, exp10 - 1, lo},
				{above, exp10 - 1, hi},
				{truncated, exp10 + len(tie), hi},
			} {
				if c.digits == truncated && (sign != "" || tt.m.Bit(0) != 0) {
					continue // an even mantissa is enough to see the difference from the midpoint
				}
				input := sign + c.digits + "e" + strconv.Itoa(c.exp10)
				want := c.want
				if sign != "" {
					want[0] |= 1 << 63
				}
				wantErr := error(nil)
				if want[0]&(mask256<<(shift256-192)) == mask256<<(shift256-192) {
					wantErr = strconv.ErrRange
				}
				got, err := ParseFloat256(input)
				if numErr, ok := err.(*strconv.NumError); ok {
					err = numErr.Err
				}
				if got != want || err != wantErr {
					t.Errorf("ParseFloat256(%.40s...e%d) with e=%d, m=%v = (%#x, %v) want (%#x, %v)",
						c.digits, c.exp10, tt.e, tt.m, [4]uint64(got), err, [4]uint64(want), wantErr)
				}
			}
		}
	}
}

// float256 scales the decimal by shifting it if the decimal point is near enough,
// float256Big does the same by the integer arithmetic. They must agree.
func TestDecimalFloat256Big(t *testing.T) {
	t.Parallel()
	check := func(input string) {
		t.Helper()
		d1 := decimal{d: make([]byte, decimalDigits256)}
		d2 := decimal{d: make([]byte, decimalDigits256)}
		if !d1.set(input) || !d2.set(input) {
			t.Fatalf("failed to set %q", input)
		}
		if d1.nd == 0 {
			return
		}
		if d1.dp > decimalLimit256 || d1.dp < -decimalLimit256 {
			t.Fatalf("%q is too large to compare with the decimal shifts", input)
		}

		mant, exp, ovf := d1.float256Big()
		if ovf {
			mant, exp = ints.Uint256{}, mask256-bias256
		}
		bits := mant.And(fracMask256)
		bits[0] |= uint64((exp+bias256)&mask256) << (shift256 - 192)
		if d1.neg {
			bits = bits.Or(signMask256)
		}

		want, wantOvf := d2.float256()
		if Float256(bits) != want || ovf != wantOvf {
			t.Errorf("%q: got (%#x, %v), want (%#x, %v)", input, [4]uint64(bits), ovf, [4]uint64(want), wantOvf)
		}
	}

	r := rand.New(rand.NewPCG(1, 2))
	for range 500 {
		var sb strings.Builder
		for range 1 + r.IntN(60) {
			sb.WriteByte(byte('0' + r.IntN(10)))
		}
		check(sb.String() + "e" + strconv.Itoa(r.IntN(2*(decimalLimit256-100))-(decimalLimit256-100)))
	}

	hidden := new(big.Int).Lsh(big.NewInt(1), shift256)
	for range 100 {
		m := new(big.Int).SetUint64(r.Uint64())
		m.Lsh(m, 64).Add(m, new(big.Int).SetUint64(r.Uint64()))
		m.Lsh(m, 64).Add(m, new(big.Int).SetUint64(r.Uint64()))
		m.Lsh(m, 64).Add(m, new(big.Int).SetUint64(r.Uint64()))
		m.Mod(m, hidden).Add(m, hidden)
		e := r.IntN(2*300) - 300
		digits, exp10 := halfway256(m, e)
		check(digits.String() + "e" + strconv.Itoa(exp10))
		check(digits.String() + "1e" + strconv.Itoa(exp10-1))
		check(new(big.Int).Sub(digits, big.NewInt(1)).String() + "9e" + strconv.Itoa(exp10-1))
	}
}

func BenchmarkParseFloat256_LargeExponent(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat256("1e78000")
		if err != nil {
			b.Fatal(err)
		}
	}
}
