package floats

import (
	"math"
	"testing"
)

func TestParseFloat32(t *testing.T) {
	tests := []struct {
		input string
		want  Float32
	}{
		{"0", NewFloat32(0)},
		{"-0", NewFloat32(math.Copysign(0, -1))},
		{"1.5", NewFloat32(1.5)},
		{"-2.75", NewFloat32(-2.75)},
		{"0x1.8p+1", NewFloat32(3.0)},
		{"3.4028234663852886e+38", NewFloat32(3.4028234663852886e+38)}, // Max float32
		{"Inf", NewFloat32(math.Inf(1))},
		{"-Inf", NewFloat32(math.Inf(-1))},
	}

	for _, tt := range tests {
		got, err := ParseFloat32(tt.input)
		if err != nil {
			t.Errorf("ParseFloat32(%q) returned error: %v", tt.input, err)
			continue
		}
		if got.Bits() != tt.want.Bits() {
			t.Errorf("ParseFloat32(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func BenchmarkParseFloat32_Decimal(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat32("33909")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFloat32_Float(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat32("339.778")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseFloat32_FloatExp(b *testing.B) {
	for b.Loop() {
		_, err := ParseFloat32("-5.09e-3")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestFloat32_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		input string
		want  Float32
	}{
		{"0", exact32(0)},
		{"1.5", exact32(1.5)},
		{"-2.75", exact32(-2.75)},
	}

	for _, tt := range tests {
		var f Float32
		err := f.UnmarshalJSON([]byte(tt.input))
		if err != nil {
			t.Errorf("Float32.UnmarshalJSON(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if !eq32(f, tt.want) {
			t.Errorf("Float32.UnmarshalJSON(%q) = %v; want %v", tt.input, f, tt.want)
		}
	}

	// invalid input does not modify the receiver
	f := exact32(1.5)
	err := f.UnmarshalJSON([]byte("invalid"))
	if err == nil {
		t.Errorf("Float32.UnmarshalJSON(%q) expected error, got nil", "invalid")
	}
	if !eq32(f, exact32(1.5)) {
		t.Errorf("Float32.UnmarshalJSON(%q) modified receiver on error: got %v, want %v", "invalid", f, exact32(1.5))
	}
}

func TestFloat32_UnmarshalText(t *testing.T) {
	tests := []struct {
		input string
		want  Float32
	}{
		{"0", exact32(0)},
		{"1.5", exact32(1.5)},
		{"-2.75", exact32(-2.75)},
	}

	for _, tt := range tests {
		var f Float32
		err := f.UnmarshalText([]byte(tt.input))
		if err != nil {
			t.Errorf("Float32.UnmarshalText(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if !eq32(f, tt.want) {
			t.Errorf("Float32.UnmarshalText(%q) = %v; want %v", tt.input, f, tt.want)
		}
	}

	// invalid input does not modify the receiver
	f := exact32(1.5)
	err := f.UnmarshalText([]byte("invalid"))
	if err == nil {
		t.Errorf("Float32.UnmarshalText(%q) expected error, got nil", "invalid")
	}
	if !eq32(f, exact32(1.5)) {
		t.Errorf("Float32.UnmarshalText(%q) modified receiver on error: got %v, want %v", "invalid", f, exact32(1.5))
	}
}
