package floats

import (
	"math"
	"testing"
)

func TestNewFloat64Pow10(t *testing.T) {
	tests := []struct {
		n    int
		want float64
	}{
		{0, 1},
		{1, 10},
		{2, 100},
		{3, 1000},
		{4, 10000},
		{308, 1e308},
		{-1, 0.1},
		{-2, 0.01},
		{-3, 0.001},
		{-4, 0.0001},
	}

	for _, test := range tests {
		got := NewFloat64Pow10(test.n)
		if !close64(got, test.want) {
			t.Errorf("NewFloat64Pow10(%d) = %v; want %v", test.n, got, test.want)
		}
	}

	strictTests := []struct {
		n    int
		want Float64
	}{
		{309, exact64(math.Inf(1))}, // overflow
		{-324, exact64(0)},          // underflow
	}

	for _, tt := range strictTests {
		got := NewFloat64Pow10(tt.n)
		if !eq64(got, tt.want) {
			t.Errorf("NewFloat64Pow10(%d) = %v; want %v", tt.n, got, tt.want)
		}
	}
}
