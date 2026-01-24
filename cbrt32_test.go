package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat32_Cbrt(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(27), 3},
		{exact32(8), 2},
		{exact32(2), math.Cbrt(2)},
		{exact32(1), 1},
		{exact32(0x1p-127), math.Cbrt(0x1p-127)},
		{exact32(0), 0},
		{exact32(-0x1p-127), math.Cbrt(-0x1p-127)},
		{exact32(-1), -1},
		{exact32(-2), math.Cbrt(-2)},
		{exact32(-8), -2},
		{exact32(-27), -3},
	}

	for _, test := range tests {
		got := test.x.Cbrt()
		if !close32(got, test.want) {
			t.Errorf("Cbrt(%v) = %v; want %v", test.x, got, test.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// Special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(math.Inf(-1))},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Cbrt()
		if !eq32(got, tt.want) {
			t.Errorf("Float32.Cbrt(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Cbrt(b *testing.B) {
	x := NewFloat32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Cbrt())
	}
}
