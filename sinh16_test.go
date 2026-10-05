package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Sinh(t *testing.T) {
	t.Parallel()
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

func TestFloat16_Cosh(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Cosh(0)},
		{exact16(1), math.Cosh(1)},
		{exact16(2), math.Cosh(2)},
		{exact16(3), math.Cosh(3)},
		{exact16(4), math.Cosh(4)},
		{exact16(11), math.Cosh(11)},
	}

	for _, tt := range tests {
		got := tt.x.Cosh()
		if !close16(got, tt.want) {
			t.Errorf("Cosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(0), exact16(1)},
		{exact16(math.Copysign(0, -1)), exact16(1)},
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(math.Inf(1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Cosh()
		if !eq16(got, tt.want) {
			t.Errorf("Cosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Cosh(b *testing.B) {
	x := exact16(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Cosh())
	}
}

func TestFloat16_Tanh(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Tanh(0)},
		{exact16(1), math.Tanh(1)},
		{exact16(2), math.Tanh(2)},
		{exact16(3), math.Tanh(3)},
		{exact16(4), math.Tanh(4)},
		{exact16(11), math.Tanh(11)},
	}

	for _, tt := range tests {
		got := tt.x.Tanh()
		if !close16(got, tt.want) {
			t.Errorf("Tanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(1)},
		{exact16(math.Inf(-1)), exact16(-1)},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Tanh()
		if !eq16(got, tt.want) {
			t.Errorf("Tanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Tanh(b *testing.B) {
	x := exact16(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Tanh())
	}
}

// TestFloat16_SinhCoshTanhAll checks Sinh, Cosh, and Tanh for all finite Float16 values.
// For every finite Float16 value, the exact sinh, cosh, and tanh are at least 2**-15 ulp away
// from the midpoint of two adjacent Float16 values (checked with mpmath),
// so math.Sinh, math.Cosh, and math.Tanh rounded to Float16 are correctly rounded.
func TestFloat16_SinhCoshTanhAll(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		x := NewFloat16FromBits(uint16(i))
		if x.IsNaN() || x.IsInf(0) {
			continue
		}
		f := x.Float64().BuiltIn()
		if got, want := x.Sinh(), NewFloat16(math.Sinh(f)); !eq16(got, want) {
			t.Errorf("Sinh(%v) = %v; want %v", x, got, want)
		}
		if got, want := x.Cosh(), NewFloat16(math.Cosh(f)); !eq16(got, want) {
			t.Errorf("Cosh(%v) = %v; want %v", x, got, want)
		}
		if got, want := x.Tanh(), NewFloat16(math.Tanh(f)); !eq16(got, want) {
			t.Errorf("Tanh(%v) = %v; want %v", x, got, want)
		}
	}
}
