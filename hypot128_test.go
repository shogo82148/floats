package floats

import (
	"math"
	"testing"
)

func TestHypot128(t *testing.T) {
	tests := []struct {
		x    Float128
		y    Float128
		want float64
	}{
		{exact128(3), exact128(4), 5},
		{exact128(5), exact128(12), 13},
		{exact128(1), exact128(1), math.Sqrt(2)},
	}

	for _, tt := range tests {
		got := Hypot128(tt.x, tt.y)
		if !close128(got, tt.want) {
			t.Errorf("Hypot128(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		y    Float128
		want Float128
	}{
		// special cases
		{exact128(0), exact128(0), exact128(0)},
		{exact128(math.Inf(1)), exact128(1), exact128(math.Inf(1))},
		{exact128(math.Inf(-1)), exact128(1), exact128(math.Inf(1))},
		{exact128(1), exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(1), exact128(math.Inf(-1)), exact128(math.Inf(1))},
		{exact128(math.NaN()), exact128(1), exact128(math.NaN())},
		{exact128(1), exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := Hypot128(tt.x, tt.y)
		if !eq128(got, tt.want) {
			t.Errorf("Hypot128(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}
