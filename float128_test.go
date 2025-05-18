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

func TestFloat128_Neg(t *testing.T) {
	tests := []struct {
		a, want Float128
	}{
		{
			Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			Float128{0xbfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -1.0
		},
		{
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
		},
		{
			Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},
		{
			Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}, // -inf
		},
		{
			Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}, // -inf
			Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
		},
		{
			Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},
	}

	for _, tt := range tests {
		got := tt.a.Neg()
		if !eq128(got, tt.want) {
			t.Errorf("Float128(%x).Neg() = %x, want %x", tt.a, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Neg(b *testing.B) {
	f := Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Neg())
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
		{
			a:    Float128{0x400c_fff7_ffff_fffe, 0x0000_0000_0000_0000},
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001},
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_3fff},
		},
		{
			a:    Float128{0x0002_ffff_ffff_ffc0, 0x0000_8000_0000_0000},
			b:    Float128{0xc061_03ff_ffff_ffff, 0xff7f_ffff_ffff_ffff},
			want: Float128{0x8065_03ff_ffff_ffdf, 0x7f80_4100_0000_0fff},
		},
		// underflow
		{
			a:    Float128{0x3ffd_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.25
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001},
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},

		// overflow
		{
			a:    Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}, // 2.0
			b:    Float128{0x7ffe_0000_0000_0000, 0x0000_0000_0000_0000},
			want: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
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

func TestFloat128_Quo(t *testing.T) {
	tests := []struct {
		a, b, want Float128
	}{
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}, // 2.0
			want: Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.5
		},
		{
			a:    Float128{0x8002_ffdf_ffff_ffff, 0xffff_f7ff_ffff_ffff},
			b:    Float128{0x4001_0000_0000_0000, 0x0000_0000_0000_0001},
			want: Float128{0x8000_ffef_ffff_ffff, 0xffff_fbff_ffff_ffff},
		},

		// overflow
		{
			a:    Float128{0x401c_f7df_ffff_ffff, 0xffff_ffff_ffff_fffe},
			b:    Float128{0x8000_ffff_ffe0_0000, 0x0000_0000_0000_00ff},
			want: Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000},
		},

		// underflow
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // 0x1p-16494
			b:    Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}, // 2.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},

		// zero division
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // Inf
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},

		// handling NaN
		{
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},

		// handling infinity
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // inf
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},
		{
			a:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // inf
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // inf
		},
		{
			a:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // inf
			b:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // inf
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},
	}

	for _, tt := range tests {
		got := tt.a.Quo(tt.b)
		if !eq128(got, tt.want) {
			t.Errorf("Float128(%x).Quo(%x) = %x, want %x", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Quo(b *testing.B) {
	f := Float128{
		0x3fff_0000_0000_0000,
		0x0000_0000_0000_0000,
	}
	for b.Loop() {
		runtime.KeepAlive(f.Quo(f))
	}
}

func TestFloat128_Add(t *testing.T) {
	tests := []struct {
		a, b, want Float128
	}{
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}, // 2.0
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // 0x1p-16494
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // 0x1p-16494
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0002}, // 0x1p-16493
		},

		// overflow
		{
			a:    Float128{0x7ffe_27ee_6958_453e, 0x5f71_b167_202f_74e1},
			b:    Float128{0x7ffd_b530_03fc_6745, 0x6d54_ad2a_e8b8_4504},
			want: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
		},

		// handling zero
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},
		{
			a:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},
		{
			a:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			b:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			want: Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
		},

		// handling NaN
		{
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
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
			a:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
		},
		{
			a:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			b:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			want: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
		},
		{
			a:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			b:    Float128{0xffff_0000_0000_0000, 0x0000_0000_0000_0000}, // -inf
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},
	}

	for _, tt := range tests {
		got := tt.a.Add(tt.b)
		if !eq128(got, tt.want) {
			t.Errorf("Float128(%x).Add(%x) = %x, want %x", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Add(b *testing.B) {
	f := Float128{
		0x3fff_0000_0000_0000,
		0x0000_0000_0000_0000,
	}
	for b.Loop() {
		runtime.KeepAlive(f.Add(f))
	}
}

func TestFloat128_Sub(t *testing.T) {
	tests := []struct {
		a, b, want Float128
	}{
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}, // 2.0
			want: Float128{0xbfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -1.0
		},

		// handling zero
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0xbfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -1.0
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},
		{
			a:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
		},
		{
			a:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			b:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},

		// handling NaN
		{
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},

		// handling infinity
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
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
			b:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},
	}

	for _, tt := range tests {
		got := tt.a.Sub(tt.b)
		if !eq128(got, tt.want) {
			t.Errorf("Float128(%x).Sub(%x) = %x, want %x", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Sub(b *testing.B) {
	f := Float128{
		0x3fff_0000_0000_0000,
		0x0000_0000_0000_0000,
	}
	for b.Loop() {
		runtime.KeepAlive(f.Sub(f))
	}
}

func TestFloat128_Sqrt(t *testing.T) {
	tests := []struct {
		a, want Float128
	}{
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
		},
		{
			a:    Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}, // 2.0
			want: Float128{0x3fff_6a09_e667_f3bc, 0xc908_b2fb_1366_ea95}, // 1.414213562373095
		},

		// special cases
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},
		{
			a:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			want: Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
		},
		{
			a:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			want: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
		},
		{
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
		},
	}

	for _, tt := range tests {
		got := tt.a.Sqrt()
		if !eq128(got, tt.want) {
			t.Errorf("Float128(%x).Sqrt() = %x, want %x", tt.a, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Sqrt(b *testing.B) {
	f := Float128{
		0x4000_0000_0000_0000,
		0x0000_0000_0000_0000,
	} // 2.0
	for b.Loop() {
		runtime.KeepAlive(f.Sqrt())
	}
}

func TestFloat128_Eq(t *testing.T) {
	tests := []struct {
		a, b Float128
		want bool
	}{
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // 0x1p-16494
			want: false,
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0001}, // 1.0
			b:    Float128{0x3fff_0000_0000_0001, 0x0000_0000_0000_0001}, // 1.5
			want: false,
		},
		{
			a:    Float128{0x3fff_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			b:    Float128{0x3fff_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			want: true,
		},

		// handling zero
		{
			a:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: true,
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			want: true,
		},

		// handling NaN
		{
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: false,
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: false,
		},
		{
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			b:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: false,
		},
	}

	for _, tt := range tests {
		got := tt.a.Eq(tt.b)
		if got != tt.want {
			t.Errorf("Float128(%x).Eq(%x) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Eq(b *testing.B) {
	f := Float128{
		0x3fff_0000_0000_0000,
		0x0000_0000_0000_0000,
	}
	for b.Loop() {
		runtime.KeepAlive(f.Eq(f))
	}
}

func TestFloat128_Ne(t *testing.T) {
	tests := []struct {
		a, b Float128
		want bool
	}{
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // 0x1p-16494
			want: true,
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0001}, // 1.0
			b:    Float128{0x3fff_0000_0000_0001, 0x0000_0000_0000_0001}, // 1.5
			want: true,
		},
		{
			a:    Float128{0x3fff_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			b:    Float128{0x3fff_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			want: false,
		},

		// handling zero
		{
			a:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0001}, // -1.5
			b:    Float128{0x8002_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			want: true,
		},

		// handling NaN
		{
			a:    Float128{0x7fff_8002_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff}, // NaN
			b:    Float128{0x8002_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			want: true,
		},
	}

	for _, tt := range tests {
		got := tt.a.Ne(tt.b)
		if got != tt.want {
			t.Errorf("Float128(%x).Ne(%x) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Ne(b *testing.B) {
	f := Float128{
		0x3fff_0000_0000_0000,
		0x0000_0000_0000_0000,
	}
	for b.Loop() {
		runtime.KeepAlive(f.Ne(f))
	}
}

func TestFloat128_Lt(t *testing.T) {
	tests := []struct {
		a, b Float128
		want bool
	}{
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // 0x1p-16494
			want: true,
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0001}, // 1.0
			b:    Float128{0x3fff_0000_0000_0001, 0x0000_0000_0000_0001}, // 1.5
			want: true,
		},
		{
			a:    Float128{0xcfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -2.0
			b:    Float128{0xbfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -1.0
			want: true,
		},
		{
			a:    Float128{0xc000_0000_0000_0000, 0x0000_0000_03ff_ff00},
			b:    Float128{0xc000_0000_0000_0000, 0x0000_0000_0000_0001},
			want: true,
		},

		// handling zero
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			want: false,
		},
		{
			a:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: false,
		},

		// handling NaN
		{
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: false,
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: false,
		},
		{
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			b:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: false,
		},
	}

	for _, tt := range tests {
		got := tt.a.Lt(tt.b)
		if got != tt.want {
			t.Errorf("Float128(%x).Lt(%x) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Lt(b *testing.B) {
	f := Float128{
		0x3fff_0000_0000_0000,
		0x0000_0000_0000_0000,
	}
	for b.Loop() {
		runtime.KeepAlive(f.Lt(f))
	}
}

func TestFloat128_Gt(t *testing.T) {
	tests := []struct {
		a, b Float128
		want bool
	}{
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // 0x1p-16494
			want: false,
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // 0x1p-16494
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: true,
		},

		// handling zero
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			want: false,
		},
		{
			a:    Float128{0x8000_0000_0000_0000, 0x0000_0000_0000_0000}, // -0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: false,
		},

		// handling NaN
		{
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: false,
		},
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: false,
		},
		{
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			b:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000}, // NaN
			want: false,
		},
	}

	for _, tt := range tests {
		got := tt.a.Gt(tt.b)
		if got != tt.want {
			t.Errorf("Float128(%x).Gt(%x) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Gt(b *testing.B) {
	f := Float128{
		0x3fff_0000_0000_0000,
		0x0000_0000_0000_0000,
	} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Gt(f))
	}
}

func TestFloat128_Le(t *testing.T) {
	tests := []struct {
		a, b Float128
		want bool
	}{
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // 0x1p-16494
			want: true,
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0001}, // 1.5
			b:    Float128{0x3fff_0000_0000_0001, 0x0000_0000_0000_0001}, // 1.5
			want: true,
		},
		{
			a:    Float128{0xcfff_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			b:    Float128{0xbfff_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			want: true,
		},

		// handling zero
		{
			a:    Float128{0x8002_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			b:    Float128{0x8002_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			want: true,
		},

		// handling NaN
		{
			a:    Float128{0x7fff_8002_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff}, // NaN
			b:    Float128{0x8002_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			want: false,
		},
	}

	for _, tt := range tests {
		got := tt.a.Le(tt.b)
		if got != tt.want {
			t.Errorf("Float128(%x).Le(%x) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Le(b *testing.B) {
	f := Float128{
		0x3fff_0000_0000_0000,
		0x0000_0000_0000_0000,
	} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Le(f))
	}
}

func TestFloat128_Ge(t *testing.T) {
	tests := []struct {
		a, b Float128
		want bool
	}{
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}, // 0x1p-16494
			want: false,
		},
		{
			a:    Float128{0x3fff_0000_0000_0001, 0x0000_0000_0000_0001}, // 1.5
			b:    Float128{0x3fff_0000_0000_0001, 0x0000_0000_0000_0001}, // 1.5
			want: true,
		},
		{
			a:    Float128{0xcfff_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			b:    Float128{0xbfff_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			want: false,
		},

		// handling zero
		{
			a:    Float128{0x8002_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			b:    Float128{0x8002_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			want: true,
		},

		// handling NaN
		{
			a:    Float128{0x7fff_8002_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff}, // NaN
			b:    Float128{0x8002_ffdf_ffdf_ffdf, 0xffff_f7ff_f7ff_f7ff},
			want: false,
		},
	}

	for _, tt := range tests {
		got := tt.a.Ge(tt.b)
		if got != tt.want {
			t.Errorf("Float128(%x).Ge(%x) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Ge(b *testing.B) {
	f := Float128{
		0x3fff_0000_0000_0000,
		0x0000_0000_0000_0000,
	} // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Ge(f))
	}
}

func TestFMA128(t *testing.T) {
	tests := []struct {
		a, b, c, want Float128
	}{
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			c:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			want: Float128{0x4000_0000_0000_0000, 0x0000_0000_0000_0000}, // 2.0
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			c:    Float128{0xc000_0000_0000_0000, 0x0000_0000_0000_0000}, // -2.0
			want: Float128{0xbfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -1.0
		},
		{
			a:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			b:    Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}, // 1.0
			c:    Float128{0xbfff_0000_0000_0000, 0x0000_0000_0000_0000}, // -1.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},

		// random values
		{
			a:    Float128{0x3ffe_6a5e_606a_9e52, 0xb658_e6e3_8a0d_b849},
			b:    Float128{0x4400_0000_0000_0000, 0x0000_03ff_ffff_ffff},
			c:    Float128{0x43ff_0d90_90d6_c6aa, 0x6b05_9e0f_0fdd_ea1a},
			want: Float128{0x4400_3bf7_78a0_b27e, 0x90af_454e_09b6_a66d},
		},
		{
			a:    Float128{0x3f80_0000_0000_0000, 0x0040_0000_4000_0000},
			b:    Float128{0x4769_0000_0000_0000, 0x0000_0000_1fff_ffef},
			c:    Float128{0xc1c0_ffff_ffff_ffff, 0xffff_ffff_ffff_ff80},
			want: Float128{0x46ea_0000_0000_0000, 0x0040_0000_5fff_ffef},
		},

		// handling zero, Inf, NaN
		{
			a:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			b:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			c:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
			want: Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000}, // 0.0
		},
		{
			a:    Float128{0x0002_ffff_ffff_ffc0, 0x0000_8000_0000_0000},
			b:    Float128{0xc061_03ff_ffff_ffff, 0xff7f_ffff_ffff_ffff},
			c:    Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0000},
			want: Float128{0x8065_03ff_ffff_ffdf, 0x7f80_4100_0000_0fff},
		},
		{
			a:    Float128{0x0002_ffff_ffff_ffc0, 0x0000_8000_0000_0000},
			b:    Float128{0xc061_03ff_ffff_ffff, 0xff7f_ffff_ffff_ffff},
			c:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
			want: Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000}, // +inf
		},
	}

	for _, tt := range tests {
		got := FMA128(tt.a, tt.b, tt.c)
		if !eq128(got, tt.want) {
			t.Errorf("FMA128(%x, %x, %x) = %x, want %x", tt.a, tt.b, tt.c, got, tt.want)
		}
	}
}
