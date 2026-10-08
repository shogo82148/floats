package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Asin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-1), math.Asin(-1)},
		{exact32(-0.75), math.Asin(-0.75)},
		{exact32(-0.5), math.Asin(-0.5)},
		{exact32(-0.25), math.Asin(-0.25)},
		{exact32(0.25), math.Asin(0.25)},
		{exact32(0.5), math.Asin(0.5)},
		{exact32(0.75), math.Asin(0.75)},
		{exact32(1), math.Asin(1)},
	}

	for _, tt := range tests {
		got := tt.x.Asin()
		if !close32(got, tt.want) {
			t.Errorf("Asin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.NaN()), exact32(math.NaN())},
		{exact32(2), exact32(math.NaN())},
		{exact32(-2), exact32(math.NaN())},
		{exact32(math.Inf(1)), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{NewFloat32FromBits(0x3f800001), exact32(math.NaN())},

		// tiny arguments
		{NewFloat32FromBits(0x00000001), NewFloat32FromBits(0x00000001)},
		{NewFloat32FromBits(0x397fffff), NewFloat32FromBits(0x397fffff)},
		{NewFloat32FromBits(0x39800000), NewFloat32FromBits(0x39800000)},

		// the boundary of the algorithms, and the endpoints
		{exact32(0.5), NewFloat32(math.Asin(0.5))},
		{NewFloat32FromBits(0x3f000001), NewFloat32(math.Asin(float64(NewFloat32FromBits(0x3f000001))))},
		{exact32(1), NewFloat32(math.Pi / 2)},
		{exact32(-1), NewFloat32(-math.Pi / 2)},
	}

	for _, tt := range strictTests {
		got := tt.x.Asin()
		if !eq32(got, tt.want) {
			t.Errorf("Asin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_AsinRandom compares Asin with math.Asin on random inputs.
// Checking all Float32 values, the results are identical to math.Asin rounded to Float32.
func TestFloat32_AsinRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-14, -1]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(14)+bias32-14)<<shift32)
		}},
		{"near1", func() Float32 {
			return NewFloat32FromBits(0x3f800000 - r.Uint32N(1<<20))
		}},
	}
	for _, g := range gens {
		for range 300000 {
			x := g.gen()
			got := x.Asin()
			want := NewFloat32(math.Asin(float64(x)))
			if !eq32(got, want) {
				t.Fatalf("%s: Asin(%v) = %v; want %v", g.name, x, got, want)
			}
		}
	}
}

func BenchmarkFloat32_Asin(b *testing.B) {
	for _, x := range []Float32{0.25, 0.75} {
		b.Run(x.String(), func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(x.Asin())
			}
		})
	}
}

func TestFloat32_Acos(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-1), math.Acos(-1)},
		{exact32(-0.75), math.Acos(-0.75)},
		{exact32(-0.5), math.Acos(-0.5)},
		{exact32(-0.25), math.Acos(-0.25)},
		{exact32(0.25), math.Acos(0.25)},
		{exact32(0.5), math.Acos(0.5)},
		{exact32(0.75), math.Acos(0.75)},
		{exact32(1), math.Acos(1)},
	}

	for _, tt := range tests {
		got := tt.x.Acos()
		if !close32(got, tt.want) {
			t.Errorf("Acos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.NaN()), exact32(math.NaN())},
		{exact32(2), exact32(math.NaN())},
		{exact32(-2), exact32(math.NaN())},
		{exact32(math.Inf(1)), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{NewFloat32FromBits(0x3f800001), exact32(math.NaN())},

		// the endpoints
		{exact32(1), exact32(0)},
		{exact32(-1), NewFloat32(math.Pi)},
		{exact32(0), NewFloat32(math.Pi / 2)},
		{exact32(math.Copysign(0, -1)), NewFloat32(math.Pi / 2)},

		// tiny arguments and the boundary of the algorithms
		{NewFloat32FromBits(0x00000001), NewFloat32(math.Pi / 2)},
		{exact32(0.5), NewFloat32(math.Acos(0.5))},
		{NewFloat32FromBits(0x3f000001), NewFloat32(math.Acos(float64(NewFloat32FromBits(0x3f000001))))},
		{NewFloat32FromBits(0xbf000001), NewFloat32(math.Acos(float64(NewFloat32FromBits(0xbf000001))))},
	}

	for _, tt := range strictTests {
		got := tt.x.Acos()
		if !eq32(got, tt.want) {
			t.Errorf("Acos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_AcosRandom compares Acos with math.Acos on random inputs.
// Checking all Float32 values, the results are identical to math.Acos rounded to Float32.
func TestFloat32_AcosRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-14, -1]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(14)+bias32-14)<<shift32)
		}},
		{"near1", func() Float32 {
			return NewFloat32FromBits(r.Uint32()&signMask32 | (0x3f800000 - r.Uint32N(1<<20)))
		}},
	}
	for _, g := range gens {
		for range 300000 {
			x := g.gen()
			got := x.Acos()
			want := NewFloat32(math.Acos(float64(x)))
			if !eq32(got, want) {
				t.Fatalf("%s: Acos(%v) = %v; want %v", g.name, x, got, want)
			}
		}
	}
}

func BenchmarkFloat32_Acos(b *testing.B) {
	for _, x := range []Float32{0.25, 0.75, -0.75} {
		b.Run(x.String(), func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(x.Acos())
			}
		})
	}
}

func TestFloat32_Atan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-0.5), math.Atan(-0.5)},
		{exact32(-0.25), math.Atan(-0.25)},
		{exact32(-0.125), math.Atan(-0.125)},
		{exact32(0.125), math.Atan(0.125)},
		{exact32(0.25), math.Atan(0.25)},
		{exact32(0.5), math.Atan(0.5)},
		{exact32(0.75), math.Atan(0.75)},
		{exact32(1), math.Atan(1)},
		{exact32(2), math.Atan(2)},
		{exact32(math.Inf(-1)), math.Atan(math.Inf(-1))},
		{exact32(math.Inf(1)), math.Atan(math.Inf(1))},
	}

	for _, tt := range tests {
		got := tt.x.Atan()
		if !close32(got, tt.want) {
			t.Errorf("Atan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.NaN()), exact32(math.NaN())},
		{exact32(math.Inf(1)), NewFloat32(math.Pi / 2)},
		{exact32(math.Inf(-1)), NewFloat32(-math.Pi / 2)},

		// tiny arguments
		{NewFloat32FromBits(0x00000001), NewFloat32FromBits(0x00000001)},
		{NewFloat32FromBits(0x397fffff), NewFloat32FromBits(0x397fffff)},
		{NewFloat32FromBits(0x39800000), NewFloat32FromBits(0x39800000)},

		// around the boundary of the algorithms, and the largest finite value
		{exact32(1), NewFloat32(math.Pi / 4)},
		{exact32(-1), NewFloat32(-math.Pi / 4)},
		{NewFloat32FromBits(0x3f800001), NewFloat32(math.Atan(float64(NewFloat32FromBits(0x3f800001))))},
		{NewFloat32FromBits(0x3f7fffff), NewFloat32(math.Atan(float64(NewFloat32FromBits(0x3f7fffff))))},
		{NewFloat32FromBits(0x7f7fffff), NewFloat32(math.Pi / 2)},

		// math.Atan returns 0x3d8d31c2, but the correctly rounded result is 0x3d8d31c3.
		{NewFloat32FromBits(0x3d8d6b23), NewFloat32FromBits(0x3d8d31c3)},
		{NewFloat32FromBits(0xbd8d6b23), NewFloat32FromBits(0xbd8d31c3)},
	}

	for _, tt := range strictTests {
		got := tt.x.Atan()
		if !eq32(got, tt.want) {
			t.Errorf("Atan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_AtanRandom compares Atan with math.Atan on random inputs.
// Checking all Float32 values, the only difference is 0x3d8d6b23 (and its negation),
// where Atan is correctly rounded but math.Atan is not.
func TestFloat32_AtanRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-14, 8]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(23)+bias32-14)<<shift32)
		}},
	}
	for _, g := range gens {
		for range 300000 {
			x := g.gen()
			if x.IsNaN() {
				continue
			}
			got := x.Atan()
			want := NewFloat32(math.Atan(float64(x)))
			if !eq32(got, want) {
				t.Fatalf("%s: Atan(%v) = %v; want %v", g.name, x, got, want)
			}
		}
	}
}

func BenchmarkFloat32_Atan(b *testing.B) {
	for _, x := range []Float32{0.25, 0.75, 3} {
		b.Run(x.String(), func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(x.Atan())
			}
		})
	}
}

func TestFloat32_Atan2(t *testing.T) {
	t.Parallel()
	tests := []struct {
		y, x Float32
		want float64
	}{
		{exact32(1), exact32(1), math.Pi / 4},
		{exact32(1), exact32(-1), 3 * math.Pi / 4},
		{exact32(-1), exact32(-1), -3 * math.Pi / 4},
		{exact32(-1), exact32(1), -math.Pi / 4},

		// special cases
		// +0.Atan2(x<=-0) = +Pi
		{exact32(0), exact32(-1), math.Pi},
		// -0.Atan2(x<=-0) = -Pi
		{exact32(math.Copysign(0, -1)), exact32(-1), -math.Pi},
		// y>0.Atan2(0) = +Pi/2
		{exact32(1), exact32(0), math.Pi / 2},
		{exact32(1), exact32(math.Copysign(0, -1)), math.Pi / 2},
		// y<0.Atan2(0) = -Pi/2
		{exact32(-1), exact32(0), -math.Pi / 2},
		{exact32(-1), exact32(math.Copysign(0, -1)), -math.Pi / 2},
		// +Inf.Atan2(+Inf) = +Pi/4
		{exact32(math.Inf(1)), exact32(math.Inf(1)), math.Pi / 4},
		// -Inf.Atan2(+Inf) = -Pi/4
		{exact32(math.Inf(-1)), exact32(math.Inf(1)), -math.Pi / 4},
		// +Inf.Atan2(-Inf) = 3*Pi/4
		{exact32(math.Inf(1)), exact32(math.Inf(-1)), 3 * math.Pi / 4},
		// -Inf.Atan2(-Inf) = -3*Pi/4
		{exact32(math.Inf(-1)), exact32(math.Inf(-1)), -3 * math.Pi / 4},
		// y.Atan2(+Inf) = 0
		{exact32(1), exact32(math.Inf(1)), 0},
		{exact32(-1), exact32(math.Inf(1)), 0},
		// (y>0).Atan2(-Inf) = +Pi
		{exact32(1), exact32(math.Inf(-1)), math.Pi},
		// (y<0).Atan2(-Inf) = -Pi
		{exact32(-1), exact32(math.Inf(-1)), -math.Pi},
		// +Inf.Atan2(x) = +Pi/2
		{exact32(math.Inf(1)), exact32(1), math.Pi / 2},
		// -Inf.Atan2(x) = -Pi/2
		{exact32(math.Inf(-1)), exact32(1), -math.Pi / 2},
	}

	for _, tt := range tests {
		got := tt.y.Atan2(tt.x)
		if !close32(got, tt.want) {
			t.Errorf("Atan2(%v, %v) = %v; want %v", tt.y, tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		y, x Float32
		want Float32
	}{
		// special cases
		// y.Atan2(NaN) = NaN
		{exact32(1), exact32(math.NaN()), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN()), exact32(math.NaN())},
		// NaN.Atan2(x) = NaN
		{exact32(math.NaN()), exact32(1), exact32(math.NaN())},
		// +0.Atan2(x>=0) = +0
		{exact32(0), exact32(1), exact32(0)},
		// -0.Atan2(x>=0) = -0
		{exact32(math.Copysign(0, -1)), exact32(1), exact32(math.Copysign(0, -1))},
	}

	for _, tt := range strictTests {
		got := tt.y.Atan2(tt.x)
		if !eq32(got, tt.want) {
			t.Errorf("Atan2(%v, %v) = %v; want %v", tt.y, tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Atan2(b *testing.B) {
	y, x := Float32(1.5), Float32(-2.25)
	for b.Loop() {
		runtime.KeepAlive(y.Atan2(x))
	}
}

func TestFloat32_Atan2_Fallback(t *testing.T) {
	t.Parallel()
	// the results are close to a rounding boundary or subnormal,
	// so the fast path falls back to the accurate path.
	tests := []struct{ y, x uint32 }{
		{0xa237d527, 0x527fa874},
		{0xbeb7ede5, 0x3cf8087a},
		{0xc1c46aa8, 0xc1e13f4c},
		{0x00000001, 0x7f7fffff}, // the result is subnormal
		{0x80000001, 0x7f7fffff},
		{0x00000001, 0x00000001}, // both are subnormal
	}
	for _, tt := range tests {
		y, x := NewFloat32FromBits(tt.y), NewFloat32FromBits(tt.x)
		want := y.Float128().Atan2(x.Float128()).Float32()
		if got := y.Atan2(x); got.Bits() != want.Bits() {
			t.Errorf("Atan2(%v, %v) = %v; want %v", y, x, got, want)
		}
	}
}
