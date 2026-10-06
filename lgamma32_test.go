package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Lgamma(t *testing.T) {
	t.Parallel()
	tests := []float64{-2.5, -0.5, 0.5, 1, 1.5, 2, 2.5, 3, 100}

	for _, x := range tests {
		want, wantSign := math.Lgamma(x)
		got, sign := exact32(x).Lgamma()
		if !close32(got, want) || sign != wantSign {
			t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", x, got, sign, want, wantSign)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
		sign int
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(math.Inf(1)), 1},
		{exact32(math.Inf(-1)), exact32(math.Inf(-1)), 1},
		{exact32(math.NaN()), exact32(math.NaN()), 1},
		{exact32(0), exact32(math.Inf(1)), 1},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(1)), 1},
		{exact32(-1), exact32(math.Inf(1)), 1},
		{exact32(-2), exact32(math.Inf(1)), 1},
		{exact32(-8388608), exact32(math.Inf(1)), 1},
		{NewFloat32(-1e30), exact32(math.Inf(1)), 1},
		{exact32(1), exact32(0), 1},
		{exact32(2), exact32(0), 1},

		// the overflow threshold
		{NewFloat32FromBits(0x7c44_af8d), NewFloat32FromBits(0x7f7f_fffe), 1},
		{NewFloat32FromBits(0x7c44_af8e), exact32(math.Inf(1)), 1},
		{NewFloat32FromBits(0x7f7f_ffff), exact32(math.Inf(1)), 1},
	}

	for _, tt := range strictTests {
		got, sign := tt.x.Lgamma()
		if !eq32(got, tt.want) || sign != tt.sign {
			t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", tt.x, got, sign, tt.want, tt.sign)
		}
	}
}

// TestFloat32_LgammaHardCases checks Lgamma on the inputs whose results are very close to the midpoint of two adjacent
// Float32 values. They are found by checking all Float32 values with math.Lgamma in float64,
// and the correctly rounded results were calculated with mpmath.
func TestFloat32_LgammaHardCases(t *testing.T) {
	t.Parallel()
	// x, and the correctly rounded Lgamma(x)
	tests := [][2]uint32{
		{0x064cb44b, 0x429e7c1a},
		{0x07c060fa, 0x429a733a},
		{0x0dc8bba4, 0x4289bac4},
		{0x0f61ff63, 0x42855565},
		{0x111c87f8, 0x428085df},
		{0x145cb6d4, 0x426f092f},
		{0x16c5ee7a, 0x42619bd4},
		{0x1a8446cb, 0x424d0a45},
		{0x1f116ab8, 0x4233b53e},
		{0x2423c085, 0x421781d1},
		{0x28e3fa26, 0x41fa75df},
		{0x29bd4f07, 0x41f0db65},
		{0x2caa0a85, 0x41d071d6},
		{0x3b7c53aa, 0x40b1d661},
		{0x3bd9a107, 0x40a056fc},
		{0x3f7f9d97, 0x3a63b3ba},
		{0x3f7fa94a, 0x3a4894d2},
		{0x3f7fc8b4, 0x39ffa724},
		{0x3f7fdf0c, 0x3998470c},
		{0x3f800e42, 0xb9839880},
		{0x3f801188, 0xb9a1c91d},
		{0x3fffbfc2, 0xba58f5ae},
		{0x3fffc68c, 0xba420ff2},
		{0x3ffff697, 0xb8fe90eb},
		{0x3ffff85a, 0xb8ceea85},
		{0x40001793, 0x3a1fa4d8},
		{0x42bc1c3d, 0x43a5fbed},
		{0x5ba4c7ca, 0x5e4407fb},
		{0x5dc98cd3, 0x608140c0},
		{0x5e961867, 0x6145a15e},
		{0x65fca09f, 0x68cead5a},
		{0x67b98485, 0x6a9eea75},
		{0x6b9ce710, 0x6e93967c},
		{0x7500b6fc, 0x78132970},
		{0x77ac5674, 0x7acf27b3},
		{0x7b54ff19, 0x7e8870c3},
		{0x864cb44b, 0x429e7c1a},
		{0x87c060fa, 0x429a733a},
		{0x8dc8bba4, 0x4289bac4},
		{0x8f61ff63, 0x42855565},
		{0x911c87f8, 0x428085df},
		{0x945cb6d4, 0x426f092f},
		{0x96c5ee7a, 0x42619bd4},
		{0x9a8446cb, 0x424d0a45},
		{0x9f116ab8, 0x4233b53e},
		{0xa2925ad4, 0x4220462d},
		{0xa423c085, 0x421781d1},
		{0xb0d6f2ca, 0x41a23559},
		{0xb90f6e0d, 0x410e5ad5},
		{0xc02fdc45, 0x39892224},
		{0xc18458a0, 0xc1f865f3},
		{0xc36c073c, 0xc483b679},
	}
	for _, tt := range tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got, _ := x.Lgamma(); !eq32(got, want) {
			t.Errorf("Lgamma(%v) = %v; want %v", x, got, want)
		}
	}
}

// TestFloat32_LgammaBoundaries checks Lgamma on the both sides of the boundaries of the segments of the calculation.
func TestFloat32_LgammaBoundaries(t *testing.T) {
	t.Parallel()
	for _, x := range []float32{
		0x1p-149, 0x1p-126, 0x1p-8, 0.5, 1, 2, 4096, 8192, 0x1p23, 0x1p30, 0x1p36,
	} {
		for _, neg := range []bool{false, true} {
			for d := -3; d <= 3; d++ {
				a := NewFloat32FromBits(math.Float32bits(x) + uint32(d))
				if neg {
					a = -a
				}
				if a == 0 {
					continue
				}
				got, sign := a.Lgamma()
				w, wantSign := math.Lgamma(float64(a))
				want := NewFloat32(w)
				if !eq32(got, want) && !within1ulp32(got, want) || sign != wantSign {
					t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", a, got, sign, want, wantSign)
				}
			}
		}
	}
}

// TestFloat32_LgammaRandom compares Lgamma with math.Lgamma on random inputs.
// math.Lgamma rounded to Float32 is correctly rounded except for lgamma32Hard,
// so the results must match except for them.
func TestFloat32_LgammaRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	gens := []struct {
		name string
		gen  func() Float32
	}{
		{"bits", func() Float32 { return NewFloat32FromBits(r.Uint32()) }},
		{"uniform", func() Float32 { return NewFloat32(r.Float64()*80 - 40) }},
		{"positive", func() Float32 { return NewFloat32(r.Float64() * 10) }},
		{"negative", func() Float32 { return NewFloat32(-r.Float64() * 20) }},
		{"large", func() Float32 { return NewFloat32(math.Exp2(r.Float64()*130 - 30)) }},
		{"exponent", func() Float32 {
			// uniformly distributed exponent in [-40, 40]
			return NewFloat32FromBits(r.Uint32()&(signMask32|fracMask32) | uint32(r.IntN(81)+bias32-40)<<shift32)
		}},
	}
	for _, g := range gens {
		for range 300000 {
			x := g.gen()
			got, sign := x.Lgamma()
			w, wantSign := math.Lgamma(float64(x))
			want := NewFloat32(w)
			if x.IsNaN() {
				if !got.IsNaN() {
					t.Fatalf("Lgamma(%v) = %v; want NaN", x, got)
				}
				continue
			}
			if !eq32(got, want) && !within1ulp32(got, want) || sign != wantSign {
				t.Fatalf("Lgamma(%v) = (%v, %d); want (%v, %d)", x, got, sign, want, wantSign)
			}
			if !eq32(got, want) {
				hard := false
				for _, h := range lgamma32Hard {
					hard = hard || h[0] == x.Bits()
				}
				if !hard {
					t.Errorf("%s: Lgamma(%v) = %v; want %v", g.name, x, got, want)
				}
			}
		}
	}
}

// TestFloat32_LgammaApprox checks the approximation of Lgamma, whose errors are hidden by the rounding to Float32, with math.Lgamma.
func TestFloat32_LgammaApprox(t *testing.T) {
	t.Parallel()
	const bound = 0x1p-34 // the relative error is less than 2**-34
	r := rand.New(rand.NewPCG(3, 4))
	for range 1000000 {
		var x float64
		switch r.IntN(4) {
		case 0:
			x = r.Float64()*20 - 10
		case 1:
			x = math.Exp2(r.Float64()*50 - 30)
		case 2:
			x = -math.Exp2(r.Float64()*23 - 10)
		default:
			x = float64(NewFloat32FromBits(r.Uint32()&0x7fff_ffff | r.Uint32()&signMask32))
		}
		if math.IsNaN(x) || math.IsInf(x, 0) || x >= 4.085e36 {
			continue
		}
		y, sign, ok := lgamma32Approx(x)
		want, wantSign := math.Lgamma(x)
		if sign != wantSign {
			t.Errorf("lgamma32Approx(%v): the sign is %d; want %d", x, sign, wantSign)
		}
		if !ok || math.IsInf(want, 0) {
			continue
		}
		if math.Abs(y-want) > bound*math.Abs(want) {
			t.Errorf("lgamma32Approx(%v) = %v; want %v", x, y, want)
		}
	}
}

func BenchmarkFloat32_Lgamma(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float32
	}{
		{"tiny", NewFloat32(1e-20)},           // Lgamma(x) ~ -log(x)
		{"small", NewFloat32(0.375)},          // 0 < x < 1/2
		{"medium", NewFloat32(1.5)},           // 1/2 <= x < 3
		{"large", NewFloat32(100.5)},          // 3 <= x < 8192
		{"huge", NewFloat32(1e20)},            // Stirling's series
		{"overflow", NewFloat32(1e37)},        // +Inf
		{"negative", NewFloat32(-2.5)},        // -2**23 < x < 0
		{"verynegative", NewFloat32(-100.25)}, // Lgamma(x) < 0
		{"nearzero", NewFloat32(-2.4609375)},  // the result is close to zero (falls back to the math package)
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				v, s := tt.x.Lgamma()
				runtime.KeepAlive(v)
				runtime.KeepAlive(s)
			}
		})
	}
}
