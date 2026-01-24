package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Log1p(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(1), math.Log1p(1)},
		{exact16(2), math.Log1p(2)},
		{exact16(3), math.Log1p(3)},
		{exact16(4), math.Log1p(4)},
		{exact16(11), math.Log1p(11)},
	}

	for _, tt := range tests {
		got := tt.x.Log1p()
		if !close16(got, tt.want) {
			t.Errorf("Log1p(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(-1), exact16(math.Inf(-1))},
		{exact16(-2), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log1p()
		if !eq16(got, tt.want) {
			t.Errorf("Log1p(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Log1p(b *testing.B) {
	x := NewFloat16(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log1p())
	}
}
