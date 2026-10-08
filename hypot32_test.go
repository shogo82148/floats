package floats

import (
	"math"
	"math/big"
	"math/rand"
	"testing"
)

func TestHypot32(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float32
		y    Float32
		want float64
	}{
		{exact32(3), exact32(4), 5},
		{exact32(5), exact32(12), 13},
		{exact32(1), exact32(1), math.Sqrt(2)},
	}

	for _, tt := range tests {
		got := Hypot32(tt.x, tt.y)
		if !close32(got, tt.want) {
			t.Errorf("Hypot32(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		y    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(0), exact32(0)},
		{exact32(math.Inf(1)), exact32(1), exact32(math.Inf(1))},
		{exact32(math.Inf(-1)), exact32(1), exact32(math.Inf(1))},
		{exact32(1), exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(1), exact32(math.Inf(-1)), exact32(math.Inf(1))},
		{exact32(math.NaN()), exact32(1), exact32(math.NaN())},
		{exact32(1), exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := Hypot32(tt.x, tt.y)
		if !eq32(got, tt.want) {
			t.Errorf("Hypot32(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestHypot32_Reference(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(1))
	for range 200000 {
		p, q := NewFloat32FromBits(r.Uint32()), NewFloat32FromBits(r.Uint32())
		if p.IsNaN() || q.IsNaN() || p.IsInf(0) || q.IsInf(0) {
			continue
		}
		x := new(big.Float).SetPrec(300).SetFloat64(p.Float64().BuiltIn())
		y := new(big.Float).SetPrec(300).SetFloat64(q.Float64().BuiltIn())
		sum := new(big.Float).SetPrec(300).Add(
			new(big.Float).SetPrec(300).Mul(x, x),
			new(big.Float).SetPrec(300).Mul(y, y),
		)
		f, _ := new(big.Float).SetPrec(300).Sqrt(sum).Float32()
		want := NewFloat32(float64(f))
		if got := Hypot32(p, q); !eq32(got, want) {
			t.Errorf("Hypot32(%#08x, %#08x) = %v; want %v", p.Bits(), q.Bits(), got, want)
		}
	}
}

func BenchmarkHypot32(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		Hypot32(NewFloat32FromBits(uint32(i)*2654435761&0x7f7fffff), NewFloat32FromBits(uint32(i)*40503&0x7f7fffff))
	}
}
