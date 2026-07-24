package floats

import (
	"math"
	"testing"
)

func TestFloat32_J1(t *testing.T) {
	tests := []float64{-50, -10, -1, -0.5, 0, 0.5, 1, 2, 5, 10, 50}

	for _, x := range tests {
		want := math.J1(x)
		got := exact32(x).J1()
		if !close32(got, want) {
			t.Errorf("J1(%v) = %v; want %v", x, got, want)
		}
	}
}
