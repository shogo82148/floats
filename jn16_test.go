package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat16_Jn(t *testing.T) {
	t.Parallel()
	type tc struct {
		n int
		x float64
	}
	tests := []tc{
		{2, -10}, {2, -1}, {2, 0}, {2, 1}, {2, 5}, {3, 10}, {5, 20}, {-2, 5},
	}

	for _, tt := range tests {
		want := math.Jn(tt.n, tt.x)
		got := exact16(tt.x).Jn(tt.n)
		if !close16(got, want) {
			t.Errorf("Jn(%d, %v) = %v; want %v", tt.n, tt.x, got, want)
		}
	}
}

func TestFloat16_JnSpecial(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		n    int
		x    Float16
		want Float16
	}{
		{2, exact16(0), exact16(0)},
		{2, exact16(math.Inf(1)), exact16(0)},
		{2, exact16(math.Inf(-1)), exact16(0)},
		{2, exact16(math.NaN()), exact16(math.NaN())},
		{0, exact16(0), exact16(1)},
		{1, exact16(0), exact16(0)},
		{-1, exact16(2), NewFloat16(-math.J1(2))},
		{-1, exact16(0), exact16(math.Copysign(0, -1))},
		{40, exact16(2), exact16(0)}, // (x/2)**n/n! underflows
		{-40, exact16(-2), exact16(0)},
		{1 << 30, exact16(2), exact16(0)},
	} {
		if got := tt.x.Jn(tt.n); !eq16(got, tt.want) {
			t.Errorf("Jn(%d, %v) = %v; want %v", tt.n, tt.x, got, tt.want)
		}
	}
}

// TestFloat16_JnHardCases checks Jn on the inputs whose results are very close to the midpoint of two adjacent
// Float16 values, or to zero. They are found by checking all the Float16 values for the Taylor series and
// the forward recurrence, and the correctly rounded results were calculated with mpmath.
func TestFloat16_JnHardCases(t *testing.T) {
	t.Parallel()
	// n, x, and the correctly rounded Jn(n, x)
	tests := [][3]uint16{
		{2, 0x0ffe, 0x0000},
		{2, 0x1000, 0x0000},
		{2, 0x1479, 0x0003},
		{2, 0x16a2, 0x0005},
		{2, 0x17bf, 0x0008},
		{2, 0x1b00, 0x0018},
		{2, 0x1d80, 0x003c},
		{2, 0x1ecd, 0x005d},
		{2, 0x1f80, 0x0070},
		{2, 0x1ffc, 0x007f},
		{2, 0x2040, 0x0090},
		{3, 0x21c5, 0x0001},
		{3, 0x2429, 0x0001},
		{3, 0x2600, 0x0004},
		{3, 0x2b40, 0x0040},
		{3, 0x2cf8, 0x00a3},
		{3, 0x2ea3, 0x0185},
		{4, 0x30b6, 0x0014},
		{4, 0x3106, 0x001b},
		{5, 0x327e, 0x0001},
		{5, 0x3331, 0x0002},
		{5, 0x3436, 0x0006},
		{5, 0x356d, 0x0013},
		{4, 0x35e6, 0x0322},
		{7, 0x388e, 0x0000},
		{6, 0x3901, 0x0016},
		{6, 0x3a3b, 0x0050},
		{6, 0x3c04, 0x0168},
		{6, 0x3c7b, 0x02b0},
		{9, 0x3e5c, 0x0005},
		{8, 0x3ee1, 0x0073},
		{9, 0x4106, 0x0132},
		{10, 0x41ed, 0x00c0},
		{13, 0x420e, 0x0001},
		{14, 0x431f, 0x0000},
		{13, 0x4382, 0x0008},
		{11, 0x442a, 0x03a1},
		{14, 0x4465, 0x0009},
		{14, 0x44b2, 0x0014},
		{3, 0x4a81, 0x169b},
		{9, 0x4e0f, 0x0504},
		{10, 0x4f38, 0x174a},
		{9, 0x504b, 0xa672},
		{10, 0x5070, 0x80c8},
		{35, 0x516f, 0xb0a5},
		{14, 0x5181, 0xa027},
		{8, 0x521f, 0x2420},
		{14, 0x5253, 0xa0c1},
		{18, 0x5295, 0x0c9a},
		{21, 0x52a4, 0x90ba},
		{48, 0x52e1, 0x8b3c},
		{21, 0x5311, 0x00aa},
		{22, 0x533a, 0x8aa4},
		{19, 0x53fd, 0x93a7},
	}
	for _, tt := range tests {
		n, x, want := int(tt[0]), NewFloat16FromBits(tt[1]), NewFloat16FromBits(tt[2])
		if got := x.Jn(n); !eq16(got, want) {
			t.Errorf("Jn(%d, %v) = %v; want %v", n, x, got, want)
		}
		// J(-n, x) = (-1)**n J(n, x) and J(n, -x) = (-1)**n J(n, x)
		wantNeg := want
		if n%2 == 1 {
			wantNeg = want.Neg()
		}
		if got := x.Neg().Jn(n); !eq16(got, wantNeg) {
			t.Errorf("Jn(%d, %v) = %v; want %v", n, x.Neg(), got, wantNeg)
		}
		if got := x.Jn(-n); !eq16(got, wantNeg) {
			t.Errorf("Jn(%d, %v) = %v; want %v", -n, x, got, wantNeg)
		}
		if got := x.Neg().Jn(-n); !eq16(got, want) {
			t.Errorf("Jn(%d, %v) = %v; want %v", -n, x.Neg(), got, want)
		}
	}
}

// TestFloat16_JnRandom compares Jn with Float256, which is correctly rounded, on random inputs.
func TestFloat16_JnRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	for range 20000 {
		var n int
		var x Float16
		switch r.IntN(5) {
		case 0, 1:
			// the Taylor series
			n = 2 + r.IntN(60)
			x = NewFloat16(math.Sqrt(2*float64(n+1)) * r.Float64())
		case 2:
			// the forward recurrence
			x = NewFloat16(2 + 62*r.Float64())
			n = 2 + r.IntN(int(x.Float64().BuiltIn())-1)
		case 3:
			// the Taylor series, where the terms increase at first: sqrt(2 (n+1)) <= x < n
			n = 2 + r.IntN(62)
			lo := math.Sqrt(2 * float64(n+1))
			x = NewFloat16(lo + (float64(n)-lo)*r.Float64())
		default:
			// the others, that is, math.Jn
			n = 2 + r.IntN(30)
			x = NewFloat16(math.Sqrt(2*float64(n+1)) + 100*r.Float64())
		}
		if r.IntN(2) == 0 {
			x = x.Neg()
		}
		if r.IntN(2) == 0 {
			n = -n
		}
		if got, want := x.Jn(n), NewFloat256(x.Float64().BuiltIn()).Jn(n).Float16(); !eq16(got, want) {
			t.Fatalf("Jn(%d, %v) = %v; want %v", n, x, got, want)
		}
	}
}

// TestFloat16_JnAll checks Jn(2, x) and Jn(7, x) for all the Float16 values with Float256.
func TestFloat16_JnAll(t *testing.T) {
	t.Parallel()
	for _, n := range []int{2, 7} {
		for i := range 0x7c00 {
			x := NewFloat16FromBits(uint16(i))
			if got, want := x.Jn(n), NewFloat256(x.Float64().BuiltIn()).Jn(n).Float16(); !eq16(got, want) {
				t.Errorf("Jn(%d, %#04x) = %v; want %v", n, i, got, want)
			}
		}
	}
}

func TestFloat16_Float16Round(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		g      float64
		window int64
		want   Float16
		ok     bool
	}{
		{1, 1, 0x3c00, true},
		{-1, 1, 0xbc00, true},
		{0, 1, 0, true},
		{1 + 0x1p-12, 1, 0x3c00, true},
		{1 + 0x1p-11, 1, 0, false},                 // the midpoint
		{1 + 0x1p-11 + 0x1p-40, 1 << 20, 0, false}, // too close to the midpoint for the window
		{1 + 0x1p-11 + 0x1p-40, 1, 0x3c01, true},
		{1 + 0x1p-11 - 0x1p-40, 1, 0x3c00, true},
		{2 - 0x1p-12, 1, 0x4000, true}, // carry to the exponent
		{0x1p-14, 1, 0x0400, true},
		{0x1p-14 - 0x1p-26, 1, 0x0400, true}, // subnormal: 1023.75 => 1024
		{0x1p-14 - 0x1p-25, 1, 0, false},     // subnormal: the midpoint 1023.5
		{0x1p-24, 1, 0x0001, true},
		{0x1p-24 * 0.5, 1, 0, false},       // the midpoint of 0 and the smallest subnormal number
		{0x1p-24 * 0.75, 1, 0x0001, true},  // subnormal
		{-0x1p-24 * 0.25, 1, 0x8000, true}, // rounded to -0
		{65504, 1, 0x7bff, true},
	} {
		got, ok := float16Round(tt.g, tt.window)
		if ok != tt.ok || (ok && got != tt.want) {
			t.Errorf("float16Round(%v, %d) = %#04x, %v; want %#04x, %v", tt.g, tt.window, got, ok, tt.want, tt.ok)
		}
	}
}

func BenchmarkFloat16_Jn(b *testing.B) {
	for _, tt := range []struct {
		name string
		n    int
		x    Float16
	}{
		{"taylor", 10, exact16(2.5)},
		{"taylor-large-n", 50, exact16(10)},
		{"forward", 5, exact16(20.5)},
		{"forward2", 10, exact16(50.5)},
		{"negative", -5, exact16(-20.5)},
		{"backward", 20, exact16(15.5)}, // math.Jn
		{"large", 3, exact16(1000)},     // math.Jn
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Jn(tt.n))
			}
		})
	}
}
