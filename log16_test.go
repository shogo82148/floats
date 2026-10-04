package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Log(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(1), math.Log(1)},
		{exact16(2), math.Log(2)},
		{exact16(3), math.Log(3)},
		{exact16(4), math.Log(4)},
		{exact16(11), math.Log(11)},
	}

	for _, tt := range tests {
		got := tt.x.Log()
		if !close16(got, tt.want) {
			t.Errorf("Log(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(0), exact16(math.Inf(-1))},
		{exact16(math.Copysign(0, -1)), exact16(math.Inf(-1))},
		{exact16(-1), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log()
		if !eq16(got, tt.want) {
			t.Errorf("Log(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Log(b *testing.B) {
	x := exact16(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log())
	}
}

func TestFloat16_Log10(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(1), math.Log10(1)},
		{exact16(2), math.Log10(2)},
		{exact16(3), math.Log10(3)},
		{exact16(4), math.Log10(4)},
		{exact16(11), math.Log10(11)},
	}

	for _, tt := range tests {
		got := tt.x.Log10()
		if !close16(got, tt.want) {
			t.Errorf("Log10(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(0), exact16(math.Inf(-1))},
		{exact16(math.Copysign(0, -1)), exact16(math.Inf(-1))},
		{exact16(-1), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log10()
		if !eq16(got, tt.want) {
			t.Errorf("Log10(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Log10(b *testing.B) {
	x := exact16(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log10())
	}
}

func TestFloat16_Log2(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(1), math.Log2(1)},
		{exact16(2), math.Log2(2)},
		{exact16(3), math.Log2(3)},
		{exact16(4), math.Log2(4)},
		{exact16(11), math.Log2(11)},
	}

	for _, tt := range tests {
		got := tt.x.Log2()
		if !close16(got, tt.want) {
			t.Errorf("Log2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(0), exact16(math.Inf(-1))},
		{exact16(math.Copysign(0, -1)), exact16(math.Inf(-1))},
		{exact16(-1), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log2()
		if !eq16(got, tt.want) {
			t.Errorf("Log2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Log2(b *testing.B) {
	x := exact16(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log2())
	}
}

// TestFloat16_LogAll checks Log, Log10, and Log2 for all Float16 values.
// For every positive finite Float16 value, the exact log, log10, and log2 are
// either exactly representable or at least 2**-14 ulp away
// from the midpoint of two adjacent Float16 values (checked with mpmath),
// so the results of the math package rounded to Float16 are correctly rounded.
func TestFloat16_LogAll(t *testing.T) {
	for i := range 1 << 16 {
		x := NewFloat16FromBits(uint16(i))
		f := x.Float64().BuiltIn()
		if got, want := x.Log(), NewFloat16(math.Log(f)); !eq16(got, want) {
			t.Errorf("Log(%v) = %v; want %v", x, got, want)
		}
		if got, want := x.Log10(), NewFloat16(math.Log10(f)); !eq16(got, want) {
			t.Errorf("Log10(%v) = %v; want %v", x, got, want)
		}
		if got, want := x.Log2(), NewFloat16(math.Log2(f)); !eq16(got, want) {
			t.Errorf("Log2(%v) = %v; want %v", x, got, want)
		}
	}
}
