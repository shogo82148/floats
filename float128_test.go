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
	f := Float128{0x3fff_0000_0000_0000, 0x0000_0000_0000_0000}
	for b.Loop() {
		runtime.KeepAlive(f.IsInf(0))
	}
}
