package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat32_IsNaN(t *testing.T) {
	tests := []struct {
		name string
		a    Float32
		want bool
	}{
		{
			name: "NaN",
			a:    Float32(math.NaN()),
			want: true,
		},
		{
			name: "Inf",
			a:    Float32(math.Inf(1)),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.IsNaN(); got != tt.want {
				t.Errorf("Float32.IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat32_IsInf(t *testing.T) {
	inf := Float32(math.Inf(1))
	neginf := Float32(math.Inf(-1))

	tests := []struct {
		in   Float32
		sign int
		want bool
	}{
		// infinity
		{inf, 1, true},
		{inf, -1, false},
		{inf, 0, true},

		// -infinity
		{neginf, 1, false},
		{neginf, -1, true},
		{neginf, 0, true},

		// +1.0(finite)
		{1.0, 1, false},
		{1.0, -1, false},
		{1.0, 0, false},
	}

	for _, tt := range tests {
		got := tt.in.IsInf(tt.sign)
		if got != tt.want {
			t.Errorf("Float32.IsInf(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat32_IsInf(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.IsInf(0))
	}
}

func TestFloat32_Int64(t *testing.T) {
	tests := []struct {
		in  Float32
		out int64
	}{
		{0, 0},
		{1, 1},
		{2, 2},
		{-1, -1},
		{0.5, 0},
		{0x1p24, 1 << 24},
		{0x1p62, 1 << 62},
	}

	for _, test := range tests {
		got := test.in.Int64()
		if got != test.out {
			t.Errorf("Float32.Int64() = %v, want %v", got, test.out)
		}
	}
}
