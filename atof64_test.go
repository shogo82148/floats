package floats

import (
	"math"
	"testing"
)

func TestParseFloat64(t *testing.T) {
	tests := []struct {
		input string
		want  Float64
	}{
		{"0", NewFloat64(0)},
		{"-0", NewFloat64(math.Copysign(0, -1))},
		{"1.5", NewFloat64(1.5)},
		{"-2.75", NewFloat64(-2.75)},
		{"0x1.8p+1", NewFloat64(3.0)},
		{"3.4028234663852886e+38", NewFloat64(3.4028234663852886e+38)},   // Max float32
		{"1.7976931348623157e+308", NewFloat64(1.7976931348623157e+308)}, // Max float64
		{"Inf", NewFloat64(math.Inf(1))},
		{"-Inf", NewFloat64(math.Inf(-1))},
	}

	for _, tt := range tests {
		got, err := ParseFloat64(tt.input)
		if err != nil {
			t.Errorf("ParseFloat64(%q) returned error: %v", tt.input, err)
			continue
		}
		if got.Bits() != tt.want.Bits() {
			t.Errorf("ParseFloat64(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func BenchmarkParseFloat64_Decimal(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat64("33909")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFloat64_Float(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat64("339.778")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFloat64_FloatExp(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat64("-5.09e-3")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFloat64_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  Float64
	}{
		{"0", exact64(0)},
		{"1.5", exact64(1.5)},
		{"-2.75", exact64(-2.75)},
	}

	for _, tt := range tests {
		var f Float64
		err := f.UnmarshalJSON([]byte(tt.input))
		if err != nil {
			t.Errorf("Float64.UnmarshalJSON(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if !eq64(f, tt.want) {
			t.Errorf("Float64.UnmarshalJSON(%q) = %v; want %v", tt.input, f, tt.want)
		}
	}

	// invalid input does not modify the receiver
	f := exact64(1.5)
	err := f.UnmarshalJSON([]byte("invalid"))
	if err == nil {
		t.Errorf("Float64.UnmarshalJSON(%q) expected error, got nil", "invalid")
	}
	if !eq64(f, exact64(1.5)) {
		t.Errorf("Float64.UnmarshalJSON(%q) modified receiver on error: got %v, want %v", "invalid", f, exact64(1.5))
	}
}

func TestFloat64_UnmarshalText(t *testing.T) {
	tests := []struct {
		input string
		want  Float64
	}{
		{"0", exact64(0)},
		{"1.5", exact64(1.5)},
		{"-2.75", exact64(-2.75)},
	}

	for _, tt := range tests {
		var f Float64
		err := f.UnmarshalText([]byte(tt.input))
		if err != nil {
			t.Errorf("Float64.UnmarshalText(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if !eq64(f, tt.want) {
			t.Errorf("Float64.UnmarshalText(%q) = %v; want %v", tt.input, f, tt.want)
		}
	}

	// invalid input does not modify the receiver
	f := exact64(1.5)
	err := f.UnmarshalText([]byte("invalid"))
	if err == nil {
		t.Errorf("Float64.UnmarshalText(%q) expected error, got nil", "invalid")
	}
	if !eq64(f, exact64(1.5)) {
		t.Errorf("Float64.UnmarshalText(%q) modified receiver on error: got %v, want %v", "invalid", f, exact64(1.5))
	}
}
