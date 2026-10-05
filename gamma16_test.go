package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat16_Gamma(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(-0.5), math.Gamma(-0.5)},
		{exact16(0.5), math.Gamma(0.5)},
		{exact16(1), math.Gamma(1)},
		{exact16(1.5), math.Gamma(1.5)},
		{exact16(2), math.Gamma(2)},
		{exact16(2.5), math.Gamma(2.5)},
		{exact16(3), math.Gamma(3)},
	}

	for _, tt := range tests {
		got := tt.x.Gamma()
		if !close16(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(0), exact16(math.Inf(1))},
		{exact16(math.Copysign(0, -1)), exact16(math.Inf(-1))},
		{exact16(-1), exact16(math.NaN())},
		{exact16(-2), exact16(math.NaN())},
		{exact16(math.Inf(-1)), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Gamma()
		if !eq16(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat16_GammaAll checks Gamma for all bit patterns of Float16.
// For every finite Float16 value, the exact Gamma is either exactly representable or at least 2**-15 ulp away
// from the midpoint of two adjacent Float16 values (checked with mpmath),
// so math.Gamma rounded to Float16 is correctly rounded.
func TestFloat16_GammaAll(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		x := NewFloat16FromBits(uint16(i))
		got, want := x.Gamma(), NewFloat16(math.Gamma(x.Float64().BuiltIn()))
		if uint16(got) != uint16(want) {
			t.Errorf("Gamma(%#04x) = %#04x; want %#04x", i, uint16(got), uint16(want))
		}
	}
}

// TestFloat16_GammaPoly checks the polynomials of Gamma, which are not fully tested by TestFloat16_GammaAll
// because the rounding to Float16 hides their errors, with the math package.
func TestFloat16_GammaPoly(t *testing.T) {
	t.Parallel()
	const bound = 0x1p-34 // the relative errors of the polynomials are less than 2**-35 and 2**-36
	rnd := rand.New(rand.NewPCG(1, 2))
	for range 100000 {
		x := 1 + 12.06*rnd.Float64() // [1, 13.06)
		if got, want := gamma16Poly(x), math.Gamma(x); math.Abs(got-want) > bound*want {
			t.Errorf("gamma16Poly(%v) = %v; want %v", x, got, want)
		}
		r := 0.5 * (1 - rnd.Float64()) // (0, 1/2]
		if got, want := gamma16Sinc(r), math.Sin(math.Pi*r)/(math.Pi*r); math.Abs(got-want) > bound*want {
			t.Errorf("gamma16Sinc(%v) = %v; want %v", r, got, want)
		}
	}
}

func BenchmarkFloat16_Gamma(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float16
	}{
		{"small", NewFloat16(0.001)},     // Gamma(x) ~ 1/x
		{"medium", exact16(1.5)},         // 1 <= x < 2
		{"large", exact16(6.5)},          // 2 <= x < 9.3
		{"negative", NewFloat16(-2.5)},   // -12.1 < x < 0
		{"verynegative", exact16(-10.5)}, // underflows to subnormal numbers
		{"overflow", exact16(20.5)},      // +Inf
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Gamma())
			}
		})
	}
}
