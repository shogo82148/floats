package floats

import (
	"math"
	"testing"
)

func TestHypot16(t *testing.T) {
	tests := []struct {
		x    Float16
		y    Float16
		want float64
	}{
		{exact16(3), exact16(4), 5},
		{exact16(5), exact16(12), 13},
		{exact16(1), exact16(1), math.Sqrt(2)},
	}

	for _, tt := range tests {
		got := Hypot16(tt.x, tt.y)
		if !close16(got, tt.want) {
			t.Errorf("Hypot16(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		y    Float16
		want Float16
	}{
		// special cases
		{exact16(0), exact16(0), exact16(0)},
		{exact16(math.Inf(1)), exact16(1), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(1), exact16(math.Inf(1))},
		{exact16(1), exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(1), exact16(math.Inf(-1)), exact16(math.Inf(1))},
		{exact16(math.NaN()), exact16(1), exact16(math.NaN())},
		{exact16(1), exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := Hypot16(tt.x, tt.y)
		if !eq16(got, tt.want) {
			t.Errorf("Hypot16(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}
