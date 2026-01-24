package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat64_Cbrt(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(27), 3},
		{exact64(8), 2},
		{exact64(2), math.Cbrt(2)},
		{exact64(1), 1},
		{exact64(0), 0},
		{exact64(-1), -1},
		{exact64(-2), math.Cbrt(-2)},
		{exact64(-8), -2},
		{exact64(-27), -3},
	}

	for _, test := range tests {
		got := test.x.Cbrt()
		if !close64(got, test.want) {
			t.Errorf("Cbrt(%v) = %v; want %v", test.x, got, test.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// Special cases
		{exact64(0), exact64(0)},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},
		{exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(math.Inf(-1)), exact64(math.Inf(-1))},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Cbrt()
		if !eq64(got, tt.want) {
			t.Errorf("Float64.Cbrt(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Cbrt(b *testing.B) {
	x := NewFloat64(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Cbrt())
	}
}
