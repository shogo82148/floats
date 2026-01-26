package floats

import (
	"math"
	"testing"
)

func TestFloat64_Asin(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(-1), math.Asin(-1)},
		{exact64(-0.75), math.Asin(-0.75)},
		{exact64(-0.5), math.Asin(-0.5)},
		{exact64(-0.25), math.Asin(-0.25)},
		{exact64(0.25), math.Asin(0.25)},
		{exact64(0.5), math.Asin(0.5)},
		{exact64(0.75), math.Asin(0.75)},
		{exact64(1), math.Asin(1)},
	}

	for _, tt := range tests {
		got := tt.x.Asin()
		if !close64(got, tt.want) {
			t.Errorf("Asin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(0), exact64(0)},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},
		{exact64(math.NaN()), exact64(math.NaN())},
		{exact64(2), exact64(math.NaN())},
		{exact64(-2), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Asin()
		if !eq64(got, tt.want) {
			t.Errorf("Asin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat64_Acos(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(-1), math.Acos(-1)},
		{exact64(-0.75), math.Acos(-0.75)},
		{exact64(-0.5), math.Acos(-0.5)},
		{exact64(-0.25), math.Acos(-0.25)},
		{exact64(0.25), math.Acos(0.25)},
		{exact64(0.5), math.Acos(0.5)},
		{exact64(0.75), math.Acos(0.75)},
		{exact64(1), math.Acos(1)},
	}

	for _, tt := range tests {
		got := tt.x.Acos()
		if !close64(got, tt.want) {
			t.Errorf("Acos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(math.NaN()), exact64(math.NaN())},
		{exact64(2), exact64(math.NaN())},
		{exact64(-2), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Acos()
		if !eq64(got, tt.want) {
			t.Errorf("Acos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat64_Atan(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(-0.5), math.Atan(-0.5)},
		{exact64(-0.25), math.Atan(-0.25)},
		{exact64(-0.125), math.Atan(-0.125)},
		{exact64(0.125), math.Atan(0.125)},
		{exact64(0.25), math.Atan(0.25)},
		{exact64(0.5), math.Atan(0.5)},
		{exact64(0.75), math.Atan(0.75)},
		{exact64(1), math.Atan(1)},
		{exact64(2), math.Atan(2)},
		{exact64(math.Inf(-1)), math.Atan(math.Inf(-1))},
		{exact64(math.Inf(1)), math.Atan(math.Inf(1))},
	}

	for _, tt := range tests {
		got := tt.x.Atan()
		if !close64(got, tt.want) {
			t.Errorf("Atan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(0), exact64(0)},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Atan()
		if !eq64(got, tt.want) {
			t.Errorf("Atan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
