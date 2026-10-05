package floats

import (
	"bufio"
	"math"
	"math/big"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestFloat256_Cbrt(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(3375), "15"},
		{exact256(27), "3"},
		{exact256(8), "2"},
		{exact256(2), "1.259921049894873164767210607278228350570251464701507980081975112155299676513959484"},
		{exact256(1), "1"},
		{exact256(0.5), "0.7937005259840997373758528196361541301957466639499265049041428809126082528121095866"},
		{exact256(0), "0"},
		{exact256(-0.5), "-0.7937005259840997373758528196361541301957466639499265049041428809126082528121095866"},
		{exact256(-1), "-1"},
		{exact256(-2), "-1.259921049894873164767210607278228350570251464701507980081975112155299676513959484"},
		{exact256(-8), "-2"},
		{exact256(-27), "-3"},
		{exact256(-3375), "-15"},
	}

	for _, test := range tests {
		got := test.x.Cbrt()
		if !close256(got, test.want) {
			t.Errorf("Cbrt(%v) = %v; want %v", test.x, got, test.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// Special cases
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(math.Inf(-1)), exact256(math.Inf(-1))},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Cbrt()
		if !eq256(got, tt.want) {
			t.Errorf("Float256.Cbrt(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat256_CbrtAccuracy(t *testing.T) {
	t.Parallel()
	testFloat256Accuracy(t, "testdata/cbrt256.txt", "Cbrt", Float256.Cbrt)
}

// TestFloat256_CbrtTestdata requires the correctly rounded result for every vector of the test data,
// while testFloat256Accuracy tolerates 1% of the results that are not correctly rounded.
func TestFloat256_CbrtTestdata(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/cbrt256.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	parse := func(s string) Float256 {
		var x Float256
		for i := range x {
			v, err := strconv.ParseUint(s[16*i:16*(i+1)], 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			x[i] = v
		}
		return x
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		x, want := parse(fields[0]), parse(fields[1])
		if got := x.Cbrt(); !eq256(got, want) {
			t.Errorf("Cbrt(%v) = %v; want %v", x, got, want)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

// cbrt256Exact returns the correctly rounded cube root of a, which must be finite and not zero, with math/big.
// frac is the distance between the exact cube root and the nearest midpoint in the unit of 2**-40 ulp.
func cbrt256Exact(a Float256) (r Float256, frac int64) {
	sign, exp, m := a.normalize()
	q, rem := exp/3, exp%3
	if rem < 0 {
		q, rem = q-1, rem+3
	}
	// cbrt(a) = cbrt(n) × 2**(q-236), where n = m × 2**(rem+472)
	n := new(big.Int)
	for _, v := range m {
		n.Lsh(n, 64).Or(n, new(big.Int).SetUint64(v))
	}
	n.Lsh(n, uint(rem+472))
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

	// the significand in [2**236, 2**237]. 2**237 carries to the exponent.
	var sig [4]uint64
	mask := new(big.Int).SetUint64(math.MaxUint64)
	for i := 3; i >= 0; i-- {
		sig[i] = new(big.Int).And(root, mask).Uint64()
		root.Rsh(root, 64)
	}
	h := sign | uint64(q+bias256)<<(shift256-192)
	h += sig[0] - 1<<(shift256-192)
	return Float256{h, sig[1], sig[2], sig[3]}, frac
}

// TestFloat256_CbrtExact compares Cbrt with the exact cube root for random numbers.
func TestFloat256_CbrtExact(t *testing.T) {
	t.Parallel()
	rnd := rand.New(rand.NewPCG(1, 2))
	var hard int
	for i := range 50000 {
		var a Float256
		switch i % 4 {
		case 0, 1: // normal numbers
			a = Float256{rnd.Uint64() & 0x7fff_ffff_ffff_ffff, rnd.Uint64(), rnd.Uint64(), rnd.Uint64()}
			if a[0]>>44 == 0x7ffff {
				a[0] &^= 0x4000_0000_0000_0000
			}
		case 2: // close to 1
			a = Float256{0x3fff_f000_0000_0000 | rnd.Uint64()&0xffff_ffff, rnd.Uint64(), rnd.Uint64(), rnd.Uint64()}
		default: // subnormal numbers
			a = Float256{rnd.Uint64() >> (20 + rnd.IntN(44)), rnd.Uint64() >> rnd.IntN(64), rnd.Uint64() >> rnd.IntN(64), rnd.Uint64()}
			if a.IsZero() {
				continue
			}
		}
		if i%2 == 1 {
			a[0] |= 1 << 63
		}
		got := a.Cbrt()
		want, frac := cbrt256Exact(a)
		if !eq256(got, want) {
			t.Fatalf("Cbrt(%v) = %v; want %v", a, got, want)
		}
		if frac < 1<<(40-12) {
			hard++
		}
	}
	t.Logf("%d cases close to a midpoint", hard)
}

func BenchmarkFloat256_Cbrt(b *testing.B) {
	benchFloat256(b, Float256.Cbrt, []struct {
		name string
		x    Float256
	}{
		{"r0", exact256(1.5)},                          // exponent mod 3 = 0
		{"r1", exact256(2.5)},                          // exponent mod 3 = 1
		{"r2", exact256(6.5)},                          // exponent mod 3 = 2
		{"negative", exact256(-10.3)},                  // cbrt(-x) = -cbrt(x)
		{"large", exact256(1e300)},                     //
		{"small", exact256(1e-300)},                    //
		{"subnormal", Float256{0, 0, 0, 0x1234567890}}, // subnormal
	})
}
