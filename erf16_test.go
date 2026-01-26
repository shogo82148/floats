package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Erf(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Erf(0)},
		{exact16(0x1p-14), math.Erf(0x1p-14)},
		{exact16(1), math.Erf(1)},
		{exact16(2), math.Erf(2)},
		{exact16(3), math.Erf(3)},
		{exact16(4), math.Erf(4)},
		{exact16(11), math.Erf(11)},
	}

	for _, tt := range tests {
		got := tt.x.Erf()
		if !close16(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(1)},
		{exact16(math.Inf(-1)), exact16(-1)},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erf()
		if !eq16(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Erf(b *testing.B) {
	x := exact16(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erf())
	}
}
