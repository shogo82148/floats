package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat128_Logb(t *testing.T) {
	tests := []struct {
		x    Float128
		want Float128
	}{
		{exact128(1), exact128(0)},
		{exact128(2), exact128(1)},
		{exact128(0.5), exact128(-1)},
		{exact128(4), exact128(2)},
		{exact128(0.25), exact128(-2)},

		// special values
		{exact128(0), exact128(math.Inf(-1))},
		{exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Logb()
		if !eq128(got, tt.want) {
			t.Errorf("Logb(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Logb(b *testing.B) {
	x := exact128(1)
	for b.Loop() {
		runtime.KeepAlive(x.Logb())
	}
}

func TestFloat128_Ilogb(t *testing.T) {
	tests := []struct {
		x    Float128
		want int
	}{
		{exact128(1), 0},
		{exact128(2), 1},
		{exact128(0.5), -1},
		{exact128(4), 2},
		{exact128(0.25), -2},

		// special values
		{exact128(0), math.MinInt32},
		{exact128(math.Inf(1)), math.MaxInt32},
		{exact128(math.NaN()), math.MaxInt32},
	}

	for _, tt := range tests {
		got := tt.x.Ilogb()
		if got != tt.want {
			t.Errorf("Ilogb(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Ilogb(b *testing.B) {
	x := exact128(1)
	for b.Loop() {
		runtime.KeepAlive(x.Ilogb())
	}
}
