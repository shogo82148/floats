package floats

import (
	"math"
	"testing"
)

func TestFloat16_Floor(t *testing.T) {
	tests := []struct {
		x    Float16
		want Float16
	}{
		{exact16(3.0), exact16(3)},
		{exact16(3.5), exact16(3)},
		{exact16(-3.0), exact16(-3)},
		{exact16(-3.5), exact16(-4)},

		// Special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(math.Inf(-1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Floor()
		if !eq16(got, tt.want) {
			t.Errorf("Float16.Floor(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
