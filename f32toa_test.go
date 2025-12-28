package floats

import (
	"fmt"
	"math"
	"testing"
)

func TestFloat32_Format(t *testing.T) {
	tests := []struct {
		format string
		x      Float32
		want   string
	}{
		// verb "%b"
		{"%b", exact32(0), "0p-149"},

		// verb "%f"
		{"%f", exact32(0.5), "0.5"},
		{"%f", exact32(-0.5), "-0.5"},
		{"%+f", exact32(0.5), "+0.5"},
		{"%+f", exact32(-0.5), "-0.5"},
		{"% f", exact32(0.5), " 0.5"},
		{"% f", exact32(-0.5), "-0.5"},
		{"%8f", exact32(0.5), "     0.5"},
		{"%-8f", exact32(0.5), "0.5     "},
		{"%.2f", exact32(0.5), "0.50"},

		// verb "%e"
		{"%.6e", exact32(0.5), "5.000000e-01"},

		// verb "%g"
		{"%g", exact32(0.5), "0.5"},
		{"%.1g", exact32(0.25), "0.2"},
		// verb "%x"
		{"%x", exact32(0.5), "0x1p-01"},
		{"%#x", exact32(0.5), "0x1p-01"},
		{"%.1x", exact32(0.5), "0x1.0p-01"},

		// verb "%X"
		{"%X", exact32(0.5), "0X1P-01"},
		{"%#X", exact32(0.5), "0X1P-01"},
		{"%.1X", exact32(0.5), "0X1.0P-01"},

		// verb "%v"
		{"%v", exact32(0.5), "0.5"},
		{"%v", exact32(math.NaN()), "NaN"},
	}

	for _, tt := range tests {
		got := fmt.Sprintf(tt.format, tt.x)
		if got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

func TestFloat32_Text(t *testing.T) {
	tests := []struct {
		x    Float32
		fmt  byte
		prec int
		s    string
	}{
		{Float32(0), 'e', 6, "0.000000e+00"},
		{Float32(1.5), 'f', 2, "1.50"},
		{Float32(-2.75), 'g', -1, "-2.75"},
		{Float32(3.1415927), 'g', 5, "3.1416"},
		{Float32(math.Inf(1)), 'g', -1, "+Inf"},
		{Float32(math.Inf(-1)), 'g', -1, "-Inf"},
		{Float32(math.NaN()), 'g', -1, "NaN"},
	}

	for _, tt := range tests {
		if got := tt.x.Text(tt.fmt, tt.prec); got != tt.s {
			t.Errorf("Float32(%g).Text(%q, %d) = %q, want %q", float32(tt.x), tt.fmt, tt.prec, got, tt.s)
		}
	}
}
