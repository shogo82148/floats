package floats

import (
	"math"
	"testing"
)

func TestFloat32_Sin(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-3), math.Sin(-3)},
		{exact32(-2), math.Sin(-2)},
		{exact32(-1), math.Sin(-1)},
		{exact32(1), math.Sin(1)},
		{exact32(2), math.Sin(2)},
		{exact32(3), math.Sin(3)},
		{exact32(0x1p28), math.Sin(0x1p28)},
	}

	for _, tt := range tests {
		got := tt.x.Sin()
		if !close32(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.Inf(1)), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Sin()
		if !eq32(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat32_Cos(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(1), math.Cos(1)},
		{exact32(2), math.Cos(2)},
		{exact32(3), math.Cos(3)},
	}

	for _, tt := range tests {
		got := tt.x.Cos()
		if !close32(got, tt.want) {
			t.Errorf("Cos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Cos()
		if !eq32(got, tt.want) {
			t.Errorf("Cos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat32_Sincos(t *testing.T) {
	tests := []struct {
		x   Float32
		sin float64
		cos float64
	}{
		{exact32(1), math.Sin(1), math.Cos(1)},
		{exact32(2), math.Sin(2), math.Cos(2)},
		{exact32(3), math.Sin(3), math.Cos(3)},
	}

	for _, tt := range tests {
		sin, cos := tt.x.Sincos()
		if !close32(sin, tt.sin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", tt.x, sin, tt.sin)
		}
		if !close32(cos, tt.cos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", tt.x, cos, tt.cos)
		}
	}

	strictTests := []struct {
		x   Float32
		sin Float32
		cos Float32
	}{
		// special cases
		{exact32(0), exact32(0), exact32(1)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1)), exact32(1)},
		{exact32(math.Inf(1)), exact32(math.NaN()), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN()), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		sin, cos := tt.x.Sincos()
		if !eq32(sin, tt.sin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", tt.x, sin, tt.sin)
		}
		if !eq32(cos, tt.cos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", tt.x, cos, tt.cos)
		}
	}
}

func TestFloat32_Tan(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(1), math.Tan(1)},
		{exact32(2), math.Tan(2)},
		{exact32(3), math.Tan(3)},
	}

	for _, tt := range tests {
		got := tt.x.Tan()
		if !close32(got, tt.want) {
			t.Errorf("Tan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.Inf(1)), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Tan()
		if !eq32(got, tt.want) {
			t.Errorf("Tan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
