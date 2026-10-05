package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Expm1(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Expm1(0)},
		{exact16(0x1p-24), math.Expm1(0x1p-24)},
		{exact16(1), math.Expm1(1)},
	}

	for _, tt := range tests {
		got := tt.x.Expm1()
		if !close16(got, tt.want) {
			t.Errorf("Expm1(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},

		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(-1)},
		{exact16(math.NaN()), exact16(math.NaN())},
		{NewFloat16FromBits(0xfe00), exact16(math.NaN())}, // negative NaN

		// tiny arguments
		{NewFloat16FromBits(0x0001), NewFloat16FromBits(0x0001)},
		{NewFloat16FromBits(0x8001), NewFloat16FromBits(0x8001)},

		// overflow
		{exact16(12), exact16(math.Inf(1))},

		// e**x - 1 rounds to -1.
		{exact16(-16), exact16(-1)},
	}

	for _, tt := range strictTests {
		got := tt.x.Expm1()
		if !eq16(got, tt.want) {
			t.Errorf("Expm1(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Expm1(b *testing.B) {
	x := exact16(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Expm1())
	}
}
