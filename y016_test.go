package floats

import (
	"math"
	"testing"
)

func TestFloat16_Y0(t *testing.T) {
	tests := []float64{0.5, 1, 2, 5, 10, 50}

	for _, x := range tests {
		want := math.Y0(x)
		got := exact16(x).Y0()
		if !close16(got, want) {
			t.Errorf("Y0(%v) = %v; want %v", x, got, want)
		}
	}
}
