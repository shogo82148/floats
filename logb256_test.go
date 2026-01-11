package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat256_Logb(t *testing.T) {
	tests := []struct {
		x    Float256
		want Float256
	}{
		{exact256(1), exact256(0)},
		{exact256(2), exact256(1)},
		{exact256(0.5), exact256(-1)},
		{exact256(4), exact256(2)},
		{exact256(0.25), exact256(-2)},

		// special values
		{exact256(0), exact256(math.Inf(-1))},
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Logb()
		if !eq256(got, tt.want) {
			t.Errorf("Logb(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Logb(b *testing.B) {
	x := exact256(1)
	for b.Loop() {
		runtime.KeepAlive(x.Logb())
	}
}

func TestFloat256_Ilogb(t *testing.T) {
	tests := []struct {
		x    Float256
		want int
	}{
		{exact256(1), 0},
		{exact256(2), 1},
		{exact256(0.5), -1},
		{exact256(4), 2},
		{exact256(0.25), -2},

		// special values
		{exact256(0), math.MinInt32},
		{exact256(math.Inf(1)), math.MaxInt32},
		{exact256(math.NaN()), math.MaxInt32},
	}

	for _, tt := range tests {
		got := tt.x.Ilogb()
		if got != tt.want {
			t.Errorf("Ilogb(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Ilogb(b *testing.B) {
	x := exact256(1)
	for b.Loop() {
		runtime.KeepAlive(x.Ilogb())
	}
}
