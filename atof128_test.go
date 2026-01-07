package floats

import (
	"math"
	"strconv"
	"testing"
)

func TestParseFloat128(t *testing.T) {
	tests := []struct {
		input string
		want  Float128
		err   error
	}{
		{"0", exact128(0), nil},
		{"-0", exact128(math.Copysign(0, -1)), nil},
		{"1", exact128(1.0), nil},
		{"1.5", exact128(1.5), nil},
		{"-2.75", exact128(-2.75), nil},
		{"10", exact128(10), nil},
		{"100", exact128(100), nil},
		{"1000", exact128(1000), nil},
		{"10000", exact128(10000), nil},
		{"1.189731495357231765085759326628007e+4932", Float128{0x7ffe_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}, nil}, // max finite value
		{"6e-4966", Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, nil},                                   // min positive denormalized

		// Hexadecimal floating-point.
		{"0x1p+0", exact128(1.0), nil},
		{"0x1p1", exact128(2.0), nil},
		{"0x1.8p+1", exact128(3.0), nil},
		{"0x1p-1", exact128(0.5), nil},
		{"-0x2p3", exact128(-16), nil},
		{"0x0.fp4", exact128(15), nil},
		{"0x1e2", exact128(0), strconv.ErrSyntax}, // missing 'p' exponent
		{"1p2", exact128(0), strconv.ErrSyntax},   // missing '0x' prefix

		// Rounding
		{
			"0x1.00000000000000000000000000008p+00",
			exact128(1.0), nil, // round down
		},
		{
			"0x1.00000000000000000000000000008000001p+00",
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0001}, nil, // round up
		},
		{
			"0x1.00000000000000000000000000017ffffffp+00",
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0001}, nil, // round down
		},
		{
			"0x1.00000000000000000000000000018p+00",
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0002}, nil, // round up
		},
		{
			"0x1.ffffffffffffffffffffffffffff8p+00",
			exact128(2.0), nil, // round up
		},

		// NaNs
		{"nan", exact128(math.NaN()), nil},
		{"NaN", exact128(math.NaN()), nil},
		{"NAN", exact128(math.NaN()), nil},

		// Infs
		{"Inf", exact128(math.Inf(1)), nil},
		{"-Inf", exact128(math.Inf(-1)), nil},
		{"+INF", exact128(math.Inf(1)), nil},
		{"-Infinity", exact128(math.Inf(-1)), nil},
		{"+INFINITY", exact128(math.Inf(1)), nil},
		{"Infinity", exact128(math.Inf(1)), nil},

		// try to overflow exponent
		{"1e-4294967296", exact128(0), nil},
		{"1e+4294967296", exact128(math.Inf(1)), strconv.ErrRange},
		{"1e-18446744073709551616", exact128(0), nil},
		{"1e+18446744073709551616", exact128(math.Inf(1)), strconv.ErrRange},
		{"0x1p-4294967296", exact128(0), nil},
		{"0x1p+4294967296", exact128(math.Inf(1)), strconv.ErrRange},
		{"0x1p-18446744073709551616", exact128(0), nil},
		{"0x1p+18446744073709551616", exact128(math.Inf(1)), strconv.ErrRange},

		// Parse errors
		{"1e", exact128(0), strconv.ErrSyntax},
		{"1e-", exact128(0), strconv.ErrSyntax},
		{".e-1", exact128(0), strconv.ErrSyntax},
		{"1\x00.2", exact128(0), strconv.ErrSyntax},
		{"0x", exact128(0), strconv.ErrSyntax},
		{"0x.", exact128(0), strconv.ErrSyntax},
		{"0x1", exact128(0), strconv.ErrSyntax},
		{"0x.1", exact128(0), strconv.ErrSyntax},
		{"0x1p", exact128(0), strconv.ErrSyntax},
		{"0x.1p", exact128(0), strconv.ErrSyntax},
		{"0x1p+", exact128(0), strconv.ErrSyntax},
		{"0x.1p+", exact128(0), strconv.ErrSyntax},
		{"0x1p-", exact128(0), strconv.ErrSyntax},
		{"0x.1p-", exact128(0), strconv.ErrSyntax},
		{"0x1p+2", exact128(4), nil},
		{"0x.1p+2", exact128(0.25), nil},
		{"0x1p-2", exact128(0.25), nil},
		{"0x.1p-2", exact128(0.015625), nil},
	}

	for _, tt := range tests {
		got, err := ParseFloat128(tt.input)
		if err != nil {
			numErr, ok := err.(*strconv.NumError)
			if !ok {
				t.Errorf("ParseFloat128(%q) unexpected error type: %v", tt.input, err)
				continue
			}
			if numErr.Func != "ParseFloat128" {
				t.Errorf("ParseFloat128(%q) unexpected Func in NumError: got %q, want %q", tt.input, numErr.Func, "ParseFloat128")
			}
			if numErr.Num != tt.input {
				t.Errorf("ParseFloat128(%q) unexpected Num in NumError: got %q, want %q", tt.input, numErr.Num, tt.input)
			}
			err = numErr.Err
		}
		if !eq128(got, tt.want) || err != tt.err {
			t.Errorf("ParseFloat128(%q) = (%v, %v) want (%v, %v)", tt.input, got, err, tt.want, tt.err)
		}
	}
}

func BenchmarkParseFloat128_Decimal(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat128("33909")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFloat128_Float(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat128("339.778")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFloat128_FloatExp(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat128("-5.09e-3")
		if err != nil {
			b.Fatal(err)
		}
	}
}
