package floats

import (
	"math"
	"math/bits"

	"github.com/shogo82148/ints"
)

// isqrt128 returns floor(sqrt(hi<<64 | lo)).
// hi must be greater than or equal to 1<<62.
func isqrt128(hi, lo uint64) uint64 {
	// Start from an overestimate of the root using the hardware square root.
	// The relative error of f is about 2^-52, so adding 2^12 is enough.
	f := math.Sqrt(float64(hi)*0x1p64 + float64(lo))
	s := uint64(math.MaxUint64)
	if f < 0x1p64-0x1p12 {
		s = uint64(f) + 1<<12
	}

	// Newton's method from above.
	// It converges to floor(sqrt(x)) monotonically.
	for {
		if hi >= s {
			// s >= floor(sqrt(x)) >= hi >= s, so s is the answer.
			return s
		}
		q, _ := bits.Div64(hi, lo, s)
		sum, carry := bits.Add64(s, q, 0)
		t := sum>>1 | carry<<63
		if t >= s {
			return s
		}
		s = t
	}
}

// sqrtRem256 returns s = floor(sqrt(n)) and r = n - s*s.
// n[0] must be greater than or equal to 1<<62.
func sqrtRem256(n ints.Uint256) (s ints.Uint128, r ints.Uint256) {
	// Karatsuba square root (Paul Zimmermann, "Karatsuba Square Root", 1999)
	// with the base b = 2^64.
	sh := isqrt128(n[0], n[1])

	// r' = (n0, n1) - sh^2. It is at most 2*sh, so it fits in 65 bits.
	ph, pl := bits.Mul64(sh, sh)
	rl, borrow := bits.Sub64(n[1], pl, 0)
	rh, _ := bits.Sub64(n[0], ph, borrow)

	// q = floor((r'*b + n2) / (2*sh)) = floor(((r'*b + n2) >> 1) / sh)
	nh := rh<<63 | rl>>1
	nl := rl<<63 | n[2]>>1
	var q uint64
	if nh >= sh {
		// the quotient overflows. clamp it.
		q = math.MaxUint64
	} else {
		q, _ = bits.Div64(nh, nl, sh)
	}

	// s is an overestimate of the root by a small amount.
	s = ints.Uint128{sh, q}
	for {
		p := s.Mul256(s)
		if p.Cmp(n) <= 0 {
			return s, n.Sub(p)
		}
		s = s.Sub(ints.Uint128{0, 1})
	}
}

// sqrtRem512 returns s = floor(sqrt(n)) and whether n - s*s is not zero.
// n[0] must be greater than or equal to 1<<62.
func sqrtRem512(n ints.Uint512) (s ints.Uint256, inexact bool) {
	// Karatsuba square root with the base b = 2^128.
	sh, rem := sqrtRem256(ints.Uint256{n[0], n[1], n[2], n[3]})

	// q = floor((r'*b + n2) / (2*sh)) = floor(((r'*b + n2) >> 1) / sh)
	// r' is at most 2*sh, so it fits in 129 bits.
	num := ints.Uint256{rem[2], rem[3], n[4], n[5]}.Rsh(1)
	num[0] |= rem[1] << 63
	var q ints.Uint256
	div := ints.Uint256{0, 0, sh[0], sh[1]}
	if (ints.Uint128{num[0], num[1]}).Cmp(sh) >= 0 {
		// the quotient overflows. clamp it.
		q = ints.Uint256{0, 0, math.MaxUint64, math.MaxUint64}
	} else {
		q, _ = num.DivMod(div)
	}

	// s is an overestimate of the root by a small amount.
	s = ints.Uint256{sh[0], sh[1], q[2], q[3]}
	for {
		p := s.Mul512(s)
		if p.Cmp(n) <= 0 {
			return s, p != n
		}
		s = s.Sub(ints.Uint256{0, 0, 0, 1})
	}
}
