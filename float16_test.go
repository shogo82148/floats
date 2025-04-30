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

func TestFloat16_Neg(t *testing.T) {
	tests := []struct {
		a, want Float16
	}{
		{0x3c00, 0xbc00}, // 1.0 = -1.0
		{0x0000, 0x8000}, // 0.0 = -0.0
		{0x8000, 0x0000}, // -0.0 = 0.0
		{0x7c00, 0xfc00}, // +Inf = -Inf
		{0xfc00, 0x7c00}, // -Inf = +Inf
		{0x7e00, 0x7e00}, // NaN = NaN
	}
	for _, tt := range tests {
		got := tt.a.Neg()
		if !eq16(got, tt.want) {
			t.Errorf("Float16(%x).Neg() = %x, want %x", tt.a, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Neg(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Neg())
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

func TestFloat16_Quo(t *testing.T) {
	tests := []struct {
		a, b, want Float16
	}{
		{0x0000, 0x3c00, 0x0000}, // 0.0 / 1.0 = 0.0
		{0x3c00, 0x3c00, 0x3c00}, // 1.0 / 1.0 = 1.0

		// zero division
		{0x3c00, 0x0000, 0x7c00}, // 1.0 / 0.0 = +Inf
		{0x0000, 0x0000, 0x7e00}, // 0.0 / 0.0 = NaN

		// overflow
		{0x7876, 0x03fe, 0x7c00},

		// the result is subnormal.
		{0x2bf0, 0x6bf8, 0x00ff},

		// handling NaN
		{0x7e00, 0x3c00, 0x7e00}, // NaN / 1.0 = NaN
		{0x3c00, 0x7e00, 0x7e00}, // 1.0 / NaN = NaN

		// handling infinity
		{0x3c00, 0x7c00, 0x0000}, // 1.0 / Inf = 0.0
		{0x7c00, 0x3c00, 0x7c00}, // Inf / 1.0 = Inf
		{0x7c00, 0x7c00, 0x7e00}, // Inf / Inf = NaN
	}

	for _, test := range tests {
		got := test.a.Quo(test.b)
		if !eq16(got, test.want) {
			t.Errorf("Float16(%x).Quo(%x) = %x, want %x", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat16_Quo(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Quo(f))
	}
}

func TestFloat16_Add(t *testing.T) {
	tests := []struct {
		a, b, want Float16
	}{
		{0x3c00, 0x3800, 0x3e00}, // 1.0 +  0.5 = 1.5
		{0x3c00, 0x3c00, 0x4000}, // 1.0 +  1.0 = 2.0
		{0x3c00, 0xbc00, 0x0000}, // 1.0 + -1.0 = 0.0
		{0x2c0d, 0x40c5, 0x40e5},
		{0x0030, 0x8400, 0x83d0},
		{0x0001, 0x03fe, 0x03ff},
		{0x0002, 0x0002, 0x0004}, // 0x1p-23 + 0x1p-23 = 0x1p-22
		{0x0001, 0x0001, 0x0002}, // 0x1p-24 + 0x1p-24 = 0x1p-23

		// overflow
		{0x7bc0, 0x77e8, 0x7c00},

		// adding zeros
		{0x0000, 0x3c00, 0x3c00}, //  0.0 +  1.0 =  1.0
		{0x3c00, 0x0000, 0x3c00}, //  1.0 +  0.0 =  1.0
		{0x0000, 0x0000, 0x0000}, //  0.0 +  0.0 =  0.0
		{0x0000, 0x8000, 0x0000}, //  0.0 + -0.0 =  0.0
		{0x8000, 0x0000, 0x0000}, // -0.0 +  0.0 =  0.0
		{0x8000, 0x8000, 0x8000}, // -0.0 + -0.0 = -0.0

		// handling NaN
		{uvnan16, 0x3c00, uvnan16}, // NaN + 1 = NaN
		{0x3c00, uvnan16, uvnan16}, // 1 + NaN = NaN

		// handling infinity
		{0x7c00, 0x3c00, 0x7c00},  //  Inf +  1.0 = Inf
		{0x3c00, 0x7c00, 0x7c00},  //  1.0 +  Inf = Inf
		{0x7c00, 0x7c00, 0x7c00},  //  Inf +  inf = Inf
		{0x7c00, 0xfc00, uvnan16}, //  Inf + -inf = NaN
		{0xfc00, 0x7c00, uvnan16}, // -inf + Inf = NaN
	}

	for _, test := range tests {
		got := test.a.Add(test.b)
		if !eq16(got, test.want) {
			t.Errorf("Float16(%x).Add(%x) = %x, want %x", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat16_Add(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Add(f))
	}
}

func TestFloat16_Sub(t *testing.T) {
	tests := []struct {
		a, b, want Float16
	}{
		{0x3c00, 0x3800, 0x3800}, // 1.0 -  0.5 = 0.5

		// handling zeros
		{0x0000, 0x3c00, 0xbc00}, //  0.0 -  1.0 =  1.0
		{0x3c00, 0x0000, 0x3c00}, //  1.0 -  0.0 =  1.0
		{0x0000, 0x0000, 0x0000}, //  0.0 -  0.0 =  0.0
		{0x0000, 0x8000, 0x0000}, //  0.0 - -0.0 =  0.0
		{0x8000, 0x0000, 0x8000}, // -0.0 -  0.0 = -0.0
		{0x8000, 0x8000, 0x0000}, // -0.0 - -0.0 =  0.0

		// handling NaN
		{uvnan16, 0x3c00, uvnan16}, // NaN - 1 = NaN
		{0x3c00, uvnan16, uvnan16}, // 1 - NaN = NaN

		// handling infinity
		{0x7c00, 0x3c00, 0x7c00}, //  Inf -  1.0 = Inf
		{0x3c00, 0x7c00, 0xfc00}, //  1.0 -  Inf = -Inf
		{0x7c00, 0x7c00, 0x7e00}, //  Inf -  inf = NaN
	}

	for _, tt := range tests {
		got := tt.a.Sub(tt.b)
		if !eq16(got, tt.want) {
			t.Errorf("Float16(%x).Sub(%x) = %x, want %x", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat16_Sub(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Sub(f))
	}
}

func TestFloat16_Eq(t *testing.T) {
	tests := []struct {
		a, b Float16
		want bool
	}{
		{0x3c00, 0x3c00, true},    // 1.0 == 1.0
		{0x3c00, 0x0000, false},   // 1.0 != 0.0
		{0x0000, 0x0000, true},    // 0.0 == 0.0
		{0x8000, 0x8000, true},    // -0.0 == -0.0
		{0x0000, 0x8000, true},    // 0.0 == -0.0
		{uvnan16, uvnan16, false}, // NaN != NaN
	}
	for _, test := range tests {
		got := test.a.Eq(test.b)
		if got != test.want {
			t.Errorf("Float16(%x).Eq(%x) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat16_Eq(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Eq(f))
	}
}

func TestFloat16_Ne(t *testing.T) {
	tests := []struct {
		a, b Float16
		want bool
	}{
		{0x3c00, 0x3c00, false},  // 1.0 == 1.0
		{0x3c00, 0x0000, true},   // 1.0 != 0.0
		{0x0000, 0x0000, false},  // 0.0 == 0.0
		{0x8000, 0x8000, false},  // -0.0 == -0.0
		{0x0000, 0x8000, false},  // 0.0 == -0.0
		{uvnan16, uvnan16, true}, // NaN != NaN
	}
	for _, test := range tests {
		got := test.a.Ne(test.b)
		if got != test.want {
			t.Errorf("Float16(%x).Ne(%x) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat16_Ne(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Ne(f))
	}
}

func TestFloat16_Lt(t *testing.T) {
	tests := []struct {
		a, b Float16
		want bool
	}{
		{0x3c00, 0x3c00, false},   // 1.0 < 1.0
		{0x3c00, 0x0000, false},   // 1.0 < 0.0
		{0x0000, 0x3c00, true},    // 0.0 < 1.0
		{0x0000, 0x0000, false},   // 0.0 < 0.0
		{0x8000, 0x8000, false},   // -0.0 < -0.0
		{0x0000, 0x8000, false},   // 0.0 < -0.0
		{0x8000, 0x0000, false},   // -0.0 < 0.0
		{uvnan16, uvnan16, false}, // NaN < NaN
	}
	for _, test := range tests {
		got := test.a.Lt(test.b)
		if got != test.want {
			t.Errorf("Float16(%x).Lt(%x) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat16_Lt(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Lt(f))
	}
}

func TestFloat16_Gt(t *testing.T) {
	tests := []struct {
		a, b Float16
		want bool
	}{
		{0x3c00, 0x3c00, false},   // 1.0 > 1.0
		{0x3c00, 0x0000, true},    // 1.0 > 0.0
		{0x0000, 0x3c00, false},   // 0.0 > 1.0
		{0x0000, 0x0000, false},   // 0.0 > 0.0
		{0x8000, 0x8000, false},   // -0.0 > -0.0
		{0x0000, 0x8000, false},   // 0.0 > -0.0
		{0x8000, 0x0000, false},   // -0.0 > 0.0
		{uvnan16, uvnan16, false}, // NaN > NaN
	}
	for _, test := range tests {
		got := test.a.Gt(test.b)
		if got != test.want {
			t.Errorf("Float16(%x).Gt(%x) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat16_Gt(b *testing.B) {
	f := Float16(0x3c00) // 1.0
	for b.Loop() {
		runtime.KeepAlive(f.Gt(f))
	}
}

func TestFloat16_Le(t *testing.T) {
	tests := []struct {
		a, b Float16
		want bool
	}{
		{0x3c00, 0x3c00, true},    // 1.0 <= 1.0
		{0x3c00, 0x0000, false},   // 1.0 <= 0.0
		{0x0000, 0x3c00, true},    // 0.0 <= 1.0
		{0x0000, 0x0000, true},    // 0.0 <= 0.0
		{0x8000, 0x8000, true},    // -0.0 <= -0.0
		{0x0000, 0x8000, true},    // 0.0 <= -0.0
		{uvnan16, uvnan16, false}, // NaN <= NaN
	}
	for _, test := range tests {
		got := test.a.Le(test.b)
		if got != test.want {
			t.Errorf("Float16(%x).Le(%x) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}
