package floats

import (
	"math"
	"testing"
)

func TestFloat16_Gamma(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(-0.5), math.Gamma(-0.5)},
		{exact16(0.5), math.Gamma(0.5)},
		{exact16(1), math.Gamma(1)},
		{exact16(1.5), math.Gamma(1.5)},
		{exact16(2), math.Gamma(2)},
		{exact16(2.5), math.Gamma(2.5)},
		{exact16(3), math.Gamma(3)},
	}

	for _, tt := range tests {
		got := tt.x.Gamma()
		if !close16(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(0), exact16(math.Inf(1))},
		{exact16(math.Copysign(0, -1)), exact16(math.Inf(-1))},
		{exact16(-1), exact16(math.NaN())},
		{exact16(-2), exact16(math.NaN())},
		{exact16(math.Inf(-1)), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Gamma()
		if !eq16(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
