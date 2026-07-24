package floats

import (
	"math"
	"testing"
)

func TestFloat16_Yn(t *testing.T) {
	type tc struct {
		n int
		x float64
	}
	tests := []tc{
		{2, 1}, {2, 5}, {3, 10}, {5, 20}, {-2, 5}, {2, 0}, {-3, 0},
	}

	for _, tt := range tests {
		want := math.Yn(tt.n, tt.x)
		got := exact16(tt.x).Yn(tt.n)
		if !close16(got, want) {
			t.Errorf("Yn(%d, %v) = %v; want %v", tt.n, tt.x, got, want)
		}
	}
}
