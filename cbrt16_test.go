package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Cbrt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(27), 3},
		{exact16(8), 2},
		{exact16(1), 1},
		{exact16(0), 0},
		{exact16(-1), -1},
		{exact16(-8), -2},
		{exact16(-27), -3},
	}

	for _, test := range tests {
		got := test.x.Cbrt()
		if !close16(got, test.want) {
			t.Errorf("Cbrt(%v) = %v; want %v", test.x, got, test.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// Special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(math.Inf(-1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Cbrt()
		if !eq16(got, tt.want) {
			t.Errorf("Float16.Cbrt(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat16_CbrtAll compares Cbrt with math.Cbrt rounded to Float16 for all bit patterns of Float16.
// No cube root of a Float16 is closer than 2**-16 ulp to a midpoint of two Float16 values (checked with mpmath),
// and the error of math.Cbrt is about 2**-52, so the result of the comparison is the correctly rounded one.
func TestFloat16_CbrtAll(t *testing.T) {
	for i := range 1 << 16 {
		a := Float16(i)
		got := a.Cbrt()
		want := NewFloat16(math.Cbrt(a.Float64().BuiltIn()))
		if uint16(got) != uint16(want) {
			t.Errorf("Cbrt(%#04x) = %#04x; want %#04x", i, uint16(got), uint16(want))
		}
	}
}

func BenchmarkFloat16_Cbrt(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float16
	}{
		{"normal", exact16(10)},
		{"small", NewFloat16(0.001)},
		{"subnormal", NewFloat16FromBits(0x0123)},
		{"large", exact16(60000)},
		{"negative", exact16(-5.5)},
		{"zero", exact16(0)},
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Cbrt())
			}
		})
	}
}
