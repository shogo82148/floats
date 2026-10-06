package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat16_Lgamma(t *testing.T) {
	t.Parallel()
	tests := []float64{-2.5, -0.5, 0.5, 1, 1.5, 2, 2.5, 3, 100}

	for _, x := range tests {
		want, wantSign := math.Lgamma(x)
		got, sign := exact16(x).Lgamma()
		if !close16(got, want) || sign != wantSign {
			t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", x, got, sign, want, wantSign)
		}
	}
}

// TestFloat16_LgammaAll checks Lgamma for all bit patterns of Float16.
// For every finite Float16 value, the exact Lgamma is either exactly zero or at least 2**-24 ulp away
// from the midpoint of two adjacent Float16 values (checked with mpmath),
// so math.Lgamma rounded to Float16 is correctly rounded.
func TestFloat16_LgammaAll(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		x := NewFloat16FromBits(uint16(i))
		got, sign := x.Lgamma()
		w, wantSign := math.Lgamma(x.Float64().BuiltIn())
		want := NewFloat16(w)
		if uint16(got) != uint16(want) || sign != wantSign {
			t.Errorf("Lgamma(%#04x) = (%#04x, %d); want (%#04x, %d)", i, uint16(got), sign, uint16(want), wantSign)
		}
	}
}

// TestFloat16_LgammaApprox checks the approximation of Lgamma for all bit patterns of Float16.
// It must be accurate if it is reported as accurate.
func TestFloat16_LgammaApprox(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a := NewFloat16FromBits(uint16(i))
		ix := a &^ signMask16
		if ix >= uvinf16 || a&signMask16 == 0 && ix >= 0x6ffe {
			continue
		}
		y, sign, ok := lgamma16Approx(a, ix)
		want, wantSign := math.Lgamma(a.Float64().BuiltIn())
		if sign != wantSign {
			t.Errorf("lgamma16Approx(%#04x): the sign is %d; want %d", i, sign, wantSign)
		}
		if !ok || math.IsInf(want, 0) {
			continue
		}
		if math.Abs(y-want) > 0x1p-28*math.Abs(want) {
			t.Errorf("lgamma16Approx(%#04x) = %v; want %v", i, y, want)
		}
	}
}

// TestFloat16_LgammaRound checks that lgamma16Round falls back to the math package
// when the approximation is close to the midpoint of two adjacent Float16 values.
func TestFloat16_LgammaRound(t *testing.T) {
	t.Parallel()
	a := exact16(1.5) // Lgamma(1.5) = -0.12078... is not close to the midpoints in the test.
	slow, _ := a.Lgamma()
	const mid = 1 + 0x1p-11 // the midpoint of 1 and 1 + 2**-10
	ulp := 0x1p-52
	for _, tt := range []struct {
		y    float64
		fall bool
	}{
		{mid, true},
		{mid + 0x1p27*ulp, true},
		{mid - 0x1p27*ulp, true},
		{mid + 0x1p29*ulp, false},
		{mid - 0x1p29*ulp, false},
	} {
		want := NewFloat16(tt.y)
		if tt.fall {
			want = slow
		}
		if got := lgamma16Round(a, tt.y); got != want {
			t.Errorf("lgamma16Round(%v, %v) = %v; want %v", a, tt.y, got, want)
		}
	}
}

// TestFloat16_LgammaPoly checks the polynomials of Lgamma, which are not fully tested by TestFloat16_LgammaAll
// because the rounding to Float16 hides their errors, with the math package.
func TestFloat16_LgammaPoly(t *testing.T) {
	t.Parallel()
	rnd := rand.New(rand.NewPCG(1, 2))
	for range 100000 {
		x := 0.5 * math.Exp2(rnd.Float64()*14) // [1/2, 8192)
		if x >= 8184 {
			continue
		}
		want, _ := math.Lgamma(x)
		got := lgamma16Poly(x)
		if want == 0 {
			continue
		}
		if math.Abs(got-want) > 0x1p-38*math.Abs(want) {
			t.Errorf("lgamma16Poly(%v) = %v; want %v", x, got, want)
		}

		x = 0.5 * rnd.Float64() // [0, 1/2)
		want, _ = math.Lgamma(1 + x)
		if got := lgamma16Small(x); math.Abs(got-want) > 0x1p-35 {
			t.Errorf("lgamma16Small(%v) = %v; want %v", x, got, want)
		}

		x = 0x1p-8 * rnd.Float64() // [0, 2**-8)
		want, _ = math.Lgamma(1 + x)
		if got := lgamma16Tiny(x); math.Abs(got-want) > 0x1p-40 {
			t.Errorf("lgamma16Tiny(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat16_Lgamma(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float16
	}{
		{"tiny", NewFloat16(0.001)},        // Lgamma(x) ~ -log(x)
		{"small", exact16(0.375)},          // 0 < x < 1/2
		{"medium", exact16(1.5)},           // 1/2 <= x < 3
		{"large", exact16(10.5)},           // 3 <= x < 8184
		{"huge", exact16(5000)},            // close to the overflow
		{"overflow", exact16(8192)},        // +Inf
		{"negative", NewFloat16(-2.5)},     // -2048 < x < 0
		{"verynegative", exact16(-100.25)}, // Lgamma(x) < 0
		{"nearzero", exact16(-2.5)},        // the result is close to zero
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				v, s := tt.x.Lgamma()
				runtime.KeepAlive(v)
				runtime.KeepAlive(s)
			}
		})
	}
}
