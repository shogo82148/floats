package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Log1p(t *testing.T) {
	t.Parallel()
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

// TestFloat16_Log1pAll checks Log1p for all Float16 values.
// For every Float16 value greater than -1, the exact log1p is
// either exactly representable or at least 2**-16 ulp away
// from the midpoint of two adjacent Float16 values (checked with mpmath),
// so math.Log1p rounded to Float16 is correctly rounded.
func TestFloat16_Log1pAll(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		x := NewFloat16FromBits(uint16(i))
		want := NewFloat16(math.Log1p(x.Float64().BuiltIn()))
		if got := x.Log1p(); !eq16(got, want) {
			t.Errorf("Log1p(%v) = %v; want %v", x, got, want)
		}
	}
}
