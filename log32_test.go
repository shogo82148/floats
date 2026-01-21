package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat32_Log(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(1), math.Log(1)},
		{exact32(2), math.Log(2)},
		{exact32(3), math.Log(3)},
		{exact32(4), math.Log(4)},
		{exact32(11), math.Log(11)},
	}

	for _, tt := range tests {
		got := tt.x.Log()
		if !close32(got, tt.want) {
			t.Errorf("Log(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(0), exact32(math.Inf(-1))},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(-1))},
		{exact32(-1), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log()
		if !eq32(got, tt.want) {
			t.Errorf("Log(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Log(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log())
	}
}
