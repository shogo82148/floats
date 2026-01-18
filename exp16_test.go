package floats

import (
	"math"
	"testing"
)

func TestFloat16_Exp(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Exp(0)},
		{exact16(1), math.Exp(1)},
		{exact16(2), math.Exp(2)},
		{exact16(3), math.Exp(3)},
		{exact16(4), math.Exp(4)},
		{exact16(11), math.Exp(11)},
	}

	for _, tt := range tests {
		got := tt.x.Exp()
		if !close16(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Exp()
		if !eq16(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat16_Exp2(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Exp2(0)},
		{exact16(1), math.Exp2(1)},
		{exact16(1.5), math.Exp2(1.5)},
	}

	for _, tt := range tests {
		got := tt.x.Exp2()
		if !close16(got, tt.want) {
			t.Errorf("Exp2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Exp2()
		if !eq16(got, tt.want) {
			t.Errorf("Exp2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
