package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat32_Expm1(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-0x1p24), math.Expm1(-0x1p24)},
		{exact32(-0x1.1aaaaep+03), math.Expm1(-0x1.1aaaaep+03)},
		{exact32(-1), math.Expm1(-1)},
		{exact32(-0x1p-25), math.Expm1(-0x1p-25)},
		{exact32(-0x1p-24), math.Expm1(-0x1p-24)},
		{exact32(0), math.Expm1(0)},
		{exact32(0x1p-25), math.Expm1(0x1p-25)},
		{exact32(0x1p-24), math.Expm1(0x1p-24)},
		{exact32(0x1.99999ap-02), math.Expm1(0x1.99999ap-02)},
		{exact32(1), math.Expm1(1)},
		{exact32(10), math.Expm1(10)},
		{exact32(36), math.Expm1(36)},
		{exact32(62), math.Expm1(62)},
		{exact32(0x1p24), math.Expm1(0x1p24)},
	}

	for _, tt := range tests {
		got := tt.x.Expm1()
		if !close32(got, tt.want) {
			t.Errorf("Expm1(%v) = %v; want %v", tt.x, got, tt.want)
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
		{exact32(math.Inf(-1)), exact32(-1)},
		{exact32(math.NaN()), exact32(math.NaN())},

		// tiny arguments
		{NewFloat32FromBits(0x00000001), NewFloat32FromBits(0x00000001)},
		{NewFloat32FromBits(0x33800000), NewFloat32FromBits(0x33800000)},
		{NewFloat32FromBits(0xb3800000), NewFloat32FromBits(0xb3800000)},

		// overflow
		{NewFloat32FromBits(0x42b17217), NewFloat32FromBits(0x7f7fff84)},
		{NewFloat32FromBits(0x42b17218), exact32(math.Inf(1))},

		// e**x < 2**-25 and e**x - 1 rounds to -1, where ln(2**-25) = -17.3286795...
		{NewFloat32FromBits(0xc18aa122), NewFloat32FromBits(0xbf7fffff)},
		{NewFloat32FromBits(0xc18aa123), exact32(-1)},
		{exact32(-104), exact32(-1)},
		{exact32(-105), exact32(-1)},
	}

	for _, tt := range strictTests {
		got := tt.x.Expm1()
		if !eq32(got, tt.want) {
			t.Errorf("Expm1(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_Expm1HardCases checks Expm1 on the inputs whose results are very close to
// the midpoint of two adjacent Float32 values.
// They are found by checking all Float32 values with math.Expm1 in float64.
// The results may not be correctly rounded, but they must be within 1 ulp.
func TestFloat32_Expm1HardCases(t *testing.T) {
	t.Parallel()
	// x, and correctly rounded e**x - 1
	tests := [][2]uint32{
		{0xbb7b3b6c, 0xbb7ac04e},
		{0x3dc252dd, 0x3dcbd76b},
		{0x36322b1b, 0x36322b2a},
		{0x34ca62c1, 0x34ca62c3},
		{0xb675cbfc, 0xb675cbdf},
		{0x3a254e7a, 0x3a255bd3},
		{0x33b504f3, 0x33b504f3},
	}
	for _, tt := range tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Expm1(); !within1ulp32(got, want) {
			t.Errorf("Expm1(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat32_Expm1(b *testing.B) {
	for _, x := range []Float32{0.01, 1.5} {
		b.Run(x.String(), func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(x.Expm1())
			}
		})
	}
}
