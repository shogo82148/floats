package floats

import (
	"math"
	"testing"
)

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
