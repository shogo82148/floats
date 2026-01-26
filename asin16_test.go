package floats

import (
	"math"
	"testing"
)

func TestFloat16_Asin(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(-1), math.Asin(-1)},
		{exact16(-0.75), math.Asin(-0.75)},
		{exact16(-0.5), math.Asin(-0.5)},
		{exact16(-0.25), math.Asin(-0.25)},
		{exact16(0.25), math.Asin(0.25)},
		{exact16(0.5), math.Asin(0.5)},
		{exact16(0.75), math.Asin(0.75)},
		{exact16(1), math.Asin(1)},
	}

	for _, tt := range tests {
		got := tt.x.Asin()
		if !close16(got, tt.want) {
			t.Errorf("Asin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.NaN()), exact16(math.NaN())},
		{exact16(2), exact16(math.NaN())},
		{exact16(-2), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Asin()
		if !eq16(got, tt.want) {
			t.Errorf("Asin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat16_Atan(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(-0.5), math.Atan(-0.5)},
		{exact16(-0.25), math.Atan(-0.25)},
		{exact16(-0.125), math.Atan(-0.125)},
		{exact16(0.125), math.Atan(0.125)},
		{exact16(0.25), math.Atan(0.25)},
		{exact16(0.5), math.Atan(0.5)},
		{exact16(0.75), math.Atan(0.75)},
		{exact16(1), math.Atan(1)},
		{exact16(2), math.Atan(2)},
		{exact16(math.Inf(-1)), math.Atan(math.Inf(-1))},
		{exact16(math.Inf(1)), math.Atan(math.Inf(1))},
	}

	for _, tt := range tests {
		got := tt.x.Atan()
		if !close16(got, tt.want) {
			t.Errorf("Atan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Atan()
		if !eq16(got, tt.want) {
			t.Errorf("Atan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
