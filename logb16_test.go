package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Logb(t *testing.T) {
	tests := []struct {
		x    Float16
		want Float16
	}{
		{exact16(1), exact16(0)},
		{exact16(2), exact16(1)},
		{exact16(0.5), exact16(-1)},
		{exact16(4), exact16(2)},
		{exact16(0.25), exact16(-2)},
		{exact16(0x1p-24), exact16(-24)}, // smallest positive subnormal number

		// special values
		{exact16(0), exact16(math.Inf(-1))},
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(math.Inf(1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Logb()
		if !eq16(got, tt.want) {
			t.Errorf("Logb(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Logb(b *testing.B) {
	x := exact16(1)
	for b.Loop() {
		runtime.KeepAlive(x.Logb())
	}
}

func TestFloat16_Ilogb(t *testing.T) {
	tests := []struct {
		x    Float16
		want int
	}{
		{exact16(1), 0},
		{exact16(2), 1},
		{exact16(0.5), -1},
		{exact16(4), 2},
		{exact16(0.25), -2},
		{exact16(0x1p-24), -24}, // smallest positive subnormal number

		// special values
		{exact16(0), math.MinInt32},
		{exact16(math.Inf(1)), math.MaxInt32},
		{exact16(math.Inf(-1)), math.MaxInt32},
		{exact16(math.NaN()), math.MaxInt32},
	}

	for _, tt := range tests {
		got := tt.x.Ilogb()
		if got != tt.want {
			t.Errorf("Ilogb(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Ilogb(b *testing.B) {
	x := exact16(1)
	for b.Loop() {
		runtime.KeepAlive(x.Ilogb())
	}
}
