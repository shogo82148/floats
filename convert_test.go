package floats

import (
	"cmp"
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Float32(t *testing.T) {
	tests := []struct {
		in   Float16
		want Float32
	}{
		// from https://en.wikipedia.org/wiki/Half-precision_floating-point_format
		{0x0000, 0},
		{0x0001, 0x1p-24},              // smallest positive subnormal number
		{0x03ff, 0x1.ff8p-15},          // largest positive subnormal number
		{0x0400, 0x1p-14},              // smallest positive normal number
		{0x3555, 0x1.554p-02},          // nearest value to 1/3
		{0x3bff, 0x1.ffcp-01},          // largest number less than one
		{0x3c00, 0x1p+00},              // one
		{0x3c01, 0x1.004p+00},          // smallest number larger than one
		{0x7bff, 0x1.ffcp+15},          // largest normal number
		{0x7c00, Float32(math.Inf(1))}, // infinity
		{0x8000, -0},
		{0xc000, -2},
		{0xfc00, Float32(math.Inf(-1))}, // negative infinity
		{0x7e00, Float32(math.NaN())},   // NaN
	}

	for _, tt := range tests {
		// we use cmp.Compare because NaN != NaN
		// and we want to check if they are equal.
		if got := tt.in.Float32(); cmp.Compare(got, tt.want) != 0 {
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
		{0x0001, 0x1p-24},              // smallest positive subnormal number
		{0x03ff, 0x1.ff8p-15},          // largest positive subnormal number
		{0x0400, 0x1p-14},              // smallest positive normal number
		{0x3555, 0x1.554p-02},          // nearest value to 1/3
		{0x3bff, 0x1.ffcp-01},          // largest number less than one
		{0x3c00, 0x1p+00},              // one
		{0x3c01, 0x1.004p+00},          // smallest number larger than one
		{0x7bff, 0x1.ffcp+15},          // largest normal number
		{0x7c00, Float64(math.Inf(1))}, // infinity
		{0x8000, -0},
		{0xc000, -2},
		{0xfc00, Float64(math.Inf(-1))}, // negative infinity
		{0x7e00, Float64(math.NaN())},   // NaN
	}

	for _, tt := range tests {
		// we use cmp.Compare because NaN != NaN
		// and we want to check if they are equal.
		if got := tt.in.Float64(); cmp.Compare(got, tt.want) != 0 {
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
		if got := tt.in.Float32(); got != tt.want {
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
		if got := tt.in.Float64(); got != tt.want {
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
		if got := tt.in.Float32(); got != tt.want {
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
		if got := tt.in.Float64(); got != tt.want {
			t.Errorf("Float64(%x).Float64() = %x, want %x", tt.in, got, tt.want)
		}
	}
}
