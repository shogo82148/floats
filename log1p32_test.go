package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Log1p(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-0.5), math.Log1p(-0.5)},
		{exact32(-0.25), math.Log1p(-0.25)},
		{exact32(0), math.Log1p(0)},
		{exact32(0x1p-25), math.Log1p(0x1p-25)},
		{exact32(0.25), math.Log1p(0.25)},
		{exact32(0.5), math.Log1p(0.5)},
		{exact32(1), math.Log1p(1)},
		{exact32(2), math.Log1p(2)},
		{exact32(3), math.Log1p(3)},
		{exact32(4), math.Log1p(4)},
		{exact32(11), math.Log1p(11)},
		{exact32(1 << 32), math.Log1p(1 << 32)},
	}

	for _, tt := range tests {
		got := tt.x.Log1p()
		if !close32(got, tt.want) {
			t.Errorf("Log1p(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(-1), exact32(math.Inf(-1))},
		{exact32(-2), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN())},

		// tiny arguments
		{NewFloat32FromBits(0x00000001), NewFloat32FromBits(0x00000001)},
		{NewFloat32FromBits(0x33800000), NewFloat32FromBits(0x33800000)},
		{NewFloat32FromBits(0xb3800000), NewFloat32FromBits(0xb3800000)},

		// near -1
		{NewFloat32FromBits(0xbf7fffff), NewFloat32FromBits(0xc1851592)},

		// 1+a is not exact for a >= 2**30.
		{NewFloat32FromBits(0x4e800000), NewFloat32FromBits(0x41a65af6)},
		{NewFloat32FromBits(0x4e800001), NewFloat32FromBits(0x41a65af7)},
		{NewFloat32FromBits(0x7f7fffff), NewFloat32FromBits(0x42b17218)},

		{exact32(1), NewFloat32FromBits(0x3f317218)},
	}

	for _, tt := range strictTests {
		got := tt.x.Log1p()
		if !eq32(got, tt.want) {
			t.Errorf("Log1p(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_Log1pHardCases checks Log1p on the inputs whose results are very close to
// the midpoint of two adjacent Float32 values.
// They are found by checking all Float32 values with math.Log1p in float64.
// The results may not be correctly rounded, but they must be within 1 ulp.
func TestFloat32_Log1pHardCases(t *testing.T) {
	// x, and correctly rounded log(1+x)
	tests := [][2]uint32{
		{0xbb0ec8c4, 0xbb0ef0a5},
		{0x35400003, 0x353fffff},
		{0x3ddbfec3, 0x3dd0f671},
		{0xb70fffe5, 0xb710000d},
		{0x6f31a8ec, 0x42845a89},
		{0x3710001b, 0x370ffff3},
		{0x41078feb, 0x400fe5e7},
		{0x3efd81ad, 0x3ecdeee1},
		{0x65d890d3, 0x4254d1f9},
		{0xb53ffffd, 0xb5400001},
	}
	for _, tt := range tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Log1p(); !within1ulp32(got, want) {
			t.Errorf("Log1p(%v) = %v; want %v", x, got, want)
		}
	}
}

// TestFloat32_Log1pRandom compares Log1p with math.Log1p on random inputs.
// math.Log1p rounded to float32 is correctly rounded except for rare cases,
// so the results should almost always match.
func TestFloat32_Log1pRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-25, 0]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(26)+bias32-25)<<shift32)
		}},
		{"near -1", func() Float32 { return NewFloat32(-1 + r.Float64()/64) }},
	}
	for _, g := range gens {
		var mismatch int
		for range 300000 {
			x := g.gen()
			if !(x > -1) || x.IsInf(0) {
				continue
			}
			got := x.Log1p()
			want := NewFloat32(math.Log1p(float64(x)))
			if !within1ulp32(got, want) {
				t.Fatalf("Log1p(%v) = %v; want %v", x, got, want)
			}
			if !eq32(got, want) {
				mismatch++
			}
		}
		// Checking all Float32 values on arm64, all results are the same.
		if mismatch > 3 {
			t.Errorf("%s: %d results are different from math.Log1p", g.name, mismatch)
		}
	}
}

func BenchmarkFloat32_Log1p(b *testing.B) {
	x := NewFloat32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log1p())
	}
}
