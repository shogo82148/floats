package floats

import (
	"math"
	"testing"
)

func TestFloat32_Jn(t *testing.T) {
	type tc struct {
		n int
		x float64
	}
	tests := []tc{
		{2, -10}, {2, -1}, {2, 0}, {2, 1}, {2, 5}, {3, 10}, {5, 20}, {-2, 5},
	}

	for _, tt := range tests {
		want := math.Jn(tt.n, tt.x)
		got := exact32(tt.x).Jn(tt.n)
		if !close32(got, want) {
			t.Errorf("Jn(%d, %v) = %v; want %v", tt.n, tt.x, got, want)
		}
	}
}
