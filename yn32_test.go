package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Yn(t *testing.T) {
	t.Parallel()
	type tc struct {
		n int
		x float64
	}
	tests := []tc{
		{2, 1}, {2, 5}, {3, 10}, {5, 20}, {-2, 5}, {2, 0}, {-3, 0},
	}

	for _, tt := range tests {
		want := math.Yn(tt.n, tt.x)
		got := exact32(tt.x).Yn(tt.n)
		if !close32(got, want) {
			t.Errorf("Yn(%d, %v) = %v; want %v", tt.n, tt.x, got, want)
		}
	}
}

func TestFloat32_YnSpecial(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		n    int
		x    Float32
		want Float32
	}{
		{2, exact32(0), exact32(math.Inf(-1))},
		{3, exact32(0), exact32(math.Inf(-1))},
		{-2, exact32(0), exact32(math.Inf(-1))},
		{-3, exact32(0), exact32(math.Inf(1))},
		{2, exact32(math.Copysign(0, -1)), exact32(math.Inf(-1))},
		{2, exact32(-1), exact32(math.NaN())},
		{-3, exact32(-5), exact32(math.NaN())},
		{2, exact32(math.Inf(1)), exact32(0)},
		{-3, exact32(math.Inf(1)), exact32(0)},
		{2, exact32(math.Inf(-1)), exact32(math.NaN())},
		{2, exact32(math.NaN()), exact32(math.NaN())},
		{-1, exact32(0), exact32(math.Inf(1))},
	} {
		if got := tt.x.Yn(tt.n); !eq32(got, tt.want) {
			t.Errorf("Yn(%d, %v) = %v; want %v", tt.n, tt.x, got, tt.want)
		}
	}
}

// TestFloat32_YnLowOrders checks that Yn(0, x), Yn(1, x) and Yn(-1, x) are Y0, Y1 and -Y1.
func TestFloat32_YnLowOrders(t *testing.T) {
	t.Parallel()
	for _, x := range []Float32{exact32(0.5), exact32(5), exact32(50), exact32(1000)} {
		if got, want := x.Yn(0), x.Y0(); !eq32(got, want) {
			t.Errorf("Yn(0, %v) = %v; want %v", x, got, want)
		}
		if got, want := x.Yn(1), x.Y1(); !eq32(got, want) {
			t.Errorf("Yn(1, %v) = %v; want %v", x, got, want)
		}
		if got, want := x.Yn(-1), -x.Y1(); !eq32(got, want) {
			t.Errorf("Yn(-1, %v) = %v; want %v", x, got, want)
		}
	}
}

// TestFloat32_YnHardCases checks Yn on the inputs whose results are very close to the midpoint of two adjacent
// Float32 values (or to zero), which are calculated by the more accurate Float256. They are found by checking
// 3,200,000 random inputs with 2 <= |n| <= 100 and 2 <= x < 64, and the correctly rounded results were calculated
// with mpmath.
func TestFloat32_YnHardCases(t *testing.T) {
	t.Parallel()
	// n, x, and the correctly rounded Yn(n, x)
	tests := [][3]int64{
		{29, 0x4276218a, 0xbc958743},
		{-7, 0x421dc0af, 0x399d1959},
		{69, 0x427fb880, 0xbf2517a6},
		{61, 0x41ff8c34, 0xd0018632},
		{23, 0x41f8bad7, 0x36a93c8e},
		{17, 0x42597f25, 0xbdda5358},
		{-20, 0x427112f8, 0xb91992b6},
		{-21, 0x42142dda, 0xbb8c89ea},
		{33, 0x4227d304, 0xb6316cba},
		{70, 0x41a76d48, 0xec6a5670},
		{-58, 0x418075f2, 0xe7912fea},
		{36, 0x42480857, 0x3af8c457},
		{60, 0x427edaeb, 0x36b5dc34},
		{-79, 0x4205adcd, 0x5fee0165},
		{-74, 0x41c42308, 0xe9d91a81},
		{17, 0x4088d01d, 0xcba6c46b},
		{90, 0x424e2a6d, 0xd39823f3},
		{57, 0x426fc0fb, 0xbd3322fc},
		{3, 0x418e56cd, 0xbba07718},
		{-51, 0x425a19c3, 0xb62f0acc},
		{16, 0x41db1de5, 0x3837f0bf},
		{69, 0x422df695, 0xcacb4e5a},
		{-36, 0x421c6999, 0xbb8bdbf5},
		{51, 0x4117743d, 0xf0e2fb0d},
		{-46, 0x425e6e81, 0x3ca9e4d4},
		{81, 0x42449b97, 0xceb73e37},
		{-97, 0x423888e2, 0x604daa7f},
		{41, 0x426fd3bd, 0xb709a77f},
	}
	for _, tt := range tests {
		n, x, want := int(tt[0]), NewFloat32FromBits(uint32(tt[1])), NewFloat32FromBits(uint32(tt[2]))
		if got := x.Yn(n); !eq32(got, want) {
			t.Errorf("Yn(%d, %v) = %v; want %v", n, x, got, want)
		}
	}
}

// TestFloat32_YnBoundaries checks Yn on the both sides of the boundaries of the range of the recurrence:
// 2 <= x < 64 and |n| <= 100.
func TestFloat32_YnBoundaries(t *testing.T) {
	t.Parallel()
	for _, n := range []int{2, 3, 50, 99, 100, 101, 102, -2, -100, -101} {
		for _, x := range []float32{1, 2, 3, 63, 64, 65} {
			for d := -3; d <= 3; d++ {
				a := NewFloat32FromBits(math.Float32bits(x) + uint32(d))
				if got, want := a.Yn(n), NewFloat256(float64(a)).Yn(n).Float32(); !eq32(got, want) {
					t.Errorf("Yn(%d, %v) = %v; want %v", n, a, got, want)
				}
			}
		}
	}
}

// TestFloat32_YnRandom compares Yn with Float256 on random inputs.
func TestFloat32_YnRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	for range 20000 {
		n := 2 + r.IntN(110)
		if r.IntN(2) == 0 {
			n = -n
		}
		var x Float32
		if r.IntN(2) == 0 {
			x = NewFloat32(2 + 62*r.Float64())
		} else {
			// near the turning point |n| ~ x
			x = NewFloat32(max(2, math.Min(63.9, math.Abs(float64(n))*(0.5+r.Float64()))))
		}
		if got, want := x.Yn(n), NewFloat256(float64(x)).Yn(n).Float32(); !eq32(got, want) {
			t.Fatalf("Yn(%d, %v) = %v; want %v", n, x, got, want)
		}
	}
}

func BenchmarkFloat32_Yn(b *testing.B) {
	for _, tt := range []struct {
		name string
		n    int
		x    Float32
	}{
		{"small-n", 2, exact32(10.5)},
		{"medium-n", 10, exact32(50.5)},
		{"large-n", 50, exact32(20.5)},
		{"negative", -5, exact32(20.5)},
		{"hard", 29, NewFloat32FromBits(0x4276218a)}, // the result is too close to the midpoint: Float256.Yn
		{"small-x", 3, exact32(1.5)},                 // math.Yn
		{"large-x", 3, exact32(1000.5)},              // math.Yn
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Yn(tt.n))
			}
		})
	}
}
