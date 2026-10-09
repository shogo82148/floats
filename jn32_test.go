package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Jn(t *testing.T) {
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
		got := exact32(tt.x).Jn(tt.n)
		if !close32(got, want) {
			t.Errorf("Jn(%d, %v) = %v; want %v", tt.n, tt.x, got, want)
		}
	}
}

func TestFloat32_JnSpecial(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		n    int
		x    Float32
		want Float32
	}{
		{2, exact32(0), exact32(0)},
		{2, exact32(math.Inf(1)), exact32(0)},
		{2, exact32(math.Inf(-1)), exact32(0)},
		{2, exact32(math.NaN()), exact32(math.NaN())},
		{0, exact32(0), exact32(1)},
		{1, exact32(0), exact32(0)},
		{-1, exact32(2), -NewFloat32(math.J1(2))},
		{2, exact32(0x1p-100), exact32(0)}, // (x/2)**2/2 underflows
		{40, exact32(2), exact32(0)},       // (x/2)**n/n! underflows
		{-40, exact32(-2), exact32(0)},
		{1 << 30, exact32(2), exact32(0)},
	} {
		if got := tt.x.Jn(tt.n); !eq32(got, tt.want) {
			t.Errorf("Jn(%d, %v) = %v; want %v", tt.n, tt.x, got, tt.want)
		}
	}
}

// TestFloat32_JnHardCases checks Jn on the inputs whose results are very close to the midpoint of two adjacent
// Float32 values, or subnormal. They are found by checking random inputs of the Taylor series and the forward recurrence,
// and the correctly rounded results were calculated with mpmath.
func TestFloat32_JnHardCases(t *testing.T) {
	t.Parallel()
	// n, x, and the correctly rounded Jn(n, x)
	tests := [][3]uint32{
		{5, 0x3ebc0d3e, 0x35e8501d},
		{9, 0x3fe908bb, 0x35921003},
		{15, 0x40aa83aa, 0x359e65ab},
		{13, 0x409b6c46, 0x37338ccf},
		{10, 0x402027fc, 0x3616b874},
		{6, 0x3f260393, 0x35d566a5},
		{15, 0x40b1e328, 0x360f89cf},
		{5, 0x3eaba6ae, 0x3593579a},
		{7, 0x3f81cea0, 0x35de3b69},
		{15, 0x40b3f8fe, 0x3628f6c0},
		{4, 0x3e16843b, 0x35a2fc0a},
		{6, 0x3f5090a5, 0x36cff065},
		{7, 0x3f77af1e, 0x35a057e5},
		{7, 0x3f7e3893, 0x35c01d9c},
		{12, 0x405dc390, 0x35a26049},
		{14, 0x40a3050a, 0x367049c0},
		{11, 0x40372c50, 0x3591f400},
		{5, 0x3ea9bd50, 0x358b553f},
		{10, 0x401edba7, 0x360b4362},
		{12, 0x40761d35, 0x368631e7},
		{14, 0x416508ef, 0x3e5384de},
		{15, 0x40ae787b, 0x35dab5f9},
		{9, 0x3feb10c7, 0x359db2f8},
		{15, 0x40abc5ce, 0x35afad13},
		{5, 0x3eaaea8e, 0x3590388d},
		{9, 0x3fe7a45b, 0x358a884d},
		{4, 0x3e2474d1, 0x35e83b3b},
		{7, 0x3f856e8f, 0x36067f13},
		{15, 0x40aaacf3, 0x35a0854c},
		{11, 0x404bf304, 0x3664406c},
		{8, 0x3fc3692b, 0x3633c572},
		{2, 0x3b70f32f, 0x35e2c8ce},
		{6, 0x3f2ea9d1, 0x36107b1e},
		{5, 0x3ec31549, 0x360b7f9b},
		{14, 0x409487ad, 0x358cc761},
		{3, 0x3d55deed, 0x3646fe63},
		{12, 0x407832dc, 0x3693c19b},
		{10, 0x4020adbf, 0x361b9065},
		{13, 0x40921405, 0x36a8add7},
		{4, 0x3e245dcf, 0x35e7b977},
		{7, 0x3f8c0808, 0x363beb34},
		{15, 0x40aa7d9b, 0x359e1673},
		{3, 0x3d7734f3, 0x3699a47b},
		{14, 0x4096a57b, 0x35a9d10e},
		{8, 0x3fe87026, 0x372f4bd5},
		{4, 0x3e7862ef, 0x3716ceb0},
		{6, 0x3f240458, 0x35c686ae},
		{9, 0x3ffa0207, 0x3607c7f8},
		{8, 0x3fae43a6, 0x3591c0cb},
		{12, 0x405cce8c, 0x359a79be},
	}
	for _, tt := range tests {
		n, x, want := int(tt[0]), NewFloat32FromBits(tt[1]), NewFloat32FromBits(tt[2])
		if got := x.Jn(n); !eq32(got, want) {
			t.Errorf("Jn(%d, %v) = %v; want %v", n, x, got, want)
		}
		// J(-n, x) = (-1)**n J(n, x) and J(n, -x) = (-1)**n J(n, x)
		wantNeg := want
		if n%2 == 1 {
			wantNeg = -want
		}
		if got := (-x).Jn(n); !eq32(got, wantNeg) {
			t.Errorf("Jn(%d, %v) = %v; want %v", n, -x, got, wantNeg)
		}
		if got := x.Jn(-n); !eq32(got, wantNeg) {
			t.Errorf("Jn(%d, %v) = %v; want %v", -n, x, got, wantNeg)
		}
		if got := (-x).Jn(-n); !eq32(got, want) {
			t.Errorf("Jn(%d, %v) = %v; want %v", -n, -x, got, want)
		}
	}
}

// TestFloat32_JnRandom compares Jn with Float256, which is correctly rounded, on random inputs.
func TestFloat32_JnRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	for range 3000 {
		var n int
		var x Float32
		switch r.IntN(5) {
		case 0, 1:
			// the Taylor series
			n = 2 + r.IntN(60)
			x = NewFloat32(math.Sqrt(2*float64(n+1)) * r.Float64())
		case 2:
			// the forward recurrence
			x = NewFloat32(2 + 62*r.Float64())
			n = 2 + r.IntN(int(x)-1)
		case 3:
			// the Taylor series, where the terms increase at first: sqrt(2 (n+1)) <= x < n
			n = 2 + r.IntN(62)
			lo := math.Sqrt(2 * float64(n+1))
			x = NewFloat32(lo + (float64(n)-lo)*r.Float64())
		default:
			// the others, that is, math.Jn: the backward recurrence and large x
			n = 2 + r.IntN(30)
			x = NewFloat32(math.Sqrt(2*float64(n+1)) + 100*r.Float64())
		}
		if r.IntN(2) == 0 {
			x = -x
		}
		if r.IntN(2) == 0 {
			n = -n
		}
		if got, want := x.Jn(n), NewFloat256(float64(x)).Jn(n).Float32(); !eq32(got, want) {
			t.Fatalf("Jn(%d, %v) = %v; want %v", n, x, got, want)
		}
	}
}

// TestFloat32_JnZeros checks Jn near the zeros of Jn, where the forward recurrence has no relative accuracy.
func TestFloat32_JnZeros(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		n int
		z float64
	}{
		{2, 5.135622301840683}, {2, 8.417244140399864}, {3, 6.380161895923984}, {3, 9.761023129981162},
		{5, 8.771483815959954}, {5, 12.338604197466944}, {10, 14.475500686554541}, {10, 18.433463666966583},
	} {
		for d := -20; d <= 20; d++ {
			x := NewFloat32FromBits(math.Float32bits(float32(tt.z)) + uint32(d))
			if got, want := x.Jn(tt.n), NewFloat256(float64(x)).Jn(tt.n).Float32(); !eq32(got, want) {
				t.Errorf("Jn(%d, %v) = %v; want %v", tt.n, x, got, want)
			}
		}
	}
}

// TestFloat32_JnBoundaries checks Jn on the both sides of the boundaries of the regions.
func TestFloat32_JnBoundaries(t *testing.T) {
	t.Parallel()
	for _, n := range []int{2, 3, 10, 40, 63, 64} {
		for _, b := range []float64{math.Sqrt(2 * float64(n+1)), float64(n), 2, 64} {
			for d := -3; d <= 3; d++ {
				x := NewFloat32FromBits(math.Float32bits(float32(b)) + uint32(d))
				if got, want := x.Jn(n), NewFloat256(float64(x)).Jn(n).Float32(); !eq32(got, want) {
					t.Errorf("Jn(%d, %v) = %v; want %v", n, x, got, want)
				}
			}
		}
	}
}

func BenchmarkFloat32_Jn(b *testing.B) {
	for _, tt := range []struct {
		name string
		n    int
		x    Float32
	}{
		{"taylor", 10, exact32(2.5)},
		{"taylor-large-n", 50, exact32(10)},
		{"forward", 5, exact32(20.5)},
		{"forward2", 10, exact32(50.5)},
		{"negative", -5, exact32(-20.5)},
		{"backward", 20, exact32(15.5)}, // math.Jn
		{"large", 3, exact32(1000.5)},   // math.Jn
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Jn(tt.n))
			}
		})
	}
}
