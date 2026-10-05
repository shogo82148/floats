package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Sin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(1), math.Sin(1)},
		{exact16(2), math.Sin(2)},
		{exact16(3), math.Sin(3)},
	}

	for _, tt := range tests {
		got := tt.x.Sin()
		if !close16(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(math.NaN())},
		{exact16(math.Inf(-1)), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Sin()
		if !eq16(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat16_Cos(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(1), math.Cos(1)},
		{exact16(2), math.Cos(2)},
		{exact16(3), math.Cos(3)},
	}

	for _, tt := range tests {
		got := tt.x.Cos()
		if !close16(got, tt.want) {
			t.Errorf("Cos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(math.NaN())},
		{exact16(math.Inf(-1)), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Cos()
		if !eq16(got, tt.want) {
			t.Errorf("Cos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat16_Sincos(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x   Float16
		sin float64
		cos float64
	}{
		{exact16(1), math.Sin(1), math.Cos(1)},
		{exact16(2), math.Sin(2), math.Cos(2)},
		{exact16(3), math.Sin(3), math.Cos(3)},
	}

	for _, tt := range tests {
		sin, cos := tt.x.Sincos()
		if !close16(sin, tt.sin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", tt.x, sin, tt.sin)
		}
		if !close16(cos, tt.cos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", tt.x, cos, tt.cos)
		}
	}

	strictTests := []struct {
		x   Float16
		sin Float16
		cos Float16
	}{
		// special cases
		{exact16(0), exact16(0), exact16(1)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1)), exact16(1)},
		{exact16(math.Inf(1)), exact16(math.NaN()), exact16(math.NaN())},
		{exact16(math.Inf(-1)), exact16(math.NaN()), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		sin, cos := tt.x.Sincos()
		if !eq16(sin, tt.sin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", tt.x, sin, tt.sin)
		}
		if !eq16(cos, tt.cos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", tt.x, cos, tt.cos)
		}
	}
}

func TestFloat16_Tan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(1), math.Tan(1)},
		{exact16(2), math.Tan(2)},
		{exact16(3), math.Tan(3)},
	}

	for _, tt := range tests {
		got := tt.x.Tan()
		if !close16(got, tt.want) {
			t.Errorf("Tan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(math.NaN())},
		{exact16(math.Inf(-1)), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},

		// overflow: tan(177.5) = -66347.4...
		{NewFloat16FromBits(0x598c), NewFloat16Inf(-1)},
		{NewFloat16FromBits(0xd98c), NewFloat16Inf(1)},
	}
	for _, tt := range strictTests {
		got := tt.x.Tan()
		if !eq16(got, tt.want) {
			t.Errorf("Tan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat16_SinCosTanAll checks Sin, Cos, Sincos, and Tan for all finite Float16 values.
// For every finite Float16 value, the exact sin, cos, and tan are at least 2**-17 ulp away
// from the midpoint of two adjacent Float16 values (checked with mpmath),
// so math.Sin, math.Cos, and math.Tan rounded to Float16 are correctly rounded.
func TestFloat16_SinCosTanAll(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		x := NewFloat16FromBits(uint16(i))
		if x.IsNaN() || x.IsInf(0) {
			continue
		}
		f := x.Float64().BuiltIn()
		wantSin := NewFloat16(math.Sin(f))
		wantCos := NewFloat16(math.Cos(f))
		wantTan := NewFloat16(math.Tan(f))
		if got := x.Sin(); !eq16(got, wantSin) {
			t.Errorf("Sin(%v) = %v; want %v", x, got, wantSin)
		}
		if got := x.Cos(); !eq16(got, wantCos) {
			t.Errorf("Cos(%v) = %v; want %v", x, got, wantCos)
		}
		sin, cos := x.Sincos()
		if !eq16(sin, wantSin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", x, sin, wantSin)
		}
		if !eq16(cos, wantCos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", x, cos, wantCos)
		}
		if got := x.Tan(); !eq16(got, wantTan) {
			t.Errorf("Tan(%v) = %v; want %v", x, got, wantTan)
		}
	}
}

func BenchmarkFloat16_Sin(b *testing.B) {
	for _, x := range []Float16{exact16(0.5), exact16(3), exact16(1000)} {
		b.Run(x.String(), func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(x.Sin())
			}
		})
	}
}

func BenchmarkFloat16_Cos(b *testing.B) {
	for _, x := range []Float16{exact16(0.5), exact16(3), exact16(1000)} {
		b.Run(x.String(), func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(x.Cos())
			}
		})
	}
}

func BenchmarkFloat16_Sincos(b *testing.B) {
	for _, x := range []Float16{exact16(0.5), exact16(3), exact16(1000)} {
		b.Run(x.String(), func(b *testing.B) {
			for b.Loop() {
				s, c := x.Sincos()
				runtime.KeepAlive(s)
				runtime.KeepAlive(c)
			}
		})
	}
}

func BenchmarkFloat16_Tan(b *testing.B) {
	for _, x := range []Float16{exact16(0.5), exact16(3), exact16(1000)} {
		b.Run(x.String(), func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(x.Tan())
			}
		})
	}
}
