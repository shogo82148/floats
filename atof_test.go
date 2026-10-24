package floats

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

// checkLongIntegerPart checks the strings whose integer part is longer than
// the capacity of the decimal buffer (size digits).
// The digits that do not fit in the buffer still have to move the decimal point,
// so that they are scaled back by the exponent.
//
// Each long string is mathematically equal to m*10^x,
// whose short form m + "e" + x never overflows the buffer.
// The results of the short form do not depend on the dropped digits.
func checkLongIntegerPart[T comparable](t *testing.T, size int, parse func(string) (T, error), ovf, udf int) {
	t.Helper()
	cases := []struct {
		m        string
		x        int
		rangeErr bool
	}{
		{"1", 0, false},
		{"12345", -3, false},
		{"7", -3, false},
		{"1", -1, false},
		{"1", ovf, true},   // too large
		{"1", -udf, false}, // too small: underflows to zero
	}
	// size-1 zeros fit in the buffer with the leading 1.
	for _, zeros := range []int{size - 1, size, size + 7} {
		z := strings.Repeat("0", zeros)
		for _, tt := range cases {
			exp := "e" + strconv.Itoa(tt.x-zeros)
			for _, sign := range []string{"", "-"} {
				short := sign + tt.m + "e" + strconv.Itoa(tt.x)
				want, wantErr := parse(short)
				if errors.Is(wantErr, strconv.ErrRange) != tt.rangeErr {
					t.Fatalf("Parse(%q) error = %v, want range error: %v", short, wantErr, tt.rangeErr)
				}
				if wantErr != nil && !tt.rangeErr {
					t.Fatalf("Parse(%q) error = %v", short, wantErr)
				}

				for _, long := range []string{
					sign + tt.m + z + exp,
					sign + tt.m + z + ".0" + exp,
				} {
					got, err := parse(long)
					if got != want || errors.Is(err, strconv.ErrRange) != tt.rangeErr || (err != nil && !tt.rangeErr) {
						t.Errorf("Parse(%d digits, exponent %s) = (%v, %v), want the same as Parse(%q) = (%v, %v)",
							len(long), exp, got, err, short, want, wantErr)
					}
				}
			}
		}
	}
}

func TestParse_LongIntegerPart(t *testing.T) {
	t.Parallel()
	t.Run("Float8E4M3", func(t *testing.T) {
		t.Parallel()
		checkLongIntegerPart(t, decimalDigits8, ParseFloat8E4M3, 3, 4)
	})
	t.Run("Float8E5M2", func(t *testing.T) {
		t.Parallel()
		checkLongIntegerPart(t, decimalDigits8, ParseFloat8E5M2, 5, 6)
	})
	t.Run("Float16", func(t *testing.T) {
		t.Parallel()
		checkLongIntegerPart(t, decimalDigits16, ParseFloat16, 5, 9)
	})
	t.Run("BFloat16", func(t *testing.T) {
		t.Parallel()
		checkLongIntegerPart(t, decimalDigitsBF16, ParseBFloat16, 39, 42)
	})
	t.Run("Float128", func(t *testing.T) {
		t.Parallel()
		checkLongIntegerPart(t, decimalDigits128, ParseFloat128, 4933, 4967)
	})
	t.Run("Float256", func(t *testing.T) {
		t.Parallel()
		checkLongIntegerPart(t, decimalDigits256, ParseFloat256, 78914, 78990)
	})
}
