package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Exp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Exp(0)},
		{exact16(1), math.Exp(1)},
		{exact16(2), math.Exp(2)},
		{exact16(3), math.Exp(3)},
		{exact16(4), math.Exp(4)},
		{exact16(11), math.Exp(11)},
	}

	for _, tt := range tests {
		got := tt.x.Exp()
		if !close16(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(0)},
		{exact16(math.NaN()), exact16(math.NaN())},
		{NewFloat16FromBits(0xfe00), exact16(math.NaN())}, // negative NaN
	}

	for _, tt := range strictTests {
		got := tt.x.Exp()
		if !eq16(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat16_Exp2(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Exp2(0)},
		{exact16(1), math.Exp2(1)},
		{exact16(1.5), math.Exp2(1.5)},
	}

	for _, tt := range tests {
		got := tt.x.Exp2()
		if !close16(got, tt.want) {
			t.Errorf("Exp2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(0)},
		{exact16(math.NaN()), exact16(math.NaN())},
		{NewFloat16FromBits(0xfe00), exact16(math.NaN())}, // negative NaN

		// overflow
		{exact16(16), exact16(math.Inf(1))},

		// underflow; 2**-25 is the midpoint of 0 and the smallest subnormal, and rounds to even.
		{exact16(-25), exact16(0)},
		{exact16(-24), NewFloat16FromBits(0x0001)},
	}

	for _, tt := range strictTests {
		got := tt.x.Exp2()
		if !eq16(got, tt.want) {
			t.Errorf("Exp2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat16_ExpAll checks Exp, Exp2, and Expm1 for all finite Float16 values.
// For every finite Float16 value, the exact exp, exp2, and expm1 are at least 2**-17 ulp away
// from the midpoint of two adjacent Float16 values (checked with mpmath),
// except for 2**-25, which is exactly the midpoint and is computed exactly by math.Exp2.
// So the results of the math package rounded to Float16 are correctly rounded.
func TestFloat16_ExpAll(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		x := NewFloat16FromBits(uint16(i))
		if x.IsNaN() || x.IsInf(0) {
			continue
		}
		f := x.Float64().BuiltIn()
		if got, want := x.Exp(), NewFloat16(math.Exp(f)); !eq16(got, want) {
			t.Errorf("Exp(%v) = %v; want %v", x, got, want)
		}
		if got, want := x.Exp2(), NewFloat16(math.Exp2(f)); !eq16(got, want) {
			t.Errorf("Exp2(%v) = %v; want %v", x, got, want)
		}
		if got, want := x.Expm1(), NewFloat16(math.Expm1(f)); !eq16(got, want) {
			t.Errorf("Expm1(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat16_Exp(b *testing.B) {
	x := exact16(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Exp())
	}
}

func BenchmarkFloat16_Exp2(b *testing.B) {
	x := exact16(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Exp2())
	}
}
