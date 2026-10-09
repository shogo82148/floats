package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat16_Yn(t *testing.T) {
	t.Parallel()
	type tc struct {
		n int
		x float64
	}
	tests := []tc{
		{2, 1}, {2, 5}, {3, 10}, {5, 20}, {-2, 5}, {2, 0}, {-3, 0},
	}

	for _, tt := range tests {
		want := math.Yn(tt.n, tt.x)
		got := exact16(tt.x).Yn(tt.n)
		if !close16(got, want) {
			t.Errorf("Yn(%d, %v) = %v; want %v", tt.n, tt.x, got, want)
		}
	}
}

func TestFloat16_YnSpecial(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		n    int
		x    Float16
		want Float16
	}{
		{2, exact16(0), exact16(math.Inf(-1))},
		{3, exact16(0), exact16(math.Inf(-1))},
		{-2, exact16(0), exact16(math.Inf(-1))},
		{-3, exact16(0), exact16(math.Inf(1))},
		{2, exact16(math.Copysign(0, -1)), exact16(math.Inf(-1))},
		{2, exact16(-1), exact16(math.NaN())},
		{-3, exact16(-5), exact16(math.NaN())},
		{2, exact16(math.Inf(1)), exact16(0)},
		{-3, exact16(math.Inf(1)), exact16(0)},
		{2, exact16(math.Inf(-1)), exact16(math.NaN())},
		{2, exact16(math.NaN()), exact16(math.NaN())},
		{-1, exact16(0), exact16(math.Inf(1))},
		{50, exact16(10), exact16(math.Inf(-1))}, // overflow
		{-51, exact16(10), exact16(math.Inf(1))}, // overflow
	} {
		if got := tt.x.Yn(tt.n); !eq16(got, tt.want) {
			t.Errorf("Yn(%d, %v) = %v; want %v", tt.n, tt.x, got, tt.want)
		}
	}
}

// TestFloat16_YnLowOrders checks that Yn(0, x), Yn(1, x) and Yn(-1, x) are Y0, Y1 and -Y1.
func TestFloat16_YnLowOrders(t *testing.T) {
	t.Parallel()
	for _, x := range []Float16{exact16(0.5), exact16(5), exact16(50), exact16(1000)} {
		if got, want := x.Yn(0), x.Y0(); !eq16(got, want) {
			t.Errorf("Yn(0, %v) = %v; want %v", x, got, want)
		}
		if got, want := x.Yn(1), x.Y1(); !eq16(got, want) {
			t.Errorf("Yn(1, %v) = %v; want %v", x, got, want)
		}
		if got, want := x.Yn(-1), x.Y1().Neg(); !eq16(got, want) {
			t.Errorf("Yn(-1, %v) = %v; want %v", x, got, want)
		}
	}
}

// TestFloat16_YnAll checks Yn for all the Float16 values in 2 <= x < 64 and some orders with Float256, which is
// correctly rounded. The other orders were checked in the same way.
func TestFloat16_YnAll(t *testing.T) {
	t.Parallel()
	for _, n := range []int{2, 3, 4, 7, 13, 30, 64, 99, 100} {
		for _, nn := range []int{n, -n} {
			for i := 0x4000; i < 0x5400; i++ {
				x := Float16(i)
				if got, want := x.Yn(nn), NewFloat256(x.Float64().BuiltIn()).Yn(nn).Float16(); !eq16(got, want) {
					t.Errorf("Yn(%d, %#04x) = %v; want %v", nn, i, got, want)
				}
			}
		}
	}
}

// TestFloat16_YnOutOfRange checks Yn on the both sides of the boundaries of the range of the recurrence:
// 2 <= x < 64 and |n| <= 100.
func TestFloat16_YnOutOfRange(t *testing.T) {
	t.Parallel()
	for _, n := range []int{2, 100, 101, 102, -100, -101} {
		for _, x := range []float64{0.001, 0.5, 1.99, 2, 63.9, 64, 100, 1000, 60000} {
			a := NewFloat16(x)
			if got, want := a.Yn(n), NewFloat256(a.Float64().BuiltIn()).Yn(n).Float16(); !eq16(got, want) {
				t.Errorf("Yn(%d, %v) = %v; want %v", n, a, got, want)
			}
		}
	}
}

func BenchmarkFloat16_Yn(b *testing.B) {
	for _, tt := range []struct {
		name string
		n    int
		x    Float16
	}{
		{"small-n", 2, exact16(10.5)},
		{"medium-n", 10, exact16(50.5)},
		{"large-n", 50, exact16(20.5)},
		{"negative", -5, exact16(20.5)},
		{"overflow", 50, exact16(10)},
		{"small-x", 3, exact16(1.5)},    // math.Yn
		{"large-x", 3, exact16(1000.5)}, // math.Yn
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Yn(tt.n))
			}
		})
	}
}
