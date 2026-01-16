package floats

import (
	"math"
	"testing"
)

func TestFloat32_Exp(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-88), math.Exp(-88)},
		{exact32(-1), math.Exp(-1)},
		{exact32(0), math.Exp(0)},
		{exact32(1), math.Exp(1)},
		{exact32(2), math.Exp(2)},
		{exact32(3), math.Exp(3)},
		{exact32(4), math.Exp(4)},
		{exact32(88), math.Exp(88)},
	}

	for _, tt := range tests {
		got := tt.x.Exp()
		if !close32(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		{exact32(0), exact32(1)},
		{exact32(0x1p-15), exact32(1 + 0x1p-15)},
		{exact32(-0x1p-15), exact32(1 - 0x1p-15)},
		{exact32(0x1.62e44p+06), exact32(math.Inf(1))},
		{exact32(-0x1.9fe36ap+06), exact32(0)},

		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(0)},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Exp()
		if !eq32(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
