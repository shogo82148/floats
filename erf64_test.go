package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat64_Erf(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(0), math.Erf(0)},
		{exact64(1), math.Erf(1)},
		{exact64(2), math.Erf(2)},
		{exact64(3), math.Erf(3)},
		{exact64(4), math.Erf(4)},
		{exact64(11), math.Erf(11)},
	}

	for _, tt := range tests {
		got := tt.x.Erf()
		if !close64(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(math.Inf(1)), exact64(1)},
		{exact64(math.Inf(-1)), exact64(-1)},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erf()
		if !eq64(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Erf(b *testing.B) {
	x := exact64(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erf())
	}
}
