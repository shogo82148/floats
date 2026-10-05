package floats

import (
	"math"
	"math/big"
	"math/rand/v2"
	"testing"
)

func TestFloat128_Cbrt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(3375), "15"},
		{exact128(27), "3"},
		{exact128(8), "2"},
		{exact128(2), "1.259921049894873164767210607278228350570251464701507980081975112155299676513959484"},
		{exact128(1), "1"},
		{exact128(0.5), "0.7937005259840997373758528196361541301957466639499265049041428809126082528121095866"},
		{exact128(0), "0"},
		{exact128(-0.5), "-0.7937005259840997373758528196361541301957466639499265049041428809126082528121095866"},
		{exact128(-1), "-1"},
		{exact128(-2), "-1.259921049894873164767210607278228350570251464701507980081975112155299676513959484"},
		{exact128(-8), "-2"},
		{exact128(-27), "-3"},
		{exact128(-3375), "-15"},
	}

	for _, test := range tests {
		got := test.x.Cbrt()
		if !close128(got, test.want) {
			t.Errorf("Cbrt(%v) = %v; want %v", test.x, got, test.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// Special cases
		{exact128(0), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},
		{exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(math.Inf(-1)), exact128(math.Inf(-1))},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Cbrt()
		if !eq128(got, tt.want) {
			t.Errorf("Float128.Cbrt(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat128_CbrtAccuracy(t *testing.T) {
	t.Parallel()
	testFloat128Accuracy(t, "testdata/cbrt128.txt", "Cbrt", Float128.Cbrt)
}

// cbrt128Exact returns the correctly rounded cube root of a, which must be finite and not zero, with math/big.
// frac is the distance between the exact cube root and the nearest midpoint in the unit of 2**-40 ulp.
func cbrt128Exact(a Float128) (r Float128, frac int64) {
	sign, exp, m := a.normalize()
	q, rem := exp/3, exp%3
	if rem < 0 {
		q, rem = q-1, rem+3
	}
	// cbrt(a) = cbrt(n) × 2**(q-112), where n = m × 2**(rem+224)
	n := new(big.Int).SetUint64(m[0])
	n.Lsh(n, 64).Or(n, new(big.Int).SetUint64(m[1]))
	n.Lsh(n, uint(rem+224))
	icbrt := func(n *big.Int) *big.Int {
		x := new(big.Int).Lsh(big.NewInt(1), uint((n.BitLen()+2)/3))
		for {
			y := new(big.Int).Quo(n, new(big.Int).Mul(x, x))
			y.Add(y, new(big.Int).Lsh(x, 1)).Quo(y, big.NewInt(3))
			if y.Cmp(x) >= 0 {
				return x
			}
			x = y
		}
	}
	root := icbrt(n)

	// the position of the cube root in [root, root+1) in the unit of 2**-40
	g := icbrt(new(big.Int).Lsh(n, 120))
	f := new(big.Int).And(g, new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 40), big.NewInt(1))).Int64()
	frac = f - 1<<39
	if frac < 0 {
		frac = -frac
	}

	// round to nearest: the cube root is above root + 1/2 if 8n > (2 root + 1)³.
	t := new(big.Int).Lsh(root, 1)
	t.Add(t, big.NewInt(1))
	t.Mul(t, new(big.Int).Mul(t, t))
	if new(big.Int).Lsh(n, 3).Cmp(t) > 0 {
		root.Add(root, big.NewInt(1))
	}

	// the significand in [2**112, 2**113]. 2**113 carries to the exponent.
	hi := new(big.Int).Rsh(root, 64).Uint64()
	lo := new(big.Int).And(root, new(big.Int).SetUint64(math.MaxUint64)).Uint64()
	h := sign | uint64(q+bias128)<<(shift128-64)
	h += hi - 1<<(shift128-64)
	return Float128{h, lo}, frac
}

// TestFloat128_CbrtExact compares Cbrt with the exact cube root for random numbers.
func TestFloat128_CbrtExact(t *testing.T) {
	t.Parallel()
	rnd := rand.New(rand.NewPCG(1, 2))
	var hard int
	for i := range 100000 {
		var a Float128
		switch i % 4 {
		case 0, 1: // normal numbers
			a = Float128{rnd.Uint64() & 0x7fff_ffff_ffff_ffff, rnd.Uint64()}
			if a[0]>>48 == 0x7fff {
				a[0] &^= 0x4000_0000_0000_0000
			}
		case 2: // close to 1
			a = Float128{0x3fff_0000_0000_0000 | rnd.Uint64()&0xffff_ffff, rnd.Uint64()}
		default: // subnormal numbers
			a = Float128{rnd.Uint64() >> (16 + rnd.IntN(48)), rnd.Uint64() >> rnd.IntN(64)}
			if a.IsZero() {
				continue
			}
		}
		if i%2 == 1 {
			a[0] |= 1 << 63
		}
		got := a.Cbrt()
		want, frac := cbrt128Exact(a)
		if !eq128(got, want) {
			t.Fatalf("Cbrt(%v) = %v; want %v", a, got, want)
		}
		if frac < 1<<(40-12) {
			hard++ // the case that needs the exact comparison
		}
	}
	if hard == 0 {
		t.Error("no hard case is tested")
	}
	t.Logf("%d hard cases", hard)
}

func BenchmarkFloat128_Cbrt(b *testing.B) {
	benchFloat128(b, Float128.Cbrt, []struct {
		name string
		x    Float128
	}{
		{"r0", exact128(1.5)},                    // exponent mod 3 = 0
		{"r1", exact128(2.5)},                    // exponent mod 3 = 1
		{"r2", exact128(6.5)},                    // exponent mod 3 = 2
		{"negative", exact128(-10.3)},            // cbrt(-x) = -cbrt(x)
		{"large", exact128(1e300)},               //
		{"small", exact128(1e-300)},              //
		{"subnormal", Float128{0, 0x1234567890}}, // subnormal
	})
}
