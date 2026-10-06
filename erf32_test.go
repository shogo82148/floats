package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Erf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0), math.Erf(0)},
		{exact32(1), math.Erf(1)},
		{exact32(2), math.Erf(2)},
		{exact32(3), math.Erf(3)},
		{exact32(4), math.Erf(4)},
		{exact32(11), math.Erf(11)},
	}

	for _, tt := range tests {
		got := tt.x.Erf()
		if !close32(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(1)},
		{exact32(math.Inf(-1)), exact32(-1)},
		{exact32(math.NaN()), exact32(math.NaN())},
		{NewFloat32FromBits(0x7f80_0001), NewFloat32FromBits(0x7f80_0001)}, // the NaN with the smallest payload
		{NewFloat32FromBits(0xffff_ffff), NewFloat32FromBits(0xffff_ffff)},
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{NewFloat32FromBits(1), NewFloat32FromBits(1)},
		{NewFloat32FromBits(0x8000_0001), NewFloat32FromBits(0x8000_0001)},
		{exact32(100), exact32(1)},
		{exact32(-100), exact32(-1)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erf()
		if !eq32(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_ErfHardCases checks Erf on the inputs whose results are very close to the midpoint of two adjacent
// Float32 values. They are found by checking all Float32 values with math.Erf in float64,
// and the correctly rounded results were calculated with mpmath.
func TestFloat32_ErfHardCases(t *testing.T) {
	t.Parallel()
	// x, and the correctly rounded Erf(x)
	tests := [][2]uint32{
		{0x0010fe88, 0x00132d0d},
		{0x001efabf, 0x0022f4e3},
		{0x002fd79c, 0x0035fbf2},
		{0x0045ef19, 0x004ee97b},
		{0x004bd132, 0x00558cef},
		{0x004d62e6, 0x00575235},
		{0x00aa320c, 0x00c00b88},
		{0x04aa320c, 0x04c00b88},
		{0x0faa320c, 0x0fc00b88},
		{0x102a320c, 0x10400b88},
		{0x358e9d2d, 0x35a0ec32},
		{0x36aea466, 0x36c51007},
		{0x3940ead6, 0x3959af14},
		{0x39c0057d, 0x39d8ac49},
		{0x3e09737d, 0x3e1a2ba7},
		{0x3e1cd895, 0x3e2f9b7f},
		{0x3e1fcc60, 0x3e32dc2d},
		{0x3e2f129d, 0x3e43a3ee},
		{0x3e35da21, 0x3e4b0fa7},
		{0x3e7165dc, 0x3e85b677},
		{0x3e7ee7ff, 0x3e8ce635},
		{0x3e97e551, 0x3ea67f65},
		{0x3e993b49, 0x3ea7e07c},
		{0x3ece26ec, 0x3edca2b4},
		{0x3eeb4a53, 0x3ef7ef0c},
		{0x3eec27cb, 0x3ef8b938},
		{0x3f043a75, 0x3f08eed3},
		{0x3f22767a, 0x3f216b2f},
		{0x3f44ddfd, 0x3f3923f8},
		{0x3f4938c4, 0x3f3bd333},
		{0x3f4a82f1, 0x3f3c9b41},
		{0x3f5a1c26, 0x3f4591e7},
		{0x3f660690, 0x3f4bd202},
		{0x3f6853a3, 0x3f4cf818},
		{0x3f79b8b3, 0x3f550f8a},
		{0x3fde4ac1, 0x3f7c6745},
		{0x3fe46451, 0x3f7d064c},
		{0x3ffafaa7, 0x3f7e93f2},
		// the threshold where erf(x) is rounded to 1
		{0x407ad444, 0x3f7fffff},
		{0x407ad445, 0x3f800000},
		{0x407ad446, 0x3f800000},
	}
	for _, tt := range tests {
		for _, neg := range []bool{false, true} {
			x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
			if neg {
				x, want = -x, -want
			}
			if got := x.Erf(); !eq32(got, want) {
				t.Errorf("Erf(%v) = %v; want %v", x, got, want)
			}
		}
	}
}

// TestFloat32_ErfBoundaries checks Erf on the both sides of the boundaries of the segments of the calculation.
func TestFloat32_ErfBoundaries(t *testing.T) {
	t.Parallel()
	for _, x := range []float32{0x1p-149, 0x1p-126, 0x1p-24, 0.125, 0.25, 0.5, 1, 2, 3.9192, 4} {
		for d := -3; d <= 3; d++ {
			for _, neg := range []bool{false, true} {
				a := NewFloat32FromBits(math.Float32bits(x) + uint32(d))
				if neg {
					a = -a
				}
				if got, want := a.Erf(), NewFloat32(math.Erf(float64(a))); !eq32(got, want) {
					t.Errorf("Erf(%v) = %v; want %v", a, got, want)
				}
			}
		}
	}
}

// TestFloat32_ErfRandom compares Erf with math.Erf on random inputs.
// math.Erf rounded to Float32 is correctly rounded for all Float32 values (it was checked
// for the arguments whose results are close to the midpoints with mpmath), so the results must match.
func TestFloat32_ErfRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"uniform", func() Float32 { return NewFloat32(r.Float64()*8 - 4) }},
		{"positive", func() Float32 { return NewFloat32(r.Float64() * 4) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-40, 3]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(44)+bias32-40)<<shift32)
		}},
	}
	for _, g := range gens {
		for range 300000 {
			x := g.gen()
			got, want := x.Erf(), NewFloat32(math.Erf(float64(x)))
			if !eq32(got, want) && !(got.IsNaN() && want.IsNaN()) {
				t.Fatalf("%s: Erf(%v) = %v; want %v", g.name, x, got, want)
			}
		}
	}
}

// TestFloat32_ErfPoly checks the polynomial of Erf, whose errors are hidden by the rounding to Float32, with math.Erf.
func TestFloat32_ErfPoly(t *testing.T) {
	t.Parallel()
	const bound = 0x1p-38 // the relative error of the polynomial is less than 2**-39
	r := rand.New(rand.NewPCG(3, 4))
	for range 200000 {
		x := 0.125 + 3.79*r.Float64() // [1/8, 3.915)
		if got, want := erf32Poly(x), math.Erf(x); math.Abs(got-want) > bound*want {
			t.Errorf("erf32Poly(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat32_Erf(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float32
	}{
		{"small", NewFloat32(0.1)},  // |x| < 1/8
		{"medium", exact32(1.5)},    // 1/8 <= |x| < 3.92
		{"negative", exact32(-1.5)}, // the sign is restored
		{"saturated", exact32(5)},   // the result is 1
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erf())
			}
		})
	}
}

func TestFloat32_Erfc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0), math.Erfc(0)},
		{exact32(0x1p-14), math.Erfc(0x1p-14)},
		{exact32(1), math.Erfc(1)},
		{exact32(2), math.Erfc(2)},
	}

	for _, tt := range tests {
		got := tt.x.Erfc()
		if !close32(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(0)},
		{exact32(math.Inf(-1)), exact32(2)},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfc()
		if !eq32(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Erfc(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfc())
	}
}

func TestFloat32_Erfinv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-0.75), math.Erfinv(-0.75)},
		{exact32(-0.5), math.Erfinv(-0.5)},
		{exact32(-0.25), math.Erfinv(-0.25)},
		{exact32(0), math.Erfinv(0)},
		{exact32(0.25), math.Erfinv(0.25)},
		{exact32(0.5), math.Erfinv(0.5)},
		{exact32(0.75), math.Erfinv(0.75)},
	}

	for _, tt := range tests {
		got := tt.x.Erfinv()
		if !close32(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(1), exact32(math.Inf(1))},
		{exact32(-1), exact32(math.Inf(-1))},
		{exact32(2), exact32(math.NaN())},
		{exact32(-2), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfinv()
		if !eq32(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Erfinv(b *testing.B) {
	x := exact32(0.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfinv())
	}
}

func TestFloat32_Erfcinv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0.25), math.Erfcinv(0.25)},
		{exact32(0.5), math.Erfcinv(0.5)},
		{exact32(0.75), math.Erfcinv(0.75)},
		{exact32(1), math.Erfcinv(1)},
		{exact32(1.25), math.Erfcinv(1.25)},
		{exact32(1.5), math.Erfcinv(1.5)},
		{exact32(1.75), math.Erfcinv(1.75)},
	}

	for _, tt := range tests {
		got := tt.x.Erfcinv()
		if !close32(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(math.Inf(1))},
		{exact32(2), exact32(math.Inf(-1))},
		{exact32(3), exact32(math.NaN())},
		{exact32(-1), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfcinv()
		if !eq32(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Erfcinv(b *testing.B) {
	x := exact32(0.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfcinv())
	}
}
