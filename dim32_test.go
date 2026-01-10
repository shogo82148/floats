package floats

import (
	"math"
	"testing"
)

func TestFloat32_Dim(t *testing.T) {
	tests := []struct {
		a    Float32
		b    Float32
		want Float32
	}{
		{exact32(5.0), exact32(3.0), exact32(2.0)},
		{exact32(3.0), exact32(5.0), exact32(0.0)},
		{exact32(3.0), exact32(3.0), exact32(0.0)},

		// Special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1)), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.Inf(-1)), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(1.0), exact32(math.NaN())},
		{exact32(1.0), exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.a.Dim(tt.b)
		if !eq32(got, tt.want) {
			t.Errorf("Float32.Dim(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
