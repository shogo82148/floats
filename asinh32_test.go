package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Asinh(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0), math.Asinh(0)},
		{exact32(0x1p-29), math.Asinh(0x1p-29)},
		{exact32(0.25), math.Asinh(0.25)},
		{exact32(0.5), math.Asinh(0.5)},
		{exact32(1), math.Asinh(1)},
		{exact32(21), math.Asinh(21)},
		{exact32(22), math.Asinh(22)},
		{exact32(0x1p29), math.Asinh(0x1p29)},

		{exact32(-0), -math.Asinh(0)},
		{exact32(-0.25), -math.Asinh(0.25)},
		{exact32(-0.5), -math.Asinh(0.5)},
		{exact32(-1), -math.Asinh(1)},
		{exact32(-21), -math.Asinh(21)},
		{exact32(-22), -math.Asinh(22)},
	}

	for _, tt := range tests {
		got := tt.x.Asinh()
		if !close32(got, tt.want) {
			t.Errorf("Asinh(%v) = %v; want %v", tt.x, got, tt.want)
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
		{NewFloat32FromBits(0x39800000), NewFloat32FromBits(0x39800000)},

		// around the boundaries of the algorithms
		{NewFloat32FromBits(0x3c75c28f), NewFloat32FromBits(0x3c75c033)}, // 0.015
		{NewFloat32FromBits(0x3c75c290), NewFloat32FromBits(0x3c75c034)},
		{NewFloat32FromBits(0x3e7fffff), NewFloat32FromBits(0x3e7d67d8)}, // 0.25
		{NewFloat32FromBits(0x3e800000), NewFloat32FromBits(0x3e7d67d9)},
		{NewFloat32FromBits(0x40000000), NewFloat32FromBits(0x3fb8c90c)}, // 2

		// the largest finite value
		{NewFloat32FromBits(0x7f7fffff), NewFloat32FromBits(0x42b2d4fc)},
		{NewFloat32FromBits(0xff7fffff), NewFloat32FromBits(0xc2b2d4fc)},
	}

	for _, tt := range strictTests {
		got := tt.x.Asinh()
		if !eq32(got, tt.want) {
			t.Errorf("Asinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_AsinhHardCases checks Asinh on the inputs whose results are very close to
// the midpoint of two adjacent Float32 values.
// They are found by checking all Float32 values with math.Asinh in float64.
// The results may not be correctly rounded, but they must be within 1 ulp.
func TestFloat32_AsinhHardCases(t *testing.T) {
	// x, and correctly rounded asinh(x)
	tests := [][2]uint32{
		{0x6eb1a8ec, 0x42845a89},
		{0xeeb1a8ec, 0xc2845a89},
		{0x4bdd65a5, 0x418f034b},
		{0xcbdd65a5, 0xc18f034b},
		{0x3ca1078c, 0x3ca104e4},
		{0xbca1078c, 0xbca104e4},
		{0x655890d3, 0x4254d1f9},
		{0xe55890d3, 0xc254d1f9},
	}
	for _, tt := range tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Asinh(); !within1ulp32(got, want) {
			t.Errorf("Asinh(%v) = %v; want %v", x, got, want)
		}
	}
}

// TestFloat32_AsinhRandom compares Asinh with math.Asinh on random inputs.
// math.Asinh rounded to float32 is correctly rounded except for rare cases,
// so the results should almost always match.
func TestFloat32_AsinhRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-13, 4]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(18)+bias32-13)<<shift32)
		}},
	}
	for _, g := range gens {
		var mismatch int
		for range 300000 {
			x := g.gen()
			if x.IsNaN() {
				continue
			}
			got := x.Asinh()
			want := NewFloat32(math.Asinh(float64(x)))
			if !within1ulp32(got, want) {
				t.Fatalf("Asinh(%v) = %v; want %v", x, got, want)
			}
			if !eq32(got, want) {
				mismatch++
			}
		}
		// Checking all Float32 values on arm64, only 2 results are different.
		if mismatch > 3 {
			t.Errorf("%s: %d results are different from math.Asinh", g.name, mismatch)
		}
	}
}

func BenchmarkFloat32_Asinh(b *testing.B) {
	for _, x := range []Float32{0.01, 1.5, 1e10} {
		b.Run(x.String(), func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(x.Asinh())
			}
		})
	}
}

func TestFloat32_Acosh(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(1), math.Acosh(1)},
		{exact32(1.5), math.Acosh(1.5)},
		{exact32(21), math.Acosh(21)},
		{exact32(22), math.Acosh(22)},
		{exact32(0x1p29), math.Acosh(0x1p29)},
	}

	for _, tt := range tests {
		got := tt.x.Acosh()
		if !close32(got, tt.want) {
			t.Errorf("Acosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
		{exact32(1), exact32(0)},
		{NewFloat32FromBits(0x3f7fffff), exact32(math.NaN())},
		{exact32(0), exact32(math.NaN())},

		// near 1
		{NewFloat32FromBits(0x3f800001), NewFloat32FromBits(0x3a000000)},
		{NewFloat32FromBits(0x3f800002), NewFloat32FromBits(0x3a3504f3)},

		// around the boundaries of the algorithms
		{NewFloat32FromBits(0x3f800346), NewFloat32FromBits(0x3c67957b)}, // 1 + 1e-4
		{NewFloat32FromBits(0x3f800347), NewFloat32FromBits(0x3c67b8d8)},
		{NewFloat32FromBits(0x3f83d70a), NewFloat32FromBits(0x3e7a346b)}, // 1.03
		{NewFloat32FromBits(0x3f83d70b), NewFloat32FromBits(0x3e7a348b)},

		// the largest finite value
		{NewFloat32FromBits(0x7f7fffff), NewFloat32FromBits(0x42b2d4fc)},
	}

	for _, tt := range strictTests {
		got := tt.x.Acosh()
		if !eq32(got, tt.want) {
			t.Errorf("Acosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat32_Atanh(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0x1p-29), math.Atanh(0x1p-29)},
		{exact32(0.25), math.Atanh(0.25)},
		{exact32(0.5), math.Atanh(0.5)},
		{exact32(0.75), math.Atanh(0.75)},
	}

	for _, tt := range tests {
		got := tt.x.Atanh()
		if !close32(got, tt.want) {
			t.Errorf("Atanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(2), exact32(math.NaN())},
		{exact32(1), exact32(math.Inf(1))},
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(-1), exact32(math.Inf(-1))},
		{exact32(-2), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},

		// tiny arguments
		{NewFloat32FromBits(0x00000001), NewFloat32FromBits(0x00000001)},
		{NewFloat32FromBits(0x397fffff), NewFloat32FromBits(0x397fffff)},
		{NewFloat32FromBits(0x39800000), NewFloat32FromBits(0x39800000)},
		{NewFloat32FromBits(0x3a000000), NewFloat32FromBits(0x3a000001)},

		// around the boundaries of the algorithms
		{NewFloat32FromBits(0x3bf5c28f), NewFloat32FromBits(0x3bf5c3bd)}, // 0.0075
		{NewFloat32FromBits(0x3bf5c290), NewFloat32FromBits(0x3bf5c3be)},
		{NewFloat32FromBits(0x3e7fffff), NewFloat32FromBits(0x3e82c577)}, // 0.25
		{NewFloat32FromBits(0x3e800000), NewFloat32FromBits(0x3e82c578)},

		// near ±1
		{NewFloat32FromBits(0x3f7fffff), NewFloat32FromBits(0x410aa123)},
		{NewFloat32FromBits(0xbf7fffff), NewFloat32FromBits(0xc10aa123)},
	}

	for _, tt := range strictTests {
		got := tt.x.Atanh()
		if !eq32(got, tt.want) {
			t.Errorf("Atanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_AcoshHardCases checks Acosh on the inputs whose results are very close to
// the midpoint of two adjacent Float32 values.
// They are found by checking all Float32 values with math.Acosh in float64.
// The results may not be correctly rounded, but they must be within 1 ulp.
func TestFloat32_AcoshHardCases(t *testing.T) {
	// x, and correctly rounded acosh(x)
	tests := [][2]uint32{
		{0x655890d3, 0x4254d1f9},
		{0x6eb1a8ec, 0x42845a89},
	}
	for _, tt := range tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Acosh(); !within1ulp32(got, want) {
			t.Errorf("Acosh(%v) = %v; want %v", x, got, want)
		}
	}
}

// TestFloat32_AcoshAtanhRandom compares Acosh and Atanh with math.Acosh and math.Atanh on random inputs.
// Checking all Float32 values on arm64, the results are the same.
func TestFloat32_AcoshAtanhRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"near 1", func() Float32 { return NewFloat32(1 + r.Float64()/16) }},
		{"(-1, 1)", func() Float32 { return NewFloat32(r.Float64()*2 - 1) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-13, 0]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(14)+bias32-13)<<shift32)
		}},
	}
	for _, g := range gens {
		var mismatch int
		for range 300000 {
			x := g.gen()
			f := float64(x)
			for _, fn := range []struct {
				name string
				got  Float32
				want Float32
			}{
				{"Acosh", x.Acosh(), NewFloat32(math.Acosh(f))},
				{"Atanh", x.Atanh(), NewFloat32(math.Atanh(f))},
			} {
				if fn.want.IsNaN() || fn.want.IsInf(0) {
					if !eq32(fn.got, fn.want) {
						t.Fatalf("%s(%v) = %v; want %v", fn.name, x, fn.got, fn.want)
					}
					continue
				}
				if !within1ulp32(fn.got, fn.want) {
					t.Fatalf("%s(%v) = %v; want %v", fn.name, x, fn.got, fn.want)
				}
				if !eq32(fn.got, fn.want) {
					mismatch++
				}
			}
		}
		if mismatch > 3 {
			t.Errorf("%s: %d results are different from the math package", g.name, mismatch)
		}
	}
}

func BenchmarkFloat32_Acosh(b *testing.B) {
	for _, x := range []Float32{1.0001, 1.5, 1e10} {
		b.Run(x.String(), func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(x.Acosh())
			}
		})
	}
}

func BenchmarkFloat32_Atanh(b *testing.B) {
	for _, x := range []Float32{0.01, 0.5} {
		b.Run(x.String(), func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(x.Atanh())
			}
		})
	}
}
