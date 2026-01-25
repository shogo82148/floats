package floats

import (
	"math"
	"testing"
)

func TestFloat64_Asinh(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(0), math.Asinh(0)},
		{exact64(0.25), math.Asinh(0.25)},
		{exact64(0.5), math.Asinh(0.5)},
		{exact64(1), math.Asinh(1)},
		{exact64(21), math.Asinh(21)},
		{exact64(22), math.Asinh(22)},

		{exact64(-0), -math.Asinh(0)},
		{exact64(-0.25), -math.Asinh(0.25)},
		{exact64(-0.5), -math.Asinh(0.5)},
		{exact64(-1), -math.Asinh(1)},
		{exact64(-21), -math.Asinh(21)},
		{exact64(-22), -math.Asinh(22)},
	}

	for _, tt := range tests {
		got := tt.x.Asinh()
		if !close64(got, tt.want) {
			t.Errorf("Asinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		{exact64(0), exact64(0)},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},

		// special cases
		{exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(math.Inf(-1)), exact64(math.Inf(-1))},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Asinh()
		if !eq64(got, tt.want) {
			t.Errorf("Asinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat64_Acosh(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(1), math.Acosh(1)},
		{exact64(21), math.Acosh(21)},
		{exact64(22), math.Acosh(22)},
	}

	for _, tt := range tests {
		got := tt.x.Acosh()
		if !close64(got, tt.want) {
			t.Errorf("Acosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(-2), exact64(math.NaN())},
		{exact64(math.Inf(-1)), exact64(math.NaN())},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Acosh()
		if !eq64(got, tt.want) {
			t.Errorf("Acosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat64_Atanh(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(0.25), math.Atanh(0.25)},
		{exact64(0.5), math.Atanh(0.5)},
	}

	for _, tt := range tests {
		got := tt.x.Atanh()
		if !close64(got, tt.want) {
			t.Errorf("Atanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(2), exact64(math.NaN())},
		{exact64(1), exact64(math.Inf(1))},
		{exact64(0), exact64(0)},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},
		{exact64(-1), exact64(math.Inf(-1))},
		{exact64(-2), exact64(math.NaN())},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Atanh()
		if !eq64(got, tt.want) {
			t.Errorf("Atanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
