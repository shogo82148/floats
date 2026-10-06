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

func TestFloat256_Gamma(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-0.5), "-3.5449077018110320545963349666822903655950989122447742564276155797058225691820644"},
		{exact256(0.25), "3.6256099082219083119306851558676720029951676828800654674333779995699192435387291"},
		{exact256(0.5), "1.7724538509055160272981674833411451827975494561223871282138077898529112845910322"},
		{exact256(0.75), "1.2254167024651776451290983033628905268512392481080706112301189382898228884267984"},
		{exact256(1), "1"},
		{exact256(1.25), "0.90640247705547707798267128896691800074879192072001636685834449989247981088468228"},
		{exact256(1.5), "0.88622692545275801364908374167057259139877472806119356410690389492645564229551609"},
		{exact256(1.75), "0.91906252684888323384682372752216789513842943608105295842258920371736716632009877"},
		{exact256(2), "1"},
		{exact256(2.5), "1.3293403881791370204736256125058588870981620920917903461603558423896834634432741"},
		{exact256(3), "2"},
		{exact256(53), "80658175170943878571660636856403766975289505440883277824000000000000"},
		{exact256(55), "230843697339241380472092742683027581083278564571807941132288000000000000"},
		{exact256(60), "1.3868311854568983573793901972038940634590287677268743254082129494016E+80"},
		{exact256(100), "9.3326215443944152681699238856266700490715968264381621468592963895217599993229916E+155"},
		{exact256(20367), "7.6986646535903249422736371129778777982736925914669264235603651433188315760456764E+78912"},

		{exact256(-150.125), "-7.6771866741547396417630494833826918584179615134189863600663179783589599438008944E-263"},
	}

	for _, tt := range tests {
		got := tt.x.Gamma()
		if !close256(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// overflow
		{exact256(20368), exact256(math.Inf(1))},
		{exact256(32768), exact256(math.Inf(1))},
		{exact256(1e6), exact256(math.Inf(1))},

		// underflow: the sign is that of sin(pi x)
		{exact256(-32768.5), exact256(math.Copysign(0, -1))},
		{exact256(-32769.5), exact256(0)},
		{exact256(-1048576.5), exact256(math.Copysign(0, -1))},
		{exact256(-1048577.5), exact256(0)},
		{exact256(-1e30), exact256(math.NaN())},

		// special cases
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(0), exact256(math.Inf(1))},
		{exact256(math.Copysign(0, -1)), exact256(math.Inf(-1))},
		{exact256(-1), exact256(math.NaN())},
		{exact256(-2), exact256(math.NaN())},
		{exact256(math.Inf(-1)), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Gamma()
		if !eq256(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat256_GammaAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat256_GammaAccuracy(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/gamma256.txt")
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
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 2 || len(fields[0]) != 64 || len(fields[1]) != 64 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		x, want := parse(fields[0]), parse(fields[1])
		if got := x.Gamma(); !eq256(got, want) {
			t.Errorf("Gamma(%v) = %v; want %v", x, got, want)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

// TestFloat256_GammaKernels checks the accuracy of the kernels of Gamma with the reference values calculated by mpmath.
// The kernels are more accurate than Float256, so that their errors are hidden by the rounding.
func TestFloat256_GammaKernels(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/gamma256_kernels.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	parse := func(s string) gammaFix256 {
		if len(s) != 96 {
			t.Fatalf("malformed number: %q", s)
		}
		var x gammaFix256
		for i := range x {
			v, err := strconv.ParseUint(s[16*i:16*(i+1)], 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			x[i] = v
		}
		return x
	}
	// the maximum errors in bits below the binary point.
	bound := map[string]int{
		"log":    312,
		"lgamma": 286,
		"exp":    313,
		"sinc":   293,
		"recip":  314,
		"gamma":  286,
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		kind := fields[0]
		if len(fields) < 3 || bound[kind] == 0 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		x, want := parse(fields[1]), parse(fields[2])
		var got gammaFix256
		switch kind {
		case "log":
			got = gammaLog256(x)
		case "lgamma":
			got = gammaLogGamma256(x)
		case "exp", "gamma":
			if len(fields) != 4 {
				t.Fatalf("malformed line: %q", sc.Text())
			}
			e, err := strconv.Atoi(fields[3])
			if err != nil {
				t.Fatal(err)
			}
			var ge int
			if kind == "exp" {
				got, ge = gammaExp256(x)
			} else {
				got, ge = gammaPos256(x)
			}
			if ge != e {
				t.Errorf("%s(%v): the exponent is %d; want %d", kind, fields[1], ge, e)
				continue
			}
		case "sinc":
			got = gammaSinc256(x)
		case "recip":
			got = gammaRecip256(x)
		}
		diff := got.sub(want)
		if got.cmp(want) < 0 {
			diff = want.sub(got)
		}
		if diff.bitLen() > 320-bound[kind] {
			t.Errorf("%s(%v): the error is 2**%d; want less than 2**-%d", kind, fields[1], diff.bitLen()-320, bound[kind])
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

func gammaFix256ToBig(x gammaFix256) *big.Int {
	v := new(big.Int)
	for _, w := range x {
		v.Lsh(v, 64)
		v.Or(v, new(big.Int).SetUint64(w))
	}
	return v
}

// TestFloat256_GammaFixArithmetic compares the fixed-point arithmetic of Gamma with math/big.
func TestFloat256_GammaFixArithmetic(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(3, 4))
	mod := new(big.Int).Lsh(big.NewInt(1), 384)
	const ones = ^uint64(0)
	// the divisors of 2**64-1. k × (2**64-1)/k = 2**64-1, which makes carries.
	divisors := []uint64{3, 5, 15, 17, 51, 85, 255, 257, 771, 1285}
	randFix := func(small bool) gammaFix256 {
		var x gammaFix256
		for i := range x {
			switch r.IntN(5) {
			case 0:
				x[i] = ones
			case 1:
				x[i] = 0
			case 2:
				x[i] = ones / divisors[r.IntN(len(divisors))]
			default:
				x[i] = r.Uint64()
			}
		}
		if small {
			x[0] = r.Uint64() >> 33
		}
		return x
	}

	for range 3000 {
		a, b := randFix(false), randFix(false)
		A, B := gammaFix256ToBig(a), gammaFix256ToBig(b)

		want := new(big.Int).Add(A, B)
		want.Mod(want, mod)
		if got := gammaFix256ToBig(a.add(b)); got.Cmp(want) != 0 {
			t.Fatalf("%v + %v = %v; want %v", a, b, got, want)
		}
		want = new(big.Int).Sub(A, B)
		want.Mod(want, mod)
		if got := gammaFix256ToBig(a.sub(b)); got.Cmp(want) != 0 {
			t.Fatalf("%v - %v = %v; want %v", a, b, got, want)
		}
		if got, want := a.cmp(b), A.Cmp(B); got != want {
			t.Fatalf("cmp(%v, %v) = %d; want %d", a, b, got, want)
		}

		k := uint64(r.IntN(1 << 10))
		if r.IntN(2) == 0 {
			k = divisors[r.IntN(len(divisors))]
		}
		want = new(big.Int).Mul(A, new(big.Int).SetUint64(k))
		want.Mod(want, mod)
		if got := gammaFix256ToBig(a.mulUint(k)); got.Cmp(want) != 0 {
			t.Fatalf("%v * %d = %v; want %v", a, k, got, want)
		}

		s := uint(r.IntN(384))
		want = new(big.Int).Lsh(A, s)
		want.Mod(want, mod)
		if got := gammaFix256ToBig(a.shl(s)); got.Cmp(want) != 0 {
			t.Fatalf("%v << %d = %v; want %v", a, s, got, want)
		}
		want = new(big.Int).Rsh(A, s)
		if got := gammaFix256ToBig(a.shr(s)); got.Cmp(want) != 0 {
			t.Fatalf("%v >> %d = %v; want %v", a, s, got, want)
		}
		if got, want := a.bitLen(), A.BitLen(); got != want {
			t.Fatalf("bitLen(%v) = %d; want %d", a, got, want)
		}

		// the multiplication with the upper n words. The error is less than a few units of the last place.
		a, b = randFix(true), randFix(true)
		for n := 2; n <= 6; n++ {
			p := gammaMul256(&a, &b, n)
			an, bn := new(big.Int), new(big.Int)
			for i := range n {
				an.Lsh(an, 64).Or(an, new(big.Int).SetUint64(a[i]))
				bn.Lsh(bn, 64).Or(bn, new(big.Int).SetUint64(b[i]))
			}
			want := new(big.Int).Mul(an, bn)
			want.Rsh(want, uint(64*(n-1)))
			got := new(big.Int)
			for i := range n {
				got.Lsh(got, 64).Or(got, new(big.Int).SetUint64(p[i]))
			}
			for i := n; i < 6; i++ {
				if p[i] != 0 {
					t.Fatalf("gammaMul256(%v, %v, %d) = %v; the lower words must be zero", a, b, n, p)
				}
			}
			diff := new(big.Int).Sub(want, got)
			if diff.Sign() < 0 || diff.Cmp(big.NewInt(8)) > 0 {
				t.Fatalf("gammaMul256(%v, %v, %d) = %v; want %v", a, b, n, got, want)
			}
		}

		// n × c for the 7 words constant c.
		var c [7]uint64
		for i := range c {
			switch r.IntN(3) {
			case 0:
				c[i] = ones
			case 1:
				c[i] = ones / divisors[r.IntN(len(divisors))]
			default:
				c[i] = r.Uint64()
			}
		}
		c[0] = 0
		n := r.Uint64() >> 36
		if r.IntN(2) == 0 {
			n = divisors[r.IntN(len(divisors))]
		}
		cb := new(big.Int)
		for _, w := range c {
			cb.Lsh(cb, 64).Or(cb, new(big.Int).SetUint64(w))
		}
		want = new(big.Int).Mul(cb, new(big.Int).SetUint64(n))
		want.Rsh(want, 64)
		if got := gammaFix256ToBig(gammaMulLn2(&c, n)); got.Cmp(want) != 0 {
			t.Fatalf("gammaMulLn2(%v, %d) = %v; want %v", c, n, got, want)
		}
	}
}

// TestFloat256_GammaExpBoundary checks gammaExp256 for the arguments just below n log(2)/256,
// where the rounding error of the estimate of n makes it larger than the exact value by one.
func TestFloat256_GammaExpBoundary(t *testing.T) {
	t.Parallel()
	for n := uint64(1 << 26); n > 1<<26-5000; n-- {
		if n%256 == 0 {
			continue // the result is close to a power of two, where the exponent is not stable.
		}
		nl2 := gammaMulLn2(&gamma256Ln2By256, n)
		m0, e0 := gammaExp256(nl2)
		for d := uint64(1); d <= 64; d++ {
			// e**(x-d) = e**x (1 - d 2**-320 + ...)
			m, e := gammaExp256(nl2.sub(gammaFix256{5: d}))
			if e != e0 {
				t.Fatalf("exp(%d × log(2)/256 - %d ulp): the exponent is %d; want %d", n, d, e, e0)
			}
			diff := m0.sub(m)
			if m.cmp(m0) >= 0 {
				diff = m.sub(m0)
			}
			if diff.bitLen() > 8 {
				t.Fatalf("exp(%d × log(2)/256 - %d ulp): the mantissa is %v; want about %v", n, d, m, m0)
			}
		}
	}
}

func BenchmarkFloat256_Gamma(b *testing.B) {
	benchFloat256(b, Float256.Gamma, []struct {
		name string
		x    Float256
	}{
		{"tiny", exact256(1e-300)},     // Gamma(x) ~ 1/x
		{"small", exact256(0.37)},      // 0 < x < 1
		{"medium", exact256(1.5)},      // 1 <= x < 3
		{"recurrence", exact256(10.3)}, // 3 <= x < the Stirling threshold
		{"stirling", exact256(100.7)},  // the Stirling threshold <= x
		{"huge", exact256(20300.5)},    // close to the overflow
		{"negative", exact256(-2.5)},   // reflection
		{"negative-large", exact256(-100.5)},
		{"overflow", exact256(30000.5)}, // +Inf
	})
}
