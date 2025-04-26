package floats

import (
	"runtime"
	"testing"
)

func TestFloat16_IsNaN(t *testing.T) {
	tests := []struct {
		name string
		a    Float16
		want bool
	}{
		{
			name: "NaN",
			a:    0x7e00,
			want: true,
		},
		{
			name: "not NaN",
			a:    0x7c00,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.IsNaN(); got != tt.want {
				t.Errorf("Float16.IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFloat16_IsInf(t *testing.T) {
	tests := []struct {
		in   Float16
		sign int
		want bool
	}{
		// infinity
		{0x7c00, 1, true},
		{0x7c00, -1, false},
		{0x7c00, 0, true},

		// -infinity
		{0xfc00, 1, false},
		{0xfc00, -1, true},
		{0xfc00, 0, true},

		// +1.0(finite)
		{0x3c00, 1, false},
		{0x3c00, -1, false},
		{0x3c00, 0, false},
	}

	for _, tt := range tests {
		got := tt.in.IsInf(tt.sign)
		if got != tt.want {
			t.Errorf("Float16.IsInf(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func BenchmarkFloat16_IsInf(b *testing.B) {
	f := Float16(0x3c00)
	for b.Loop() {
		runtime.KeepAlive(f.IsInf(0))
	}
}
