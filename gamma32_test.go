package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Gamma(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-0.5), math.Gamma(-0.5)},
		{exact32(0.5), math.Gamma(0.5)},
		{exact32(1), math.Gamma(1)},
		{exact32(1.5), math.Gamma(1.5)},
		{exact32(2), math.Gamma(2)},
		{exact32(2.5), math.Gamma(2.5)},
		{exact32(3), math.Gamma(3)},
	}

	for _, tt := range tests {
		got := tt.x.Gamma()
		if !close32(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(0), exact32(math.Inf(1))},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(-1))},
		{exact32(-1), exact32(math.NaN())},
		{exact32(-2), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Gamma()
		if !eq32(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_GammaHardCases checks Gamma on the inputs whose results are very close to the midpoint of two adjacent
// Float32 values. They are found by checking all Float32 values with math.Gamma in float64,
// and the correctly rounded results were calculated with mpmath.
func TestFloat32_GammaHardCases(t *testing.T) {
	t.Parallel()
	// x, and the correctly rounded Gamma(x)
	tests := [][2]uint32{
		{0x27dc23dd, 0x5714d9c6},
		{0x27de86a9, 0x57134133},
		{0x27e05475, 0x57121211},
		{0x27e1368b, 0x57117f6e},
		{0x27e25473, 0x5710c7a2},
		{0x3c7bb570, 0x42810ec8},
		{0x27d25d25, 0x571bc4aa},
		{0x27d43d43, 0x571a644b},
		{0x27ec7ec7, 0x570a8e84},
		{0x27ee4bc5, 0x57098279},
		{0x27f00ff1, 0x57087f77},
		{0x27f02a3b, 0x57087086},
		{0x27f0ff0f, 0x5707f808},
		{0x29de9bd0, 0x55133335},
		{0xb4cec13b, 0xca1e7cbf},
		{0xba268ee2, 0xc4c4cef0},
		{0xb2b278ac, 0xcc379a8a},
		{0xa7c100c1, 0xd729c7a0},
		{0xaa718c03, 0xd487a8b0},
		{0xb6072a66, 0xc8f26dc5},
		{0xa96401c8, 0xd58fb705},
		{0xb2cbe92d, 0xcc20b29c},
		{0xb50b4c06, 0xc9eb3d0e},
		{0xbd745227, 0xc18b3c6a},
		{0xaf85faf4, 0xcf7492c3},
		{0xb368449e, 0xcb8d1412},
		{0xa912abab, 0xd5df6988},
		{0xb7ae08f6, 0xc73c4946},
	}
	for _, tt := range tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		// The results may not be correctly rounded if math.Gamma is not, but they must be within 1 ulp.
		if got := x.Gamma(); !within1ulp32(got, want) {
			t.Errorf("Gamma(%v) = %v; want %v", x, got, want)
		}
	}

	// math.Gamma rounded to Float32 is not correctly rounded for these arguments, but Gamma is.
	for _, tt := range [][2]uint32{
		{0x27de86a9, 0x57134133},
		{0x27e05475, 0x57121211},
		{0x41e886d1, 0x709989b5},
	} {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Gamma(); !eq32(got, want) {
			t.Errorf("Gamma(%v) = %v; want %v", x, got, want)
		}
	}
}

// TestFloat32_GammaBoundaries checks Gamma around the boundaries of the algorithm.
func TestFloat32_GammaBoundaries(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name string
		x    float64
	}{
		{"overflow of 1/x", 0x1p-128},
		{"tiny", 0x1p-36},
		{"between the polynomial and 1/x", 1},
		{"overflow", 35.04},
		{"the limit of the polynomial", 35.1},
		{"underflow to subnormal numbers", -33.5},
		{"underflow to zero", -41.5},
		{"the limit of the polynomial for x < 0", -43},
		{"subnormal results", -35.5},
		{"around poles", -10},
		{"around poles", -30},
		{"around poles", -40},
		{"min normal", -0x1p-126},
	} {
		for _, sign := range []float64{1, -1} {
			if c.x < 0 && sign < 0 {
				continue
			}
			base := NewFloat32(sign * c.x).Bits()
			for d := range uint32(64) {
				// the arguments next to the boundary on both sides, including the ones of the opposite sign
				for _, b := range []uint32{base + d - 32, base - d + 32} {
					x := NewFloat32FromBits(b)
					got, want := x.Gamma(), NewFloat32(math.Gamma(float64(x)))
					if !within1ulp32(got, want) && !(got.IsNaN() && want.IsNaN()) {
						t.Errorf("%s: Gamma(%v) = %v; want %v", c.name, x, got, want)
					}
				}
			}
		}
	}
}

// TestFloat32_GammaRanges compares Gamma with math.Gamma for all the Float32 values in the ranges,
// where the results are close to the boundaries of the normal numbers and the overflow, or very small:
// the results for x in [-42, -32] are subnormal numbers or close to them, and those for x in [32, 35.2] are close to the maximum.
func TestFloat32_GammaRanges(t *testing.T) {
	t.Parallel()
	for _, r := range []struct {
		name   string
		lo, hi float32
	}{
		{"negative", -42, -32},
		{"large", 32, 35.2},
	} {
		lo, hi := math.Float32bits(r.lo), math.Float32bits(r.hi)
		if r.lo < 0 {
			lo, hi = hi, lo // the bits of the negative numbers decrease as they increase
		}
		var mismatch int
		for b := lo; b <= hi; b++ {
			x := NewFloat32FromBits(b)
			got, want := x.Gamma(), NewFloat32(math.Gamma(float64(x)))
			if !within1ulp32(got, want) && !(got.IsNaN() && want.IsNaN()) {
				t.Fatalf("%s: Gamma(%v) = %v; want %v", r.name, x, got, want)
			}
			if !eq32(got, want) {
				mismatch++
			}
		}
		// Checking all Float32 values, only 3 results are different.
		if mismatch > 1 {
			t.Errorf("%s: %d results are different from math.Gamma", r.name, mismatch)
		}
	}
}

// TestFloat32_GammaRandom compares Gamma with math.Gamma on random inputs.
// math.Gamma rounded to Float32 is correctly rounded except for rare cases,
// so the results should almost always match.
func TestFloat32_GammaRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"uniform", func() Float32 { return NewFloat32(r.Float64()*80 - 44) }},
		{"positive", func() Float32 { return NewFloat32(r.Float64() * 36) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-40, 5]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(46)+bias32-40)<<shift32)
		}},
	}
	for _, g := range gens {
		var mismatch int
		for range 300000 {
			x := g.gen()
			got, want := x.Gamma(), NewFloat32(math.Gamma(float64(x)))
			if !within1ulp32(got, want) && !(got.IsNaN() && want.IsNaN()) {
				t.Fatalf("Gamma(%v) = %v; want %v", x, got, want)
			}
			if !eq32(got, want) {
				mismatch++
			}
		}
		// Checking all Float32 values, only 3 results are different.
		if mismatch > 1 {
			t.Errorf("%s: %d results are different from math.Gamma", g.name, mismatch)
		}
	}
}

// TestFloat32_GammaPoly checks the polynomials of Gamma, whose errors are hidden by the rounding to Float32, with math.Gamma.
func TestFloat32_GammaPoly(t *testing.T) {
	t.Parallel()
	const bound = 0x1p-34 // the relative error of the polynomial is less than 2**-35
	rnd := rand.New(rand.NewPCG(1, 2))
	for range 200000 {
		x := 1 + 43*rnd.Float64() // [1, 44)
		if got, want := gamma32Poly(x), math.Gamma(x); math.Abs(got-want) > bound*want {
			t.Errorf("gamma32Poly(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat32_Gamma(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float32
	}{
		{"tiny", NewFloat32(1e-20)},         // Gamma(x) ~ 1/x
		{"medium", NewFloat32(1.5)},         // 1 <= x < 2
		{"large", NewFloat32(20.5)},         // 2 <= x < 35.1
		{"negative", NewFloat32(-2.5)},      // -43 < x < 0
		{"verynegative", NewFloat32(-35.5)}, // reflection with large |x|
		{"overflow", NewFloat32(50.5)},      // +Inf
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Gamma())
			}
		})
	}
}
