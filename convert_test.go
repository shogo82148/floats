package floats

import (
	"math"
	"runtime"
	"testing"
)

// eq16 reports whether a and b are equal.
// It returns true if both a and b are NaN.
// It distinguishes between +0 and -0.
func eq16(a, b Float16) bool {
	if a.IsNaN() && b.IsNaN() {
		return true
	}
	return a == b
}

// eq32 reports whether a and b are equal.
// It returns true if both a and b are NaN.
// It distinguishes between +0 and -0.
func eq32(a, b Float32) bool {
	if a.IsNaN() && b.IsNaN() {
		return true
	}
	return math.Float32bits(float32(a)) == math.Float32bits(float32(b))
}

// eq64 reports whether a and b are equal.
// It returns true if both a and b are NaN.
// It distinguishes between +0 and -0.
func eq64(a, b Float64) bool {
	if a.IsNaN() && b.IsNaN() {
		return true
	}
	return math.Float64bits(float64(a)) == math.Float64bits(float64(b))
}

// eq128 reports whether a and b are equal.
// It returns true if both a and b are NaN.
// It distinguishes between +0 and -0.
func eq128(a, b Float128) bool {
	if a.IsNaN() && b.IsNaN() {
		return true
	}
	return a[0] == b[0] && a[1] == b[1]
}

// eq256 reports whether a and b are equal.
// It returns true if both a and b are NaN.
// It distinguishes between +0 and -0.
func eq256(a, b Float256) bool {
	if a.IsNaN() && b.IsNaN() {
		return true
	}
	return a[0] == b[0] && a[1] == b[1] && a[2] == b[2] && a[3] == b[3]
}

func TestFloat16_Float32(t *testing.T) {
	tests := []struct {
		in   Float16
		want Float32
	}{
		// from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
		{0x0000, 0},
		{0x0001, 0x1p-24},                       // smallest positive subnormal number
		{0x03ff, 0x1.ff8p-15},                   // largest positive subnormal number
		{0x0400, 0x1p-14},                       // smallest positive normal number
		{0x3555, 0x1.554p-02},                   // nearest value to 1/3
		{0x3bff, 0x1.ffcp-01},                   // largest number less than one
		{0x3c00, 0x1p+00},                       // one
		{0x3c01, 0x1.004p+00},                   // smallest number larger than one
		{0x7bff, 0x1.ffcp+15},                   // largest normal number
		{0x7c00, Float32(math.Inf(1))},          // infinity
		{0x8000, Float32(math.Copysign(0, -1))}, // -0
		{0xc000, -2},
		{0xfc00, Float32(math.Inf(-1))}, // negative infinity
		{0x7e00, Float32(math.NaN())},   // NaN
	}

	for _, tt := range tests {
		if got := tt.in.Float32(); !eq32(got, tt.want) {
			t.Errorf("Float16(%x).Float32() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Float32(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Float32())
	}
}

func TestFloat16_Float64(t *testing.T) {
	tests := []struct {
		in   Float16
		want Float64
	}{
		// from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
		{0x0000, 0},
		{0x0001, 0x1p-24},                       // smallest positive subnormal number
		{0x03ff, 0x1.ff8p-15},                   // largest positive subnormal number
		{0x0400, 0x1p-14},                       // smallest positive normal number
		{0x3555, 0x1.554p-02},                   // nearest value to 1/3
		{0x3bff, 0x1.ffcp-01},                   // largest number less than one
		{0x3c00, 0x1p+00},                       // one
		{0x3c01, 0x1.004p+00},                   // smallest number larger than one
		{0x7bff, 0x1.ffcp+15},                   // largest normal number
		{0x7c00, Float64(math.Inf(1))},          // infinity
		{0x8000, Float64(math.Copysign(0, -1))}, // -0
		{0xc000, -2},
		{0xfc00, Float64(math.Inf(-1))}, // negative infinity
		{0x7e00, Float64(math.NaN())},   // NaN
	}

	for _, tt := range tests {
		if got := tt.in.Float64(); !eq64(got, tt.want) {
			t.Errorf("Float16(%x).Float64() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Float64(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Float64())
	}
}

func TestFloat16_Float128(t *testing.T) {
	tests := []struct {
		in   Float16
		want Float128
	}{
		// from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
		// and https://en.wikipedia.org/wiki/Quadruple-precision_floating-point_format
		{0x0000, Float128{0, 0}},
		{0x7c00, Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}}, // infinity
		{0x8000, Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}}, // -0
		{0xc000, Float128{0xc000_0000_0000_0000, 0x0000_0000_0000_0000}}, // -2
		{0xfc00, Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}}, // negative infinity

		{0x0001, Float128{0x3fe7_0000_0000_0000, 0x0000_0000_0000_0000}}, // 0x1p-24
	}

	for _, tt := range tests {
		if got := tt.in.Float128(); !eq128(got, tt.want) {
			t.Errorf("Float16(%x).Float128() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Float128(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Float128())
	}
}

func TestFloat16_Float256(t *testing.T) {
	tests := []struct {
		in   Float16
		want Float256
	}{
		// from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
		// and https://en.wikipedia.org/wiki/Octuple-precision_floating-point_format
		{
			// +0
			0x0000,
			Float256{0, 0, 0, 0},
		},
		{
			// -0
			0x8000,
			Float256{0x8000_0000_0000_0000, 0, 0, 0},
		},
		{
			// infinity
			0x7c00,
			Float256{
				0x7fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// -infinity
			0xfc00,
			Float256{
				0xffff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// 1
			0x3c00,
			Float256{
				0x3fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},

		// 0x1p-24
		{
			0x0001,
			Float256{
				0x3ffe_7000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float256(); !eq256(got, tt.want) {
			t.Errorf("Float16(%x).Float256() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Float256(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Float256())
	}
}

func TestFloat32_Float16(t *testing.T) {
	tests := []struct {
		in   Float32
		want Float16
	}{
		// from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
		{0, 0x0000},
		{Float32(math.Copysign(0, -1)), 0x8000}, // -0
		{0x1p-24, 0x0001},                       // smallest positive subnormal number
		{0x1.ff8p-15, 0x03ff},                   // largest positive subnormal number
		{0x1p-14, 0x0400},                       // smallest positive normal number
		{0x1.554p-02, 0x3555},                   // nearest value to 1/3
		{0x1.ffcp-01, 0x3bff},                   // largest number less than one
		{0x1p+00, 0x3c00},                       // one
		{0x1.004p+00, 0x3c01},                   // smallest number larger than one
		{0x1.ffcp+15, 0x7bff},                   // largest normal number
		{Float32(math.Inf(1)), 0x7c00},          // infinity
		{-2, 0xc000},
		{Float32(math.Inf(-1)), 0xfc00}, // negative infinity
		{Float32(math.NaN()), 0x7e00},   // NaN

		// underflow
		{0x1p-25, 0x0000},

		// overflow
		{0x1.ffep+15, 0x7c00},
	}

	for _, tt := range tests {
		if got := tt.in.Float16(); !eq16(got, tt.want) {
			t.Errorf("Float32(%x).Float16() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Float16(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Float16())
	}
}

func TestFloat32_Float32(t *testing.T) {
	tests := []struct {
		in   Float32
		want Float32
	}{
		{
			in:   1.0,
			want: 1.0,
		},
		{
			in:   2.0,
			want: 2.0,
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float32(); !eq32(got, tt.want) {
			t.Errorf("Float32(%x).Float32() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func TestFloat32_Float64(t *testing.T) {
	tests := []struct {
		in   Float32
		want Float64
	}{
		{
			in:   1.0,
			want: 1.0,
		},
		{
			in:   2.0,
			want: 2.0,
		},
		{
			in:   math.MaxFloat32,
			want: math.MaxFloat32,
		},
		{
			in:   math.SmallestNonzeroFloat32,
			want: math.SmallestNonzeroFloat32,
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float64(); !eq64(got, tt.want) {
			t.Errorf("Float32(%x).Float64() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Float64(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Float64())
	}
}

func TestFloat32_Float128(t *testing.T) {
	tests := []struct {
		in   Float32
		want Float128
	}{
		// https://en.wikipedia.org/wiki/Single-precision_floating-point_format
		{
			in:   0.0,
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   Float32(math.Copysign(0, -1)), // -0
			want: Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   1.0,
			want: Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   2.0,
			want: Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   -2.0,
			want: Float128{0xc000_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   0x1p-149, // smallest positive subnormal number
			want: Float128{0x3f6a_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   0x1.fffffcp-127, // largest subnormal number
			want: Float128{0x3f80_ffff_fc00_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   0x1p-126, // smallest positive normal number
			want: Float128{0x3f81_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   0x1.fffffep+127, // largest normal number
			want: Float128{0x407e_ffff_fe00_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   Float32(math.Inf(1)),
			want: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   Float32(math.Inf(-1)),
			want: Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   Float32(math.NaN()),
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000},
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float128(); !eq128(got, tt.want) {
			t.Errorf("Float32(%x).Float128() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Float128(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Float128())
	}
}

func TestFloat32_Float256(t *testing.T) {
	tests := []struct {
		in   Float32
		want Float256
	}{
		// https://en.wikipedia.org/wiki/Single-precision_floating-point_format
		{
			in: 0.0,
			want: Float256{
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: Float32(math.Copysign(0, -1)), // -0
			want: Float256{
				0x8000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: 1.0,
			want: Float256{
				0x3fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: 2.0,
			want: Float256{
				0x4000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: -2.0,
			want: Float256{
				0xc000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: 0x1p-149, // smallest positive subnormal number
			want: Float256{
				0x3ff6_a000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: 0x1.fffffcp-127, // largest subnormal number
			want: Float256{
				0x3ff8_0fff_ffc0_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: 0x1p-126, // smallest positive normal number
			want: Float256{
				0x3ff8_1000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: 0x1.fffffep+127, // largest normal number
			want: Float256{
				0x4007_efff_ffe0_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: Float32(math.Inf(1)),
			want: Float256{
				0x7fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: Float32(math.Inf(-1)),
			want: Float256{
				0xffff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: Float32(math.NaN()),
			want: Float256{
				0x7fff_f800_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float256(); !eq256(got, tt.want) {
			t.Errorf("Float32(%x).Float256() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Float256(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Float256())
	}
}

func TestFloat64_Float16(t *testing.T) {
	tests := []struct {
		in   Float64
		want Float16
	}{
		// from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
		{0, 0x0000},
		{Float64(math.Copysign(0, -1)), 0x8000}, // -0
		{0x1p-24, 0x0001},                       // smallest positive subnormal number
		{0x1.ff8p-15, 0x03ff},                   // largest positive subnormal number
		{0x1p-14, 0x0400},                       // smallest positive normal number
		{0x1.554p-02, 0x3555},                   // nearest value to 1/3
		{0x1.ffcp-01, 0x3bff},                   // largest number less than one
		{0x1p+00, 0x3c00},                       // one
		{0x1.004p+00, 0x3c01},                   // smallest number larger than one
		{0x1.ffcp+15, 0x7bff},                   // largest normal number
		{Float64(math.Inf(1)), 0x7c00},          // infinity
		{-2, 0xc000},
		{Float64(math.Inf(-1)), 0xfc00}, // negative infinity
		{Float64(math.NaN()), 0x7e00},   // NaN

		// underflow
		{0x1p-25, 0x0000},

		// overflow
		{0x1.ffep+15, 0x7c00},
	}

	for _, tt := range tests {
		if got := tt.in.Float16(); !eq16(got, tt.want) {
			t.Errorf("Float64(%x).Float16() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Float16(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Float16())
	}
}

func TestFloat64_Float32(t *testing.T) {
	tests := []struct {
		in   Float64
		want Float32
	}{
		{
			in:   1.0,
			want: 1.0,
		},
		{
			in:   2.0,
			want: 2.0,
		},
		{
			in:   math.MaxFloat32,
			want: math.MaxFloat32,
		},
		{
			in:   math.SmallestNonzeroFloat32,
			want: math.SmallestNonzeroFloat32,
		},

		// overflow
		{
			in:   math.MaxFloat64,
			want: Float32(math.Inf(1)),
		},

		// underflow
		{
			in:   math.SmallestNonzeroFloat64,
			want: 0,
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float32(); !eq32(got, tt.want) {
			t.Errorf("Float64(%x).Float32() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Float32(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Float32())
	}
}

func TestFloat64_Float64(t *testing.T) {
	tests := []struct {
		in   Float64
		want Float64
	}{
		{
			in:   1.0,
			want: 1.0,
		},
		{
			in:   2.0,
			want: 2.0,
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float64(); !eq64(got, tt.want) {
			t.Errorf("Float64(%x).Float64() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func TestFloat64_Float128(t *testing.T) {
	tests := []struct {
		in   Float64
		want Float128
	}{
		// from https://en.wikipedia.org/wiki/Double-precision_floating-point_format
		{
			in:   0.0,
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   Float64(math.Copysign(0, -1)), // -0
			want: Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   1.0,
			want: Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   2.0,
			want: Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   -2.0,
			want: Float128{0xc000_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   0x1p-1074, // smallest positive subnormal number
			want: Float128{0x3bcd_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   0x1.ffffffffffffep-1023, // largest subnormal number
			want: Float128{0x3c00_ffff_ffff_ffff, 0xe000_0000_0000_0000},
		},
		{
			in:   0x1p-1022, // smallest positive normal number
			want: Float128{0x3c01_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   0x1.fffffffffffffp+1023, // largest normal number
			want: Float128{0x43fe_ffff_ffff_ffff, 0xf000_0000_0000_0000},
		},
		{
			in:   Float64(math.Inf(1)),
			want: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   Float64(math.Inf(-1)),
			want: Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000},
		},
		{
			in:   Float64(math.NaN()),
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000},
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float128(); !eq128(got, tt.want) {
			t.Errorf("Float64(%x).Float128() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkTestFloat64_Float128(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Float128())
	}
}

func TestFloat64_Float256(t *testing.T) {
	tests := []struct {
		in   Float64
		want Float256
	}{
		// from https://en.wikipedia.org/wiki/Double-precision_floating-point_format
		{
			in: 0.0,
			want: Float256{
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: Float64(math.Copysign(0, -1)), // -0
			want: Float256{
				0x8000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: 1.0,
			want: Float256{
				0x3fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: 2.0,
			want: Float256{
				0x4000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: -2.0,
			want: Float256{
				0xc000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: 0x1p-1074, // smallest positive subnormal number
			want: Float256{
				0x3fbc_d000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: 0x1.ffffffffffffep-1023, // largest subnormal number
			want: Float256{
				0x3fc0_0fff_ffff_ffff,
				0xfe00_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: 0x1p-1022, // smallest positive normal number
			want: Float256{
				0x3fc0_1000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: 0x1.fffffffffffffp+1023, // largest normal number
			want: Float256{
				0x403f_efff_ffff_ffff,
				0xff00_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: Float64(math.Inf(1)),
			want: Float256{
				0x7fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: Float64(math.Inf(-1)),
			want: Float256{
				0xffff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: Float64(math.NaN()),
			want: Float256{
				0x7fff_f800_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float256(); !eq256(got, tt.want) {
			t.Errorf("Float64(%x).Float256() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Float256(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Float256())
	}
}

func TestFloat128_Float16(t *testing.T) {
	tests := []struct {
		in   Float128
		want Float16
	}{
		{
			in:   Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0
			want: 0x0000,
		},
		{
			in:   Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0
			want: 0x8000,
		},
		{
			in:   Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1
			want: 0x3c00,
		},
		{
			in:   Float128{0xc000_0000_0000_0000, 0x0000_0000_0000_0000}, // -2
			want: 0xc000,
		},
		{
			in:   Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // infinity
			want: 0x7c00,
		},
		{
			in:   Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}, // -infinity
			want: 0xfc00,
		},
		{
			in:   Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: 0x7e00,
		},
		{
			in:   Float128{0x3fe7_0000_0000_0000, 0x0000_0000_0000_0000}, // 0x1p-24
			want: 0x0001,
		},

		// test rounding to even
		{
			in:   Float128{0x3fff_0020_0000_0000, 0x0000_0000_0000_0000}, // 0x1.002p+00
			want: 0x3c00,                                                 // 0x1p+00
		},
		{
			in:   Float128{0x3fff_0020_0000_0000, 0x0000_0000_0000_0001}, // 0x1.0020000000000000000000000001p+00
			want: 0x3c01,                                                 // 0x1.004p+00
		},
		{
			in:   Float128{0x3fff_005f_ffff_ffff, 0xffff_ffff_ffff_ffff}, // 0x1.005fffffffffffffffffffffffffp+00
			want: 0x3c01,                                                 // 0x1.004p+00
		},
		{
			in:   Float128{0x3fff_0060_0000_0000, 0x0000_0000_0000_0000}, // 0x1.006p+00
			want: 0x3c02,                                                 // 0x1.008p+00
		},
		{
			in:   Float128{0x3fe7_7fff_ffff_ffff, 0xffff_ffff_ffff_ffff}, // 0x1.7fffffffffffffffffffffffffffp-24
			want: 0x0001,                                                 // 0x1p-24
		},
		{
			in:   Float128{0x3fe7_8000_0000_0000, 0x0000_0000_0000_0000}, // 0x1.8p-24
			want: 0x0002,                                                 // 0x1p-23
		},

		// overflow
		{
			in:   Float128{0x43fe_7fd1_da78_7c72, 0xdb6b_d758_0349_b96f},
			want: 0x7c00,
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float16(); !eq16(got, tt.want) {
			t.Errorf("Float128(%x).Float16() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Float16(b *testing.B) {
	f := Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Float16())
	}
}

func TestFloat128_Float32(t *testing.T) {
	tests := []struct {
		in   Float128
		want Float32
	}{
		{
			in:   Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0
			want: 0,
		},
		{
			in:   Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0
			want: Float32(math.Copysign(0, -1)),
		},
		{
			in:   Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1
			want: 1,
		},
		{
			in:   Float128{0xc000_0000_0000_0000, 0x0000_0000_0000_0000}, // -2
			want: -2,
		},
		{
			in:   Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // infinity
			want: Float32(math.Inf(1)),
		},
		{
			in:   Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}, // -infinity
			want: Float32(math.Inf(-1)),
		},
		{
			in:   Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: Float32(math.NaN()),
		},

		// test rounding to even
		{
			in:   Float128{0x3fff_0000_0100_0000, 0x0000_0000_0000_0000}, // 1.000001p+00
			want: 0x1p+00,
		},
		{
			in:   Float128{0x3fff_0000_0100_0000, 0x0000_0000_0000_0001}, // 1.0000010000000000000000000001p+00
			want: 0x1.000002p+00,
		},
		{
			in:   Float128{0x3fff_0000_02ff_ffff, 0xffff_ffff_ffff_ffff}, // 1.000002ffffffffffffffffffffffp+00
			want: 0x1.000002p+00,
		},
		{
			in:   Float128{0x3fff_0000_0300_0000, 0x0000_0000_0000_0000}, // 1.000003p+00
			want: 0x1.000004p+00,
		},
		{
			in:   Float128{0x3f80_0000_0200_0000, 0x0000_0000_0000_0000}, // 0x1.000002p-127
			want: 0x1.000000p-127,
		},
		{
			in:   Float128{0x3f80_0000_0200_0000, 0x0000_0000_0000_0001}, // 0x1.0000020000000000000000000001p-127
			want: 0x1.000004p-127,
		},

		// overflow
		{
			in:   Float128{0x407e_ffff_fe00_0000, 0x0000_0000_0000_0000},
			want: 0x1.fffffep+127, // largest normal number
		},
		{
			in:   Float128{0x407e_ffff_ff00_0000, 0x0000_0000_0000_0000},
			want: Float32(math.Inf(1)),
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float32(); !eq32(got, tt.want) {
			t.Errorf("Float128(%x).Float32() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Float32(b *testing.B) {
	f := Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Float32())
	}
}

func TestFloat128_Float64(t *testing.T) {
	tests := []struct {
		in   Float128
		want Float64
	}{
		{
			in:   Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0
			want: 0,
		},
		{
			in:   Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0
			want: Float64(math.Copysign(0, -1)),
		},
		{
			in:   Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1
			want: 1,
		},
		{
			in:   Float128{0xc000_0000_0000_0000, 0x0000_0000_0000_0000}, // -2
			want: -2,
		},
		{
			in:   Float128{0x3bcd_0000_0000_0000, 0x0000_0000_0000_0000},
			want: 0x1p-1074, // smallest positive subnormal number
		},
		{
			in:   Float128{0x3c00_ffff_ffff_ffff, 0xe000_0000_0000_0000},
			want: 0x1.ffffffffffffep-1023, // largest subnormal number
		},
		{
			in:   Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // infinity
			want: Float64(math.Inf(1)),
		},
		{
			in:   Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}, // -infinity
			want: Float64(math.Inf(-1)),
		},
		{
			in:   Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: Float64(math.NaN()),
		},
		{
			in:   Float128{0x3e52ffffffffffff, 0xf801ffffffffffff},
			want: 0x1p-428,
		},

		// test rounding to even
		{
			in:   Float128{0x3fff_0000_0000_0000, 0x0800_0000_0000_0000}, // 0x1.00000000000008p+00
			want: 0x1p+00,
		},
		{
			in:   Float128{0x3fff_0000_0000_0000, 0x0800_0000_0000_0001}, // 0x1.0000000000000800000000000001p+00
			want: 0x1.0000000000001p+00,
		},
		{
			in:   Float128{0x3fff_0000_0000_0000, 0x17ff_ffff_ffff_ffff}, // 0x1.00000000000017ffffffffffffffp+00
			want: 0x1.0000000000001p+00,
		},
		{
			in:   Float128{0x3fff_0000_0000_0000, 0x1800_0000_0000_0000}, // 0x1.00000000000018p+00
			want: 0x1.0000000000002p+00,
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float64(); !eq64(got, tt.want) {
			t.Errorf("Float128(%x).Float64() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Float64(b *testing.B) {
	f := Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Float64())
	}
}

func TestFloat128_Float256(t *testing.T) {
	tests := []struct {
		in   Float128
		want Float256
	}{
		// from https://en.wikipedia.org/wiki/Quadruple-precision_floating-point_format
		{
			in: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0
			want: Float256{
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0
			want: Float256{
				0x8000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			in: Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1
			want: Float256{
				0x3fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// smallest positive subnormal number
			in: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001},
			want: Float256{
				0x3bf9_1000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// largest subnormal number
			in: Float128{0x0000_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff},
			want: Float256{
				0x3c00_0fff_ffff_ffff,
				0xffff_ffff_ffff_ffff,
				0xe000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// smallest positive normal number
			in: Float128{0x0001_0000_0000_0000, 0x0000_0000_0000_0000},
			want: Float256{
				0x3c00_1000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// largest normal number
			in: Float128{0x7ffe_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff},
			want: Float256{
				0x43ff_efff_ffff_ffff,
				0xffff_ffff_ffff_ffff,
				0xf000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// largest number less than one
			in: Float128{0x3ffe_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff},
			want: Float256{
				0x3fff_efff_ffff_ffff,
				0xffff_ffff_ffff_ffff,
				0xf000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// smallest number larger than one
			in: Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0001},
			want: Float256{
				0x3fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x1000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// 2
			in: Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000},
			want: Float256{
				0x4000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// -2
			in: Float128{0xc000_0000_0000_0000, 0x0000_0000_0000_0000},
			want: Float256{
				0xc000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// infinity
			in: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000},
			want: Float256{
				0x7fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// -infinity
			in: Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000},
			want: Float256{
				0xffff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// NaN
			in: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000},
			want: Float256{
				0x7fff_f800_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// closest approximation to π
			in: Float128{0x4000_921f_b544_42d1, 0x8469_898c_c517_01b8},
			want: Float256{
				0x4000_0921_fb54_442d,
				0x1846_9898_cc51_701b,
				0x8000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
		{
			// closest approximation to 1/3
			in: Float128{0x3ffd_5555_5555_5555, 0x5555_5555_5555_5555},
			want: Float256{
				0x3fff_d555_5555_5555,
				0x5555_5555_5555_5555,
				0x5000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float256(); !eq256(got, tt.want) {
			t.Errorf("Float128(%x).Float256() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Float256(b *testing.B) {
	f := Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Float256())
	}
}

func TestFloat256_Float16(t *testing.T) {
	tests := []struct {
		in   Float256
		want Float16
	}{
		{
			// 0.0
			in: Float256{
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x0000,
		},
		{
			// 1.0
			in: Float256{
				0x3fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x3c00,
		},
		{
			// -0.0
			in: Float256{
				0x8000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x8000,
		},
		{
			// Infinity
			in: Float256{
				0x7fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x7c00,
		},
		{
			// -Infinity
			in: Float256{
				0xffff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0xfc00,
		},
		{
			// NaN
			in: Float256{
				0x7fff_f800_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x7e00,
		},

		// test rounding to even
		{
			// 0x1.002p+00
			in: Float256{
				0x3fff_f002_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x3c00, // 1.0
		},
		{
			// 0x1.00200000000000000000000000000000000000000000000000000000001p+00
			in: Float256{
				0x3fff_f002_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0001,
			},
			want: 0x3c01, // 0x1.004p+00
		},
		{
			// 0x1.005ffffffffffffffffffffffffffffffffffffffffffffffffffffffffp+00
			in: Float256{
				0x3fff_f005_ffff_ffff,
				0xffff_ffff_ffff_ffff,
				0xffff_ffff_ffff_ffff,
				0xffff_ffff_ffff_ffff,
			},
			want: 0x3c01, // 0x1.004p+00
		},
		{
			// 0x1.006p+00
			in: Float256{
				0x3fff_f006_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x3c02, // 0x1.008p+00
		},
		{
			// 0x1.004p-15
			in: Float256{
				0x3fff_0004_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x0200, // 0x1p-15
		},
		{
			// 0x1.00400000000000000000000000000000000000000000000000000000001p-15
			in: Float256{
				0x3fff_0004_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0001,
			},
			want: 0x0201, // 0x1.008p-15
		},
		{
			// 0x1.00bffffffffffffffffffffffffffffffffffffffffffffffffffffffffp-15
			in: Float256{
				0x3fff_000b_ffff_ffff,
				0xffff_ffff_ffff_ffff,
				0xffff_ffff_ffff_ffff,
				0xffff_ffff_ffff_ffff,
			},
			want: 0x0201, // 0x1.008p-15
		},
		{
			// 0x1.00cp-15
			in: Float256{
				0x3fff_000c_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x0202, // 0x1.01p-15
		},

		// overflow
		{
			// 65520
			in: Float256{
				0x4000_effe_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x7c00,
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float16(); !eq16(got, tt.want) {
			t.Errorf("Float256(%x).Float16() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Float16(b *testing.B) {
	f := Float256{
		0x3fff_f000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
	} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Float16())
	}
}

func TestFloat256_Float32(t *testing.T) {
	tests := []struct {
		in   Float256
		want Float32
	}{
		{
			in: Float256{
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0.0,
		},
		{
			in: Float256{
				0x3fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 1.0,
		},
		{
			// -0.0
			in: Float256{
				0x8000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: Float32(math.Copysign(0, -1)),
		},
		{
			// Infinity
			in: Float256{
				0x7fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: Float32(math.Inf(1)),
		},
		{
			// -Infinity
			in: Float256{
				0xffff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: Float32(math.Inf(-1)),
		},
		{
			// NaN
			in: Float256{
				0x7fff_f800_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: Float32(math.NaN()),
		},

		// test rounding to nearest even
		{
			// 1.000001p+00
			in: Float256{
				0x3fff_f000_0010_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x1p+00,
		},
		{
			// 1.00000100000000000000000000000000000000000000000000000000001p+00
			in: Float256{
				0x3fff_f000_0010_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0001,
			},
			want: 0x1.000002p+00,
		},
		{
			// 1.000002fffffffffffffffffffffffffffffffffffffffffffffffffffffp+00
			in: Float256{
				0x3fff_f000_002f_ffff,
				0xffff_ffff_ffff_ffff,
				0xffff_ffff_ffff_ffff,
				0xffff_ffff_ffff_ffff,
			},
			want: 0x1.000002p+00,
		},
		{
			// 1.000003p+00
			in: Float256{
				0x3fff_f000_0030_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x1.000004p+00,
		},
		{
			// 1.000002p-127
			in: Float256{
				0x3ff8_0000_0020_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x1.000000p-127,
		},
		{
			// 1.00000200000000000000000000000000000000000000000000000000001p+00
			in: Float256{
				0x3ff8_0000_0020_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0001,
			},
			want: 0x1.000004p-127,
		},
		{
			// 1.000005fffffffffffffffffffffffffffffffffffffffffffffffffffffp+00
			in: Float256{
				0x3ff8_0000_005f_ffff,
				0xffff_ffff_ffff_ffff,
				0xffff_ffff_ffff_ffff,
				0xffff_ffff_ffff_ffff,
			},
			want: 0x1.000004p-127,
		},
		{
			// 1.000006p-127
			in: Float256{
				0x3ff8_0000_0060_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x1.000008p-127,
		},

		// overflow
		{
			// 0x1.fffffep+127
			in: Float256{
				0x4007_efff_ffe0_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: 0x1.fffffep+127, // largest normal number
		},
		{
			// 0x1.fffffep+127
			in: Float256{
				0x4007_efff_fff0_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: Float32(math.Inf(1)),
		},
	}

	for _, tt := range tests {
		if got := tt.in.Float32(); !eq32(got, tt.want) {
			t.Errorf("Float256(%x).Float32() = %x, want %x", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Float32(b *testing.B) {
	f := Float256{
		0x3fff_f000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
	} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Float32())
	}
}
