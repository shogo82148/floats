package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Y0(t *testing.T) {
	t.Parallel()
	tests := []float64{0.5, 1, 2, 5, 10, 50}

	for _, x := range tests {
		want := math.Y0(x)
		got := exact16(x).Y0()
		if !close16(got, want) {
			t.Errorf("Y0(%v) = %v; want %v", x, got, want)
		}
	}
}

func TestFloat16_Y0Special(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		x    Float16
		want Float16
	}{
		{exact16(0), exact16(math.Inf(-1))},
		{exact16(math.Copysign(0, -1)), exact16(math.Inf(-1))},
		{exact16(-1), exact16(math.NaN())},
		{exact16(-5), exact16(math.NaN())},
		{exact16(math.Inf(1)), exact16(0)},
		{exact16(math.Inf(-1)), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	} {
		if got := tt.x.Y0(); !eq16(got, tt.want) {
			t.Errorf("Y0(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat16_Y0All checks Y0 for all the Float16 values with Float256, which is correctly rounded.
func TestFloat16_Y0All(t *testing.T) {
	t.Parallel()
	for i := range 0x7c00 {
		for _, sign := range []uint16{0, 1 << 15} {
			x := NewFloat16FromBits(uint16(i) | sign)
			if got, want := x.Y0(), NewFloat256(x.Float64().BuiltIn()).Y0().Float16(); !eq16(got, want) {
				t.Errorf("Y0(%#04x) = %v; want %v", uint16(i)|sign, got, want)
			}
		}
	}
}

// TestFloat16_Y0Poly checks the absolute error of the polynomials, which are hidden by the rounding to Float16,
// with Float256.
func TestFloat16_Y0Poly(t *testing.T) {
	t.Parallel()
	for i := 0x4000; i < 0x5400; i++ {
		x := normal16ToFloat64(Float16(i))
		k := 0
		if x < 3 {
			k = int(x*2) - 4
		} else {
			k = int(x) - 1
		}
		c := &y016Coeffs[k]
		tt := x - y016Centers[k]
		got := 0.0
		for j := len(c) - 1; j >= 0; j-- {
			got = got*tt + c[j]
		}
		want := NewFloat256(x).Y0().Float64().BuiltIn()
		if math.Abs(got-want) > 0x1p-30 {
			t.Errorf("polynomial(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat16_Y0(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float16
	}{
		{"small", exact16(0.5)},    // math.Y0
		{"medium", exact16(5)},     // the polynomials
		{"medium2", exact16(50.5)}, // the polynomials
		{"large", exact16(1000)},   // math.Y0
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Y0())
			}
		})
	}
}
