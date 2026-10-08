package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_J1(t *testing.T) {
	t.Parallel()
	tests := []float64{-50, -10, -1, -0.5, 0, 0.5, 1, 2, 5, 10, 50}

	for _, x := range tests {
		want := math.J1(x)
		got := exact16(x).J1()
		if !close16(got, want) {
			t.Errorf("J1(%v) = %v; want %v", x, got, want)
		}
	}
}

func TestFloat16_J1Special(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		x    Float16
		want Float16
	}{
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(0)},
		{exact16(math.Inf(-1)), exact16(0)},
		{exact16(math.NaN()), exact16(math.NaN())},
		{NewFloat16FromBits(0x0001), NewFloat16FromBits(0x0000)}, // J1(x) = x/2 (1 - x**2/8) is rounded down to 0
		{NewFloat16FromBits(0x0003), NewFloat16FromBits(0x0001)}, // x/2 is a tie of 1 and 2, and J1(x) is less than it
	} {
		if got := tt.x.J1(); !eq16(got, tt.want) {
			t.Errorf("J1(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat16_J1All checks J1 for all the Float16 values. The previous implementation, which rounds math.J1,
// is correctly rounded for all the Float16 values, that was checked with mpmath, so that it is the reference.
func TestFloat16_J1All(t *testing.T) {
	t.Parallel()
	for i := range 0x10000 {
		x := NewFloat16FromBits(uint16(i))
		if x.IsZero() {
			continue // the sign of zero is tested in TestFloat16_J1Special
		}
		if got, want := x.J1(), NewFloat16(math.J1(x.Float64().BuiltIn())); !eq16(got, want) {
			t.Errorf("J1(%#04x) = %v; want %v", i, got, want)
		}
	}
}

// TestFloat16_J1Poly checks the absolute error of the polynomials, which are hidden by the rounding to Float16,
// with math.J1.
func TestFloat16_J1Poly(t *testing.T) {
	t.Parallel()
	for i := 0x4000; i < 0x5400; i++ {
		x := normal16ToFloat64(Float16(i))
		k := int(x)
		c := &j116Coeffs[k-2]
		tt := x - float64(k) - 0.5
		got := 0.0
		for j := len(c) - 1; j >= 0; j-- {
			got = got*tt + c[j]
		}
		if want := math.J1(x); math.Abs(got-want) > 0x1p-30 {
			t.Errorf("polynomial(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat16_J1(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float16
	}{
		{"small", exact16(0.5)},      // math.J1
		{"medium", exact16(5)},       // the polynomials
		{"medium2", exact16(50.5)},   // the polynomials
		{"negative", exact16(-10.5)}, // the polynomials
		{"large", exact16(1000)},     // math.J1
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.J1())
			}
		})
	}
}
