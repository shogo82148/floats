package floats

import (
	"runtime"
	"testing"
)

func TestFloat256_IsNaN(t *testing.T) {
	tests := []struct {
		name string
		a    Float256
		want bool
	}{
		{
			name: "NaN",
			a: Float256{
				0x7fff_f800_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: true,
		},
		{
			name: "infinity",
			a: Float256{
				0x7fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.IsNaN(); got != tt.want {
				t.Errorf("Float256.IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat256_IsInf(t *testing.T) {
	inf := Float256{
		0x7fff_f000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
	}
	neginf := Float256{
		0xffff_f000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
	}
	one := Float256{
		0x3fff_f000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
	}

	tests := []struct {
		in   Float256
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

		// +1.1(finite)
		{one, 1, false},
	}

	for _, tt := range tests {
		got := tt.in.IsInf(tt.sign)
		if got != tt.want {
			t.Errorf("Float256.IsInf(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat256_IsInf(b *testing.B) {
	f := Float256{
		0x3fff_f000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
	}
	for b.Loop() {
		runtime.KeepAlive(f.IsInf(0))
	}
}

func TestFloat256_Int64(t *testing.T) {
	tests := []struct {
		in  Float256
		out int64
	}{
		{
			// 0.0
			in: Float256{
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			out: 0,
		},
		{
			// 1.0
			in: Float256{
				0x3fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			out: 1,
		},
		{
			// 0.5
			in: Float256{
				0x3fff_e000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			out: 0,
		},
		{
			// -1.0
			in: Float256{
				0xbfff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			out: -1,
		},
	}

	for _, tt := range tests {
		got := tt.in.Int64()
		if got != tt.out {
			t.Errorf("Float256.Int64() = %v, want %v", got, tt.out)
		}
	}
}

func BenchmarkFloat256_Int64(b *testing.B) {
	f := Float256{
		0x3fff_f000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
		0x0000_0000_0000_0000,
	}
	for b.Loop() {
		runtime.KeepAlive(f.Int64())
	}
}
