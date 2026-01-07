package floats

import (
	"math"
	"strconv"
	"testing"
)

func TestParseFloat16(t *testing.T) {
	tests := []struct {
		input string
		want  Float16
		err   error
	}{
		{"0", exact16(0), nil},
		{"-0", exact16(math.Copysign(0, -1)), nil},
		{"1", exact16(1.0), nil},
		{"1.5", exact16(1.5), nil},
		{"-2.75", exact16(-2.75), nil},
		{"10", exact16(10), nil},
		{"100", exact16(100), nil},
		{"1000", exact16(1000), nil},
		{"10000", exact16(10000), nil},
		{"65504", exact16(65504), nil}, // max finite value

		// denormalized
		{"6e-8", exact16(0x1p-24), nil}, // min positive denormalized
		{"1e-7", exact16(0x2p-24), nil},
		{"2e-7", exact16(0x3p-24), nil},
		{"2.4e-7", exact16(0x4p-24), nil},
		{"3e-7", exact16(0x5p-24), nil},
		{"3.6e-7", exact16(0x6p-24), nil},
		{"4e-7", exact16(0x7p-24), nil},
		{"5e-7", exact16(0x8p-24), nil},
		{"5.4e-7", exact16(0x9p-24), nil},
		{"6e-7", exact16(0xap-24), nil},

		// Hexadecimal floating-point.
		{"0x1p+0", exact16(1.0), nil},
		{"0x1p1", exact16(2.0), nil},
		{"0x1.8p+1", exact16(3.0), nil},
		{"0x1p-1", exact16(0.5), nil},
		{"-0x2p3", exact16(-16), nil},
		{"0x0.fp4", exact16(15), nil},
		{"0x1e2", exact16(0), strconv.ErrSyntax}, // missing 'p' exponent
		{"1p2", exact16(0), strconv.ErrSyntax},   // missing '0x' prefix

		// NaNs
		{"nan", exact16(math.NaN()), nil},
		{"NaN", exact16(math.NaN()), nil},
		{"NAN", exact16(math.NaN()), nil},

		// Infs
		{"Inf", exact16(math.Inf(1)), nil},
		{"-Inf", exact16(math.Inf(-1)), nil},
		{"+INF", exact16(math.Inf(1)), nil},
		{"-Infinity", exact16(math.Inf(-1)), nil},
		{"+INFINITY", exact16(math.Inf(1)), nil},
		{"Infinity", exact16(math.Inf(1)), nil},

		// try to overflow exponent
		{"1e-4294967296", exact16(0), nil},
		{"1e+4294967296", exact16(math.Inf(1)), strconv.ErrRange},
		{"1e-18446744073709551616", exact16(0), nil},
		{"1e+18446744073709551616", exact16(math.Inf(1)), strconv.ErrRange},
		{"0x1p-4294967296", exact16(0), nil},
		{"0x1p+4294967296", exact16(math.Inf(1)), strconv.ErrRange},
		{"0x1p-18446744073709551616", exact16(0), nil},
		{"0x1p+18446744073709551616", exact16(math.Inf(1)), strconv.ErrRange},

		// Parse errors
		{"1e", exact16(0), strconv.ErrSyntax},
		{"1e-", exact16(0), strconv.ErrSyntax},
		{".e-1", exact16(0), strconv.ErrSyntax},
		{"1\x00.2", exact16(0), strconv.ErrSyntax},
		{"0x", exact16(0), strconv.ErrSyntax},
		{"0x.", exact16(0), strconv.ErrSyntax},
		{"0x1", exact16(0), strconv.ErrSyntax},
		{"0x.1", exact16(0), strconv.ErrSyntax},
		{"0x1p", exact16(0), strconv.ErrSyntax},
		{"0x.1p", exact16(0), strconv.ErrSyntax},
		{"0x1p+", exact16(0), strconv.ErrSyntax},
		{"0x.1p+", exact16(0), strconv.ErrSyntax},
		{"0x1p-", exact16(0), strconv.ErrSyntax},
		{"0x.1p-", exact16(0), strconv.ErrSyntax},
		{"0x1p+2", exact16(4), nil},
		{"0x.1p+2", exact16(0.25), nil},
		{"0x1p-2", exact16(0.25), nil},
		{"0x.1p-2", exact16(0.015625), nil},
	}

	for _, tt := range tests {
		got, err := ParseFloat16(tt.input)
		if err != nil {
			numErr, ok := err.(*strconv.NumError)
			if !ok {
				t.Errorf("ParseFloat16(%q) unexpected error type: %v", tt.input, err)
				continue
			}
			if numErr.Func != "ParseFloat16" {
				t.Errorf("ParseFloat16(%q) unexpected Func in NumError: got %q, want %q", tt.input, numErr.Func, "ParseFloat16")
			}
			if numErr.Num != tt.input {
				t.Errorf("ParseFloat16(%q) unexpected Num in NumError: got %q, want %q", tt.input, numErr.Num, tt.input)
			}
			err = numErr.Err
		}
		if !eq16(got, tt.want) || err != tt.err {
			t.Errorf("ParseFloat16(%q) = (%v, %v) want (%v, %v)", tt.input, got, err, tt.want, tt.err)
		}
	}
}
