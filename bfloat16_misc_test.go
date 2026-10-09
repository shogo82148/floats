package floats

import (
	"math"
	"math/rand/v2"
	"testing"
)

func bf16Value(i int) (BFloat16, float64) {
	a := BFloat16(i)
	return a, float64(math.Float32frombits(uint32(a) << 16))
}

func TestBFloat16_Compare(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	check := func(i, j int) {
		a, x := bf16Value(i)
		b, y := bf16Value(j)
		if got, want := a.Eq(b), x == y; got != want {
			t.Fatalf("BFloat16(%#04x).Eq(%#04x) = %v, want %v", i, j, got, want)
		}
		if got, want := a.Ne(b), x != y; got != want {
			t.Fatalf("BFloat16(%#04x).Ne(%#04x) = %v, want %v", i, j, got, want)
		}
		if got, want := a.Lt(b), x < y; got != want {
			t.Fatalf("BFloat16(%#04x).Lt(%#04x) = %v, want %v", i, j, got, want)
		}
		if got, want := a.Gt(b), x > y; got != want {
			t.Fatalf("BFloat16(%#04x).Gt(%#04x) = %v, want %v", i, j, got, want)
		}
		if got, want := a.Le(b), x <= y; got != want {
			t.Fatalf("BFloat16(%#04x).Le(%#04x) = %v, want %v", i, j, got, want)
		}
		if got, want := a.Ge(b), x >= y; got != want {
			t.Fatalf("BFloat16(%#04x).Ge(%#04x) = %v, want %v", i, j, got, want)
		}
	}
	special := []int{0, 0x8000, 0x7f80, 0xff80, 0x7fc0, 0xffc0, 0x7f81, 0x3f80, 0xbf80, 0x0001, 0x8001}
	for i := range 1 << 16 {
		for _, j := range special {
			check(i, j)
			check(j, i)
		}
		for range 100 {
			check(i, r.IntN(1<<16))
		}
		check(i, i)
	}
}

func TestBFloat16_Round(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a, x := bf16Value(i)
		for _, tt := range []struct {
			name string
			got  BFloat16
			want float64
		}{
			{"Floor", a.Floor(), math.Floor(x)},
			{"Ceil", a.Ceil(), math.Ceil(x)},
			{"Trunc", a.Trunc(), math.Trunc(x)},
			{"Round", a.Round(), math.Round(x)},
			{"RoundToEven", a.RoundToEven(), math.RoundToEven(x)},
		} {
			want := NewBFloat16(tt.want)
			if !eqbf16(tt.got, want) {
				t.Errorf("BFloat16(%#04x).%s() = %#04x, want %#04x", i, tt.name, uint16(tt.got), uint16(want))
			}
		}
	}
}

func TestBFloat16_Modf(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a, x := bf16Value(i)
		gotInt, gotFrac := a.Modf()
		wi, wf := math.Modf(x)
		if !eqbf16(gotInt, NewBFloat16(wi)) || !eqbf16(gotFrac, NewBFloat16(wf)) {
			t.Errorf("BFloat16(%#04x).Modf() = (%#04x, %#04x), want (%#04x, %#04x)", i, uint16(gotInt), uint16(gotFrac), uint16(NewBFloat16(wi)), uint16(NewBFloat16(wf)))
		}
	}
}

func TestBFloat16_FrexpLdexp(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a, x := bf16Value(i)
		gotFrac, gotExp := a.Frexp()
		wf, we := math.Frexp(x)
		if !eqbf16(gotFrac, NewBFloat16(wf)) || gotExp != we {
			t.Errorf("BFloat16(%#04x).Frexp() = (%#04x, %d), want (%#04x, %d)", i, uint16(gotFrac), gotExp, uint16(NewBFloat16(wf)), we)
		}
		for e := -300; e <= 300; e += 7 {
			got := a.Ldexp(e)
			want := NewBFloat16(math.Ldexp(x, e))
			if !eqbf16(got, want) {
				t.Errorf("BFloat16(%#04x).Ldexp(%d) = %#04x, want %#04x", i, e, uint16(got), uint16(want))
			}
		}
		for _, e := range []int{math.MaxInt32, math.MinInt32, math.MaxInt32 - 1, math.MinInt32 + 1} {
			if got, want := a.Ldexp(e), NewBFloat16(math.Ldexp(x, e)); !eqbf16(got, want) {
				t.Errorf("BFloat16(%#04x).Ldexp(%d) = %#04x, want %#04x", i, e, uint16(got), uint16(want))
			}
		}
	}
}

func TestBFloat16_ModRemainder(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(3, 4))
	check := func(i, j int) {
		a, x := bf16Value(i)
		b, y := bf16Value(j)
		if got, want := a.Mod(b), NewBFloat16(math.Mod(x, y)); !eqbf16(got, want) {
			t.Fatalf("BFloat16(%#04x).Mod(%#04x) = %#04x, want %#04x", i, j, uint16(got), uint16(want))
		}
		if got, want := a.Remainder(b), NewBFloat16(math.Remainder(x, y)); !eqbf16(got, want) {
			t.Fatalf("BFloat16(%#04x).Remainder(%#04x) = %#04x, want %#04x", i, j, uint16(got), uint16(want))
		}
		if got, want := a.Dim(b), NewBFloat16(math.Dim(x, y)); !eqbf16(got, want) {
			t.Fatalf("BFloat16(%#04x).Dim(%#04x) = %#04x, want %#04x", i, j, uint16(got), uint16(want))
		}
		if got, want := a.Max(b), NewBFloat16(math.Max(x, y)); !eqbf16(got, want) {
			t.Fatalf("BFloat16(%#04x).Max(%#04x) = %#04x, want %#04x", i, j, uint16(got), uint16(want))
		}
		if got, want := a.Min(b), NewBFloat16(math.Min(x, y)); !eqbf16(got, want) {
			t.Fatalf("BFloat16(%#04x).Min(%#04x) = %#04x, want %#04x", i, j, uint16(got), uint16(want))
		}
	}
	special := []int{0, 0x8000, 0x7f80, 0xff80, 0x7fc0, 0x3f80, 0xbf80, 0x0001, 0x8001, 0x7f7f}
	for i := range 1 << 16 {
		for _, j := range special {
			check(i, j)
			check(j, i)
		}
		for range 20 {
			check(i, r.IntN(1<<16))
		}
	}
}

func TestBFloat16_Logb(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a, x := bf16Value(i)
		if got, want := a.Logb(), NewBFloat16(math.Logb(x)); !eqbf16(got, want) {
			t.Errorf("BFloat16(%#04x).Logb() = %#04x, want %#04x", i, uint16(got), uint16(want))
		}
		if got, want := a.Ilogb(), math.Ilogb(x); got != want {
			t.Errorf("BFloat16(%#04x).Ilogb() = %d, want %d", i, got, want)
		}
	}
}

func TestBFloat16_Nextafter(t *testing.T) {
	t.Parallel()
	// the neighbors are found from the sorted list of all finite values.
	r := rand.New(rand.NewPCG(5, 6))
	check := func(i, j int) {
		a, x := bf16Value(i)
		b, y := bf16Value(j)
		got := a.Nextafter(b)
		var want BFloat16
		switch {
		case math.IsNaN(x) || math.IsNaN(y):
			want = uvnanbf16
		case x == y:
			want = a
		default:
			// the nearest value to x in the direction of y
			best := math.NaN()
			var bestBits BFloat16
			for k := range 1 << 16 {
				c, z := bf16Value(k)
				if math.IsNaN(z) || k == 0x8000 { // -0 is the same as 0
					continue
				}
				if (y > x && z > x) || (y < x && z < x) {
					if math.IsNaN(best) || math.Abs(z-x) < math.Abs(best-x) {
						best, bestBits = z, c
					}
				}
			}
			want = bestBits
			if best == 0 {
				// the sign of zero follows the sign of the original value
				want = a & 0x8000
			}
			if math.IsInf(x, 0) {
				// the largest finite value
				want = 0x7f7f | a&0x8000
			}
			if x == 0 {
				// the smallest subnormal number in the direction of y
				want = BFloat16(1)
				if y < 0 {
					want = 0x8001
				}
			}
		}
		if !eqbf16(got, want) {
			t.Fatalf("BFloat16(%#04x).Nextafter(%#04x) = %#04x, want %#04x", i, j, uint16(got), uint16(want))
		}
	}
	for range 3000 {
		i := r.IntN(1 << 16)
		check(i, r.IntN(1<<16))
		check(i, 0x7f80)
		check(i, 0xff80)
		check(i, 0x7fc0)
		check(i, i)
	}
	for _, i := range []int{0, 0x8000, 1, 0x8001, 0x007f, 0x0080, 0x7f7f, 0xff7f, 0x7f80, 0xff80, 0x3f80, 0xbf80} {
		for _, j := range []int{0, 0x8000, 0x7f80, 0xff80, 0x3f80, 0xbf80, 1, 0x8001} {
			check(i, j)
		}
	}
}
