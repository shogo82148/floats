package floats

import (
	"math"
	"testing"
)

func TestFloat64_Expm1(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(0), math.Expm1(0)},
		{exact64(0x1p-24), math.Expm1(0x1p-24)},
		{exact64(1), math.Expm1(1)},
	}

	for _, tt := range tests {
		got := tt.x.Expm1()
		if !close64(got, tt.want) {
			t.Errorf("Expm1(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		{exact64(0), exact64(0)},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},

		// special cases
		{exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(math.Inf(-1)), exact64(-1)},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Expm1()
		if !eq64(got, tt.want) {
			t.Errorf("Expm1(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
