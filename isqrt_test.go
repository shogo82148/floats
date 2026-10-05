package floats

import (
	"math/bits"
	"math/rand/v2"
	"testing"

	"github.com/shogo82148/ints"
)

// sqrtBitByBit128 is the reference implementation of Float128.Sqrt
// that generates the square root bit by bit.
func sqrtBitByBit128(a Float128) Float128 {
	_, exp, frac := a.normalize()
	if exp%2 != 0 {
		frac = frac.Lsh(1)
	}
	exp >>= 1
	frac = frac.Lsh(1)
	var q, s ints.Uint128
	r := ints.Uint128{1 << (shift128 + 1 - 64), 0}
	for !r.IsZero() {
		t := s.Add(r)
		if t.Cmp(frac) <= 0 {
			s = t.Add(r)
			frac = frac.Sub(t)
			q = q.Add(r)
		}
		frac = frac.Lsh(1)
		r = r.Rsh(1)
	}
	if !frac.IsZero() {
		q = q.Add(q.And(ints.Uint128{0, 1}))
	}
	q = q.Rsh(1)
	q = q.Add(ints.Uint128{uint64(exp-1+bias128) << (shift128 - 64), 0})
	return Float128(q)
}

// sqrtBitByBit256 is the reference implementation of Float256.Sqrt
// that generates the square root bit by bit.
func sqrtBitByBit256(a Float256) Float256 {
	_, exp, frac := a.normalize()
	if exp%2 != 0 {
		frac = frac.Lsh(1)
	}
	exp >>= 1
	frac = frac.Lsh(1)
	var q, s ints.Uint256
	r := ints.Uint256{1 << (shift256 + 1 - 192), 0, 0, 0}
	for !r.IsZero() {
		t := s.Add(r)
		if t.Cmp(frac) <= 0 {
			s = t.Add(r)
			frac = frac.Sub(t)
			q = q.Add(r)
		}
		frac = frac.Lsh(1)
		r = r.Rsh(1)
	}
	if !frac.IsZero() {
		q = q.Add(q.And(ints.Uint256{0, 0, 0, 1}))
	}
	q = q.Rsh(1)
	q = q.Add(ints.Uint256{uint64(exp-1+bias256) << (shift256 - 192), 0, 0, 0})
	return Float256(q)
}

func TestIsqrt128(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	check := func(hi, lo uint64) {
		s := isqrt128(hi, lo)
		x := ints.Uint256{0, 0, hi, lo}
		p := ints.Uint128{0, s}.Mul256(ints.Uint128{0, s})
		if x.Cmp(p) < 0 {
			t.Fatalf("isqrt128(%x, %x) = %x: too large", hi, lo, s)
		}
		if s != ^uint64(0) {
			p = ints.Uint128{0, s + 1}.Mul256(ints.Uint128{0, s + 1})
			if x.Cmp(p) >= 0 {
				t.Fatalf("isqrt128(%x, %x) = %x: too small", hi, lo, s)
			}
		}
	}
	check(1<<62, 0)
	check(^uint64(0), ^uint64(0))
	check(^uint64(0)-1, 1) // (2^64-1)^2
	check(^uint64(0)-1, 0)
	for range 1_000_000 {
		s := r.Uint64() | 1<<63
		// perfect squares and their predecessors
		h, l := bits.Mul64(s, s)
		check(h, l)
		l, borrow := bits.Sub64(l, 1, 0)
		check(h-borrow, l)
		check(r.Uint64()|1<<62, r.Uint64())
	}
}

func randomFloat128(r *rand.Rand) Float128 {
	a := Float128{r.Uint64() &^ signMask128[0], r.Uint64()}
	switch r.IntN(4) {
	case 0:
		// subnormal
		a[0] &= fracMask128[0] >> r.IntN(48)
		a[1] >>= r.IntN(64)
	case 1:
		// sparse fraction
		a[0] &^= fracMask128[0]
		a[1] = 1 << r.IntN(64)
	}
	return a
}

func randomFloat256(r *rand.Rand) Float256 {
	a := Float256{r.Uint64() &^ signMask256[0], r.Uint64(), r.Uint64(), r.Uint64()}
	switch r.IntN(4) {
	case 0:
		// subnormal
		a[0] &= fracMask256[0] >> r.IntN(44)
		a[3] >>= r.IntN(64)
	case 1:
		// sparse fraction
		a[0] &^= fracMask256[0]
		a[1], a[2], a[3] = 0, 0, 1<<r.IntN(64)
	}
	return a
}

func TestFloat128_SqrtRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(3, 4))
	for i := range 200_000 {
		a := randomFloat128(r)
		if a.IsZero() || a.IsNaN() || a.IsInf(0) {
			continue
		}
		// square of a value with a short fraction: exact results
		if i%4 == 0 {
			b := a.Sqrt()
			b[1] &^= 0xffffffff
			a = b.Mul(b)
			if a.IsZero() || a.IsInf(0) {
				continue
			}
		}
		got := a.Sqrt()
		want := sqrtBitByBit128(a)
		if got != want {
			t.Fatalf("Float128(%x).Sqrt() = %x, want %x", a, got, want)
		}
	}
}

func TestFloat256_SqrtRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(5, 6))
	for i := range 100_000 {
		a := randomFloat256(r)
		if a.IsZero() || a.IsNaN() || a.IsInf(0) {
			continue
		}
		if i%4 == 0 {
			b := a.Sqrt()
			b[3] = 0
			a = b.Mul(b)
			if a.IsZero() || a.IsInf(0) {
				continue
			}
		}
		got := a.Sqrt()
		want := sqrtBitByBit256(a)
		if got != want {
			t.Fatalf("Float256(%x).Sqrt() = %x, want %x", a, got, want)
		}
	}
}
