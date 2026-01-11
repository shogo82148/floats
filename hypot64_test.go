package floats

import (
	"math"
	"testing"
)

func TestHypot64(t *testing.T) {
	tests := []struct {
		x    Float64
		y    Float64
		want float64
	}{
		{exact64(3), exact64(4), 5},
		{exact64(5), exact64(12), 13},
		{exact64(1), exact64(1), math.Sqrt(2)},
	}

	for _, tt := range tests {
		got := Hypot64(tt.x, tt.y)
		if !close64(got, tt.want) {
			t.Errorf("Hypot64(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		y    Float64
		want Float64
	}{
		// special cases
		{exact64(0), exact64(0), exact64(0)},
		{exact64(math.Inf(1)), exact64(1), exact64(math.Inf(1))},
		{exact64(math.Inf(-1)), exact64(1), exact64(math.Inf(1))},
		{exact64(1), exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(1), exact64(math.Inf(-1)), exact64(math.Inf(1))},
		{exact64(math.NaN()), exact64(1), exact64(math.NaN())},
		{exact64(1), exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := Hypot64(tt.x, tt.y)
		if !eq64(got, tt.want) {
			t.Errorf("Hypot64(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}
