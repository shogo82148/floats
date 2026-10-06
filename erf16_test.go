package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat16_Erf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Erf(0)},
		{exact16(0x1p-14), math.Erf(0x1p-14)},
		{exact16(1), math.Erf(1)},
		{exact16(2), math.Erf(2)},
		{exact16(3), math.Erf(3)},
		{exact16(4), math.Erf(4)},
		{exact16(11), math.Erf(11)},
	}

	for _, tt := range tests {
		got := tt.x.Erf()
		if !close16(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(1)},
		{exact16(math.Inf(-1)), exact16(-1)},
		{exact16(math.NaN()), exact16(math.NaN())},
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{NewFloat16FromBits(0x0001), NewFloat16FromBits(0x0001)},
		{NewFloat16FromBits(0x8001), NewFloat16FromBits(0x8001)},

		// erf(x) is rounded to 1 for |x| >= 2.5957
		{NewFloat16FromBits(0x4130), NewFloat16FromBits(0x3bff)},
		{NewFloat16FromBits(0x4131), exact16(1)},
		{NewFloat16FromBits(0xc130), NewFloat16FromBits(0xbbff)},
		{NewFloat16FromBits(0xc131), exact16(-1)},
		{NewFloat16FromBits(0x7bff), exact16(1)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erf()
		if !eq16(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat16_ErfAll checks Erf for all bit patterns of Float16.
// For every finite Float16 value, the exact erf is at least 2**-15 ulp away from the midpoint of two adjacent Float16 values
// (checked with mpmath), so math.Erf rounded to Float16 is correctly rounded.
func TestFloat16_ErfAll(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		x := NewFloat16FromBits(uint16(i))
		got, want := x.Erf(), NewFloat16(math.Erf(x.Float64().BuiltIn()))
		if uint16(got) != uint16(want) && !(got.IsNaN() && want.IsNaN()) {
			t.Errorf("Erf(%#04x) = %#04x; want %#04x", i, uint16(got), uint16(want))
		}
	}
}

// TestFloat16_ErfPoly checks the polynomials of Erf, whose errors are hidden by the rounding to Float16, with the math package.
func TestFloat16_ErfPoly(t *testing.T) {
	t.Parallel()
	const bound = 0x1p-34 // the relative errors of the polynomials are less than 2**-35 and 2**-56
	rnd := rand.New(rand.NewPCG(1, 2))
	for range 100000 {
		x := 0.125 + 2.4607*rnd.Float64() // [1/8, 2.5857)
		if got, want := erf16Poly(x), math.Erf(x); math.Abs(got-want) > bound*want {
			t.Errorf("erf16Poly(%v) = %v; want %v", x, got, want)
		}

		x = 0.125 * rnd.Float64() // [0, 1/8)
		c := &erf16SmallCoeffs
		u := x * x
		got := x * (c[0] + u*(c[1]+u*(c[2]+u*(c[3]+u*(c[4]+u*c[5])))))
		if want := math.Erf(x); math.Abs(got-want) > bound*want {
			t.Errorf("erf16SmallCoeffs(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat16_Erf(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float16
	}{
		{"small", NewFloat16(0.1)},  // |x| < 1/8
		{"medium", exact16(1.5)},    // 1/8 <= |x| < 2.6
		{"negative", exact16(-1.5)}, // the sign is restored
		{"saturated", exact16(3)},   // the result is 1
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erf())
			}
		})
	}
}

func TestFloat16_Erfc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Erfc(0)},
		{exact16(0x1p-14), math.Erfc(0x1p-14)},
		{exact16(1), math.Erfc(1)},
		{exact16(2), math.Erfc(2)},
	}

	for _, tt := range tests {
		got := tt.x.Erfc()
		if !close16(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(0)},
		{exact16(math.Inf(-1)), exact16(2)},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfc()
		if !eq16(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Erfc(b *testing.B) {
	x := exact16(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfc())
	}
}

func TestFloat16_Erfinv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(-0.75), math.Erfinv(-0.75)},
		{exact16(-0.5), math.Erfinv(-0.5)},
		{exact16(-0.25), math.Erfinv(-0.25)},
		{exact16(0), math.Erfinv(0)},
		{exact16(0.25), math.Erfinv(0.25)},
		{exact16(0.5), math.Erfinv(0.5)},
		{exact16(0.75), math.Erfinv(0.75)},
	}

	for _, tt := range tests {
		got := tt.x.Erfinv()
		if !close16(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(1), exact16(math.Inf(1))},
		{exact16(-1), exact16(math.Inf(-1))},
		{exact16(2), exact16(math.NaN())},
		{exact16(-2), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfinv()
		if !eq16(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Erfinv(b *testing.B) {
	x := exact16(0.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfinv())
	}
}

func TestFloat16_Erfcinv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0.25), math.Erfcinv(0.25)},
		{exact16(0.5), math.Erfcinv(0.5)},
		{exact16(0.75), math.Erfcinv(0.75)},
		{exact16(1), math.Erfcinv(1)},
		{exact16(1.25), math.Erfcinv(1.25)},
		{exact16(1.5), math.Erfcinv(1.5)},
		{exact16(1.75), math.Erfcinv(1.75)},
	}

	for _, tt := range tests {
		got := tt.x.Erfcinv()
		if !close16(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(0), exact16(math.Inf(1))},
		{exact16(2), exact16(math.Inf(-1))},
		{exact16(3), exact16(math.NaN())},
		{exact16(-1), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfcinv()
		if !eq16(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Erfcinv(b *testing.B) {
	x := exact16(0.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfcinv())
	}
}
