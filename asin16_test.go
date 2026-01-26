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

func TestFloat16_Acos(t *testing.T) {
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(-1), math.Acos(-1)},
		{exact16(-0.75), math.Acos(-0.75)},
		{exact16(-0.5), math.Acos(-0.5)},
		{exact16(-0.25), math.Acos(-0.25)},
		{exact16(0.25), math.Acos(0.25)},
		{exact16(0.5), math.Acos(0.5)},
		{exact16(0.75), math.Acos(0.75)},
		{exact16(1), math.Acos(1)},
	}

	for _, tt := range tests {
		got := tt.x.Acos()
		if !close16(got, tt.want) {
			t.Errorf("Acos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.NaN()), exact16(math.NaN())},
		{exact16(2), exact16(math.NaN())},
		{exact16(-2), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Acos()
		if !eq16(got, tt.want) {
			t.Errorf("Acos(%v) = %v; want %v", tt.x, got, tt.want)
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

func TestFloat16_Atan2(t *testing.T) {
	tests := []struct {
		y, x Float16
		want float64
	}{
		{exact16(1), exact16(1), math.Pi / 4},
		{exact16(1), exact16(-1), 3 * math.Pi / 4},
		{exact16(-1), exact16(-1), -3 * math.Pi / 4},
		{exact16(-1), exact16(1), -math.Pi / 4},

		// special cases
		// +0.Atan2(x<=-0) = +Pi
		{exact16(0), exact16(-1), math.Pi},
		// -0.Atan2(x<=-0) = -Pi
		{exact16(math.Copysign(0, -1)), exact16(-1), -math.Pi},
		// y>0.Atan2(0) = +Pi/2
		{exact16(1), exact16(0), math.Pi / 2},
		{exact16(1), exact16(math.Copysign(0, -1)), math.Pi / 2},
		// y<0.Atan2(0) = -Pi/2
		{exact16(-1), exact16(0), -math.Pi / 2},
		{exact16(-1), exact16(math.Copysign(0, -1)), -math.Pi / 2},
		// +Inf.Atan2(+Inf) = +Pi/4
		{exact16(math.Inf(1)), exact16(math.Inf(1)), math.Pi / 4},
		// -Inf.Atan2(+Inf) = -Pi/4
		{exact16(math.Inf(-1)), exact16(math.Inf(1)), -math.Pi / 4},
		// +Inf.Atan2(-Inf) = 3*Pi/4
		{exact16(math.Inf(1)), exact16(math.Inf(-1)), 3 * math.Pi / 4},
		// -Inf.Atan2(-Inf) = -3*Pi/4
		{exact16(math.Inf(-1)), exact16(math.Inf(-1)), -3 * math.Pi / 4},
		// y.Atan2(+Inf) = 0
		{exact16(1), exact16(math.Inf(1)), 0},
		{exact16(-1), exact16(math.Inf(1)), 0},
		// (y>0).Atan2(-Inf) = +Pi
		{exact16(1), exact16(math.Inf(-1)), math.Pi},
		// (y<0).Atan2(-Inf) = -Pi
		{exact16(-1), exact16(math.Inf(-1)), -math.Pi},
		// +Inf.Atan2(x) = +Pi/2
		{exact16(math.Inf(1)), exact16(1), math.Pi / 2},
		// -Inf.Atan2(x) = -Pi/2
		{exact16(math.Inf(-1)), exact16(1), -math.Pi / 2},
	}

	for _, tt := range tests {
		got := tt.y.Atan2(tt.x)
		if !close16(got, tt.want) {
			t.Errorf("Atan2(%v, %v) = %v; want %v", tt.y, tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		y, x Float16
		want Float16
	}{
		// special cases
		// y.Atan2(NaN) = NaN
		{exact16(1), exact16(math.NaN()), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN()), exact16(math.NaN())},
		// NaN.Atan2(x) = NaN
		{exact16(math.NaN()), exact16(1), exact16(math.NaN())},
		// +0.Atan2(x>=0) = +0
		{exact16(0), exact16(1), exact16(0)},
		// -0.Atan2(x>=0) = -0
		{exact16(math.Copysign(0, -1)), exact16(1), exact16(math.Copysign(0, -1))},
	}

	for _, tt := range strictTests {
		got := tt.y.Atan2(tt.x)
		if !eq16(got, tt.want) {
			t.Errorf("Atan2(%v, %v) = %v; want %v", tt.y, tt.x, got, tt.want)
		}
	}
}
