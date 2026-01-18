package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat64_Sinh(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(0), math.Sinh(0)},
		{exact64(1), math.Sinh(1)},
		{exact64(2), math.Sinh(2)},
		{exact64(3), math.Sinh(3)},
		{exact64(4), math.Sinh(4)},
		{exact64(11), math.Sinh(11)},
	}

	for _, tt := range tests {
		got := tt.x.Sinh()
		if !close64(got, tt.want) {
			t.Errorf("Sinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Sinh()
		if !eq64(got, tt.want) {
			t.Errorf("Sinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Sinh(b *testing.B) {
	x := exact64(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Sinh())
	}
}
