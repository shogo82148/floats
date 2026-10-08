package floats

import (
	"math"
	"math/big"
	"math/rand"
	"testing"
)

func TestHypot64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float64
		y    Float64
		want float64
	}{
		{exact64(3), exact64(4), 5},
		{exact64(5), exact64(12), 13},
		{exact64(1), exact64(1), math.Sqrt(2)},
	}

	for _, tt := range tests {
		got := Hypot64(tt.x, tt.y)
		if !close64(got, tt.want) {
			t.Errorf("Hypot64(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		y    Float64
		want Float64
	}{
		// special cases
		{exact64(0), exact64(0), exact64(0)},
		{exact64(math.Inf(1)), exact64(1), exact64(math.Inf(1))},
		{exact64(math.Inf(-1)), exact64(1), exact64(math.Inf(1))},
		{exact64(1), exact64(math.Inf(1)), exact64(math.Inf(1))},
		{exact64(1), exact64(math.Inf(-1)), exact64(math.Inf(1))},
		{exact64(math.NaN()), exact64(1), exact64(math.NaN())},
		{exact64(1), exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := Hypot64(tt.x, tt.y)
		if !eq64(got, tt.want) {
			t.Errorf("Hypot64(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestHypot64_Edge(t *testing.T) {
	t.Parallel()
	tiny := Float64(math.Float64frombits(1))
	max := Float64(math.MaxFloat64)
	if got := Hypot64(tiny, tiny); !eq64(got, tiny) {
		t.Errorf("Hypot64(tiny, tiny) = %v; want %v", got, tiny)
	}
	if got := Hypot64(max, max); !got.IsInf(1) {
		t.Errorf("Hypot64(max, max) = %v; want +Inf", got)
	}
	if got := Hypot64(max, exact64(1)); !eq64(got, max) {
		t.Errorf("Hypot64(max, 1) = %v; want max", got)
	}
	if got := Hypot64(exact64(-3), exact64(0)); !eq64(got, exact64(3)) {
		t.Errorf("Hypot64(-3, 0) = %v; want 3", got)
	}
}

func TestHypot64_Reference(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(1))
	const prec = 400
	setExp := func(v uint64, e uint64) uint64 {
		return v&^(0x7ff<<52) | e<<52
	}
	for range 200000 {
		a, b := r.Uint64(), r.Uint64()
		if r.Intn(2) == 0 {
			// make q's exponent near p's
			b = setExp(b, (a>>52)&0x7ff)
		}
		// exercise subnormal and near-overflow exponents
		for _, v := range []*uint64{&a, &b} {
			switch r.Intn(8) {
			case 0:
				*v = setExp(*v, uint64(r.Intn(3)))
			case 1:
				*v = setExp(*v, uint64(0x7fe-r.Intn(3)))
			}
		}
		p, q := Float64(math.Float64frombits(a)), Float64(math.Float64frombits(b))
		if p.IsNaN() || q.IsNaN() || p.IsInf(0) || q.IsInf(0) {
			continue
		}
		x := new(big.Float).SetPrec(prec).SetFloat64(p.BuiltIn())
		y := new(big.Float).SetPrec(prec).SetFloat64(q.BuiltIn())
		sum := new(big.Float).SetPrec(prec).Add(
			new(big.Float).SetPrec(prec).Mul(x, x),
			new(big.Float).SetPrec(prec).Mul(y, y),
		)
		f, _ := new(big.Float).SetPrec(prec).Sqrt(sum).Float64()
		want := Float64(f)
		if got := Hypot64(p, q); !eq64(got, want) {
			t.Errorf("Hypot64(%#016x, %#016x) = %v; want %v", a, b, got, want)
		}
	}
}

func BenchmarkHypot64(b *testing.B) {
	p := exact64(3).Quo(exact64(7))
	q := exact64(5).Quo(exact64(11))
	for b.Loop() {
		Hypot64(p, q)
	}
}
