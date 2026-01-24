package floats

import (
	"math"
	"testing"
)

func TestFloat16_Cbrt(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(27), 3},
		{exact16(8), 2},
		{exact16(1), 1},
		{exact16(0), 0},
		{exact16(-1), -1},
		{exact16(-8), -2},
		{exact16(-27), -3},
	}

	for _, test := range tests {
		got := test.x.Cbrt()
		if !close16(got, test.want) {
			t.Errorf("Cbrt(%v) = %v; want %v", test.x, got, test.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// Special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(math.Inf(-1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Cbrt()
		if !eq16(got, tt.want) {
			t.Errorf("Float16.Cbrt(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
