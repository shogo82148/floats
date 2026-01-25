package floats

import (
	"math"
	"testing"
)

func TestFloat64_Sin(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(1), math.Sin(1)},
		{exact64(2), math.Sin(2)},
		{exact64(3), math.Sin(3)},
	}

	for _, tt := range tests {
		got := tt.x.Sin()
		if !close64(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(0), exact64(0)},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},
		{exact64(math.Inf(1)), exact64(math.NaN())},
		{exact64(math.Inf(-1)), exact64(math.NaN())},
		{exact64(math.NaN()), exact64(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Sin()
		if !eq64(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat64_Cos(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(1), math.Cos(1)},
		{exact64(2), math.Cos(2)},
		{exact64(3), math.Cos(3)},
	}

	for _, tt := range tests {
		got := tt.x.Cos()
		if !close64(got, tt.want) {
			t.Errorf("Cos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(math.Inf(1)), exact64(math.NaN())},
		{exact64(math.Inf(-1)), exact64(math.NaN())},
		{exact64(math.NaN()), exact64(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Cos()
		if !eq64(got, tt.want) {
			t.Errorf("Cos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat64_Sincos(t *testing.T) {
	tests := []struct {
		x   Float64
		sin float64
		cos float64
	}{
		{exact64(1), math.Sin(1), math.Cos(1)},
		{exact64(2), math.Sin(2), math.Cos(2)},
		{exact64(3), math.Sin(3), math.Cos(3)},
	}

	for _, tt := range tests {
		sin, cos := tt.x.Sincos()
		if !close64(sin, tt.sin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", tt.x, sin, tt.sin)
		}
		if !close64(cos, tt.cos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", tt.x, cos, tt.cos)
		}
	}

	strictTests := []struct {
		x   Float64
		sin Float64
		cos Float64
	}{
		// special cases
		{exact64(0), exact64(0), exact64(1)},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1)), exact64(1)},
		{exact64(math.Inf(1)), exact64(math.NaN()), exact64(math.NaN())},
		{exact64(math.Inf(-1)), exact64(math.NaN()), exact64(math.NaN())},
		{exact64(math.NaN()), exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		sin, cos := tt.x.Sincos()
		if !eq64(sin, tt.sin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", tt.x, sin, tt.sin)
		}
		if !eq64(cos, tt.cos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", tt.x, cos, tt.cos)
		}
	}
}
