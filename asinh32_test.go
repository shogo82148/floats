package floats

import (
	"math"
	"testing"
)

func TestFloat32_Asinh(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0), math.Asinh(0)},
		{exact32(0.25), math.Asinh(0.25)},
		{exact32(0.5), math.Asinh(0.5)},
		{exact32(1), math.Asinh(1)},
		{exact32(21), math.Asinh(21)},
		{exact32(22), math.Asinh(22)},

		{exact32(-0), -math.Asinh(0)},
		{exact32(-0.25), -math.Asinh(0.25)},
		{exact32(-0.5), -math.Asinh(0.5)},
		{exact32(-1), -math.Asinh(1)},
		{exact32(-21), -math.Asinh(21)},
		{exact32(-22), -math.Asinh(22)},
	}

	for _, tt := range tests {
		got := tt.x.Asinh()
		if !close32(got, tt.want) {
			t.Errorf("Asinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},

		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(math.Inf(-1))},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Asinh()
		if !eq32(got, tt.want) {
			t.Errorf("Asinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
