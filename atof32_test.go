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
		{"-0", NewFloat32(-0)},
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
		if got != tt.want {
			t.Errorf("ParseFloat32(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
