package floats

import "testing"

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
}
