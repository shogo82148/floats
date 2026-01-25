package floats

import (
	"math"
	"testing"
)

func TestFloat32_Sin(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-3), math.Sin(-3)},
		{exact32(-2), math.Sin(-2)},
		{exact32(-1), math.Sin(-1)},
		{exact32(1), math.Sin(1)},
		{exact32(2), math.Sin(2)},
		{exact32(3), math.Sin(3)},
		{exact32(0x1p28), math.Sin(0x1p28)},
	}

	for _, tt := range tests {
		got := tt.x.Sin()
		if !close32(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.Inf(1)), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Sin()
		if !eq32(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
