package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat64_Log1p(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(1), math.Log1p(1)},
		{exact64(2), math.Log1p(2)},
		{exact64(3), math.Log1p(3)},
		{exact64(4), math.Log1p(4)},
		{exact64(11), math.Log1p(11)},
	}

	for _, tt := range tests {
		got := tt.x.Log1p()
		if !close64(got, tt.want) {
			t.Errorf("Log1p(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(0), exact64(0)},
		{exact64(math.Copysign(0, -1)), exact64(math.Copysign(0, -1))},
		{exact64(-1), exact64(math.Inf(-1))},
		{exact64(-2), exact64(math.NaN())},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log1p()
		if !eq64(got, tt.want) {
			t.Errorf("Log1p(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Log1p(b *testing.B) {
	x := NewFloat64(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log1p())
	}
}
