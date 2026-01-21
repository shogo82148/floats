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

func TestFloat16_Cosh(t *testing.T) {
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
