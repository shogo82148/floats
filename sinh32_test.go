package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Sinh(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0), math.Sinh(0)},
		{exact32(0.25), math.Sinh(0.25)},
		{exact32(0.5), math.Sinh(0.5)},
		{exact32(1), math.Sinh(1)},
		{exact32(21), math.Sinh(21)},
		{exact32(22), math.Sinh(22)},

		{exact32(-0), -math.Sinh(0)},
		{exact32(-0.25), -math.Sinh(0.25)},
		{exact32(-0.5), -math.Sinh(0.5)},
		{exact32(-1), -math.Sinh(1)},
		{exact32(-21), -math.Sinh(21)},
		{exact32(-22), -math.Sinh(22)},
	}

	for _, tt := range tests {
		got := tt.x.Sinh()
		if !close32(got, tt.want) {
			t.Errorf("Sinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},

		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(math.Inf(-1))},
		{exact32(math.NaN()), exact32(math.NaN())},

		// tiny arguments
		{NewFloat32FromBits(0x00000001), NewFloat32FromBits(0x00000001)},
		{NewFloat32FromBits(0x397fffff), NewFloat32FromBits(0x397fffff)},

		// overflow
		{NewFloat32FromBits(0x42b2d4fc), NewFloat32FromBits(0x7f7fffec)},
		{NewFloat32FromBits(0x42b2d4fd), NewFloat32Inf(1)},
		{NewFloat32FromBits(0xc2b2d4fd), NewFloat32Inf(-1)},
	}

	for _, tt := range strictTests {
		got := tt.x.Sinh()
		if !eq32(got, tt.want) {
			t.Errorf("Sinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Sinh(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Sinh())
	}
}

func TestFloat32_Cosh(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0), math.Cosh(0)},
		{exact32(1), math.Cosh(1)},
		{exact32(2), math.Cosh(2)},
		{exact32(3), math.Cosh(3)},
		{exact32(4), math.Cosh(4)},
		{exact32(22), math.Cosh(22)},
	}

	for _, tt := range tests {
		got := tt.x.Cosh()
		if !close32(got, tt.want) {
			t.Errorf("Cosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(1)},
		{exact32(math.Copysign(0, -1)), exact32(1)},
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(math.Inf(1))},
		{exact32(math.NaN()), exact32(math.NaN())},

		// tiny arguments
		{NewFloat32FromBits(0x397fffff), exact32(1)},

		// overflow
		{NewFloat32FromBits(0x42b2d4fc), NewFloat32FromBits(0x7f7fffec)},
		{NewFloat32FromBits(0x42b2d4fd), NewFloat32Inf(1)},
		{NewFloat32FromBits(0xc2b2d4fd), NewFloat32Inf(1)},
	}

	for _, tt := range strictTests {
		got := tt.x.Cosh()
		if !eq32(got, tt.want) {
			t.Errorf("Cosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Cosh(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Cosh())
	}
}

func TestFloat32_Tanh(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0), math.Tanh(0)},
		{exact32(1), math.Tanh(1)},
		{exact32(2), math.Tanh(2)},
		{exact32(3), math.Tanh(3)},
		{exact32(4), math.Tanh(4)},
		{exact32(45), math.Tanh(45)},
	}

	for _, tt := range tests {
		got := tt.x.Tanh()
		if !close32(got, tt.want) {
			t.Errorf("Tanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.Inf(1)), exact32(1)},
		{exact32(math.Inf(-1)), exact32(-1)},
		{exact32(math.NaN()), exact32(math.NaN())},

		// tiny arguments
		{NewFloat32FromBits(0x397fffff), NewFloat32FromBits(0x397fffff)},

		// tanh(9) = 1 - 3.05e-8 rounds to 1 - 2**-24, and tanh(10) rounds to 1.
		{exact32(9), NewFloat32FromBits(0x3f7fffff)},
		{exact32(10), exact32(1)},
		{exact32(-10), exact32(-1)},
	}

	for _, tt := range strictTests {
		got := tt.x.Tanh()
		if !eq32(got, tt.want) {
			t.Errorf("Tanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Tanh(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Tanh())
	}
}

// TestFloat32_SinhCoshTanhHardCases checks Sinh, Cosh, and Tanh on the inputs
// whose results are very close to the midpoint of two adjacent Float32 values.
// They are found by checking all Float32 values with math.Sinh, math.Cosh, and math.Tanh in float64.
// The results may not be correctly rounded, but they must be within 1 ulp.
func TestFloat32_SinhCoshTanhHardCases(t *testing.T) {
	// x, and correctly rounded sinh(x), cosh(x), and tanh(x)
	tests := [][4]uint32{
		{0x3a1285ff, 0x3a1285ff, 0x3f800001, 0x3a1285fe},
		{0x3a6f7750, 0x3a6f7752, 0x3f800004, 0x3a6f774c},
		{0x3a87c3b6, 0x3a87c3b8, 0x3f800004, 0x3a87c3b3},
		{0x3b36aa1f, 0x3b36aa2f, 0x3f800021, 0x3b36aa00},
		{0x3bc8b605, 0x3bc8b657, 0x3f80009d, 0x3bc8b561},
		{0x3c46f746, 0x3c46f887, 0x3f80026b, 0x3c46f4c5},
		{0x3d09c4d8, 0x3d09cb7f, 0x3f80128a, 0x3d09b78d},
		{0x3d609528, 0x3d60b1f8, 0x3f803145, 0x3d605b9d},
	}
	for _, tt := range tests {
		for _, neg := range []bool{false, true} {
			x := NewFloat32FromBits(tt[0])
			wantSinh, wantCosh, wantTanh := NewFloat32FromBits(tt[1]), NewFloat32FromBits(tt[2]), NewFloat32FromBits(tt[3])
			if neg {
				x, wantSinh, wantTanh = -x, -wantSinh, -wantTanh
			}
			if got := x.Sinh(); !within1ulp32(got, wantSinh) {
				t.Errorf("Sinh(%v) = %v; want %v", x, got, wantSinh)
			}
			if got := x.Cosh(); !within1ulp32(got, wantCosh) {
				t.Errorf("Cosh(%v) = %v; want %v", x, got, wantCosh)
			}
			if got := x.Tanh(); !within1ulp32(got, wantTanh) {
				t.Errorf("Tanh(%v) = %v; want %v", x, got, wantTanh)
			}
		}
	}
}

// TestFloat32_SinhCoshTanhRandom compares Sinh, Cosh, and Tanh with math.Sinh, math.Cosh, and math.Tanh on random inputs.
// math.Sinh, math.Cosh, and math.Tanh rounded to float32 are correctly rounded except for rare cases,
// so the results should almost always match.
func TestFloat32_SinhCoshTanhRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"uniform", func() Float32 { return NewFloat32(r.Float64()*24 - 12) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-13, 6]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(20)+bias32-13)<<shift32)
		}},
	}
	for _, g := range gens {
		var mismatch int
		for range 300000 {
			x := g.gen()
			if x.IsNaN() || x.IsInf(0) {
				continue
			}
			f := float64(x)
			for _, fn := range []struct {
				name string
				got  Float32
				want Float32
			}{
				{"Sinh", x.Sinh(), NewFloat32(math.Sinh(f))},
				{"Cosh", x.Cosh(), NewFloat32(math.Cosh(f))},
				{"Tanh", x.Tanh(), NewFloat32(math.Tanh(f))},
			} {
				if !within1ulp32(fn.got, fn.want) {
					t.Fatalf("%s(%v) = %v; want %v", fn.name, x, fn.got, fn.want)
				}
				if !eq32(fn.got, fn.want) {
					mismatch++
				}
			}
		}
		// Checking all Float32 values, 16 of 3*2**32 results are different.
		if mismatch > 3 {
			t.Errorf("%s: %d results are different from math.Sinh, math.Cosh, and math.Tanh", g.name, mismatch)
		}
	}
}
