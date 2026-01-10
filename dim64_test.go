package floats

import (
	"math"
	"testing"
)

func TestFloat64_Dim(t *testing.T) {
	tests := []struct {
		a    Float64
		b    Float64
		want Float64
	}{
		{exact64(5.0), exact64(3.0), exact64(2.0)},
		{exact64(3.0), exact64(5.0), exact64(0.0)},
		{exact64(3.0), exact64(3.0), exact64(0.0)},

		// Special cases
		{exact64(math.Inf(1)), exact64(math.Inf(1)), exact64(math.NaN())},
		{exact64(math.Inf(-1)), exact64(math.Inf(-1)), exact64(math.NaN())},
		{exact64(math.NaN()), exact64(1.0), exact64(math.NaN())},
		{exact64(1.0), exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.a.Dim(tt.b)
		if !eq64(got, tt.want) {
			t.Errorf("Float64.Dim(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestFloat64_Max(t *testing.T) {
	tests := []struct {
		a    Float64
		b    Float64
		want Float64
	}{
		{exact64(5.0), exact64(3.0), exact64(5.0)},
		{exact64(3.0), exact64(5.0), exact64(5.0)},
		{exact64(3.0), exact64(3.0), exact64(3.0)},

		// Special cases
		{exact64(1), exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(math.Inf(1)), exact64(1), exact64(math.Inf(1))},
		{exact64(math.NaN()), exact64(1), exact64(math.NaN())},
		{exact64(1), exact64(math.NaN()), exact64(math.NaN())},
		{exact64(0), exact64(math.Copysign(0, -1)), exact64(0)},
		{exact64(math.Copysign(0, -1)), exact64(0), exact64(0)},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},
	}
	for _, tt := range tests {
		got := tt.a.Max(tt.b)
		if !eq64(got, tt.want) {
			t.Errorf("Float16.Max(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestFloat64_Min(t *testing.T) {
	tests := []struct {
		a    Float64
		b    Float64
		want Float64
	}{
		{exact64(5.0), exact64(3.0), exact64(3.0)},
		{exact64(3.0), exact64(5.0), exact64(3.0)},
		{exact64(3.0), exact64(3.0), exact64(3.0)},

		// Special cases
		{exact64(1), exact64(math.Inf(-1)), exact64(math.Inf(-1))},
		{exact64(math.Inf(-1)), exact64(1), exact64(math.Inf(-1))},
		{exact64(math.NaN()), exact64(1), exact64(math.NaN())},
		{exact64(1), exact64(math.NaN()), exact64(math.NaN())},
		{exact64(0), exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},
		{exact64(math.Copysign(0, -1)), exact64(0), exact64(math.Copysign(0, -1))},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},
	}
	for _, tt := range tests {
		got := tt.a.Min(tt.b)
		if !eq64(got, tt.want) {
			t.Errorf("Float64.Min(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
