package floats

import (
	"math"
	"runtime"
	"testing"
)

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
