package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat32_Log1p(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-0.5), math.Log1p(-0.5)},
		{exact32(-0.25), math.Log1p(-0.25)},
		{exact32(0), math.Log1p(0)},
		{exact32(0x1p-25), math.Log1p(0x1p-25)},
		{exact32(0.25), math.Log1p(0.25)},
		{exact32(0.5), math.Log1p(0.5)},
		{exact32(1), math.Log1p(1)},
		{exact32(2), math.Log1p(2)},
		{exact32(3), math.Log1p(3)},
		{exact32(4), math.Log1p(4)},
		{exact32(11), math.Log1p(11)},
		{exact32(1 << 32), math.Log1p(1 << 32)},
	}

	for _, tt := range tests {
		got := tt.x.Log1p()
		if !close32(got, tt.want) {
			t.Errorf("Log1p(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(-1), exact32(math.Inf(-1))},
		{exact32(-2), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log1p()
		if !eq32(got, tt.want) {
			t.Errorf("Log1p(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Log1p(b *testing.B) {
	x := NewFloat32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log1p())
	}
}
