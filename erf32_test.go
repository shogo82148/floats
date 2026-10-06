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
		{exact32(-1), math.Erfc(-1)},
		{exact32(-3), math.Erfc(-3)},
		{exact32(5), math.Erfc(5)},
		{exact32(9), math.Erfc(9)},
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
		{exact32(0), exact32(1)},
		{exact32(math.Copysign(0, -1)), exact32(1)},

		// erfc(x) is rounded to 0 for x >= 10.0541
		{NewFloat32FromBits(0x4120ddfb), NewFloat32FromBits(0x00000001)},
		{NewFloat32FromBits(0x4120ddfc), exact32(0)},
		{NewFloat32FromBits(0x7f7fffff), exact32(0)},

		// erfc(x) is rounded to 2 for x <= -3.8325
		{NewFloat32FromBits(0xc07547ca), NewFloat32FromBits(0x3fffffff)},
		{NewFloat32FromBits(0xc07547cb), exact32(2)},
		{NewFloat32FromBits(0xff7fffff), exact32(2)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfc()
		if !eq32(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_ErfcHardCases checks Erfc on the inputs whose results are very close to the midpoint of two adjacent
// Float32 values, which are found by checking all Float32 values with math.Erfc in float64.
// The correctly rounded results were calculated with mpmath.
func TestFloat32_ErfcHardCases(t *testing.T) {
	t.Parallel()
	// x, and the correctly rounded Erfc(x)
	tests := [][2]uint32{
		{0x32e2d197, 0x3f800000},
		{0x32e2e907, 0x3f7fffff},
		{0x34d4b27d, 0x3f7ffff8},
		{0x3886353e, 0x3f7ffb45},
		{0x3941c770, 0x3f7ff256},
		{0x3b71cde7, 0x3f7eef27},
		{0x3c2e50df, 0x3f7ced41},
		{0x3dabbd85, 0x3f67d549},
		{0x3f26187d, 0x3eb7bb42},
		{0x40bc972b, 0x24b34d41},
		{0xb362e41b, 0x3f800001},
		{0xb42a273e, 0x3f800001},
		{0xb594e30f, 0x3f80000b},
		{0xb6512676, 0x3f80001e},
		{0xb70eaec0, 0x3f800051},
		{0xb76c9f62, 0x3f800085},
		{0xb7ca80be, 0x3f8000e5},
		{0xb9fcef4a, 0x3f8011d6},
		{0xbab8309f, 0x3f8033f6},
		{0xbceedc0a, 0x3f8435cb},
		{0xbe67faf8, 0x3fa02b43},
		{0xbfdde4a2, 0x3ffe2df5},
	}
	for _, tt := range tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Erfc(); !eq32(got, want) {
			t.Errorf("Erfc(%v) = %v; want %v", x, got, want)
		}
	}
}

// TestFloat32_ErfcBoundaries checks Erfc on the both sides of the boundaries of the segments of the calculation.
func TestFloat32_ErfcBoundaries(t *testing.T) {
	t.Parallel()
	for _, x := range []float32{0x1p-149, 0x1p-126, 0x1p-25, 0x1p-24, 1.0 / 32, 0.125, 0.25, 1, 2, 3.8325, 4, 9.1875, 10.0541} {
		for d := -3; d <= 3; d++ {
			for _, neg := range []bool{false, true} {
				a := NewFloat32FromBits(math.Float32bits(x) + uint32(d))
				if neg {
					a = -a
				}
				if got, want := a.Erfc(), NewFloat32(math.Erfc(float64(a))); !eq32(got, want) {
					t.Errorf("Erfc(%v) = %v; want %v", a, got, want)
				}
			}
		}
	}
}

// TestFloat32_ErfcRandom compares Erfc with math.Erfc on random inputs.
// math.Erfc rounded to Float32 is correctly rounded for all Float32 values but one (0xb76c9f62, which is
// checked by TestFloat32_ErfcHardCases), so the results must match.
func TestFloat32_ErfcRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"uniform", func() Float32 { return NewFloat32(r.Float64()*20 - 6) }},
		{"positive", func() Float32 { return NewFloat32(r.Float64() * 10.1) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-40, 4]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(45)+bias32-40)<<shift32)
		}},
	}
	for _, g := range gens {
		for range 300000 {
			x := g.gen()
			got, want := x.Erfc(), NewFloat32(math.Erfc(float64(x)))
			if !eq32(got, want) && !(got.IsNaN() && want.IsNaN()) {
				t.Fatalf("%s: Erfc(%v) = %v; want %v", g.name, x, got, want)
			}
		}
	}
}

// TestFloat32_ErfcPoly checks the polynomials of Erfc, whose errors are hidden by the rounding to Float32, with math.Erfc.
func TestFloat32_ErfcPoly(t *testing.T) {
	t.Parallel()
	const bound = 0x1p-40 // the relative error of the polynomial is less than 2**-41
	r := rand.New(rand.NewPCG(3, 4))
	for range 200000 {
		x := 10.05 * r.Float64() // [0, 10.05)
		if got, want := erfc32Poly(x), math.Erfc(x); math.Abs(got-want) > bound*want {
			t.Errorf("erfc32Poly(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat32_Erfc(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float32
	}{
		{"small", NewFloat32(0.01)}, // close to 1
		{"medium", exact32(1.5)},    // 0 < x < 10.05
		{"large", exact32(8.5)},     // 2**-100
		{"negative", exact32(-1.5)}, // 2 - erfc(-x)
		{"saturated", exact32(12)},  // the result is 0
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erfc())
			}
		})
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
		{exact32(math.Inf(1)), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{NewFloat32FromBits(0x3f800001), exact32(math.NaN())},
		{NewFloat32FromBits(0xbf800001), exact32(math.NaN())},
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{NewFloat32FromBits(0x3f7fffff), NewFloat32FromBits(math.Float32bits(float32(math.Erfinv(float64(math.Float32frombits(0x3f7fffff))))))},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfinv()
		if !eq32(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_ErfinvHardCases checks Erfinv on the inputs whose results are very close to the midpoint of two adjacent
// Float32 values. They are found by checking all Float32 values with math.Erfinv in float64,
// and the correctly rounded results were calculated with mpmath.
func TestFloat32_ErfinvHardCases(t *testing.T) {
	t.Parallel()
	// x, and the correctly rounded Erfinv(x)
	tests := [][2]uint32{
		{0x00000cf6, 0x00000b7d},
		{0x005dbd7c, 0x00531337},
		{0x01e9dc5e, 0x01cf40f6},
		{0x04d8e690, 0x04c03922},
		{0x07c8151e, 0x07b15189},
		{0x0ab71ea9, 0x0aa24922},
		{0x0da64dde, 0x0d93621c},
		{0x10955769, 0x108459b5},
		{0x13846b59, 0x136ab507},
		{0x167398b8, 0x1657e1bf},
		{0x1962a2ea, 0x1948d9ec},
		{0x1c51ac75, 0x1c39d185},
		{0x1f40b600, 0x1f2ac91e},
		{0x222fe535, 0x221be218},
		{0x251eeec0, 0x250cd9b1},
		{0x280e1437, 0x27fbd413},
		{0x2afd546b, 0x2ae081f6},
		{0x2dec83a0, 0x2dd19af0},
		{0x30db8d2b, 0x30c29289},
		{0x33ca96b6, 0x33b38a22},
		{0x36b84dee, 0x36a355e5},
		{0x39a50efc, 0x39924782},
		{0x3bba61fd, 0x3ba52dc8},
		{0x3c95bac3, 0x3c84b4b9},
		{0x3f7f8ec2, 0x400dcd2e},
	}
	for _, tt := range tests {
		for _, neg := range []bool{false, true} {
			x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
			if neg {
				x, want = -x, -want
			}
			if got := x.Erfinv(); !eq32(got, want) {
				t.Errorf("Erfinv(%v) = %v; want %v", x, got, want)
			}
		}
	}
}

// TestFloat32_ErfinvBoundaries checks Erfinv on the both sides of the boundaries of the segments of the calculation.
func TestFloat32_ErfinvBoundaries(t *testing.T) {
	t.Parallel()
	for _, x := range []float32{0x1p-149, 0x1p-126, 0x1p-24, 0.125, 0.25, 0.5, 0.75, 0.875, 1 - 0x1p-10, 1 - 0x1p-20, 1 - 0x1p-24} {
		for d := -3; d <= 3; d++ {
			for _, neg := range []bool{false, true} {
				a := NewFloat32FromBits(math.Float32bits(x) + uint32(d))
				if a.Abs().Gt(1) {
					continue
				}
				if neg {
					a = -a
				}
				if got, want := a.Erfinv(), NewFloat32(math.Erfinv(float64(a))); !eq32(got, want) {
					t.Errorf("Erfinv(%v) = %v; want %v", a, got, want)
				}
			}
		}
	}
}

// TestFloat32_ErfinvRandom compares Erfinv with math.Erfinv on random inputs.
// math.Erfinv rounded to Float32 is correctly rounded for all Float32 values but one (0x3bba61fd, which is
// checked by TestFloat32_ErfinvHardCases), so the results must match.
func TestFloat32_ErfinvRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()&^signMask32%0x3f800000 | r.Uint32()&signMask32) }},
		{"uniform", func() Float32 { return NewFloat32(r.Float64()*2 - 1) }},
		{"close-to-one", func() Float32 { return NewFloat32(1 - r.Float64()*0x1p-10) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-40, -1]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(40)+bias32-40)<<shift32)
		}},
	}
	for _, g := range gens {
		for range 300000 {
			x := g.gen()
			got, want := x.Erfinv(), NewFloat32(math.Erfinv(float64(x)))
			if !eq32(got, want) && !(got.IsNaN() && want.IsNaN()) {
				t.Fatalf("%s: Erfinv(%v) = %v; want %v", g.name, x, got, want)
			}
		}
	}
}

// TestFloat32_ErfinvPoly checks the polynomials of Erfinv, whose errors are hidden by the rounding to Float32, with math.Erfinv.
func TestFloat32_ErfinvPoly(t *testing.T) {
	t.Parallel()
	const bound = 0x1p-42 // the relative errors of the polynomials are less than 2**-43
	r := rand.New(rand.NewPCG(3, 4))
	for range 200000 {
		x := 0.125 + 0.375*r.Float64() // [1/8, 1/2)
		if got, want := erfinv16Poly(&erfinv16MidCoeffs[math.Float64bits(x)>>49-8160], x), math.Erfinv(x); math.Abs(got-want) > bound*want {
			t.Errorf("erfinv16Poly(%v) = %v; want %v", x, got, want)
		}

		// t = 1 - x is a multiple of 2**-24 (so that x is exactly representable), and it is distributed uniformly in the exponent: [2**-24, 1/2]
		tt := float64(1+r.Uint32N(1<<(1+r.IntN(23)))) * 0x1p-24
		if got, want := erfinv32Poly(tt), math.Erfinv(1-tt); math.Abs(got-want) > bound*want {
			t.Errorf("erfinv32Poly(%v) = %v; want %v", tt, got, want)
		}
	}
}

func BenchmarkFloat32_Erfinv(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float32
	}{
		{"small", NewFloat32(0.1)},            // |x| < 1/8
		{"medium", exact32(0.25)},             // 1/8 <= |x| < 1/2
		{"large", exact32(0.75)},              // 1/2 <= |x| < 1
		{"close-to-one", NewFloat32(0.99999)}, // the polynomial of the shortest segment
		{"negative", exact32(-0.75)},          // the sign is restored
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erfinv())
			}
		})
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
		{exact32(math.Inf(1)), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(1))},
		{NewFloat32FromBits(0x4000_0001), exact32(math.NaN())},
		{exact32(1), exact32(0)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfcinv()
		if !eq32(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_ErfcinvHardCases checks Erfcinv on the inputs whose results are very close to the midpoint of two adjacent
// Float32 values, and some others. They are found by checking all Float32 values less than 1/2 with the polynomials,
// and the correctly rounded results were calculated with mpmath.
// TestFloat32_ErfcinvHardCasesGreaterThanOne checks Erfcinv(x) for 1 < x < 2 on the inputs whose results are very close to
// the midpoint of two adjacent Float32 values. The correctly rounded results were calculated with mpmath.
func TestFloat32_ErfcinvHardCasesGreaterThanOne(t *testing.T) {
	t.Parallel()
	// x, and the correctly rounded Erfcinv(x)
	tests := [][2]uint32{
		{0x3fff369a, 0xbff7fc48},
		{0x3ff2e4f4, 0xbf93d652},
		{0x3fe3fc33, 0xbf5e9279},
		{0x3fd88202, 0xbf385547},
	}
	for _, tt := range tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Erfcinv(); !eq32(got, want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", x, got, want)
		}
	}
}

func TestFloat32_ErfcinvHardCases(t *testing.T) {
	t.Parallel()
	// x, and the correctly rounded Erfcinv(x)
	tests := [][2]uint32{
		{0x0000168b, 0x411948a4},
		{0x027fab02, 0x4110b229},
		{0x04e90404, 0x410db30e},
		{0x076f1e23, 0x410a87b7},
		{0x09d8ae0c, 0x41076789},
		{0x0c683da5, 0x41040b5d},
		{0x0eda2fde, 0x4100ba5c},
		{0x1144b05b, 0x40fabbe8},
		{0x13c3b590, 0x40f39fc9},
		{0x1645614b, 0x40ec466f},
		{0x18d0f8f0, 0x40e4962d},
		{0x1b47cc56, 0x40dcdea0},
		{0x1dcdb927, 0x40d4b3d5},
		{0x2051d2ff, 0x40cc3d5e},
		{0x22ddd8e7, 0x40c3539c},
		{0x254fcd89, 0x40ba5396},
		{0x27cd7b20, 0x40b0ba57},
		{0x2a4715b6, 0x40a6a80d},
		{0x2cc8c9bf, 0x409bd485},
		{0x2f3a2009, 0x409084e4},
		{0x31ab7827, 0x4084504a},
		{0x34303f1a, 0x406cf3ac},
		{0x36b09334, 0x404e17b9},
		{0x391b046c, 0x402bba3b},
		{0x3ba51b85, 0x3ffdd6c4},
		{0x3e146d41, 0x3f83ed85},
		{0x34800531, 0x4069cfb1},
		{0x35d7f845, 0x4059188a},
		{0x3723f946, 0x40481f57},
		{0x38698174, 0x403665a0},
		{0x39b83668, 0x4021bf76},
		{0x3b091260, 0x400b3ecb},
		{0x3c65043f, 0x3fde7503},
		{0x3da8fb47, 0x3f9d26ef},
		{0x35bbd003, 0x405a5bf6},
		{0x36e1f5a8, 0x404bbbda},
		{0x37eced1d, 0x403d7d82},
		{0x390b6eed, 0x402ce9db},
		{0x3a2e13db, 0x401a0c37},
		{0x3b623aea, 0x4004570c},
		{0x3c9ec729, 0x3fd39ee1},
		{0x3dc666ef, 0x3f96436b},
		{0x3ef682df, 0x3efedaef},
	}
	for _, tt := range tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Erfcinv(); !eq32(got, want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", x, got, want)
		}
		// erfcinv(x) = -erfcinv(2-x) if 2 - x is exactly representable.
		y := NewFloat32(2 - float64(x))
		if x < 0x1p-22 || float64(y) != 2-float64(x) { // 2 - x is exact in float64 for x >= 2**-22
			continue
		}
		if got := y.Erfcinv(); !eq32(got, -want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", y, got, -want)
		}
	}
}

// erfcinv32Reference returns the correctly rounded Erfcinv(x) for 2**-149 <= x <= 2 by Float256.
func erfcinv32Reference(x Float32) Float32 {
	return Float256(uvone256).Sub(NewFloat256(float64(x))).Erfinv().Float32()
}

// TestFloat32_ErfcinvBoundaries checks Erfcinv on the both sides of the boundaries of the segments of the calculation.
func TestFloat32_ErfcinvBoundaries(t *testing.T) {
	t.Parallel()
	for _, x := range []float32{0x1p-149, 0x1p-126, 0x1p-100, 0x1p-25, 0x1p-24, 0.125, 0.25, 0.5, 0.75, 1, 1.5, 2} {
		for d := -3; d <= 3; d++ {
			a := NewFloat32FromBits(math.Float32bits(x) + uint32(d))
			if a.IsNaN() || a.IsInf(0) || a <= 0 || a > 2 {
				continue
			}
			if got, want := a.Erfcinv(), erfcinv32Reference(a); !eq32(got, want) {
				t.Errorf("Erfcinv(%v) = %v; want %v", a, got, want)
			}
		}
	}
}

// TestFloat32_ErfcinvRandom compares Erfcinv with Float256 on random inputs.
func TestFloat32_ErfcinvRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"uniform", func() Float32 { return NewFloat32(r.Float64() * 2) }},
		{"small", func() Float32 { return NewFloat32(r.Float64() * 0.5) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-149, 1)
			return NewFloat32FromBits(r.Uint32()&fracMask32 | uint32(r.IntN(126)+1)<<shift32)
		}},
		{"subnormal", func() Float32 { return NewFloat32FromBits(r.Uint32()&fracMask32 | 1) }},
	}
	for _, g := range gens {
		for range 3000 {
			x := g.gen()
			if x.IsZero() || x.Gt(2) {
				continue
			}
			if got, want := x.Erfcinv(), erfcinv32Reference(x); !eq32(got, want) {
				t.Fatalf("%s: Erfcinv(%v) = %v; want %v", g.name, x, got, want)
			}
		}
	}
}

// TestFloat32_ErfcinvPoly checks the polynomials of Erfcinv, whose errors are hidden by the rounding to Float32, with Float256.
func TestFloat32_ErfcinvPoly(t *testing.T) {
	t.Parallel()
	const bound = 0x1p-44 // the relative error of the polynomial is less than 2**-47
	r := rand.New(rand.NewPCG(3, 4))
	for range 3000 {
		// t is distributed uniformly in the exponent: [2**-149, 2**-24)
		tt := math.Ldexp(1+r.Float64(), -149+r.IntN(125))
		want := Float256(uvone256).Sub(NewFloat256(tt)).Erfinv().Float64().BuiltIn()
		if got := erfcinv32Poly(tt); math.Abs(got-want) > bound*want {
			t.Errorf("erfcinv32Poly(%v) = %v; want %v", tt, got, want)
		}
	}
}

func BenchmarkFloat32_Erfcinv(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float32
	}{
		{"tiny", NewFloat32(1e-30)}, // x < 2**-24
		{"small", exact32(0.25)},    // 2**-24 <= x < 1/2
		{"medium", exact32(0.75)},   // 1/2 <= x <= 1
		{"large", exact32(1.5)},     // 1 < x < 2, the sign is restored
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erfcinv())
			}
		})
	}
}
