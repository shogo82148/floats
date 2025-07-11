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

func TestFloat64_Signbit(t *testing.T) {
	tests := []struct {
		in   Float64
		want bool
	}{
		{0, false},
		{1, false},
		{-1, true},
		{Float64(math.Copysign(0, -1)), true}, // negative zero
	}

	for _, test := range tests {
		got := test.in.Signbit()
		if got != test.want {
			t.Errorf("Float64.Signbit(%v) = %v, want %v", test.in, got, test.want)
		}
	}
}

func BenchmarkFloat64_Signbit(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Signbit())
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

func TestFloat64_Neg(t *testing.T) {
	negZero := Float64(math.Copysign(0, -1))
	nan := Float64(math.NaN())
	inf := Float64(math.Inf(1))

	tests := []struct {
		a    Float64
		want Float64
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
		if !eq64(got, tt.want) {
			t.Errorf("Float64(%x).Neg() = %x, want %x", tt.a, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Neg(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Neg())
	}
}

func TestFloat64_Mul(t *testing.T) {
	nan := Float64(math.NaN())
	negZero := Float64(math.Copysign(0, -1))
	inf := Float64(math.Inf(1))

	tests := []struct {
		a, b Float64
		want Float64
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

func TestFloat64_Quo(t *testing.T) {
	nan := Float64(math.NaN())
	inf := Float64(math.Inf(1))

	tests := []struct {
		a, b Float64
		want Float64
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
		if !eq64(got, test.want) {
			t.Errorf("Float64(%x).Quo(%x) = %x, want %x", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat64_Quo(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Quo(f))
	}
}

func TestFloat64_Add(t *testing.T) {
	nan := Float64(math.NaN())
	inf := Float64(math.Inf(1))

	tests := []struct {
		a, b, want Float64
	}{
		{1, 1, 2},
		{2, 3, 5},

		// adding zero
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
		if !eq64(got, test.want) {
			t.Errorf("Float64(%x).Add(%x) = %x, want %x", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat64_Add(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Add(f))
	}
}

func TestFloat64_Sub(t *testing.T) {
	nan := Float64(math.NaN())
	inf := Float64(math.Inf(1))

	tests := []struct {
		a, b, want Float64
	}{
		{1, 1, 0},
		{2, 3, -1},

		// subtracting zero
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
		if !eq64(got, test.want) {
			t.Errorf("Float64(%x).Sub(%x) = %x, want %x", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat64_Sub(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Sub(f))
	}
}

func TestFloat64_Sqrt(t *testing.T) {
	nan := Float64(math.NaN())
	negZero := Float64(math.Copysign(0, -1))
	inf := Float64(math.Inf(1))

	tests := []struct {
		a    Float64
		want Float64
	}{
		{0, 0},
		{1, 1},
		{4, 2},
		{9, 3},

		// handling NaN
		{nan, nan},

		// handling negative numbers
		{-1, nan},
		{-4, nan},

		// handling zero
		{negZero, negZero},

		// handling infinity
		{inf, inf},
	}

	for _, test := range tests {
		got := test.a.Sqrt()
		if !eq64(got, test.want) {
			t.Errorf("Float64(%x).Sqrt() = %x, want %x", test.a, got, test.want)
		}
	}
}

func BenchmarkFloat64_Sqrt(b *testing.B) {
	f := Float64(2.0)
	for b.Loop() {
		runtime.KeepAlive(f.Sqrt())
	}
}

func TestFloat64_Eq(t *testing.T) {
	nan := Float64(math.NaN())
	negZero := Float64(math.Copysign(0, -1))

	tests := []struct {
		a, b Float64
		want bool
	}{
		{1, 1, true},
		{2, 3, false},
		{0, 1, false},
		{negZero, 1, false},
		{0, negZero, true},
		{negZero, 0, true},
		{nan, 1, false},
		{1, nan, false},
		{nan, nan, false},
	}

	for _, test := range tests {
		got := test.a.Eq(test.b)
		if got != test.want {
			t.Errorf("Float64(%x).Eq(%x) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat64_Eq(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Eq(f))
	}
}

func TestFloat64_Ne(t *testing.T) {
	nan := Float64(math.NaN())
	negZero := Float64(math.Copysign(0, -1))

	tests := []struct {
		a, b Float64
		want bool
	}{
		{1, 1, false},
		{2, 3, true},
		{0, 1, true},
		{negZero, 1, true},
		{0, negZero, false},
		{negZero, 0, false},
		{nan, 1, true},
		{1, nan, true},
		{nan, nan, true},
	}

	for _, test := range tests {
		got := test.a.Ne(test.b)
		if got != test.want {
			t.Errorf("Float64(%x).Ne(%x) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat64_Ne(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Ne(f))
	}
}

func TestFloat64_Lt(t *testing.T) {
	nan := Float64(math.NaN())
	negZero := Float64(math.Copysign(0, -1))
	inf := Float64(math.Inf(1))

	tests := []struct {
		a, b Float64
		want bool
	}{
		{1, 1, false},
		{2, 3, true},
		{0, 1, true},
		{negZero, 1, true},
		{0, negZero, false},
		{negZero, 0, false},
		{nan, 1, false},
		{1, nan, false},
		{nan, nan, false},
		{inf, 1, false},
		{1, inf, true},
		{inf, inf, false},
	}

	for _, test := range tests {
		got := test.a.Lt(test.b)
		if got != test.want {
			t.Errorf("Float64(%x).Lt(%x) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat64_Lt(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Lt(f))
	}
}

func TestFloat64_Gt(t *testing.T) {
	nan := Float64(math.NaN())
	negZero := Float64(math.Copysign(0, -1))
	inf := Float64(math.Inf(1))

	tests := []struct {
		a, b Float64
		want bool
	}{
		{1, 1, false},
		{2, 3, false},
		{0, 1, false},
		{negZero, 1, false},
		{0, negZero, false},
		{negZero, 0, false},
		{nan, 1, false},
		{1, nan, false},
		{nan, nan, false},
		{inf, 1, true},
	}

	for _, test := range tests {
		got := test.a.Gt(test.b)
		if got != test.want {
			t.Errorf("Float64(%x).Gt(%x) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat64_Gt(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Gt(f))
	}
}

func TestFloat64_Le(t *testing.T) {
	nan := Float64(math.NaN())
	negZero := Float64(math.Copysign(0, -1))
	inf := Float64(math.Inf(1))

	tests := []struct {
		a, b Float64
		want bool
	}{
		{1, 1, true},
		{2, 3, true},
		{0, 1, true},
		{negZero, 1, true},
		{0, negZero, true},
		{negZero, 0, true},
		{nan, 1, false},
		{1, nan, false},
		{nan, nan, false},
		{inf, 1, false},
	}

	for _, test := range tests {
		got := test.a.Le(test.b)
		if got != test.want {
			t.Errorf("Float64(%x).Le(%x) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat64_Le(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Le(f))
	}
}

func TestFloat64_Ge(t *testing.T) {
	nan := Float64(math.NaN())
	negZero := Float64(math.Copysign(0, -1))
	inf := Float64(math.Inf(1))

	tests := []struct {
		a, b Float64
		want bool
	}{
		{1, 1, true},
		{2, 3, false},
		{0, 1, false},
		{negZero, 1, false},
		{0, negZero, true},
		{negZero, 0, true},
		{nan, 1, false},
		{1, nan, false},
		{nan, nan, false},
		{inf, 1, true},
	}

	for _, test := range tests {
		got := test.a.Ge(test.b)
		if got != test.want {
			t.Errorf("Float64(%x).Ge(%x) = %v, want %v", test.a, test.b, got, test.want)
		}
	}
}

func BenchmarkFloat64_Ge(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(f.Ge(f))
	}
}

func TestFMA64(t *testing.T) {
	nan := Float64(math.NaN())
	negZero := Float64(math.Copysign(0, -1))
	//inf := Float64(math.Inf(1))

	tests := []struct {
		x, y, z, want Float64
	}{
		{1, 2, 3, 5},
		{1, 2, -3, -1},
		{1, -2, 3, 1},
		{1, -2, -3, -5},
		{0, 0, 0, 0},
		{negZero, negZero, negZero, 0},
		{nan, 1, 2, nan},
		{1, nan, 2, nan},
		{1, 2, nan, nan},
	}

	for _, test := range tests {
		got := FMA64(test.x, test.y, test.z)
		if !eq64(got, test.want) {
			t.Errorf("FMA64(%x,%x,%x) = %x want %x", test.x, test.y, test.z, got, test.want)
		}
	}
}

func BenchmarkFMA64(b *testing.B) {
	f := Float64(1.0)
	for b.Loop() {
		runtime.KeepAlive(FMA64(f, f, f))
	}
}
