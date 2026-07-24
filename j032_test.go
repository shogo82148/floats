package floats

import (
	"math"
	"testing"
)

func TestFloat32_J0(t *testing.T) {
	tests := []float64{-10, -1, -0.5, 0, 0.5, 1, 2, 5, 10, 50}

	for _, x := range tests {
		want := math.J0(x)
		got := exact32(x).J0()
		if !close32(got, want) {
			t.Errorf("J0(%v) = %v; want %v", x, got, want)
		}
	}
}
