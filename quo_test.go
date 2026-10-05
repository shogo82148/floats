package floats

import (
	"math/rand/v2"
	"testing"

	"github.com/shogo82148/ints"
)

// quoReference128 is the reference implementation of Float128.Quo
// that uses the generic 256-bit division.
func quoReference128(a, b Float128) Float128 {
	if a.IsNaN() || b.IsNaN() {
		// a / NaN = NaN
		// NaN / b = NaN
		return Float128(uvnan128)
	}

	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()
	sign := signA ^ signB

	if b.IsZero() {
		if a.IsZero() {
			// 0 / 0 = NaN
			return Float128(uvnan128)
		}
		// ±finite / 0 = ±inf
		return Float128{sign | uvinf128[0], uvinf128[1]}
	}
	if a.IsZero() {
		// 0 / finite = 0
		return Float128{sign, 0}
	}
	if expA == mask128-bias128 {
		// NaN check is done above; a is ±inf
		if expB == mask128-bias128 {
			// ±inf / ±inf = NaN
			return Float128(uvnan128)
		} else {
			// ±inf / finite = ±inf
			return Float128{sign | uvinf128[0], uvinf128[1]}
		}
	}
	if expB == mask128-bias128 {
		// NaN check is done above; b is ±inf
		// NaN and Inf checks are done above; a is finite.
		// ±finite / ±inf = 0
		return Float128{sign, 0}
	}

	exp := expA - expB + bias128
	if fracA.Cmp(fracB) < 0 {
		exp--
		fracA = fracA.Lsh(1)
	}
	if exp >= mask128 {
		// overflow
		return Float128{sign | uvinf128[0], uvinf128[1]}
	}

	shift := shift128 + 3 // 1 for the implicit bit, 1 for the rounding bit, 1 for the guard bit
	fracA256 := fracA.Uint256().Lsh(uint(shift))
	fracB256 := fracB.Uint256()
	frac256, mod := fracA256.DivMod(fracB256)
	frac256[3] |= nonzero64(mod[0]) | nonzero64(mod[1]) | nonzero64(mod[2]) | nonzero64(mod[3])
	frac := frac256.Uint128()

	if exp <= 0 {
		// the result is subnormal
		shift := -exp + 3 + 1
		frac = roundToNearestEvenReference128(frac, uint(shift))
		return Float128{sign | frac[0], frac[1]}
	}

	// round-to-nearest-even (guard+round+sticky are in the low 3 bits)
	frac = roundToNearestEvenReference128(frac, uint(3))
	// detect carry-out caused by rounding
	if frac[0]&(1<<(shift128-64+1)) != 0 {
		frac = frac.Rsh(1)
		exp++
		if exp >= mask128 { // overflow -> ±Inf
			return Float128{sign | uvinf128[0], uvinf128[1]}
		}
	}
	return Float128{sign | uint64(exp)<<(shift128-64) | frac[0]&fracMask128[0], frac[1] & fracMask128[1]}
}

// quoReference256 is the reference implementation of Float256.Quo
// that uses the generic 512-bit division.
func quoReference256(a, b Float256) Float256 {
	if a.IsNaN() || b.IsNaN() {
		// a / NaN = NaN
		// NaN / b = NaN
		return Float256(uvnan256)
	}

	signA, expA, fracA := a.normalize()
	signB, expB, fracB := b.normalize()
	sign := signA ^ signB

	if b.IsZero() {
		if a.IsZero() {
			// 0 / 0 = NaN
			return Float256(uvnan256)
		}
		// ±finite / 0 = ±inf
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}
	if a.IsZero() {
		// 0 / finite = 0
		return Float256{sign, 0, 0, 0}
	}
	if expA == mask256-bias256 {
		// NaN check is done above; a is ±inf
		if expB == mask256-bias256 {
			// ±inf / ±inf = NaN
			return Float256(uvnan256)
		}
		// ±inf / finite = ±inf
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}
	if expB == mask256-bias256 {
		// NaN check is done above; b is ±inf
		// NaN and Inf checks are done above; a is finite.
		// ±finite / ±inf = 0
		return Float256{sign, 0, 0, 0}
	}

	exp := expA - expB + bias256
	if fracA.Cmp(fracB) < 0 {
		exp--
		fracA = fracA.Lsh(1)
	}
	if exp >= mask256 {
		// overflow
		return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
	}

	shift := shift256 + 3 // 1 for the implicit bit, 1 for the rounding bit, 1 for the guard bit
	fracA512 := fracA.Uint512().Lsh(uint(shift))
	fracB512 := fracB.Uint512()
	frac512, mod := fracA512.DivMod(fracB512)
	frac512[7] |= squash512(mod)
	frac := frac512.Uint256()

	if exp <= 0 {
		// the result is subnormal
		shift := -exp + 3 + 1
		frac = roundToNearestEvenReference256(frac, uint(shift))
		return Float256{
			sign | frac[0],
			frac[1],
			frac[2],
			frac[3],
		}
	}

	// round-to-nearest-even (guard+round+sticky are in the low 3 bits)
	frac = roundToNearestEvenReference256(frac, 3)
	// detect carry-out caused by rounding
	if frac.BitLen() > shift256+1 {
		frac = frac.Rsh(1)
		exp++
		if exp >= mask256 {
			// overflow
			return Float256{sign | uvinf256[0], uvinf256[1], uvinf256[2], uvinf256[3]}
		}
	}
	return Float256{
		sign | uint64(exp)<<(shift256-192) | frac[0]&fracMask256[0],
		frac[1],
		frac[2],
		frac[3],
	}
}

// randomWords returns random words that tend to have
// long runs of zeros and ones, to cover the corner cases of division.
func randomWords(r *rand.Rand, w []uint64) {
	for i := range w {
		switch r.IntN(4) {
		case 0:
			w[i] = 0
		case 1:
			w[i] = ^uint64(0)
		case 2:
			w[i] = ^uint64(0) << r.IntN(64)
		default:
			w[i] = r.Uint64()
		}
	}
}

func TestDiv3by2(t *testing.T) {
	r := rand.New(rand.NewPCG(17, 18))
	for range 2_000_000 {
		var d ints.Uint128
		randomWords(r, d[:])
		d[0] |= 1 << 63
		var u ints.Uint256
		randomWords(r, u[1:])
		if (ints.Uint128{u[1], u[2]}).Cmp(d) >= 0 {
			// make (u2, u1) less than d
			hi := ints.Uint128{u[1], u[2]}.Mod(d)
			u[1], u[2] = hi[0], hi[1]
		}
		q, r1, r0 := div3by2(u[1], u[2], u[3], d[0], d[1])
		wantQ, wantR := u.DivMod(ints.Uint256{0, 0, d[0], d[1]})
		if (ints.Uint256{0, 0, 0, q}) != wantQ || (ints.Uint256{0, 0, r1, r0}) != wantR {
			t.Fatalf("div3by2(%x, %x) = %x, (%x, %x), want %x, %x", u, d, q, r1, r0, wantQ, wantR)
		}

		// quo256by128 divides (u2, u1) << 128 by d
		gotQ, gotInexact := quo256by128(ints.Uint128{u[1], u[2]}, d)
		wantQ, wantR = ints.Uint256{u[1], u[2], 0, 0}.DivMod(ints.Uint256{0, 0, d[0], d[1]})
		if (ints.Uint256{0, 0, gotQ[0], gotQ[1]}) != wantQ || gotInexact != !wantR.IsZero() {
			t.Fatalf("quo256by128(%x, %x) = %x, %v, want %x, %v", u[1:3], d, gotQ, gotInexact, wantQ, !wantR.IsZero())
		}
	}
}

func TestDiv5by4(t *testing.T) {
	r := rand.New(rand.NewPCG(19, 20))
	for range 1_000_000 {
		var d ints.Uint256
		randomWords(r, d[:])
		d[0] |= 1 << 63
		var u ints.Uint512
		randomWords(r, u[3:])
		hi := ints.Uint256{u[3], u[4], u[5], u[6]}
		if hi.Cmp(d) >= 0 {
			// make u[3:7] less than d
			hi = hi.Mod(d)
			u[3], u[4], u[5], u[6] = hi[0], hi[1], hi[2], hi[3]
		}
		rem := hi
		q := div5by4(&rem, u[7], d)
		wantQ, wantR := u.DivMod(ints.Uint512{0, 0, 0, 0, d[0], d[1], d[2], d[3]})
		if (ints.Uint512{0, 0, 0, 0, 0, 0, 0, q}) != wantQ || (ints.Uint512{0, 0, 0, 0, rem[0], rem[1], rem[2], rem[3]}) != wantR {
			t.Fatalf("div5by4(%x, %x) = %x, %x, want %x, %x", u, d, q, rem, wantQ, wantR)
		}

		// quo512by256 divides hi << 256 by d
		gotQ, gotInexact := quo512by256(hi, d)
		wantQ, wantR = ints.Uint512{hi[0], hi[1], hi[2], hi[3], 0, 0, 0, 0}.DivMod(ints.Uint512{0, 0, 0, 0, d[0], d[1], d[2], d[3]})
		if (ints.Uint512{0, 0, 0, 0, gotQ[0], gotQ[1], gotQ[2], gotQ[3]}) != wantQ || gotInexact != !wantR.IsZero() {
			t.Fatalf("quo512by256(%x, %x) = %x, %v, want %x, %v", hi, d, gotQ, gotInexact, wantQ, !wantR.IsZero())
		}
	}
}

func TestFloat128_QuoRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(21, 22))
	for range 2_000_000 {
		a := randomFloat128(r)
		if r.IntN(2) == 0 {
			a = a.Neg()
		}
		// the quotient is near the subnormal / overflow ranges, or b divides a exactly
		b := randomQuoOperand128(r, a)
		if r.IntN(8) == 0 {
			b = a.Quo(b)
		}
		got := a.Quo(b)
		want := quoReference128(a, b)
		if got != want && !(got.IsNaN() && want.IsNaN()) {
			t.Fatalf("Float128(%x).Quo(%x) = %x, want %x", a, b, got, want)
		}
	}
}

func TestFloat256_QuoRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(23, 24))
	for range 1_000_000 {
		a := randomFloat256(r)
		if r.IntN(2) == 0 {
			a = a.Neg()
		}
		b := randomQuoOperand256(r, a)
		if r.IntN(8) == 0 {
			b = a.Quo(b)
		}
		got := a.Quo(b)
		want := quoReference256(a, b)
		if got != want && !(got.IsNaN() && want.IsNaN()) {
			t.Fatalf("Float256(%x).Quo(%x) = %x, want %x", a, b, got, want)
		}
	}
}

func randomQuoOperand128(r *rand.Rand, a Float128) Float128 {
	b := randomFloat128(r)
	switch r.IntN(4) {
	case 0:
		// the quotient is close to the subnormal range or the overflow threshold.
		// the biased exponent of the quotient is about ea - eb + bias.
		e := uint64(r.IntN(260) - 130)
		if r.IntN(2) == 0 {
			e += 2 * bias128
		}
		ea := a[0] >> (shift128 - 64) & mask128
		b[0] = b[0]&^(mask128<<(shift128-64)) | (ea-e+bias128)&mask128<<(shift128-64)
	case 1:
		// short fractions
		b[1] &^= 1<<r.IntN(64) - 1
	}
	if r.IntN(2) == 0 {
		b = b.Neg()
	}
	return b
}

func randomQuoOperand256(r *rand.Rand, a Float256) Float256 {
	b := randomFloat256(r)
	switch r.IntN(4) {
	case 0:
		// the quotient is close to the subnormal range or the overflow threshold.
		// the biased exponent of the quotient is about ea - eb + bias.
		e := uint64(r.IntN(520) - 260)
		if r.IntN(2) == 0 {
			e += 2 * bias256
		}
		ea := a[0] >> (shift256 - 192) & mask256
		b[0] = b[0]&^(mask256<<(shift256-192)) | (ea-e+bias256)&mask256<<(shift256-192)
	case 1:
		// short fractions
		b[2], b[3] = 0, 0
	}
	if r.IntN(2) == 0 {
		b = b.Neg()
	}
	return b
}
