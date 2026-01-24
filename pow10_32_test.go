package floats

import (
	"math"
	"testing"
)

func TestNewFloat32Pow10(t *testing.T) {
	tests := []struct {
		n    int
		want float64
	}{
		{0, 1},
		{1, 10},
		{2, 100},
		{3, 1000},
		{4, 10000},
		{38, 1e38},
		{-1, 0.1},
		{-2, 0.01},
		{-3, 0.001},
		{-4, 0.0001},
	}

	for _, test := range tests {
		got := NewFloat32Pow10(test.n)
		if !close32(got, test.want) {
			t.Errorf("NewFloat32Pow10(%d) = %v; want %v", test.n, got, test.want)
		}
	}

	strictTests := []struct {
		n    int
		want Float32
	}{
		{39, exact32(math.Inf(1))}, // overflow
		{-46, exact32(0)},          // underflow
	}

	for _, tt := range strictTests {
		got := NewFloat32Pow10(tt.n)
		if !eq32(got, tt.want) {
			t.Errorf("NewFloat32Pow10(%d) = %v; want %v", tt.n, got, tt.want)
		}
	}
}
