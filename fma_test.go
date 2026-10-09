package floats

import (
	"math"
	"math/bits"
	"math/rand/v2"
	"testing"

	"github.com/shogo82148/ints"
)

// fmaReference128 is the reference implementation of FMA128
// that uses the constant-time shifts of the ints package.
func fmaReference128(x, y, z Float128) Float128 {
	if x.IsZero() || y.IsZero() || x[0]&(mask128<<(shift128-64)) == (mask128<<(shift128-64)) || y[0]&(mask128<<(shift128-64)) == (mask128<<(shift128-64)) {
		return x.Mul(y).Add(z)
	}
	if z.IsZero() {
		return x.Mul(y)
	}
	// Handle non-finite z separately. Evaluating x*y+z where
	// x and y are finite, but z is infinite, should always result in z.
	if z[0]&(mask128<<(shift128-64)) == (mask128 << (shift128 - 64)) {
		return z
	}

	// Split x, y, z into sign, exponent, mantissa.
	signX, expX, fracX := x.normalize()
	signY, expY, fracY := y.normalize()
	signZ, expZ, fracZ0 := z.normalize()

	// Compute product p = x*y as sign, exponent, mantissa.
	expP := expX + expY + 1
	fracP := fracX.Lsh(14).Mul256(fracY.Lsh(15))
	signP := signX ^ signY // product sign

	// Normalize the product
	is254zero := uint((^fracP[0] >> 62) & 1)
	fracP = fracP.Lsh(is254zero)
	expP -= int(is254zero)

	fracZ := fracZ0.Uint256().Lsh(14 + 128)

	// Swap addition operands so |p| >= |z|
	if expP < expZ || expP == expZ && fracP.Cmp(fracZ) < 0 {
		signP, signZ = signZ, signP
		expP, expZ = expZ, expP
		fracP, fracZ = fracZ, fracP
	}

	// Special case: if p == -z the result is always +0 since neither operand is zero.
	if signP != signZ && expP == expZ && fracP.Cmp(fracZ) == 0 {
		return Float128{0, 0}
	}

	// Align mantissa
	fracZ = shrcompressReference256(fracZ, uint(expP-expZ))

	// Compute resulting significands, normalizing if necessary.
	var frac ints.Uint128
	if signP == signZ {
		// Adding fracP + fracZ
		fracP = fracP.Add(fracZ)
		expP += int(fracP[0] >> 63)
		frac = shrcompressReference256(fracP, uint(128+fracP[0]>>63)).Uint128()
	} else {
		// Subtracting fracP - fracZ
		fracP = fracP.Sub(fracZ)
		nz := fracP.LeadingZeros() - 1
		expP -= nz
		frac = shrcompressReference256(fracP.Lsh(uint(nz)), 128).Uint128()
	}

	// check for underflow
	expP += bias128
	if expP <= 0 {
		n := uint(1 - expP)
		frac = roundToNearestEvenReference128(frac, n+14)
		return Float128{signP | frac[0], frac[1]}
	}

	// Round and break ties to even
	frac = roundToNearestEvenReference128(frac, 14)
	if frac[0]&(1<<(shift128+1-64)) != 0 {
		expP++
		frac = frac.Rsh(1)
	}
	if expP >= mask128 {
		// overflow
		return Float128{signP | uvinf128[0], uvinf128[1]}
	}
	return Float128{
		signP | uint64(expP)<<(shift128-64) | frac[0]&fracMask128[0],
		frac[1] & fracMask128[1],
	}
}

// fmaReference256 is the reference implementation of FMA256
// that uses the constant-time shifts of the ints package.
func fmaReference256(x, y, z Float256) Float256 {
	if x.IsZero() || y.IsZero() || x[0]&(mask256<<(shift256-192)) == mask256<<(shift256-192) || y[0]&(mask256<<(shift256-192)) == mask256<<(shift256-192) {
		return x.Mul(y).Add(z)
	}
	if z.IsZero() {
		return x.Mul(y)
	}
	// Handle non-finite z separately. Evaluating x*y+z where
	// x and y are finite, but z is infinite, should always result in z.
	if z[0]&(mask256<<(shift256-192)) == mask256<<(shift256-192) {
		return z
	}

	// Split x, y, z into sign, exponent, mantissa.
	signX, expX, fracX := x.normalize()
	signY, expY, fracY := y.normalize()
	signZ, expZ, fracZ0 := z.normalize()

	// Compute product p = x*y as sign, exponent, mantissa.
	expP := expX + expY + 1
	fracP := fracX.Lsh(18).Mul512(fracY.Lsh(19))
	signP := signX ^ signY // product sign

	// Normalize the product
	is510zero := uint((^fracP[0] >> 62) & 1)
	fracP = fracP.Lsh(is510zero)
	expP -= int(is510zero)

	fracZ := fracZ0.Uint512().Lsh(18 + 256)

	// Swap addition operands so |p| >= |z|
	if expP < expZ || expP == expZ && fracP.Cmp(fracZ) < 0 {
		signP, signZ = signZ, signP
		expP, expZ = expZ, expP
		fracP, fracZ = fracZ, fracP
	}

	// Special case: if p == -z the result is always +0 since neither operand is zero.
	if signP != signZ && expP == expZ && fracP.Cmp(fracZ) == 0 {
		return Float256{0, 0, 0, 0}
	}

	// Align mantissa
	fracZ = shrcompressReference512(fracZ, uint(expP-expZ))

	// Compute resulting significands, normalizing if necessary.
	var frac ints.Uint256
	if signP == signZ {
		// Adding fracP + fracZ
		fracP = fracP.Add(fracZ)
		expP += int(fracP[0] >> 63)
		frac = shrcompressReference512(fracP, uint(256+fracP[0]>>63)).Uint256()
	} else {
		// Subtracting fracP - fracZ
		fracP = fracP.Sub(fracZ)
		nz := fracP.LeadingZeros() - 1
		expP -= nz
		frac = shrcompressReference512(fracP.Lsh(uint(nz)), 256).Uint256()
	}

	// check for underflow
	expP += bias256
	if expP <= 0 {
		n := uint(1 - expP)
		frac = roundToNearestEvenReference256(frac, n+18)
		return Float256{signP | frac[0], frac[1], frac[2], frac[3]}
	}

	// Round and break ties to even
	frac = roundToNearestEvenReference256(frac, 18)
	if frac[0]&(1<<(shift256+1-192)) != 0 {
		expP++
		frac = frac.Rsh(1)
	}
	if expP >= mask256 {
		// Overflow
		return Float256{signP | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}
	return Float256{
		signP | uint64(expP)<<(shift256-192) | frac[0]&fracMask256[0],
		frac[1],
		frac[2],
		frac[3],
	}
}

func shrcompressReference512(x ints.Uint512, n uint) ints.Uint512 {
	if n >= 512 {
		return nonzero512(x)
	}
	one := ints.Uint512{0, 0, 0, 0, 0, 0, 0, 1}
	mask := one.Lsh(n).Sub(one)
	y := x.Rsh(n)
	y = y.Or(nonzero512(x.And(mask)))
	return y
}

func TestShift(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(25, 26))
	for range 1_000_000 {
		var x ints.Uint512
		randomWords(r, x[:])
		n := uint(r.IntN(530))
		if got, want := lsh512(x, n), x.Lsh(n); got != want {
			t.Fatalf("lsh512(%x, %d) = %x, want %x", x, n, got, want)
		}
		if got, want := shrcompress512(x, n), shrcompressReference512(x, n); got != want {
			t.Fatalf("shrcompress512(%x, %d) = %x, want %x", x, n, got, want)
		}

		x256 := ints.Uint256{x[0], x[1], x[2], x[3]}
		n256 := n / 2
		if got, want := lsh256(x256, n256), x256.Lsh(n256); got != want {
			t.Fatalf("lsh256(%x, %d) = %x, want %x", x256, n256, got, want)
		}
		if got, want := shrcompress256(x256, n256), shrcompressReference256(x256, n256); got != want {
			t.Fatalf("shrcompress256(%x, %d) = %x, want %x", x256, n256, got, want)
		}
	}
}

// randomFMAOperands128 returns x, y, z for testing FMA128.
// z is chosen to be close to -x*y or to the subnormal / overflow ranges.
func randomFMAOperands128(r *rand.Rand) (x, y, z Float128) {
	x = randomFloat128(r)
	y = randomMulOperand128(r, x)
	if r.IntN(2) == 0 {
		x = x.Neg()
	}
	if r.IntN(2) == 0 {
		// short fractions, to make exact products
		x = shortenFraction128(r, x)
		y = shortenFraction128(r, y)
	}
	p := x.Mul(y)
	switch r.IntN(4) {
	case 0:
		// cancellation: z is close to -x*y
		z = p.Neg()
		z[1] ^= r.Uint64() >> r.IntN(64)
	case 1:
		// the exponent of z is close to the one of x*y
		z = randomFloat128(r)
		z[0] = z[0]&^(mask128<<(shift128-64)) | p[0]&(mask128<<(shift128-64))
		z[0] += uint64(r.IntN(260)-130) << (shift128 - 64)
	case 2:
		// z is much smaller than x*y
		z = shortenFraction128(r, randomFloat128(r))
		e := int(p[0]>>(shift128-64)&mask128) - r.IntN(400)
		if e <= 0 {
			z = Float128{0, r.Uint64() >> r.IntN(64)}
		} else {
			z[0] = z[0]&^(mask128<<(shift128-64)) | uint64(e)<<(shift128-64)
		}
	default:
		z = randomFloat128(r)
	}
	if r.IntN(2) == 0 {
		z = z.Neg()
	}
	return
}

// shortenFraction128 clears the random number of low bits of the fraction of x.
func shortenFraction128(r *rand.Rand, x Float128) Float128 {
	n := uint(r.IntN(shift128))
	return Float128(ints.Uint128(x).Rsh(n).Lsh(n))
}

func TestFMA128Random(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(27, 28))
	for range 2_000_000 {
		x, y, z := randomFMAOperands128(r)
		got := FMA128(x, y, z)
		want := fmaReference128(x, y, z)
		if got != want && !(got.IsNaN() && want.IsNaN()) {
			t.Fatalf("FMA128(%x, %x, %x) = %x, want %x", x, y, z, got, want)
		}
	}
}

// randomFMAOperands256 returns x, y, z for testing FMA256.
// z is chosen to be close to -x*y or to the subnormal / overflow ranges.
func randomFMAOperands256(r *rand.Rand) (x, y, z Float256) {
	x = randomFloat256(r)
	y = randomMulOperand256(r, x)
	if r.IntN(2) == 0 {
		x = x.Neg()
	}
	if r.IntN(2) == 0 {
		// short fractions, to make exact products
		x = shortenFraction256(r, x)
		y = shortenFraction256(r, y)
	}
	p := x.Mul(y)
	switch r.IntN(4) {
	case 0:
		// cancellation: z is close to -x*y
		z = p.Neg()
		z[3] ^= r.Uint64() >> r.IntN(64)
		if r.IntN(2) == 0 {
			z[2] ^= r.Uint64() >> r.IntN(64)
		}
	case 1:
		// the exponent of z is close to the one of x*y
		z = randomFloat256(r)
		z[0] = z[0]&^(mask256<<(shift256-192)) | p[0]&(mask256<<(shift256-192))
		z[0] += uint64(r.IntN(520)-260) << (shift256 - 192)
	case 2:
		// z is much smaller than x*y
		z = shortenFraction256(r, randomFloat256(r))
		e := int(p[0]>>(shift256-192)&mask256) - r.IntN(800)
		if e <= 0 {
			z = Float256{0, 0, 0, r.Uint64() >> r.IntN(64)}
		} else {
			z[0] = z[0]&^(mask256<<(shift256-192)) | uint64(e)<<(shift256-192)
		}
	default:
		z = randomFloat256(r)
	}
	if r.IntN(2) == 0 {
		z = z.Neg()
	}
	return
}

// shortenFraction256 clears the random number of low bits of the fraction of x.
func shortenFraction256(r *rand.Rand, x Float256) Float256 {
	n := uint(r.IntN(shift256))
	return Float256(ints.Uint256(x).Rsh(n).Lsh(n))
}

func TestFMA256Random(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(29, 30))
	for range 1_000_000 {
		x, y, z := randomFMAOperands256(r)
		got := FMA256(x, y, z)
		want := fmaReference256(x, y, z)
		if got != want && !(got.IsNaN() && want.IsNaN()) {
			t.Fatalf("FMA256(%x, %x, %x) = %x, want %x", x, y, z, got, want)
		}
	}
}

// TestFMA256Sticky tests the cases where the sum is exactly halfway between
// two representable values except for the bits far below,
// so that the sticky bit decides the rounding direction.
func TestFMA256Sticky(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x, y, z, want Float256
	}{
		// the sum carries out, and the bit shifted out by the normalization is the sticky bit.
		{
			Float256{0x4000000000000000, 0x0, 0x800020000, 0x0},
			Float256{0x3fffe00000000000, 0x200, 0x0, 0x0},
			Float256{0x3ffff00000000000, 0x0, 0x0, 0x0},
			Float256{0x4000000000000000, 0x100, 0x400010000, 0x1},
		},
		// the sticky bit is in the least significant word.
		{
			Float256{0x3ffff08000000000, 0x0, 0x0, 0x2},
			Float256{0x3fffe00000000000, 0x0, 0x0, 0x10},
			Float256{0xbfe0600000000000, 0x0, 0x1000000000000, 0x0},
			Float256{0x3fffe08000000000, 0x0, 0x0, 0x13},
		},
	}
	for _, tt := range tests {
		if got := FMA256(tt.x, tt.y, tt.z); got != tt.want {
			t.Errorf("FMA256(%x, %x, %x) = %x, want %x", tt.x, tt.y, tt.z, got, tt.want)
		}
		if got := fmaReference256(tt.x, tt.y, tt.z); got != tt.want {
			t.Errorf("fmaReference256(%x, %x, %x) = %x, want %x", tt.x, tt.y, tt.z, got, tt.want)
		}
	}
}

// fmaReference32 is the reference implementation of FMA32. It returns x * y + z, computed with only one rounding.
// (That is, FMA32 returns the fused multiply-add of x, y, and z.)
func fmaReference32(x, y, z Float32) Float32 {
	// Split x, y, z into sign, exponent, mantissa.
	signX, expX, fracX := x.normalize()
	signY, expY, fracY := y.normalize()
	signZ, expZ, fracZ0 := z.normalize()

	// Inf or NaN involved. At most one rounding will occur.
	if x == 0 || y == 0 || expX == mask32-bias32 || expY == mask32-bias32 {
		return x*y + z
	}
	if z == 0 {
		return x * y
	}
	// Handle non-finite z separately. Evaluating x*y+z where
	// x and y are finite, but z is infinite, should always result in z.
	if expZ == mask32-bias32 {
		return z
	}

	// Compute product p = x*y as sign, exponent, mantissa.
	expP := expX + expY + 1
	fracP := uint64(fracX<<7) * uint64(fracY<<8)
	signP := signX ^ signY // product sign

	// Normalize product.
	is62zero := uint((^fracP >> 62) & 1)
	fracP <<= is62zero
	expP -= int(is62zero)

	fracZ := uint64(fracZ0) << (7 + 32)

	// Swap addition operands so |p| >= |z|
	if expP < expZ || expP == expZ && fracP < fracZ {
		signP, signZ = signZ, signP
		expP, expZ = expZ, expP
		fracP, fracZ = fracZ, fracP
	}

	// Special case: if p == -z the result is always +0 since neither operand is zero.
	if signP != signZ && expP == expZ && fracP == fracZ {
		return 0
	}

	// Align mantissa
	fracZ = shrcompress64(fracZ, uint(expP-expZ))

	// Compute resulting significands, normalizing if necessary.
	var frac uint32
	if signP == signZ {
		// Adding fracP + fracZ
		fracP += fracZ
		expP += int(fracP >> 63)
		frac = uint32(shrcompress64(fracP, uint(32+fracP>>63)))
	} else {
		// Subtracting fracP - fracZ
		fracP -= fracZ
		nz := bits.LeadingZeros64(fracP) - 1
		expP -= nz
		frac = uint32(shrcompress64(fracP<<uint(nz), 32))
	}

	// check for underflow
	expP += bias32
	if expP <= 0 {
		n := uint(1 - expP)
		frac = roundToNearestEven32(frac, n+7)
		return Float32(math.Float32frombits(signP | frac))
	}

	// Round and break ties to even
	frac = roundToNearestEven32(frac, 7)
	if frac&(1<<(shift32+1)) != 0 {
		expP++
		frac >>= 1
	}
	if expP >= mask32 {
		// overflow
		return Float32(math.Float32frombits(signP | uvinf32))
	}
	return Float32(math.Float32frombits(signP | uint32(expP<<shift32) | frac&fracMask32))
}

// randomFMAOperands32 returns x, y, z for testing FMA32.
// z is chosen to be close to -x*y or to the subnormal / overflow ranges.
func randomFMAOperands32(r *rand.Rand) (x, y, z Float32) {
	rnd := func() Float32 {
		b := r.Uint32()
		switch r.IntN(8) {
		case 0:
			b &^= 0x7f80_0000 // subnormal
		case 1:
			b |= 0x7f80_0000 // Inf, NaN
		case 2:
			b = b&^0x7f80_0000 | uint32(r.IntN(40)+110)<<23 // close to 1
		}
		return Float32(math.Float32frombits(b))
	}
	short := func(a Float32) Float32 {
		n := uint(r.IntN(24))
		return Float32(math.Float32frombits(math.Float32bits(float32(a)) >> n << n))
	}
	x, y = rnd(), rnd()
	if r.IntN(2) == 0 {
		// short fractions, to make exact products
		x, y = short(x), short(y)
	}
	if r.IntN(2) == 0 {
		// the exponent of y is chosen so that x*y is in the normal range
		e := int(math.Float32bits(float32(x))>>23&0xff) - 127
		b := math.Float32bits(float32(y))&^0x7f80_0000 | uint32(127-e+r.IntN(40)-20)&0xff<<23
		y = Float32(math.Float32frombits(b))
	}
	p := float64(x) * float64(y)
	switch r.IntN(4) {
	case 0:
		// cancellation: z is close to -x*y
		z = Float32(-p)
		z = Float32(math.Float32frombits(math.Float32bits(float32(z)) ^ uint32(r.Uint64()>>r.IntN(64))&0xffff))
	case 1:
		// the exponent of z is close to the one of x*y
		z = rnd()
		e := int(math.Float64bits(p)>>52&0x7ff) - 1023 + r.IntN(60) - 30
		z = Float32(math.Ldexp(float64(math.Float32frombits(math.Float32bits(float32(z))&0x807f_ffff|127<<23)), e))
	case 2:
		// z is much smaller than x*y
		z = short(rnd())
		e := int(math.Float64bits(p)>>52&0x7ff) - 1023 - r.IntN(100)
		z = Float32(math.Ldexp(float64(math.Float32frombits(math.Float32bits(float32(z))&0x807f_ffff|127<<23)), e))
	default:
		z = rnd()
	}
	if r.IntN(2) == 0 {
		z = -z
	}
	return
}

func TestFMA32Random(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(31, 32))
	for range 5_000_000 {
		x, y, z := randomFMAOperands32(r)
		got := FMA32(x, y, z)
		want := fmaReference32(x, y, z)
		if math.Float32bits(float32(got)) != math.Float32bits(float32(want)) && !(got.IsNaN() && want.IsNaN()) {
			t.Fatalf("FMA32(%x, %x, %x) = %x, want %x", x, y, z, got, want)
		}
	}
}
