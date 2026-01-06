package floats

import (
	"math"
	"testing"
)

func TestParseFloat16(t *testing.T) {
	tests := []struct {
		input string
		want  Float16
	}{
		{"0", exact16(0)},
		{"-0", exact16(math.Copysign(0, -1))},
		{"1.5", exact16(1.5)},
		{"-2.75", exact16(-2.75)},
		{"0x1.8p+1", exact16(3.0)},
		{"Inf", exact16(math.Inf(1))},
		{"-Inf", exact16(math.Inf(-1))},
	}

	for _, tt := range tests {
		got, err := ParseFloat16(tt.input)
		if err != nil {
			t.Errorf("ParseFloat16(%q) returned error: %v", tt.input, err)
			continue
		}
		if got.Bits() != tt.want.Bits() {
			t.Errorf("ParseFloat16(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
