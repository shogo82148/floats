package floats

import (
	"math"
	"testing"
)

func TestFloat32_Floor(t *testing.T) {
	tests := []struct {
		x    Float32
		want Float32
	}{
		{exact32(3.0), exact32(3)},
		{exact32(3.5), exact32(3)},
		{exact32(-3.0), exact32(-3)},
		{exact32(-3.5), exact32(-4)},

		// Special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(math.Inf(-1))},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Floor()
		if !eq32(got, tt.want) {
			t.Errorf("Float32.Floor(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat32_Ceil(t *testing.T) {
	tests := []struct {
		x    Float32
		want Float32
	}{
		{exact32(3.0), exact32(3)},
		{exact32(3.5), exact32(4)},
		{exact32(-3.0), exact32(-3)},
		{exact32(-3.5), exact32(-3)},

		// Special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(math.Inf(-1))},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Ceil()
		if !eq32(got, tt.want) {
			t.Errorf("Float32.Ceil(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat32_Trunc(t *testing.T) {
	tests := []struct {
		x    Float32
		want Float32
	}{
		{exact32(3.0), exact32(3)},
		{exact32(3.5), exact32(3)},
		{exact32(-3.0), exact32(-3)},
		{exact32(-3.5), exact32(-3)},

		// Special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(math.Inf(-1))},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Trunc()
		if !eq32(got, tt.want) {
			t.Errorf("Float32.Trunc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat32_Round(t *testing.T) {
	tests := []struct {
		x    Float32
		want Float32
	}{
		{exact32(3.0), exact32(3)},
		{exact32(3.5), exact32(4)},
		{exact32(4.5), exact32(5)},
		{exact32(-3.0), exact32(-3)},
		{exact32(-3.5), exact32(-4)},
		{exact32(-4.5), exact32(-5)},

		// Special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(math.Inf(-1))},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Round()
		if !eq32(got, tt.want) {
			t.Errorf("Float32.Round(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat32_RoundToEven(t *testing.T) {
	tests := []struct {
		x    Float32
		want Float32
	}{
		{exact32(3.0), exact32(3)},
		{exact32(3.5), exact32(4)},
		{exact32(4.5), exact32(4)},
		{exact32(-3.0), exact32(-3)},
		{exact32(-3.5), exact32(-4)},
		{exact32(-4.5), exact32(-4)},

		// Special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(math.Inf(-1))},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.RoundToEven()
		if !eq32(got, tt.want) {
			t.Errorf("Float32.RoundToEven(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
