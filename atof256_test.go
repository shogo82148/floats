package floats

import (
	"math"
	"strconv"
	"testing"
)

func TestParseFloat256(t *testing.T) {
	tests := []struct {
		input string
		want  Float256
		err   error
	}{
		{"0", exact256(0), nil},
		{"-0", exact256(math.Copysign(0, -1)), nil},
		{"1", exact256(1.0), nil},
		{"1.5", exact256(1.5), nil},
		{"-2.75", exact256(-2.75), nil},
		{"10", exact256(10), nil},
		{"100", exact256(100), nil},
		{"1000", exact256(1000), nil},
		{"10000", exact256(10000), nil},
		{
			"1.61132571748576047361957211845200501064402387454966951747637125049607183e+78913",
			Float256{
				0x7fff_efff_ffff_ffff, 0xffff_ffff_ffff_ffff,
				0xffff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff,
			},
			nil,
		},

		// NaNs
		{"nan", exact256(math.NaN()), nil},
		{"NaN", exact256(math.NaN()), nil},
		{"NAN", exact256(math.NaN()), nil},

		// Infs
		{"Inf", exact256(math.Inf(1)), nil},
		{"-Inf", exact256(math.Inf(-1)), nil},
		{"+INF", exact256(math.Inf(1)), nil},
		{"-Infinity", exact256(math.Inf(-1)), nil},
		{"+INFINITY", exact256(math.Inf(1)), nil},
		{"Infinity", exact256(math.Inf(1)), nil},

		// try to overflow exponent
		{"1e-4294967296", exact256(0), nil},
		{"1e+4294967296", exact256(math.Inf(1)), strconv.ErrRange},
		{"1e-18446744073709551616", exact256(0), nil},
		{"1e+18446744073709551616", exact256(math.Inf(1)), strconv.ErrRange},
		{"0x1p-4294967296", exact256(0), nil},
		{"0x1p+4294967296", exact256(math.Inf(1)), strconv.ErrRange},
		{"0x1p-18446744073709551616", exact256(0), nil},
		{"0x1p+18446744073709551616", exact256(math.Inf(1)), strconv.ErrRange},

		// Parse errors
		{"1e", exact256(0), strconv.ErrSyntax},
		{"1e-", exact256(0), strconv.ErrSyntax},
		{".e-1", exact256(0), strconv.ErrSyntax},
		{"1\x00.2", exact256(0), strconv.ErrSyntax},
		{"0x", exact256(0), strconv.ErrSyntax},
		{"0x.", exact256(0), strconv.ErrSyntax},
		{"0x1", exact256(0), strconv.ErrSyntax},
		{"0x.1", exact256(0), strconv.ErrSyntax},
		{"0x1p", exact256(0), strconv.ErrSyntax},
		{"0x.1p", exact256(0), strconv.ErrSyntax},
		{"0x1p+", exact256(0), strconv.ErrSyntax},
		{"0x.1p+", exact256(0), strconv.ErrSyntax},
		{"0x1p-", exact256(0), strconv.ErrSyntax},
		{"0x.1p-", exact256(0), strconv.ErrSyntax},
		{"0x1p+2", exact256(4), nil},
		{"0x.1p+2", exact256(0.25), nil},
		{"0x1p-2", exact256(0.25), nil},
		{"0x.1p-2", exact256(0.015625), nil},
	}

	for _, tt := range tests {
		got, err := ParseFloat256(tt.input)
		if err != nil {
			numErr, ok := err.(*strconv.NumError)
			if !ok {
				t.Errorf("ParseFloat256(%q) unexpected error type: %v", tt.input, err)
				continue
			}
			if numErr.Func != "ParseFloat256" {
				t.Errorf("ParseFloat256(%q) unexpected Func in NumError: got %q, want %q", tt.input, numErr.Func, "ParseFloat256")
			}
			if numErr.Num != tt.input {
				t.Errorf("ParseFloat256(%q) unexpected Num in NumError: got %q, want %q", tt.input, numErr.Num, tt.input)
			}
			err = numErr.Err
		}
		if !eq256(got, tt.want) || err != tt.err {
			t.Errorf("ParseFloat256(%q) = (%v, %v) want (%v, %v)", tt.input, got, err, tt.want, tt.err)
		}
	}
}

func BenchmarkParseFloat256_Decimal(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat256("33909")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFloat256_Float(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat256("339.778")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFloat256_FloatExp(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat256("-5.09e-3")
		if err != nil {
			b.Fatal(err)
		}
	}
}
