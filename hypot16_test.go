package floats

import (
	"math"
	"math/big"
	"testing"
)

func TestHypot16(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		y    Float16
		want float64
	}{
		{exact16(3), exact16(4), 5},
		{exact16(5), exact16(12), 13},
		{exact16(1), exact16(1), math.Sqrt(2)},
	}

	for _, tt := range tests {
		got := Hypot16(tt.x, tt.y)
		if !close16(got, tt.want) {
			t.Errorf("Hypot16(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		y    Float16
		want Float16
	}{
		// special cases
		{exact16(0), exact16(0), exact16(0)},
		{exact16(math.Inf(1)), exact16(1), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(1), exact16(math.Inf(1))},
		{exact16(1), exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(1), exact16(math.Inf(-1)), exact16(math.Inf(1))},
		{exact16(math.NaN()), exact16(1), exact16(math.NaN())},
		{exact16(1), exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := Hypot16(tt.x, tt.y)
		if !eq16(got, tt.want) {
			t.Errorf("Hypot16(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestHypot16_Reference(t *testing.T) {
	t.Parallel()
	check := func(a, b uint16) {
		p, q := NewFloat16FromBits(a), NewFloat16FromBits(b)
		if p.IsNaN() || q.IsNaN() || p.IsInf(0) || q.IsInf(0) {
			return
		}
		x := new(big.Float).SetPrec(200).SetFloat64(p.Float64().BuiltIn())
		y := new(big.Float).SetPrec(200).SetFloat64(q.Float64().BuiltIn())
		sum := new(big.Float).SetPrec(200).Add(new(big.Float).SetPrec(200).Mul(x, x), new(big.Float).SetPrec(200).Mul(y, y))
		root := new(big.Float).SetPrec(200).Sqrt(sum)
		want := NewFloat16(func() float64 { f, _ := root.Float64(); return f }())
		// big.Float.Float64 rounds to nearest even in float64 first; acceptable reference
		if got := Hypot16(p, q); !eq16(got, want) {
			t.Errorf("Hypot16(%#04x, %#04x) = %v; want %v", a, b, got, want)
		}
	}
	for a := 0; a < 1<<16; a += 7 {
		for b := 0; b < 1<<16; b += 251 {
			check(uint16(a), uint16(b))
		}
	}
}

func BenchmarkHypot16(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		Hypot16(NewFloat16FromBits(uint16(i*7919)&0x7bff), NewFloat16FromBits(uint16(i*104729)&0x7bff))
	}
}
