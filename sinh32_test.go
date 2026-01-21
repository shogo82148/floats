package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat32_Sinh(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0), math.Sinh(0)},
		{exact32(0.25), math.Sinh(0.25)},
		{exact32(0.5), math.Sinh(0.5)},
		{exact32(1), math.Sinh(1)},
		{exact32(21), math.Sinh(21)},
		{exact32(22), math.Sinh(22)},

		{exact32(-0), -math.Sinh(0)},
		{exact32(-0.25), -math.Sinh(0.25)},
		{exact32(-0.5), -math.Sinh(0.5)},
		{exact32(-1), -math.Sinh(1)},
		{exact32(-21), -math.Sinh(21)},
		{exact32(-22), -math.Sinh(22)},
	}

	for _, tt := range tests {
		got := tt.x.Sinh()
		if !close32(got, tt.want) {
			t.Errorf("Sinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},

		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(math.Inf(-1))},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Sinh()
		if !eq32(got, tt.want) {
			t.Errorf("Sinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Sinh(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Sinh())
	}
}

func TestFloat32_Cosh(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0), math.Cosh(0)},
		{exact32(1), math.Cosh(1)},
		{exact32(2), math.Cosh(2)},
		{exact32(3), math.Cosh(3)},
		{exact32(4), math.Cosh(4)},
		{exact32(22), math.Cosh(22)},
	}

	for _, tt := range tests {
		got := tt.x.Cosh()
		if !close32(got, tt.want) {
			t.Errorf("Cosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(1)},
		{exact32(math.Copysign(0, -1)), exact32(1)},
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(math.Inf(1))},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Cosh()
		if !eq32(got, tt.want) {
			t.Errorf("Cosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Cosh(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Cosh())
	}
}

func TestFloat32_Tanh(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0), math.Tanh(0)},
		{exact32(1), math.Tanh(1)},
		{exact32(2), math.Tanh(2)},
		{exact32(3), math.Tanh(3)},
		{exact32(4), math.Tanh(4)},
		{exact32(45), math.Tanh(45)},
	}

	for _, tt := range tests {
		got := tt.x.Tanh()
		if !close32(got, tt.want) {
			t.Errorf("Tanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.Inf(1)), exact32(1)},
		{exact32(math.Inf(-1)), exact32(-1)},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Tanh()
		if !eq32(got, tt.want) {
			t.Errorf("Tanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Tanh(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Tanh())
	}
}
