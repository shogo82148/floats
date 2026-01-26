package floats

import (
	"math"
	"testing"
)

func TestFloat32_Asin(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-1), math.Asin(-1)},
		{exact32(-0.75), math.Asin(-0.75)},
		{exact32(-0.5), math.Asin(-0.5)},
		{exact32(-0.25), math.Asin(-0.25)},
		{exact32(0.25), math.Asin(0.25)},
		{exact32(0.5), math.Asin(0.5)},
		{exact32(0.75), math.Asin(0.75)},
		{exact32(1), math.Asin(1)},
	}

	for _, tt := range tests {
		got := tt.x.Asin()
		if !close32(got, tt.want) {
			t.Errorf("Asin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.NaN()), exact32(math.NaN())},
		{exact32(2), exact32(math.NaN())},
		{exact32(-2), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Asin()
		if !eq32(got, tt.want) {
			t.Errorf("Asin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat32_Atan(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-0.5), math.Atan(-0.5)},
		{exact32(-0.25), math.Atan(-0.25)},
		{exact32(-0.125), math.Atan(-0.125)},
		{exact32(0.125), math.Atan(0.125)},
		{exact32(0.25), math.Atan(0.25)},
		{exact32(0.5), math.Atan(0.5)},
		{exact32(0.75), math.Atan(0.75)},
		{exact32(1), math.Atan(1)},
		{exact32(2), math.Atan(2)},
		{exact32(math.Inf(-1)), math.Atan(math.Inf(-1))},
		{exact32(math.Inf(1)), math.Atan(math.Inf(1))},
	}

	for _, tt := range tests {
		got := tt.x.Atan()
		if !close32(got, tt.want) {
			t.Errorf("Atan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Atan()
		if !eq32(got, tt.want) {
			t.Errorf("Atan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
