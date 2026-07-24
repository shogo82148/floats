package floats

import (
	"math"
	"testing"
)

func TestFloat64_Yn(t *testing.T) {
	type tc struct {
		n int
		x Float64
	}
	tests := []tc{
		{2, 1}, {2, 5}, {3, 10}, {5, 20}, {-2, 5}, {2, 0}, {-3, 0},
	}

	for _, tt := range tests {
		want := math.Yn(tt.n, float64(tt.x))
		got := tt.x.Yn(tt.n)
		if !close64(got, want) {
			t.Errorf("Yn(%d, %v) = %v; want %v", tt.n, tt.x, got, want)
		}
	}
}
