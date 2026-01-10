package floats

import (
	"math"
	"testing"
)

func TestFloat128_Dim(t *testing.T) {
	tests := []struct {
		a    Float128
		b    Float128
		want Float128
	}{
		{exact128(5.0), exact128(3.0), exact128(2.0)},
		{exact128(3.0), exact128(5.0), exact128(0.0)},
		{exact128(3.0), exact128(3.0), exact128(0.0)},

		// Special cases
		{exact128(math.Inf(1)), exact128(math.Inf(1)), exact128(math.NaN())},
		{exact128(math.Inf(-1)), exact128(math.Inf(-1)), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(1.0), exact128(math.NaN())},
		{exact128(1.0), exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.a.Dim(tt.b)
		if !eq128(got, tt.want) {
			t.Errorf("Float128.Dim(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestFloat128_Max(t *testing.T) {
	tests := []struct {
		a    Float128
		b    Float128
		want Float128
	}{
		{exact128(5.0), exact128(3.0), exact128(5.0)},
		{exact128(3.0), exact128(5.0), exact128(5.0)},
		{exact128(3.0), exact128(3.0), exact128(3.0)},

		// Special cases
		{exact128(1), exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(math.Inf(1)), exact128(1), exact128(math.Inf(1))},
		{exact128(math.NaN()), exact128(1), exact128(math.NaN())},
		{exact128(1), exact128(math.NaN()), exact128(math.NaN())},
		{exact128(0), exact128(math.Copysign(0, -1)), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(0), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},
	}
	for _, tt := range tests {
		got := tt.a.Max(tt.b)
		if !eq128(got, tt.want) {
			t.Errorf("Float128.Max(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestFloat128_Min(t *testing.T) {
	tests := []struct {
		a    Float128
		b    Float128
		want Float128
	}{
		{exact128(5.0), exact128(3.0), exact128(3.0)},
		{exact128(3.0), exact128(5.0), exact128(3.0)},
		{exact128(3.0), exact128(3.0), exact128(3.0)},

		// Special cases
		{exact128(1), exact128(math.Inf(-1)), exact128(math.Inf(-1))},
		{exact128(math.Inf(-1)), exact128(1), exact128(math.Inf(-1))},
		{exact128(math.NaN()), exact128(1), exact128(math.NaN())},
		{exact128(1), exact128(math.NaN()), exact128(math.NaN())},
		{exact128(0), exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},
		{exact128(math.Copysign(0, -1)), exact128(0), exact128(math.Copysign(0, -1))},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},
	}
	for _, tt := range tests {
		got := tt.a.Min(tt.b)
		if !eq128(got, tt.want) {
			t.Errorf("Float128.Min(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
