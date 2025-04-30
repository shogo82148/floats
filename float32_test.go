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

func TestFloat32_Neg(t *testing.T) {
	negZero := Float32(math.Copysign(0, -1))
	nan := Float32(math.NaN())
	inf := Float32(math.Inf(1))
	tests := []struct {
		a, want Float32
	}{
		{1, -1},
		{0, negZero},
		{negZero, 0},
		{inf, -inf},
		{-inf, inf},
		{nan, nan},
	}

	for _, tt := range tests {
		got := tt.a.Neg()
		if !eq32(got, tt.want) {
			t.Errorf("Float32.Neg() = %x, want %x", got, tt.want)
		}
	}
}

func BenchmarkFloat32_Neg(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Neg())
	}
}

func TestFloat32_Mul(t *testing.T) {
	nan := Float32(math.NaN())
	negZero := Float32(math.Copysign(0, -1))
	inf := Float32(math.Inf(1))

	tests := []struct {
		a, b, want Float32
	}{
		{0, 0, 0},
		{1, 1, 1},
		{2, 3, 6},

		// handling zero
		{0, 1, 0},
		{negZero, 1, negZero},
		{0, -1, negZero},
		{negZero, -1, 0},

		// handling NaN
		{nan, 0, nan},
		{0, nan, nan},

		// handling infinity
		{1, inf, inf},
		{-1, inf, -inf},
		{inf, 1, inf},
		{inf, -1, -inf},
		{inf, 0, nan},
		{0, inf, nan},
	}

	for _, test := range tests {
		got := test.a.Mul(test.b)
		if !eq32(got, test.want) {
			t.Errorf("Float32(%x).Mul(%x) = %x, want %x", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat32_Mul(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Mul(f))
	}
}

func TestFloat32_Quo(t *testing.T) {
	nan := Float32(math.NaN())
	inf := Float32(math.Inf(1))

	tests := []struct {
		a, b, want Float32
	}{
		{0, 1, 0},
		{1, 1, 1},

		// zero division
		{0, 0, nan},
		{1, 0, inf},

		// handling NaN
		{nan, 1, nan},
		{1, nan, nan},

		// handling infinity
		{1, inf, 0},
		{inf, 1, inf},
		{inf, inf, nan},
	}

	for _, test := range tests {
		got := test.a.Quo(test.b)
		if !eq32(got, test.want) {
			t.Errorf("Float32(%x).Quo(%x) = %x, want %x", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat32_Quo(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Quo(f))
	}
}

func TestFloat32_Add(t *testing.T) {
	nan := Float32(math.NaN())
	inf := Float32(math.Inf(1))
	tests := []struct {
		a, b, want Float32
	}{
		{1, 1, 2},
		{2, 3, 5},

		// adding zeros
		{0, 1, 1},
		{1, 0, 1},

		// handling NaN
		{nan, 1, nan},
		{1, nan, nan},

		// handling infinity
		{inf, 1, inf},
		{1, inf, inf},
		{inf, inf, inf},
		{inf, -inf, nan},
		{-inf, inf, nan},
	}

	for _, test := range tests {
		got := test.a.Add(test.b)
		if !eq32(got, test.want) {
			t.Errorf("Float32(%x).Add(%x) = %x, want %x", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat32_Add(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Add(f))
	}
}

func TestFloat32_Sub(t *testing.T) {
	nan := Float32(math.NaN())
	inf := Float32(math.Inf(1))
	tests := []struct {
		a, b, want Float32
	}{
		{1, 1, 0},
		{2, 3, -1},

		// subtracting zeros
		{0, 1, -1},
		{1, 0, 1},

		// handling NaN
		{nan, 1, nan},
		{1, nan, nan},

		// handling infinity
		{inf, 1, inf},
		{1, inf, -inf},
	}

	for _, test := range tests {
		got := test.a.Sub(test.b)
		if !eq32(got, test.want) {
			t.Errorf("Float32(%x).Sub(%x) = %x, want %x", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat32_Sub(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Sub(f))
	}
}

func TestFloat32_Eq(t *testing.T) {
	nan := Float32(math.NaN())
	negZero := Float32(math.Copysign(0, -1))
	inf := Float32(math.Inf(1))

	tests := []struct {
		a, b Float32
		want bool
	}{
		{1, 1, true},
		{2, 3, false},
		{0, 1, false},
		{1, 0, false},
		{0, 0, true},
		{0, negZero, true}, // -0.0 == 0.0
		{negZero, 0, true}, // 0.0 == -0.0
		{nan, 1, false},    // NaN != 1
		{1, nan, false},    // 1 != NaN
		{nan, nan, false},  // NaN != NaN
		{inf, 1, false},    // Inf != 1
		{1, inf, false},    // 1 != Inf
		{inf, inf, true},   // Inf == Inf
	}

	for _, tt := range tests {
		got := tt.a.Eq(tt.b)
		if got != tt.want {
			t.Errorf("Float32(%x).Eq(%x) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Eq(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Eq(f))
	}
}

func TestFloat32_Ne(t *testing.T) {
	nan := Float32(math.NaN())
	negZero := Float32(math.Copysign(0, -1))
	inf := Float32(math.Inf(1))

	tests := []struct {
		a, b Float32
		want bool
	}{
		{1, 1, false},
		{2, 3, true},
		{0, 1, true},
		{1, 0, true},
		{0, 0, false},
		{0, negZero, false}, // -0.0 == 0.0
		{negZero, 0, false}, // 0.0 == -0.0
		{nan, 1, true},      // NaN != 1
		{1, nan, true},      // 1 != NaN
		{nan, nan, true},    // NaN != NaN
		{inf, 1, true},      // Inf != 1
		{1, inf, true},      // 1 != Inf
	}

	for _, tt := range tests {
		got := tt.a.Ne(tt.b)
		if got != tt.want {
			t.Errorf("Float32(%x).Ne(%x) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Ne(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Ne(f))
	}
}

func TestFloat32_Lt(t *testing.T) {
	nan := Float32(math.NaN())
	negZero := Float32(math.Copysign(0, -1))
	inf := Float32(math.Inf(1))

	tests := []struct {
		a, b Float32
		want bool
	}{
		{1, 1, false},
		{2, 3, true},
		{0, 1, true},
		{1, 0, false},
		{0, 0, false},
		{0, negZero, false},
		{negZero, 0, false},
		{nan, 1, false},
		{1, nan, false},
		{nan, nan, false},
		{inf, 1, false},
	}

	for _, tt := range tests {
		got := tt.a.Lt(tt.b)
		if got != tt.want {
			t.Errorf("Float32(%x).Lt(%x) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Lt(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Lt(f))
	}
}

func TestFloat32_Gt(t *testing.T) {
	nan := Float32(math.NaN())
	negZero := Float32(math.Copysign(0, -1))
	inf := Float32(math.Inf(1))

	tests := []struct {
		a, b Float32
		want bool
	}{
		{1, 1, false},
		{2, 3, false},
		{0, 1, false},
		{1, 0, true},
		{0, 0, false},
		{0, negZero, false},
		{negZero, 0, false},
		{nan, 1, false},
		{1, nan, false},
		{nan, nan, false},
		{inf, 1, true},
	}

	for _, tt := range tests {
		got := tt.a.Gt(tt.b)
		if got != tt.want {
			t.Errorf("Float32(%x).Gt(%x) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Gt(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Gt(f))
	}
}

func TestFloat32_Le(t *testing.T) {
	nan := Float32(math.NaN())
	negZero := Float32(math.Copysign(0, -1))
	inf := Float32(math.Inf(1))

	tests := []struct {
		a, b Float32
		want bool
	}{
		{1, 1, true},
		{2, 3, true},
		{0, 1, true},
		{1, 0, false},
		{0, 0, true},
		{0, negZero, true},
		{negZero, 0, true},
		{nan, 1, false},
		{1, nan, false},
		{nan, nan, false},
		{inf, 1, false},
		{1, inf, true},
	}

	for _, tt := range tests {
		got := tt.a.Le(tt.b)
		if got != tt.want {
			t.Errorf("Float32(%x).Le(%x) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Le(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Le(f))
	}
}

func TestFloat32_Ge(t *testing.T) {
	nan := Float32(math.NaN())
	negZero := Float32(math.Copysign(0, -1))
	inf := Float32(math.Inf(1))

	tests := []struct {
		a, b Float32
		want bool
	}{
		{1, 1, true},
		{2, 3, false},
		{0, 1, false},
		{1, 0, true},
		{0, 0, true},
		{0, negZero, true},
		{negZero, 0, true},
		{nan, 1, false},
		{1, nan, false},
		{nan, nan, false},
		{inf, 1, true},
	}

	for _, tt := range tests {
		got := tt.a.Ge(tt.b)
		if got != tt.want {
			t.Errorf("Float32(%x).Ge(%x) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Ge(b *testing.B) {
	f := Float32(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Ge(f))
	}
}
