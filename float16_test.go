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

func TestFloat16_Mul(t *testing.T) {
	tests := []struct {
		a, b, want Float16
	}{
		{0x3c00, 0x0000, 0x0000}, // 1.0 * 0.0 = 0.0
		{0x3c00, 0x3c00, 0x3c00}, // 1.0 * 1.0 = 1.0

		// handling zero
		{0x0000, 0x3c00, 0x0000}, //  0.0 *  1.0 =  0.0
		{0x8000, 0x3c00, 0x8000}, // -0.0 *  1.0 = -0.0
		{0x0000, 0xbc00, 0x8000}, //  0.0 * -1.0 = -0.0
		{0x8000, 0xbc00, 0x0000}, // -0.0 * -1.0 =  0.0

		// handling NaN
		{0x7e00, 0x0000, 0x7e00}, // NaN * 0.0 = NaN
		{0x0000, 0x7e00, 0x7e00}, // 0.0 * NaN = NaN

		// handling infinity
		{0x3c00, 0x7c00, 0x7c00}, //  1.0 *  Inf =  Inf
		{0xbc00, 0x7c00, 0xfc00}, // -1.0 *  Inf = -Inf
		{0x7c00, 0x3c00, 0x7c00}, //  Inf *  1.0 =  Inf
		{0x7c00, 0xbc00, 0xfc00}, //  Inf * -1.0 = -Inf
		{0x7c00, 0x0000, 0x7e00}, //  Inf *  0.0 =  NaN
		{0x0000, 0x7c00, 0x7e00}, //  0.0 *  Inf =  NaN
	}

	for _, test := range tests {
		got := test.a.Mul(test.b)
		if !eq16(got, test.want) {
			t.Errorf("Float16(%x).Mul(%x) = %x, want %x", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat16_Mul(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Mul(f))
	}
}
