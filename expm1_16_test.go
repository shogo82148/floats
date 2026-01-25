package floats

import (
	"math"
	"testing"
)

func TestFloat16_Expm1(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Expm1(0)},
		{exact16(0x1p-24), math.Expm1(0x1p-24)},
		{exact16(1), math.Expm1(1)},
	}

	for _, tt := range tests {
		got := tt.x.Expm1()
		if !close16(got, tt.want) {
			t.Errorf("Expm1(%v) = %v; want %v", tt.x, got, tt.want)
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
		{exact16(math.Inf(-1)), exact16(-1)},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Expm1()
		if !eq16(got, tt.want) {
			t.Errorf("Expm1(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
