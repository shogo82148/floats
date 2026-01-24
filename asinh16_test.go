package floats

import (
	"math"
	"testing"
)

func TestFloat16_Asinh(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Asinh(0)},
		{exact16(0.25), math.Asinh(0.25)},
		{exact16(0.5), math.Asinh(0.5)},
		{exact16(1), math.Asinh(1)},
		{exact16(21), math.Asinh(21)},
		{exact16(22), math.Asinh(22)},

		{exact16(-0), -math.Asinh(0)},
		{exact16(-0.25), -math.Asinh(0.25)},
		{exact16(-0.5), -math.Asinh(0.5)},
		{exact16(-1), -math.Asinh(1)},
		{exact16(-21), -math.Asinh(21)},
		{exact16(-22), -math.Asinh(22)},
	}

	for _, tt := range tests {
		got := tt.x.Asinh()
		if !close16(got, tt.want) {
			t.Errorf("Asinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},

		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(math.Inf(-1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Asinh()
		if !eq16(got, tt.want) {
			t.Errorf("Asinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
