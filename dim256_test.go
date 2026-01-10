package floats

import (
	"math"
	"testing"
)

func TestFloat256_Dim(t *testing.T) {
	tests := []struct {
		a    Float256
		b    Float256
		want Float256
	}{
		{exact256(5.0), exact256(3.0), exact256(2.0)},
		{exact256(3.0), exact256(5.0), exact256(0.0)},
		{exact256(3.0), exact256(3.0), exact256(0.0)},

		// Special cases
		{exact256(math.Inf(1)), exact256(math.Inf(1)), exact256(math.NaN())},
		{exact256(math.Inf(-1)), exact256(math.Inf(-1)), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(1.0), exact256(math.NaN())},
		{exact256(1.0), exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.a.Dim(tt.b)
		if !eq256(got, tt.want) {
			t.Errorf("Float256.Dim(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestFloat256_Max(t *testing.T) {
	tests := []struct {
		a    Float256
		b    Float256
		want Float256
	}{
		{exact256(5.0), exact256(3.0), exact256(5.0)},
		{exact256(3.0), exact256(5.0), exact256(5.0)},
		{exact256(3.0), exact256(3.0), exact256(3.0)},

		// Special cases
		{exact256(1), exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(math.Inf(1)), exact256(1), exact256(math.Inf(1))},
		{exact256(math.NaN()), exact256(1), exact256(math.NaN())},
		{exact256(1), exact256(math.NaN()), exact256(math.NaN())},
		{exact256(0), exact256(math.Copysign(0, -1)), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
	}
	for _, tt := range tests {
		got := tt.a.Max(tt.b)
		if !eq256(got, tt.want) {
			t.Errorf("Float256.Max(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestFloat256_Min(t *testing.T) {
	tests := []struct {
		a    Float256
		b    Float256
		want Float256
	}{
		{exact256(5.0), exact256(3.0), exact256(3.0)},
		{exact256(3.0), exact256(5.0), exact256(3.0)},
		{exact256(3.0), exact256(3.0), exact256(3.0)},

		// Special cases
		{exact256(1), exact256(math.Inf(-1)), exact256(math.Inf(-1))},
		{exact256(math.Inf(-1)), exact256(1), exact256(math.Inf(-1))},
		{exact256(math.NaN()), exact256(1), exact256(math.NaN())},
		{exact256(1), exact256(math.NaN()), exact256(math.NaN())},
		{exact256(0), exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
		{exact256(math.Copysign(0, -1)), exact256(0), exact256(math.Copysign(0, -1))},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
	}
	for _, tt := range tests {
		got := tt.a.Min(tt.b)
		if !eq256(got, tt.want) {
			t.Errorf("Float256.Min(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
