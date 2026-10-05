package floats

import (
	"math/rand/v2"
	"testing"

	"github.com/shogo82148/ints"
)

// roundToNearestEvenReference128 is the reference implementation of roundToNearestEven128.
func roundToNearestEvenReference128(x ints.Uint128, shift uint) ints.Uint128 {
	one := ints.Uint128{0, 1}
	mask := one.Lsh(shift - 1).Sub(one)
	x = x.Add(mask).Add(x.Rsh(shift).And(one))
	return x.Rsh(shift)
}

// roundToNearestEvenReference256 is the reference implementation of roundToNearestEven256.
func roundToNearestEvenReference256(x ints.Uint256, shift uint) ints.Uint256 {
	one := ints.Uint256{0, 0, 0, 1}
	mask := one.Lsh(shift - 1).Sub(one)
	x = x.Add(mask).Add(x.Rsh(shift).And(one))
	return x.Rsh(shift)
}

// roundToNearestEvenReference512 is the reference implementation of roundToNearestEven512.
func roundToNearestEvenReference512(x ints.Uint512, shift uint) ints.Uint512 {
	one := ints.Uint512{0, 0, 0, 0, 0, 0, 0, 1}
	mask := one.Lsh(uint(shift - 1)).Sub(one)
	return x.Add(mask).Add(x.Rsh(uint(shift)).And(one))
}

func TestRoundToNearestEven(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(11, 12))
	for range 1_000_000 {
		// keep the top bit clear so that the reference implementation doesn't overflow
		x := ints.Uint256{r.Uint64() >> 1, r.Uint64(), r.Uint64(), r.Uint64()}
		switch r.IntN(3) {
		case 0:
			// exact half
			x = x.Rsh(uint(r.IntN(256))).Or(ints.Uint256{0, 0, 0, 1}.Lsh(uint(r.IntN(255))))
		case 1:
			x = x.Rsh(uint(r.IntN(256)))
		}
		shift := uint(r.IntN(255) + 1)
		if got, want := roundToNearestEven256(x, shift), roundToNearestEvenReference256(x, shift); got != want {
			t.Fatalf("roundToNearestEven256(%x, %d) = %x, want %x", x, shift, got, want)
		}

		x128 := ints.Uint128{x[2] >> 1, x[3]}
		shift128 := shift%128 + 1
		if got, want := roundToNearestEven128(x128, shift128), roundToNearestEvenReference128(x128, shift128); got != want {
			t.Fatalf("roundToNearestEven128(%x, %d) = %x, want %x", x128, shift128, got, want)
		}

		if got, want := shrcompress256(x, shift), shrcompressReference256(x, shift); got != want {
			t.Fatalf("shrcompress256(%x, %d) = %x, want %x", x, shift, got, want)
		}
		if got, want := shrcompress128(x128, shift128-1), shrcompressReference128(x128, shift128-1); got != want {
			t.Fatalf("shrcompress128(%x, %d) = %x, want %x", x128, shift128-1, got, want)
		}
	}
}

func shrcompressReference128(x ints.Uint128, n uint) ints.Uint128 {
	if n >= 128 {
		return nonzero128(x)
	}
	one := ints.Uint128{0, 1}
	mask := one.Lsh(n).Sub(one)
	y := x.Rsh(n)
	y = y.Or(nonzero128(x.And(mask)))
	return y
}

func shrcompressReference256(x ints.Uint256, n uint) ints.Uint256 {
	if n >= 256 {
		return nonzero256(x)
	}
	one := ints.Uint256{0, 0, 0, 1}
	mask := one.Lsh(n).Sub(one)
	y := x.Rsh(n)
	y = y.Or(nonzero256(x.And(mask)))
	return y
}

// addReference128 is the reference implementation of Float128.Add
// that adds the fractions using 256-bit signed integers.
func addReference128(a, b Float128) Float128 {
	if a.IsNaN() || b.IsNaN() {
		// a + NaN = NaN
		// NaN + b = NaN
		return Float128(uvnan128)
	}
	if a.IsZero() {
		if b.IsZero() {
			//  0 +  0 =  0
			//  0 + -0 =  0
			// -0 +  0 =  0
			// -0 + -0 = -0
			return Float128{a[0] & b[0], 0}
		}
		// ±0 + b = b
		return b
	}
	if b.IsZero() {
		// a + ±0 = a
		return a
	}

	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()

	// handle special cases
	if expA == mask128-bias128 {
		// NaN check is done above; a is ±inf
		if expB == mask128-bias128 {
			if signA == signB {
				// ±inf + ±inf = ±inf
				return Float128{signA | uvinf128[0], uvinf128[1]}
			}
			// ±inf + ∓inf = NaN
			return Float128(uvnan128)
		}
		// b is finite, the result is ±inf
		return a
	}
	if expB == mask128-bias128 {
		// NaN check is done above; b is ±inf
		// NaN and Inf checks are done above; a is finite.
		return b
	}

	if expA < expB {
		// swap a and b
		signA, signB = signB, signA
		expA, expB = expB, expA
		fracA, fracB = fracB, fracA
	}

	// add the fractions
	const offset = 128
	fracA256 := ints.Int256{fracA[0], fracA[1], 0, 0}
	fracB256 := ints.Int256{fracB[0], fracB[1], 0, 0}
	fracB256 = fracB256.Rsh(uint(expA - expB))
	if signA != 0 {
		fracA256 = fracA256.Neg()
	}
	if signB != 0 {
		fracB256 = fracB256.Neg()
	}
	frac256 := fracA256.Add(fracB256)
	sign := uint64(0)
	if frac256.Sign() < 0 {
		sign = signMask128[0]
		frac256 = frac256.Neg()
	}

	shift := ints.Uint256(frac256).BitLen() - (shift128 + 1)
	exp := expA + shift - offset

	if frac256.IsZero() || exp < -(bias128+shift128) {
		// underflow
		return Float128{sign, 0}
	}
	if exp <= -bias128 {
		// the result is subnormal
		shift := offset - (expA + bias128) + 1
		frac256 = ints.Int256(roundToNearestEvenReference256(ints.Uint256(frac256), uint(shift)))
		return Float128{sign | frac256[2]&fracMask128[0], frac256[3]}
	}
	if exp >= mask128-bias128 {
		// overflow
		return Float128{sign | uvinf128[0], uvinf128[1]}
	}

	frac256 = ints.Int256(roundToNearestEvenReference256(ints.Uint256(frac256), uint(shift)))
	// detect carry-out caused by rounding
	if ints.Uint256(frac256).BitLen() > shift128+1 {
		frac256 = frac256.Rsh(1)
		exp++
		if exp >= mask128 {
			// overflow
			return Float128{sign | uvinf128[0], uvinf128[1]}
		}
	}
	return Float128{sign | uint64(exp+bias128)<<(shift128-64) | frac256[2]&fracMask128[0], frac256[3]}
}

// addReference256 is the reference implementation of Float256.Add
// that adds the fractions using 512-bit signed integers.
func addReference256(a, b Float256) Float256 {
	if a.IsNaN() || b.IsNaN() {
		// a + NaN = NaN
		// NaN + b = NaN
		return Float256(uvnan256)
	}
	if a.IsZero() {
		if b.IsZero() {
			//  0 +  0 =  0
			//  0 + -0 =  0
			// -0 +  0 =  0
			// -0 + -0 = -0
			return Float256{a[0] & b[0], 0, 0, 0}
		}
		// ±0 + b = b
		return b
	}
	if b.IsZero() {
		// a + ±0 = a
		return a
	}

	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()

	// handle special cases
	if expA == mask256-bias256 {
		// NaN check is done above; a is ±inf
		if expB == mask256-bias256 {
			if signA == signB {
				// ±inf + ±inf = ±inf
				return Float256{signA | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
			}
			// ±inf + ∓inf = NaN
			return Float256(uvnan256)
		}
		// b is finite, the result is ±inf
		return a
	}
	if expB == mask256-bias256 {
		// NaN check is done above; b is ±inf
		// NaN and Inf checks are done above; a is finite.
		return b
	}

	if expA < expB {
		// swap a and b
		signA, signB = signB, signA
		expA, expB = expB, expA
		fracA, fracB = fracB, fracA
	}

	// add the fractions
	const offset = 256
	fracA512 := ints.Uint512{fracA[0], fracA[1], fracA[2], fracA[3], 0, 0, 0, 0}
	fracB512 := ints.Uint512{fracB[0], fracB[1], fracB[2], fracB[3], 0, 0, 0, 0}
	fracB512 = fracB512.Rsh(uint(expA - expB))
	if signA != 0 {
		fracA512 = fracA512.Neg()
	}
	if signB != 0 {
		fracB512 = fracB512.Neg()
	}
	frac512 := fracA512.Add(fracB512)
	sign := uint64(0)
	if ints.Int512(frac512).Sign() < 0 {
		sign = signMask256[0]
		frac512 = frac512.Neg()
	}

	shift := frac512.BitLen() - (shift256 + 1)
	exp := expA + shift - offset

	if frac512.IsZero() || exp < -(bias256+shift256) {
		// underflow
		return Float256{sign, 0, 0, 0}
	}
	if exp <= -bias256 {
		// the result is subnormal
		shift := offset - (expA + bias256) + 1
		frac512 = roundToNearestEvenReference512(frac512, uint(shift))
		frac512 = frac512.Rsh(uint(shift))
		return Float256{sign | frac512[4], frac512[5], frac512[6], frac512[7]}
	}
	if exp >= mask256-bias256 {
		// overflow
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}

	frac512 = roundToNearestEvenReference512(frac512, uint(shift))
	// detect carry-out caused by rounding
	if frac512.BitLen() > shift256+shift+1 {
		frac512 = frac512.Rsh(1)
		exp++
		if exp >= mask256 {
			// overflow
			return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
		}
	}
	frac512 = frac512.Rsh(uint(shift))
	return Float256{
		sign | uint64(exp+bias256)<<(shift256-192) | frac512[4]&fracMask256[0],
		frac512[5],
		frac512[6],
		frac512[7],
	}
}

func randomAddOperand128(r *rand.Rand, a Float128) Float128 {
	b := randomFloat128(r)
	switch r.IntN(4) {
	case 0:
		// close to a, to cause cancellation
		b = a
		b[1] ^= r.Uint64() >> r.IntN(64)
	case 1:
		// exponent close to a
		b[0] = b[0]&^(mask128<<(shift128-64)) | a[0]&(mask128<<(shift128-64))
		b[0] += uint64(r.IntN(260)-130) << (shift128 - 64)
	}
	if r.IntN(2) == 0 {
		b = b.Neg()
	}
	return b
}

func TestFloat128_AddRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(7, 8))
	for range 2_000_000 {
		a := randomFloat128(r)
		if r.IntN(2) == 0 {
			a = a.Neg()
		}
		b := randomAddOperand128(r, a)
		got := a.Add(b)
		want := addReference128(a, b)
		if got != want && !(got.IsNaN() && want.IsNaN()) {
			t.Fatalf("Float128(%x).Add(%x) = %x, want %x", a, b, got, want)
		}
	}
}

func randomAddOperand256(r *rand.Rand, a Float256) Float256 {
	b := randomFloat256(r)
	switch r.IntN(4) {
	case 0:
		// close to a, to cause cancellation
		b = a
		b[3] ^= r.Uint64() >> r.IntN(64)
	case 1:
		// exponent close to a
		b[0] = b[0]&^(mask256<<(shift256-192)) | a[0]&(mask256<<(shift256-192))
		b[0] += uint64(r.IntN(520)-260) << (shift256 - 192)
	}
	if r.IntN(2) == 0 {
		b = b.Neg()
	}
	return b
}

func TestFloat256_AddRandom(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(9, 10))
	for range 1_000_000 {
		a := randomFloat256(r)
		if r.IntN(2) == 0 {
			a = a.Neg()
		}
		b := randomAddOperand256(r, a)
		got := a.Add(b)
		want := addReference256(a, b)
		if got != want && !(got.IsNaN() && want.IsNaN()) {
			t.Fatalf("Float256(%x).Add(%x) = %x, want %x", a, b, got, want)
		}
	}
}
