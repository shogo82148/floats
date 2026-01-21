package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat32_Pow(t *testing.T) {
	tests := []struct {
		x    Float32
		y    Float32
		want float64
	}{
		{exact32(2), exact32(3), math.Pow(2, 3)},
		{exact32(5), exact32(0.5), math.Pow(5, 0.5)},
		{exact32(5), exact32(1.5), math.Pow(5, 1.5)},
		{exact32(5), exact32(-1.5), math.Pow(5, -1.5)},
	}

	for _, tt := range tests {
		got := tt.x.Pow(tt.y)
		if !close32(got, tt.want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		y    Float32
		want Float32
	}{
		// special cases
		// a.Pow(±0) = 1 for any a
		{exact32(2), exact32(0), exact32(1)},
		{exact32(2), exact32(math.Copysign(0, -1)), exact32(1)},
		{exact32(-2), exact32(0), exact32(1)},
		{exact32(-2), exact32(math.Copysign(0, -1)), exact32(1)},
		{exact32(math.Inf(1)), exact32(0), exact32(1)},
		{exact32(math.Inf(1)), exact32(math.Copysign(0, -1)), exact32(1)},
		{exact32(math.NaN()), exact32(0), exact32(1)},
		{exact32(math.NaN()), exact32(math.Copysign(0, -1)), exact32(1)},

		// 1.Pow(b) = 1 for any b
		{exact32(1), exact32(3), exact32(1)},
		{exact32(1), exact32(-3), exact32(1)},
		{exact32(1), exact32(math.Inf(1)), exact32(1)},
		{exact32(1), exact32(math.Inf(-1)), exact32(1)},
		{exact32(1), exact32(math.NaN()), exact32(1)},

		// a.Pow(1) = a for any a
		{exact32(2), exact32(1), exact32(2)},
		{exact32(-2), exact32(1), exact32(-2)},
		{exact32(math.Inf(1)), exact32(1), exact32(math.Inf(1))},
		{exact32(math.NaN()), exact32(1), exact32(math.NaN())},

		// NaN.Pow(b) = NaN
		{exact32(math.NaN()), exact32(3), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(-3), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.Inf(1)), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.Inf(-1)), exact32(math.NaN())},

		// a.Pow(NaN) = NaN
		{exact32(2), exact32(math.NaN()), exact32(math.NaN())},
		{exact32(-2), exact32(math.NaN()), exact32(math.NaN())},
		{exact32(math.Inf(1)), exact32(math.NaN()), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN()), exact32(math.NaN())},

		// ±0.Pow(b) = ±Inf for b an odd integer < 0
		{exact32(0), exact32(-3), exact32(math.Inf(1))},
		{exact32(math.Copysign(0, -1)), exact32(-3), exact32(math.Inf(-1))},

		// ±0.Pow(-Inf) = +Inf
		{exact32(0), exact32(math.Inf(-1)), exact32(math.Inf(1))},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(-1)), exact32(math.Inf(1))},

		// ±0.Pow(+Inf) = +0
		{exact32(0), exact32(math.Inf(1)), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(1)), exact32(0)},

		// ±0.Pow(b) = +Inf for finite b < 0 and not an odd integer
		{exact32(0), exact32(-2), exact32(math.Inf(1))},
		{exact32(math.Copysign(0, -1)), exact32(-2), exact32(math.Inf(1))},
		{exact32(math.Copysign(0, -1)), exact32(-0.5), exact32(math.Inf(1))},

		// ±0.Pow(b) = ±0 for b an odd integer > 0
		{exact32(0), exact32(3), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(3), exact32(math.Copysign(0, -1))},

		// ±0.Pow(b) = +0 for finite b > 0 and not an odd integer
		{exact32(0), exact32(2), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(2), exact32(0)},

		// -1.Pow(±Inf) = 1
		{exact32(-1), exact32(math.Inf(1)), exact32(1)},
		{exact32(-1), exact32(math.Inf(-1)), exact32(1)},

		// a.Pow(+Inf) = +Inf for |a| > 1
		{exact32(2), exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(-2), exact32(math.Inf(1)), exact32(math.Inf(1))},

		// a.Pow(-Inf) = +0 for |a| > 1
		{exact32(2), exact32(math.Inf(-1)), exact32(0)},
		{exact32(-2), exact32(math.Inf(-1)), exact32(0)},

		// a.Pow(+Inf) = +0 for |a| < 1
		{exact32(0.5), exact32(math.Inf(1)), exact32(0)},
		{exact32(-0.5), exact32(math.Inf(1)), exact32(0)},

		// a.Pow(-Inf) = +Inf for |a| < 1
		{exact32(0.5), exact32(math.Inf(-1)), exact32(math.Inf(1))},
		{exact32(-0.5), exact32(math.Inf(-1)), exact32(math.Inf(1))},

		// +Inf.Pow(b) = +Inf for b > 0
		{exact32(math.Inf(1)), exact32(2), exact32(math.Inf(1))},

		// +Inf.Pow(b) = +0 for b < 0
		{exact32(math.Inf(1)), exact32(-2), exact32(0)},

		// -Inf.Pow(b) = (-0).Pow(-b)
		{exact32(math.Inf(-1)), exact32(3), exact32(math.Inf(-1))},
		{exact32(math.Inf(-1)), exact32(2), exact32(math.Inf(1))},

		// a.Pow(b) = NaN for finite a < 0 and finite non-integer b
		{exact32(-2), exact32(0.5), exact32(math.NaN())},
		{exact32(-2), exact32(-0.5), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Pow(tt.y)
		if !eq32(got, tt.want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Pow(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Pow(x))
	}
}
