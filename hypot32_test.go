package floats

import (
	"math"
	"testing"
)

func TestHypot32(t *testing.T) {
	tests := []struct {
		x    Float32
		y    Float32
		want float64
	}{
		{exact32(3), exact32(4), 5},
		{exact32(5), exact32(12), 13},
		{exact32(1), exact32(1), math.Sqrt(2)},
	}

	for _, tt := range tests {
		got := Hypot32(tt.x, tt.y)
		if !close32(got, tt.want) {
			t.Errorf("Hypot32(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		y    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(0), exact32(0)},
		{exact32(math.Inf(1)), exact32(1), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(1), exact32(math.Inf(1))},
		{exact32(1), exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(1), exact32(math.Inf(-1)), exact32(math.Inf(1))},
		{exact32(math.NaN()), exact32(1), exact32(math.NaN())},
		{exact32(1), exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := Hypot32(tt.x, tt.y)
		if !eq32(got, tt.want) {
			t.Errorf("Hypot32(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}
