package floats

import (
	"math"
	"testing"
)

func TestFloat64_J0(t *testing.T) {
	tests := []Float64{-10, -1, -0.5, 0, 0.5, 1, 2, 5, 10, 50}

	for _, x := range tests {
		want := math.J0(float64(x))
		got := x.J0()
		if !close64(got, want) {
			t.Errorf("J0(%v) = %v; want %v", x, got, want)
		}
	}
}
