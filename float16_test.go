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

func TestFloat16_Int64(t *testing.T) {
	tests := []struct {
		in  Float16
		out int64
	}{
		{0x0000, 0},
		{0x3c00, 1},
	}
	for _, test := range tests {
		got := test.in.Int64()
		if got != test.out {
			t.Errorf("Float16.Int64() = %v, want %v", got, test.out)
		}
	}
}

func BenchmarkFloat16_Int64(b *testing.B) {
	f := Float16(0x3c00)
	for b.Loop() {
		runtime.KeepAlive(f.Int64())
	}
}
