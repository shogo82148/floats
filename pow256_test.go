package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat2356_Pow(t *testing.T) {
	tests := []struct {
		x    Float256
		y    Float256
		want string
	}{
		{exact256(2), exact256(3), "8"},
		{exact256(5), exact256(0.5), "2.236067977499789696409173668731276235440618359611525724270897245410520925637804899414414408378782274969508176150773783504253267724"},
		{exact256(5), exact256(1.5), "11.18033988749894848204586834365638117720309179805762862135448622705260462818902449707207204189391137484754088075386891752126633862"},
	}

	for _, tt := range tests {
		got := tt.x.Pow(tt.y)
		if !close256(got, tt.want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		y    Float256
		want Float256
	}{
		// special cases
		// a.Pow(±0) = 1 for any a
		{exact256(2), exact256(0), exact256(1)},
		{exact256(2), exact256(math.Copysign(0, -1)), exact256(1)},
		{exact256(-2), exact256(0), exact256(1)},
		{exact256(-2), exact256(math.Copysign(0, -1)), exact256(1)},
		{exact256(math.Inf(1)), exact256(0), exact256(1)},
		{exact256(math.Inf(1)), exact256(math.Copysign(0, -1)), exact256(1)},
		{exact256(math.NaN()), exact256(0), exact256(1)},
		{exact256(math.NaN()), exact256(math.Copysign(0, -1)), exact256(1)},

		// 1.Pow(b) = 1 for any b
		{exact256(1), exact256(3), exact256(1)},
		{exact256(1), exact256(-3), exact256(1)},
		{exact256(1), exact256(math.Inf(1)), exact256(1)},
		{exact256(1), exact256(math.Inf(-1)), exact256(1)},
		{exact256(1), exact256(math.NaN()), exact256(1)},

		// a.Pow(1) = a for any a
		{exact256(2), exact256(1), exact256(2)},
		{exact256(-2), exact256(1), exact256(-2)},
		{exact256(math.Inf(1)), exact256(1), exact256(math.Inf(1))},
		{exact256(math.NaN()), exact256(1), exact256(math.NaN())},

		// NaN.Pow(b) = NaN
		{exact256(math.NaN()), exact256(3), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(-3), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.Inf(1)), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.Inf(-1)), exact256(math.NaN())},

		// a.Pow(NaN) = NaN
		{exact256(2), exact256(math.NaN()), exact256(math.NaN())},
		{exact256(-2), exact256(math.NaN()), exact256(math.NaN())},
		{exact256(math.Inf(1)), exact256(math.NaN()), exact256(math.NaN())},
		{exact256(math.Inf(-1)), exact256(math.NaN()), exact256(math.NaN())},

		// ±0.Pow(b) = ±Inf for b an odd integer < 0
		{exact256(0), exact256(-3), exact256(math.Inf(1))},
		{exact256(math.Copysign(0, -1)), exact256(-3), exact256(math.Inf(-1))},

		// ±0.Pow(-Inf) = +Inf
		{exact256(0), exact256(math.Inf(-1)), exact256(math.Inf(1))},
		{exact256(math.Copysign(0, -1)), exact256(math.Inf(-1)), exact256(math.Inf(1))},

		// ±0.Pow(+Inf) = +0
		{exact256(0), exact256(math.Inf(1)), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Inf(1)), exact256(0)},

		// ±0.Pow(b) = +Inf for finite b < 0 and not an odd integer
		{exact256(0), exact256(-2), exact256(math.Inf(1))},
		{exact256(math.Copysign(0, -1)), exact256(-2), exact256(math.Inf(1))},
		{exact256(math.Copysign(0, -1)), exact256(-0.5), exact256(math.Inf(1))},

		// ±0.Pow(b) = ±0 for b an odd integer > 0
		{exact256(0), exact256(3), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(3), exact256(math.Copysign(0, -1))},

		// ±0.Pow(b) = +0 for finite b > 0 and not an odd integer
		{exact256(0), exact256(2), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(2), exact256(0)},

		// -1.Pow(±Inf) = 1
		{exact256(-1), exact256(math.Inf(1)), exact256(1)},
		{exact256(-1), exact256(math.Inf(-1)), exact256(1)},

		// a.Pow(+Inf) = +Inf for |a| > 1
		{exact256(2), exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(-2), exact256(math.Inf(1)), exact256(math.Inf(1))},

		// a.Pow(-Inf) = +0 for |a| > 1
		{exact256(2), exact256(math.Inf(-1)), exact256(0)},
		{exact256(-2), exact256(math.Inf(-1)), exact256(0)},

		// a.Pow(+Inf) = +0 for |a| < 1
		{exact256(0.5), exact256(math.Inf(1)), exact256(0)},
		{exact256(-0.5), exact256(math.Inf(1)), exact256(0)},

		// a.Pow(-Inf) = +Inf for |a| < 1
		{exact256(0.5), exact256(math.Inf(-1)), exact256(math.Inf(1))},
		{exact256(-0.5), exact256(math.Inf(-1)), exact256(math.Inf(1))},

		// +Inf.Pow(b) = +Inf for b > 0
		{exact256(math.Inf(1)), exact256(2), exact256(math.Inf(1))},

		// +Inf.Pow(b) = +0 for b < 0
		{exact256(math.Inf(1)), exact256(-2), exact256(0)},

		// -Inf.Pow(b) = (-0).Pow(-b)
		{exact256(math.Inf(-1)), exact256(3), exact256(math.Inf(-1))},
		{exact256(math.Inf(-1)), exact256(2), exact256(math.Inf(1))},

		// a.Pow(b) = NaN for finite a < 0 and finite non-integer b
		{exact256(-2), exact256(0.5), exact256(math.NaN())},
		{exact256(-2), exact256(-0.5), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Pow(tt.y)
		if !eq256(got, tt.want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Pow(b *testing.B) {
	x := exact256(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Pow(x))
	}
}
