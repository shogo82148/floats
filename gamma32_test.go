package floats

import (
	"math"
	"testing"
)

func TestFloat32_Gamma(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-0.5), math.Gamma(-0.5)},
		{exact32(0.5), math.Gamma(0.5)},
		{exact32(1), math.Gamma(1)},
		{exact32(1.5), math.Gamma(1.5)},
		{exact32(2), math.Gamma(2)},
		{exact32(2.5), math.Gamma(2.5)},
		{exact32(3), math.Gamma(3)},
	}

	for _, tt := range tests {
		got := tt.x.Gamma()
		if !close32(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(0), exact32(math.Inf(1))},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(-1))},
		{exact32(-1), exact32(math.NaN())},
		{exact32(-2), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Gamma()
		if !eq32(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
