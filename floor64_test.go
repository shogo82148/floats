package floats

import (
	"math"
	"testing"
)

func TestFloat64_Floor(t *testing.T) {
	tests := []struct {
		x    Float64
		want Float64
	}{
		{exact64(3.0), exact64(3)},
		{exact64(3.5), exact64(3)},
		{exact64(-3.0), exact64(-3)},
		{exact64(-3.5), exact64(-4)},

		// Special cases
		{exact64(0), exact64(0)},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},
		{exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(math.Inf(-1)), exact64(math.Inf(-1))},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Floor()
		if !eq64(got, tt.want) {
			t.Errorf("Float64.Floor(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat64_Ceil(t *testing.T) {
	tests := []struct {
		x    Float64
		want Float64
	}{
		{exact64(3.0), exact64(3)},
		{exact64(3.5), exact64(4)},
		{exact64(-3.0), exact64(-3)},
		{exact64(-3.5), exact64(-3)},

		// Special cases
		{exact64(0), exact64(0)},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},
		{exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(math.Inf(-1)), exact64(math.Inf(-1))},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Ceil()
		if !eq64(got, tt.want) {
			t.Errorf("Float64.Ceil(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
