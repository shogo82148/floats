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

// TestFloat32_ExpRandom compares Exp, Exp2, and Expm1 with math.Exp, math.Exp2, and math.Expm1 on random inputs.
// The results of the math package rounded to float32 are correctly rounded except for rare cases,
// so the results should almost always match.
func TestFloat32_ExpRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"uniform", func() Float32 { return NewFloat32(r.Float64()*300 - 155) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-26, 7]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(34)+bias32-26)<<shift32)
		}},
	}
	for _, g := range gens {
		var mismatch int
		for range 300000 {
			x := g.gen()
			if x.IsNaN() {
				continue
			}
			f := float64(x)
			for _, fn := range []struct {
				name string
				got  Float32
				want Float32
			}{
				{"Exp", x.Exp(), NewFloat32(math.Exp(f))},
				{"Exp2", x.Exp2(), NewFloat32(math.Exp2(f))},
				{"Expm1", x.Expm1(), NewFloat32(math.Expm1(f))},
			} {
				if !within1ulp32(fn.got, fn.want) {
					t.Fatalf("%s(%v) = %v; want %v", fn.name, x, fn.got, fn.want)
				}
				if !eq32(fn.got, fn.want) {
					mismatch++
				}
			}
		}
		// Checking all Float32 values on arm64, only 1 result is different.
		if mismatch > 3 {
			t.Errorf("%s: %d results are different from the math package", g.name, mismatch)
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

		// overflow
		{NewFloat32FromBits(0x42ffffff), NewFloat32FromBits(0x7f7fffa7)},

		// underflow; 2**-150 is the midpoint of 0 and the smallest subnormal, and rounds to even.
		{exact32(-150), exact32(0)},
		{NewFloat32FromBits(0xc315ffff), NewFloat32FromBits(0x00000001)},
		{exact32(-149), NewFloat32FromBits(0x00000001)},
		{exact32(-127), NewFloat32FromBits(0x00400000)},
		{exact32(-126), NewFloat32FromBits(0x00800000)},

		{exact32(0.5), NewFloat32FromBits(0x3fb504f3)},

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

// TestFloat32_Exp2HardCases checks Exp2 on the inputs whose results are very close to
// the midpoint of two adjacent Float32 values.
// They are found by checking all Float32 values with math.Exp2 in float64.
// The results may not be correctly rounded, but they must be within 1 ulp.
func TestFloat32_Exp2HardCases(t *testing.T) {
	// x, and correctly rounded 2**x
	tests := [][2]uint32{
		{0xbaec2b40, 0x3f7fae34},
		{0xb52d1f9a, 0x3f7ffff8},
		{0xbe1f29de, 0x3f65da56},
		{0x3c02a9ad, 0x3f80b5a3},
		{0xb8d3d026, 0x3f7ffb69},
		{0x33b8aa3b, 0x3f800001},
		{0x3a07857c, 0x3f800bbe},
		{0x36879cf7, 0x3f800018},
		{0x3b429d37, 0x3f804385},
		{0xbcf3a937, 0x3f7ac6b1},
	}
	for _, tt := range tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Exp2(); !within1ulp32(got, want) {
			t.Errorf("Exp2(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat32_Exp2(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Exp2())
	}
}
