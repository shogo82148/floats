package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Sinh(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Sinh(0)},
		{exact16(0.5), math.Sinh(0.5)},
		{exact16(1), math.Sinh(1)},

		{exact16(-0), -math.Sinh(0)},
		{exact16(-0.5), -math.Sinh(0.5)},
		{exact16(-1), -math.Sinh(1)},
	}

	for _, tt := range tests {
		got := tt.x.Sinh()
		if !close16(got, tt.want) {
			t.Errorf("Sinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(22), exact16(math.Inf(1))},

		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(math.Inf(-1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Sinh()
		if !eq16(got, tt.want) {
			t.Errorf("Sinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Sinh(b *testing.B) {
	x := exact16(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Sinh())
	}
}
