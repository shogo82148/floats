package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Exp(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-88), math.Exp(-88)},
		{exact32(-1), math.Exp(-1)},
		{exact32(0), math.Exp(0)},
		{exact32(1), math.Exp(1)},
		{exact32(2), math.Exp(2)},
		{exact32(3), math.Exp(3)},
		{exact32(4), math.Exp(4)},
		{exact32(88), math.Exp(88)},
	}

	for _, tt := range tests {
		got := tt.x.Exp()
		if !close32(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		{exact32(0), exact32(1)},
		{exact32(0x1p-15), exact32(1 + 0x1p-15)},
		{exact32(-0x1p-15), exact32(1 - 0x1p-15)},
		{exact32(0x1.62e44p+06), exact32(math.Inf(1))},
		{exact32(-0x1.9fe36ap+06), exact32(0)},

		// tiny arguments
		{NewFloat32FromBits(0x337fffff), exact32(1)},
		{NewFloat32FromBits(0x33800000), NewFloat32FromBits(0x3f800001)},
		{NewFloat32FromBits(0xb3000000), exact32(1)},
		{NewFloat32FromBits(0xb3000001), NewFloat32FromBits(0x3f7fffff)},

		// overflow
		{NewFloat32FromBits(0x42b17217), NewFloat32FromBits(0x7f7fff84)},
		{NewFloat32FromBits(0x42b17218), exact32(math.Inf(1))},

		// underflow
		{NewFloat32FromBits(0xc2cff1b4), NewFloat32FromBits(0x00000001)},
		{NewFloat32FromBits(0xc2cff1b5), exact32(0)},

		// hard-to-round cases, found by checking all Float32 values with math.Exp in float64
		{NewFloat32FromBits(0x377eff81), NewFloat32FromBits(0x3f800080)},
		{NewFloat32FromBits(0x38e69cc1), NewFloat32FromBits(0x3f80039a)},
		{NewFloat32FromBits(0x39c6be5b), NewFloat32FromBits(0x3f800c6d)},
		{NewFloat32FromBits(0xbae0e25c), NewFloat32FromBits(0x3f7f8fa7)},
		{NewFloat32FromBits(0xbbf0edf1), NewFloat32FromBits(0x3f7e1fe9)},
		{NewFloat32FromBits(0xc16912cd), NewFloat32FromBits(0x34fd331b)},

		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(0)},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Exp()
		if !eq32(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_ExpRandom compares Exp with math.Exp on random inputs.
// math.Exp rounded to float32 is correctly rounded except for rare cases,
// so the results should almost always match.
func TestFloat32_ExpRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"uniform", func() Float32 { return NewFloat32(r.Float64()*200 - 105) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-26, 6]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(33)+bias32-26)<<shift32)
		}},
	}
	for _, g := range gens {
		var mismatch int
		for range 300000 {
			x := g.gen()
			if x.IsNaN() {
				continue
			}
			got := x.Exp()
			want := NewFloat32(math.Exp(float64(x)))
			if !within1ulp32(got, want) {
				t.Fatalf("Exp(%v) = %v; want %v", x, got, want)
			}
			if !eq32(got, want) {
				mismatch++
			}
		}
		// Checking all Float32 values on arm64, Exp is correctly rounded,
		// and so is math.Exp rounded to float32.
		if mismatch > 3 {
			t.Errorf("%s: %d results are different from math.Exp", g.name, mismatch)
		}
	}
}

func BenchmarkFloat32_Exp(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Exp())
	}
}

func TestFloat32_Exp2(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-1), math.Exp2(-1)},
		{exact32(0), math.Exp2(0)},
		{exact32(1), math.Exp2(1)},
		{exact32(1.5), math.Exp2(1.5)},
	}

	for _, tt := range tests {
		got := tt.x.Exp2()
		if !close32(got, tt.want) {
			t.Errorf("Exp2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		{exact32(128), exact32(math.Inf(1))}, // overflow
		{exact32(-159), exact32(0)},          // underflow

		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(0)},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Exp2()
		if !eq32(got, tt.want) {
			t.Errorf("Exp2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
