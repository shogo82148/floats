package floats

import (
	"log"
	"math"
	"runtime"
	"testing"
)

func TestFloat128_Pow(t *testing.T) {
	x, _ := ParseFloat128("10384593717069655257060992658440192")
	log.Printf("%016x, %016x", x[0], x[1])
	log.Printf("%g", x)

	tests := []struct {
		x    Float128
		y    Float128
		want string
	}{
		{exact128(2), exact128(3), "8"},
		{exact128(5), exact128(0.5), "2.236067977499789696409173668731276235440618359611525724270897245410520925637804899"},
		{exact128(5), exact128(1.5), "11.1803398874989484820458683436563811772030917980576286213544862270526046281890245"},
	}

	for _, tt := range tests {
		got := tt.x.Pow(tt.y)
		if !close128(got, tt.want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		y    Float128
		want Float128
	}{
		// special cases
		// a.Pow(±0) = 1 for any a
		{exact128(2), exact128(0), exact128(1)},
		{exact128(2), exact128(math.Copysign(0, -1)), exact128(1)},
		{exact128(-2), exact128(0), exact128(1)},
		{exact128(-2), exact128(math.Copysign(0, -1)), exact128(1)},
		{exact128(math.Inf(1)), exact128(0), exact128(1)},
		{exact128(math.Inf(1)), exact128(math.Copysign(0, -1)), exact128(1)},
		{exact128(math.NaN()), exact128(0), exact128(1)},
		{exact128(math.NaN()), exact128(math.Copysign(0, -1)), exact128(1)},

		// 1.Pow(b) = 1 for any b
		{exact128(1), exact128(3), exact128(1)},
		{exact128(1), exact128(-3), exact128(1)},
		{exact128(1), exact128(math.Inf(1)), exact128(1)},
		{exact128(1), exact128(math.Inf(-1)), exact128(1)},
		{exact128(1), exact128(math.NaN()), exact128(1)},

		// a.Pow(1) = a for any a
		{exact128(2), exact128(1), exact128(2)},
		{exact128(-2), exact128(1), exact128(-2)},
		{exact128(math.Inf(1)), exact128(1), exact128(math.Inf(1))},
		{exact128(math.NaN()), exact128(1), exact128(math.NaN())},

		// NaN.Pow(b) = NaN
		{exact128(math.NaN()), exact128(3), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(-3), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.Inf(1)), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.Inf(-1)), exact128(math.NaN())},

		// a.Pow(NaN) = NaN
		{exact128(2), exact128(math.NaN()), exact128(math.NaN())},
		{exact128(-2), exact128(math.NaN()), exact128(math.NaN())},
		{exact128(math.Inf(1)), exact128(math.NaN()), exact128(math.NaN())},
		{exact128(math.Inf(-1)), exact128(math.NaN()), exact128(math.NaN())},

		// ±0.Pow(b) = ±Inf for b an odd integer < 0
		{exact128(0), exact128(-3), exact128(math.Inf(1))},
		{exact128(math.Copysign(0, -1)), exact128(-3), exact128(math.Inf(-1))},

		// ±0.Pow(-Inf) = +Inf
		{exact128(0), exact128(math.Inf(-1)), exact128(math.Inf(1))},
		{exact128(math.Copysign(0, -1)), exact128(math.Inf(-1)), exact128(math.Inf(1))},

		// ±0.Pow(+Inf) = +0
		{exact128(0), exact128(math.Inf(1)), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Inf(1)), exact128(0)},

		// ±0.Pow(b) = +Inf for finite b < 0 and not an odd integer
		{exact128(0), exact128(-2), exact128(math.Inf(1))},
		{exact128(math.Copysign(0, -1)), exact128(-2), exact128(math.Inf(1))},
		{exact128(math.Copysign(0, -1)), exact128(-0.5), exact128(math.Inf(1))},

		// ±0.Pow(b) = ±0 for b an odd integer > 0
		{exact128(0), exact128(3), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(3), exact128(math.Copysign(0, -1))},

		// ±0.Pow(b) = +0 for finite b > 0 and not an odd integer
		{exact128(0), exact128(2), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(2), exact128(0)},

		// -1.Pow(±Inf) = 1
		{exact128(-1), exact128(math.Inf(1)), exact128(1)},
		{exact128(-1), exact128(math.Inf(-1)), exact128(1)},

		// a.Pow(+Inf) = +Inf for |a| > 1
		{exact128(2), exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(-2), exact128(math.Inf(1)), exact128(math.Inf(1))},

		// a.Pow(-Inf) = +0 for |a| > 1
		{exact128(2), exact128(math.Inf(-1)), exact128(0)},
		{exact128(-2), exact128(math.Inf(-1)), exact128(0)},

		// a.Pow(+Inf) = +0 for |a| < 1
		{exact128(0.5), exact128(math.Inf(1)), exact128(0)},
		{exact128(-0.5), exact128(math.Inf(1)), exact128(0)},

		// a.Pow(-Inf) = +Inf for |a| < 1
		{exact128(0.5), exact128(math.Inf(-1)), exact128(math.Inf(1))},
		{exact128(-0.5), exact128(math.Inf(-1)), exact128(math.Inf(1))},

		// +Inf.Pow(b) = +Inf for b > 0
		{exact128(math.Inf(1)), exact128(2), exact128(math.Inf(1))},

		// +Inf.Pow(b) = +0 for b < 0
		{exact128(math.Inf(1)), exact128(-2), exact128(0)},

		// -Inf.Pow(b) = (-0).Pow(-b)
		{exact128(math.Inf(-1)), exact128(3), exact128(math.Inf(-1))},
		{exact128(math.Inf(-1)), exact128(2), exact128(math.Inf(1))},

		// a.Pow(b) = NaN for finite a < 0 and finite non-integer b
		{exact128(-2), exact128(0.5), exact128(math.NaN())},
		{exact128(-2), exact128(-0.5), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Pow(tt.y)
		if !eq128(got, tt.want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Pow(b *testing.B) {
	x := exact128(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Pow(x))
	}
}
