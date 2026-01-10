package floats

import (
	"math"
	"testing"
)

func TestFloat128_Floor(t *testing.T) {
	tests := []struct {
		x    Float128
		want Float128
	}{
		{exact128(3.0), exact128(3)},
		{exact128(3.5), exact128(3)},
		{exact128(-3.0), exact128(-3)},
		{exact128(-3.5), exact128(-4)},

		// Special cases
		{exact128(0), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},
		{exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(math.Inf(-1)), exact128(math.Inf(-1))},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Floor()
		if !eq128(got, tt.want) {
			t.Errorf("Float128.Floor(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat128_Ceil(t *testing.T) {
	tests := []struct {
		x    Float128
		want Float128
	}{
		{exact128(3.0), exact128(3)},
		{exact128(3.5), exact128(4)},
		{exact128(-3.0), exact128(-3)},
		{exact128(-3.5), exact128(-3)},

		// Special cases
		{exact128(0), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},
		{exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(math.Inf(-1)), exact128(math.Inf(-1))},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Ceil()
		if !eq128(got, tt.want) {
			t.Errorf("Float128.Ceil(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
