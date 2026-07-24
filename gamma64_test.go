package floats

import (
	"math"
	"testing"
)

func TestFloat64_Gamma(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(-0.5), math.Gamma(-0.5)},
		{exact64(0.5), math.Gamma(0.5)},
		{exact64(1), math.Gamma(1)},
		{exact64(1.5), math.Gamma(1.5)},
		{exact64(2), math.Gamma(2)},
		{exact64(2.5), math.Gamma(2.5)},
		{exact64(3), math.Gamma(3)},
	}

	for _, tt := range tests {
		got := tt.x.Gamma()
		if !close64(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(0), exact64(math.Inf(1))},
		{exact64(math.Copysign(0, -1)), exact64(math.Inf(-1))},
		{exact64(-1), exact64(math.NaN())},
		{exact64(-2), exact64(math.NaN())},
		{exact64(math.Inf(-1)), exact64(math.NaN())},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Gamma()
		if !eq64(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
