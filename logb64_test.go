package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat64_Logb(t *testing.T) {
	tests := []struct {
		x    Float64
		want Float64
	}{
		{exact64(1), exact64(0)},
		{exact64(2), exact64(1)},
		{exact64(0.5), exact64(-1)},
		{exact64(4), exact64(2)},
		{exact64(0.25), exact64(-2)},
		{
			// smallest positive subnormal number
			exact64(0x1p-1022), exact64(-1022),
		},

		// special values
		{exact64(0), exact64(math.Inf(-1))},
		{exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(math.Inf(-1)), exact64(math.Inf(1))},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Logb()
		if !eq64(got, tt.want) {
			t.Errorf("Logb(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Logb(b *testing.B) {
	x := exact64(1)
	for b.Loop() {
		runtime.KeepAlive(x.Logb())
	}
}

func TestFloat64_Ilogb(t *testing.T) {
	tests := []struct {
		x    Float64
		want int
	}{
		{exact64(1), 0},
		{exact64(2), 1},
		{exact64(0.5), -1},
		{exact64(4), 2},
		{exact64(0.25), -2},
		{
			// smallest positive subnormal number
			exact64(0x1p-1022), -1022,
		},

		// special values
		{exact64(0), math.MinInt32},
		{exact64(math.Inf(1)), math.MaxInt32},
		{exact64(math.Inf(-1)), math.MaxInt32},
		{exact64(math.NaN()), math.MaxInt32},
	}

	for _, tt := range tests {
		got := tt.x.Ilogb()
		if got != tt.want {
			t.Errorf("Ilogb(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Ilogb(b *testing.B) {
	x := exact64(1)
	for b.Loop() {
		runtime.KeepAlive(x.Ilogb())
	}
}
