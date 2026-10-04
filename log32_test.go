package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Log(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(1), math.Log(1)},
		{exact32(2), math.Log(2)},
		{exact32(3), math.Log(3)},
		{exact32(4), math.Log(4)},
		{exact32(11), math.Log(11)},
	}

	for _, tt := range tests {
		got := tt.x.Log()
		if !close32(got, tt.want) {
			t.Errorf("Log(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(0), exact32(math.Inf(-1))},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(-1))},
		{exact32(-1), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{NewFloat32FromBits(0x80000001), exact32(math.NaN())},

		{exact32(1), exact32(0)},

		// subnormal
		{NewFloat32FromBits(0x00000001), NewFloat32FromBits(0xc2ce8ed0)},
		{NewFloat32FromBits(0x00400000), NewFloat32FromBits(0xc2b00f34)},
		{NewFloat32FromBits(0x007fffff), NewFloat32FromBits(0xc2aeac50)},
	}

	for _, tt := range strictTests {
		got := tt.x.Log()
		if !eq32(got, tt.want) {
			t.Errorf("Log(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Log(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log())
	}
}

func TestFloat32_Log10(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(1), math.Log10(1)},
		{exact32(2), math.Log10(2)},
		{exact32(3), math.Log10(3)},
		{exact32(4), math.Log10(4)},
		{exact32(11), math.Log10(11)},
	}

	for _, tt := range tests {
		got := tt.x.Log10()
		if !close32(got, tt.want) {
			t.Errorf("Log10(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(0), exact32(math.Inf(-1))},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(-1))},
		{exact32(-1), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},

		// subnormal
		{NewFloat32FromBits(0x00000001), NewFloat32FromBits(0xc23369f4)},
		{NewFloat32FromBits(0x00400000), NewFloat32FromBits(0xc218ec59)},
		{NewFloat32FromBits(0x007fffff), NewFloat32FromBits(0xc217b818)},
	}

	// powers of ten
	for n := range 11 {
		x := exact32(math.Pow10(n))
		if got, want := x.Log10(), exact32(float64(n)); !eq32(got, want) {
			t.Errorf("Log10(%v) = %v; want %v", x, got, want)
		}
	}

	for _, tt := range strictTests {
		got := tt.x.Log10()
		if !eq32(got, tt.want) {
			t.Errorf("Log10(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Log10(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log10())
	}
}

func TestFloat32_Log2(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(1), math.Log2(1)},
		{exact32(2), math.Log2(2)},
		{exact32(3), math.Log2(3)},
		{exact32(4), math.Log2(4)},
		{exact32(11), math.Log2(11)},
	}

	for _, tt := range tests {
		got := tt.x.Log2()
		if !close32(got, tt.want) {
			t.Errorf("Log2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(0), exact32(math.Inf(-1))},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(-1))},
		{exact32(-1), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	// powers of two give exact answers
	for n := -149; n <= 127; n++ {
		x := exact32(math.Ldexp(1, n))
		if got, want := x.Log2(), exact32(float64(n)); !eq32(got, want) {
			t.Errorf("Log2(%v) = %v; want %v", x, got, want)
		}
	}

	for _, tt := range strictTests {
		got := tt.x.Log2()
		if !eq32(got, tt.want) {
			t.Errorf("Log2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Log2(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log2())
	}
}

// TestFloat32_LogHardCases checks Log and Log10 on the inputs whose results are very close to
// the midpoint of two adjacent Float32 values.
// They are found by checking all Float32 values with math.Log and math.Log10 in float64.
// The results may not be correctly rounded, but they must be within 1 ulp.
func TestFloat32_LogHardCases(t *testing.T) {
	// x, and correctly rounded log(x)
	logTests := [][2]uint32{
		{0x3c413d3a, 0xc08e158f},
		{0x4c5d65a5, 0x418f034b},
		{0x65d890d3, 0x4254d1f9},
		{0x41178feb, 0x400fe5e7},
		{0x6f31a8ec, 0x42845a89},
	}
	for _, tt := range logTests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Log(); !within1ulp32(got, want) {
			t.Errorf("Log(%v) = %v; want %v", x, got, want)
		}
	}

	// x, and correctly rounded log10(x)
	log10Tests := [][2]uint32{
		{0x120b93dc, 0xc1dad957},
		{0x0efeee7a, 0xc1e99d23},
		{0x13ae78d3, 0xc1d2d957},
	}
	for _, tt := range log10Tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Log10(); !within1ulp32(got, want) {
			t.Errorf("Log10(%v) = %v; want %v", x, got, want)
		}
	}
}

// TestFloat32_LogRandom compares Log, Log10, and Log2 with math.Log, math.Log10, and math.Log2 on random inputs.
// The results of the math package rounded to float32 are correctly rounded except for rare cases,
// so the results should almost always match.
func TestFloat32_LogRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32() &^ signMask32) }},
		{"near one", func() Float32 { return NewFloat32(1 + (r.Float64()-0.5)/16) }},
		{"subnormal", func() Float32 { return NewFloat32FromBits(r.Uint32() & fracMask32) }},
	}
	for _, g := range gens {
		var mismatch int
		for range 300000 {
			x := g.gen()
			if x.IsNaN() || x.IsInf(0) || x == 0 {
				continue
			}
			f := float64(x)
			for _, fn := range []struct {
				name string
				got  Float32
				want Float32
			}{
				{"Log", x.Log(), NewFloat32(math.Log(f))},
				{"Log10", x.Log10(), NewFloat32(math.Log10(f))},
				{"Log2", x.Log2(), NewFloat32(math.Log2(f))},
			} {
				if !within1ulp32(fn.got, fn.want) {
					t.Fatalf("%s(%v) = %v; want %v", fn.name, x, fn.got, fn.want)
				}
				if !eq32(fn.got, fn.want) {
					mismatch++
				}
			}
		}
		// Checking all positive Float32 values on arm64, all results are the same.
		if mismatch > 3 {
			t.Errorf("%s: %d results are different from the math package", g.name, mismatch)
		}
	}
}
