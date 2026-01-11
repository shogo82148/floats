package floats

import (
	"math"
	"testing"
)

func TestHypot256(t *testing.T) {
	tests := []struct {
		x    Float256
		y    Float256
		want string
	}{
		{exact256(3), exact256(4), "5"},
		{exact256(5), exact256(12), "13"},
		{exact256(1), exact256(1), "1.414213562373095048801688724209698078569671875376948073176679737990733"},
	}

	for _, tt := range tests {
		got := Hypot256(tt.x, tt.y)
		if !close256(got, tt.want) {
			t.Errorf("Hypot256(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		y    Float256
		want Float256
	}{
		// special cases
		{exact256(0), exact256(0), exact256(0)},
		{exact256(math.Inf(1)), exact256(1), exact256(math.Inf(1))},
		{exact256(math.Inf(-1)), exact256(1), exact256(math.Inf(1))},
		{exact256(1), exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(1), exact256(math.Inf(-1)), exact256(math.Inf(1))},
		{exact256(math.NaN()), exact256(1), exact256(math.NaN())},
		{exact256(1), exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := Hypot256(tt.x, tt.y)
		if !eq256(got, tt.want) {
			t.Errorf("Hypot256(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}
