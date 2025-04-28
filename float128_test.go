package floats

import (
	"runtime"
	"testing"
)

func TestFloat128_IsNaN(t *testing.T) {
	tests := []struct {
		name string
		a    Float128
		want bool
	}{
		{
			name: "NaN",
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000},
			want: true,
		},
		{
			name: "infinity",
			a:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.IsNaN(); got != tt.want {
				t.Errorf("Float128.IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat128_IsInf(t *testing.T) {
	inf := Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}
	neginf := Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}
	one := Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}

	tests := []struct {
		in   Float128
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
		{one, 1, false},
		{one, -1, false},
		{one, 0, false},
	}

	for _, tt := range tests {
		got := tt.in.IsInf(tt.sign)
		if got != tt.want {
			t.Errorf("Float128.IsInf(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat128_IsInf(b *testing.B) {
	f := Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.IsInf(0))
	}
}

func TestFloat128_Int64(t *testing.T) {
	tests := []struct {
		in  Float128
		out int64
	}{
		{Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, 0},
		{Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, 1},
		{Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}, 2},
	}

	for _, tt := range tests {
		if got := tt.in.Int64(); got != tt.out {
			t.Errorf("Float128.Int64() = %v, want %v", got, tt.out)
		}
	}
}

func BenchmarkFloat128_Int64(b *testing.B) {
	f := Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Int64())
	}
}

func TestFloat128_Mul(t *testing.T) {
	tests := []struct {
		a, b, want Float128
	}{
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}, // 2.0
			want: Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}, // 2.0
		},

		// handling zero
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},
		{
			a:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0xbfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -1.0
			want: Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
		},
		{
			a:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			b:    Float128{0xbfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -1.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},

		// handling NaN
		{
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},

		// handling infinity
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			want: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
		},
		{
			a:    Float128{0xbfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -1.0
			b:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			want: Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}, // -inf
		},
		{
			a:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
		},
		{
			a:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			b:    Float128{0xbfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -1.0
			want: Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}, // -inf
		},
		{
			a:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},
	}

	for _, tt := range tests {
		got := tt.a.Mul(tt.b)
		if !eq128(got, tt.want) {
			t.Errorf("Float128(%x).Mul(%x) = %x, want %x", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Mul(b *testing.B) {
	f := Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Mul(f))
	}
}
