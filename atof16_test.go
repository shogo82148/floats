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
		{"1.5", exact16(1.5), nil},
		{"-2.75", exact16(-2.75), nil},

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
			t.Errorf("ParseFloat16(%q) = %v, %v want %v, %v", tt.input, got, err, tt.want, tt.err)
		}
	}
}
