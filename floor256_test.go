package floats

import (
	"math"
	"testing"
)

func TestFloat256_Floor(t *testing.T) {
	tests := []struct {
		x    Float256
		want Float256
	}{
		{exact256(3.0), exact256(3)},
		{exact256(3.5), exact256(3)},
		{exact256(-3.0), exact256(-3)},
		{exact256(-3.5), exact256(-4)},

		// Special cases
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(math.Inf(-1)), exact256(math.Inf(-1))},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Floor()
		if !eq256(got, tt.want) {
			t.Errorf("Float256.Floor(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat256_Ceil(t *testing.T) {
	tests := []struct {
		x    Float256
		want Float256
	}{
		{exact256(3.0), exact256(3)},
		{exact256(3.5), exact256(4)},
		{exact256(-3.0), exact256(-3)},
		{exact256(-3.5), exact256(-3)},

		// Special cases
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(math.Inf(-1)), exact256(math.Inf(-1))},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Ceil()
		if !eq256(got, tt.want) {
			t.Errorf("Float256.Ceil(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat256_Trunc(t *testing.T) {
	tests := []struct {
		x    Float256
		want Float256
	}{
		{exact256(3.0), exact256(3)},
		{exact256(3.5), exact256(3)},
		{exact256(-3.0), exact256(-3)},
		{exact256(-3.5), exact256(-3)},

		// Special cases
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(math.Inf(-1)), exact256(math.Inf(-1))},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Trunc()
		if !eq256(got, tt.want) {
			t.Errorf("Float256.Trunc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
