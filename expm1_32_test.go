package floats

import (
	"math"
	"testing"
)

func TestFloat32_Expm1(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-0x1p24), math.Expm1(-0x1p24)},
		{exact32(-0x1.1aaaaep+03), math.Expm1(-0x1.1aaaaep+03)},
		{exact32(-1), math.Expm1(-1)},
		{exact32(-0x1p-25), math.Expm1(-0x1p-25)},
		{exact32(-0x1p-24), math.Expm1(-0x1p-24)},
		{exact32(0), math.Expm1(0)},
		{exact32(0x1p-25), math.Expm1(0x1p-25)},
		{exact32(0x1p-24), math.Expm1(0x1p-24)},
		{exact32(0x1.99999ap-02), math.Expm1(0x1.99999ap-02)},
		{exact32(1), math.Expm1(1)},
		{exact32(10), math.Expm1(10)},
		{exact32(36), math.Expm1(36)},
		{exact32(62), math.Expm1(62)},
		{exact32(0x1p24), math.Expm1(0x1p24)},
	}

	for _, tt := range tests {
		got := tt.x.Expm1()
		if !close32(got, tt.want) {
			t.Errorf("Expm1(%v) = %v; want %v", tt.x, got, tt.want)
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
		{exact32(math.Inf(-1)), exact32(-1)},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Expm1()
		if !eq32(got, tt.want) {
			t.Errorf("Expm1(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
