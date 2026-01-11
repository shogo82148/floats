package floats

import (
	"math"
	"testing"
)

func TestFloat64_Exp(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(0), math.Exp(0)},
		{exact64(1), math.Exp(1)},
		{exact64(2), math.Exp(2)},
		{exact64(3), math.Exp(3)},
		{exact64(4), math.Exp(4)},
		{exact64(11), math.Exp(11)},
	}

	for _, tt := range tests {
		got := tt.x.Exp()
		if !close64(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Exp()
		if !eq64(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
