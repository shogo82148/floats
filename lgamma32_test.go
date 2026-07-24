package floats

import (
	"math"
	"testing"
)

func TestFloat32_Lgamma(t *testing.T) {
	tests := []float64{-2.5, -0.5, 0.5, 1, 1.5, 2, 2.5, 3, 100}

	for _, x := range tests {
		want, wantSign := math.Lgamma(x)
		got, sign := exact32(x).Lgamma()
		if !close32(got, want) || sign != wantSign {
			t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", x, got, sign, want, wantSign)
		}
	}
}
