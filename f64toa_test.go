package floats

import (
	"math"
	"testing"
)

func TestFloat64_Text(t *testing.T) {
	tests := []struct {
		x    Float64
		fmt  byte
		prec int
		s    string
	}{
		{Float64(0), 'e', 6, "0.000000e+00"},
		{Float64(1.5), 'f', 2, "1.50"},
		{Float64(-2.75), 'g', -1, "-2.75"},
		{Float64(3.1415927), 'g', 5, "3.1416"},
		{Float64(math.Inf(1)), 'g', -1, "+Inf"},
		{Float64(math.Inf(-1)), 'g', -1, "-Inf"},
		{Float64(math.NaN()), 'g', -1, "NaN"},
	}

	for _, tt := range tests {
		if got := tt.x.Text(tt.fmt, tt.prec); got != tt.s {
			t.Errorf("Float64(%g).Text(%q, %d) = %q, want %q", float64(tt.x), tt.fmt, tt.prec, got, tt.s)
		}
	}
}
