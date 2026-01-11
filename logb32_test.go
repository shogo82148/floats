package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat32_Logb(t *testing.T) {
	tests := []struct {
		x    Float32
		want Float32
	}{
		{exact32(1), exact32(0)},
		{exact32(2), exact32(1)},
		{exact32(0.5), exact32(-1)},
		{exact32(4), exact32(2)},
		{exact32(0.25), exact32(-2)},

		// special values
		{exact32(0), exact32(math.Inf(-1))},
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Logb()
		if !eq32(got, tt.want) {
			t.Errorf("Logb(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Logb(b *testing.B) {
	x := exact32(1)
	for b.Loop() {
		runtime.KeepAlive(x.Logb())
	}
}

func TestFloat32_Ilogb(t *testing.T) {
	tests := []struct {
		x    Float32
		want int
	}{
		{exact32(1), 0},
		{exact32(2), 1},
		{exact32(0.5), -1},
		{exact32(4), 2},
		{exact32(0.25), -2},

		// special values
		{exact32(0), math.MinInt32},
		{exact32(math.Inf(1)), math.MaxInt32},
		{exact32(math.NaN()), math.MaxInt32},
	}

	for _, tt := range tests {
		got := tt.x.Ilogb()
		if got != tt.want {
			t.Errorf("Ilogb(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Ilogb(b *testing.B) {
	x := exact32(1)
	for b.Loop() {
		runtime.KeepAlive(x.Ilogb())
	}
}
