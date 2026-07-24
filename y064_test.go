package floats

import (
	"math"
	"testing"
)

func TestFloat64_Y0(t *testing.T) {
	tests := []Float64{0.5, 1, 2, 5, 10, 50}

	for _, x := range tests {
		want := math.Y0(float64(x))
		got := x.Y0()
		if !close64(got, want) {
			t.Errorf("Y0(%v) = %v; want %v", x, got, want)
		}
	}
}
