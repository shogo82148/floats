package floats

import (
	"math"
	"testing"
)

func TestNewFloat256Pow10(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "1"},
		{1, "10"},
		{2, "100"},
		{3, "1000"},
		{4, "10000"},
		{78913, "1e78913"},
		{-1, "0.1"},
		{-2, "0.01"},
		{-3, "0.001"},
		{-4, "0.0001"},
		{-78913, "1e-78913"},
	}

	for _, test := range tests {
		got := NewFloat256Pow10(test.n)
		if !close256(got, test.want) {
			t.Errorf("NewFloat256Pow10(%d) = %v; want %v", test.n, got, test.want)
		}
	}

	strictTests := []struct {
		n    int
		want Float256
	}{
		{78914, exact256(math.Inf(1))}, // overflow
		{-78985, exact256(0)},          // underflow
	}

	for _, tt := range strictTests {
		got := NewFloat256Pow10(tt.n)
		if !eq256(got, tt.want) {
			t.Errorf("NewFloat256Pow10(%d) = %v; want %v", tt.n, got, tt.want)
		}
	}
}
