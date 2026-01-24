package floats

import (
	"math"
	"testing"
)

func TestNewFloat128Pow10(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "1"},
		{1, "10"},
		{2, "100"},
		{3, "1000"},
		{4, "10000"},
		{4932, "1e4932"},
		{-1, "0.1"},
		{-2, "0.01"},
		{-3, "0.001"},
		{-4, "0.0001"},
		{-4932, "1e-4932"},
	}

	for _, test := range tests {
		got := NewFloat128Pow10(test.n)
		if !close128(got, test.want) {
			t.Errorf("NewFloat128Pow10(%d) = %v; want %v", test.n, got, test.want)
		}
	}

	strictTests := []struct {
		n    int
		want Float128
	}{
		{4933, exact128(math.Inf(1))}, // overflow
		{-4966, exact128(0)},          // underflow
	}

	for _, tt := range strictTests {
		got := NewFloat128Pow10(tt.n)
		if !eq128(got, tt.want) {
			t.Errorf("NewFloat128Pow10(%d) = %v; want %v", tt.n, got, tt.want)
		}
	}
}
