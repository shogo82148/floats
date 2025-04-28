package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat64_IsNaN(t *testing.T) {
	tests := []struct {
		name string
		a    Float64
		want bool
	}{
		{
			name: "NaN",
			a:    Float64(math.NaN()),
			want: true,
		},
		{
			name: "Inf",
			a:    Float64(math.Inf(1)),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.IsNaN(); got != tt.want {
				t.Errorf("Float64.IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat64_IsInf(t *testing.T) {
	inf := Float64(math.Inf(1))
	neginf := Float64(math.Inf(-1))

	tests := []struct {
		in   Float64
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
			t.Errorf("Float64.IsInf(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat64_IsInf(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.IsInf(0))
	}
}

func TestFloat64_Int64(t *testing.T) {
	tests := []struct {
		in   Float64
		want int64
	}{
		{0, 0},
		{1, 1},
		{0.5, 0},
		{-1, -1},
	}

	for _, test := range tests {
		got := test.in.Int64()
		if got != test.want {
			t.Errorf("Float64.Int64(%v) = %v, want %v", test.in, got, test.want)
		}
	}
}

func TestFloat64_Mul(t *testing.T) {
	tests := []struct {
		a, b Float64
		want Float64
	}{
		{1.0, 2.0, 2.0},
		{1.0, 0.0, 0.0},
		{0.0, 1.0, 0.0},
	}

	for _, test := range tests {
		got := test.a.Mul(test.b)
		if !eq64(got, test.want) {
			t.Errorf("Float64(%x).Mul(%x) = %x, want %x", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat64_Mul(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Mul(f))
	}
}
