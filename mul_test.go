package floats

import (
	"math/rand/v2"
	"testing"
)

// mulReference128 is the reference implementation of Float128.Mul
// that rounds the full 256-bit product.
func mulReference128(a, b Float128) Float128 {
	if a.IsNaN() || b.IsNaN() {
		// a * NaN = NaN
		// NaN * b = NaN
		return Float128(uvnan128)
	}

	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()
	sign := signA ^ signB

	// handle special cases
	if expA == mask128-bias128 {
		// NaN check is done above; a is ±inf
		if b.IsZero() {
			// ±inf * 0 = NaN
			return Float128(uvnan128)
		} else {
			// ±inf * +finite = ±inf
			// ±inf * -finite = ∓inf
			return Float128{sign | uvinf128[0], uvinf128[1]}
		}
	}
	if expB == mask128-bias128 {
		// NaN check is done above; b is ±inf
		if a.IsZero() {
			// 0 * ±inf = NaN
			return Float128(uvnan128)
		} else {
			// +finite * ±inf = ±inf
			// -finite * ±inf = ∓inf
			return Float128{sign | uvinf128[0], uvinf128[1]}
		}
	}
	if a.IsZero() || b.IsZero() {
		// 0 * finite = 0
		return Float128{sign, 0}
	}

	exp := expA + expB
	frac := fracA.Mul256(fracB)
	shift := frac.BitLen() - (shift128 + 1)
	exp += shift - shift128

	if exp < -(bias128 + shift128) {
		// underflow
		return Float128{sign, 0}
	} else if exp <= -bias128 {
		// the result is subnormal
		// normalize
		shift := shift128 - (expA + expB + bias128) + 1
		frac = roundToNearestEvenReference256(frac, uint(shift))
		return Float128{sign | frac[2], frac[3]}
	}

	exp = expA + expB + bias128
	frac = roundToNearestEvenReference256(frac, uint(shift))
	shift = frac.BitLen() + shift - (shift128 + 1)
	exp += shift - shift128
	if exp >= mask128 {
		// overflow
		return Float128{sign | uvinf128[0], uvinf128[1]}
	}
	return Float128{
		sign | uint64(exp)<<(shift128-64) | frac[2]&fracMask128[0],
		frac[3],
	}
}

// mulReference256 is the reference implementation of Float256.Mul
// that rounds the full 512-bit product.
func mulReference256(a, b Float256) Float256 {
	if a.IsNaN() || b.IsNaN() {
		// a * NaN = NaN
		// NaN * b = NaN
		return Float256(uvnan256)
	}
	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()
	sign := signA ^ signB

	// handle special cases
	if expA == mask256-bias256 {
		// NaN check is done above; a is ±inf
		if b.IsZero() {
			// ±inf * 0 = NaN
			return Float256(uvnan256)
		} else {
			// ±inf * +finite = ±inf
			// ±inf * -finite = ∓inf
			return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
		}
	}
	if expB == mask256-bias256 {
		// NaN check is done above; b is ±inf
		if a.IsZero() {
			// 0 * ±inf = NaN
			return Float256(uvnan256)
		} else {
			// +finite * ±inf = ±inf
			// -finite * ±inf = ∓inf
			return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
		}
	}
	if a.IsZero() || b.IsZero() {
		// +0 * ±finite = ±0
		// -0 * ±finite = ∓0
		return Float256{sign, 0, 0, 0}
	}

	// normal case
	exp := expA + expB
	frac := fracA.Mul512(fracB)
	shift := frac.BitLen() - (shift256 + 1)
	exp += shift - shift256

	if exp < -(bias256 + shift256) {
		// underflow
		return Float256{sign, 0, 0, 0}
	} else if exp <= -bias256 {
		// the result is subnormal
		// normalize
		shift := shift256 - (expA + expB + bias256) + 1
		frac = roundToNearestEvenReference512(frac, uint(shift))
		frac = frac.Rsh(uint(shift))
		return Float256{sign | frac[4], frac[5], frac[6], frac[7]}
	}

	exp = expA + expB + bias256
	frac = roundToNearestEvenReference512(frac, uint(shift))
	shift = frac.BitLen() - (shift256 + 1)
	exp += shift - shift256
	if exp >= mask256 {
		// overflow
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}

	frac = frac.Rsh(uint(shift))
	return Float256{
		sign | uint64(exp)<<(shift256-192) | frac[4]&fracMask256[0],
		frac[5],
		frac[6],
		frac[7],
	}
}

func randomMulOperand128(r *rand.Rand, a Float128) Float128 {
	b := randomFloat128(r)
	switch r.IntN(4) {
	case 0:
		// the product is close to the subnormal range or the overflow threshold
		// the biased exponent of the product is about ea + eb - bias.
		e := uint64(r.IntN(260) - 130 + bias128)
		if r.IntN(2) == 0 {
			e += 2 * bias128
		}
		ea := a[0] >> (shift128 - 64) & mask128
		b[0] = b[0]&^(mask128<<(shift128-64)) | (e-ea)&mask128<<(shift128-64)
	case 1:
		// short fractions, to make exact products and ties
		b[1] &^= 1<<r.IntN(64) - 1
	}
	if r.IntN(2) == 0 {
		b = b.Neg()
	}
	return b
}

func TestFloat128_MulRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(13, 14))
	for range 2_000_000 {
		a := randomFloat128(r)
		if r.IntN(2) == 0 {
			a = a.Neg()
		}
		if r.IntN(4) == 0 {
			a[1] &^= 1<<r.IntN(64) - 1
		}
		b := randomMulOperand128(r, a)
		got := a.Mul(b)
		want := mulReference128(a, b)
		if !eq128(got, want) {
			t.Fatalf("Float128(%x).Mul(%x) = %x, want %x", a, b, got, want)
		}
	}
}

func randomMulOperand256(r *rand.Rand, a Float256) Float256 {
	b := randomFloat256(r)
	switch r.IntN(4) {
	case 0:
		// the product is close to the subnormal range or the overflow threshold
		// the biased exponent of the product is about ea + eb - bias.
		e := uint64(r.IntN(520) - 260 + bias256)
		if r.IntN(2) == 0 {
			e += 2 * bias256
		}
		ea := a[0] >> (shift256 - 192) & mask256
		b[0] = b[0]&^(mask256<<(shift256-192)) | (e-ea)&mask256<<(shift256-192)
	case 1:
		// short fractions, to make exact products and ties
		b[2], b[3] = 0, 0
	}
	if r.IntN(2) == 0 {
		b = b.Neg()
	}
	return b
}

func TestFloat256_MulRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(15, 16))
	for range 1_000_000 {
		a := randomFloat256(r)
		if r.IntN(2) == 0 {
			a = a.Neg()
		}
		if r.IntN(4) == 0 {
			a[2], a[3] = 0, 0
		}
		b := randomMulOperand256(r, a)
		got := a.Mul(b)
		want := mulReference256(a, b)
		if !eq256(got, want) {
			t.Fatalf("Float256(%x).Mul(%x) = %x, want %x", a, b, got, want)
		}
	}
}
