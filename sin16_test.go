package floats

import (
	"math"
	"testing"
)

func TestFloat16_Sin(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(1), math.Sin(1)},
		{exact16(2), math.Sin(2)},
		{exact16(3), math.Sin(3)},
	}

	for _, tt := range tests {
		got := tt.x.Sin()
		if !close16(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(math.NaN())},
		{exact16(math.Inf(-1)), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Sin()
		if !eq16(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat16_Cos(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(1), math.Cos(1)},
		{exact16(2), math.Cos(2)},
		{exact16(3), math.Cos(3)},
	}

	for _, tt := range tests {
		got := tt.x.Cos()
		if !close16(got, tt.want) {
			t.Errorf("Cos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(math.NaN())},
		{exact16(math.Inf(-1)), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Cos()
		if !eq16(got, tt.want) {
			t.Errorf("Cos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat16_Sincos(t *testing.T) {
	tests := []struct {
		x   Float16
		sin float64
		cos float64
	}{
		{exact16(1), math.Sin(1), math.Cos(1)},
		{exact16(2), math.Sin(2), math.Cos(2)},
		{exact16(3), math.Sin(3), math.Cos(3)},
	}

	for _, tt := range tests {
		sin, cos := tt.x.Sincos()
		if !close16(sin, tt.sin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", tt.x, sin, tt.sin)
		}
		if !close16(cos, tt.cos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", tt.x, cos, tt.cos)
		}
	}

	strictTests := []struct {
		x   Float16
		sin Float16
		cos Float16
	}{
		// special cases
		{exact16(0), exact16(0), exact16(1)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1)), exact16(1)},
		{exact16(math.Inf(1)), exact16(math.NaN()), exact16(math.NaN())},
		{exact16(math.Inf(-1)), exact16(math.NaN()), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		sin, cos := tt.x.Sincos()
		if !eq16(sin, tt.sin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", tt.x, sin, tt.sin)
		}
		if !eq16(cos, tt.cos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", tt.x, cos, tt.cos)
		}
	}
}
