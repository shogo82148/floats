package floats

import (
	"bufio"
	"math"
	"math/rand/v2"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/shogo82148/ints"
)

func TestFloat128_Lgamma(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float128
		want string
		sign int
	}{
		{exact128(0.5), "0.5723649429247000870717136756765293558236", 1},
		{exact128(2.5), "0.2846828704729191596324946696827019243201", 1},
		{exact128(-0.5), "1.265512123484645396488945797134705923899", -1},
		{exact128(-2.5), "-0.05624371649767405067259453009765428412294", -1},
		{exact128(100), "359.1342053695753987760440104602869096126", 1},

		// around MaxStirling = 1756, straddling the Gamma()-delegate /
		// log-Stirling boundary
		{exact128(1755), "11352.42723245307564687249568604776301785", 1},
		{exact128(1755.5), "11356.16227329595536842840606592791090684", 1},
		{exact128(1756), "11359.89745658897561305733566853332210121", 1},
		{exact128(1756.5), "11363.63278229156385536179632102015464708", 1},
		{exact128(5000), "37582.6263156853503317465661476968580828", 1},
		{exact128(-1755.5), "-11362.48805240571445518765289366880158837", 1},
		{exact128(-1756.5), "-11369.95913087742033691146413996420582514", -1},
		{exact128(-5000.5), "-37594.25745058162566268740632072478674085", -1},
	}

	for _, tt := range tests {
		got, sign := tt.x.Lgamma()
		if !close128(got, tt.want) || sign != tt.sign {
			t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", tt.x, got, sign, tt.want, tt.sign)
		}
	}

	strictTests := []struct {
		x        Float128
		want     Float128
		wantSign int
	}{
		// special cases
		{exact128(math.Inf(1)), exact128(math.Inf(1)), 1},
		{exact128(math.Inf(-1)), exact128(math.Inf(-1)), 1},
		{exact128(0), exact128(math.Inf(1)), 1},
		{exact128(math.Copysign(0, -1)), exact128(math.Inf(1)), 1},
		{exact128(-1), exact128(math.Inf(1)), 1},
		{exact128(-2), exact128(math.Inf(1)), 1},
		{exact128(math.NaN()), exact128(math.NaN()), 1},
	}

	for _, tt := range strictTests {
		got, sign := tt.x.Lgamma()
		if !eq128(got, tt.want) || sign != tt.wantSign {
			t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", tt.x, got, sign, tt.want, tt.wantSign)
		}
	}
}

// TestFloat128_LgammaAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat128_LgammaAccuracy(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/lgamma128.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	parse := func(s string) Float128 {
		var x Float128
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
		if len(fields) != 3 || len(fields[0]) != 32 || len(fields[1]) != 32 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		x, want := parse(fields[0]), parse(fields[1])
		wantSign, err := strconv.Atoi(fields[2])
		if err != nil {
			t.Fatal(err)
		}
		if got, sign := x.Lgamma(); !eq128(got, want) || sign != wantSign {
			t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", x, got, sign, want, wantSign)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

// TestFloat128_LgammaKernels compares the fixed point kernels of Lgamma with Float256.Lgamma,
// whose errors are hidden by the rounding to Float128.
func TestFloat128_LgammaKernels(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(5, 6))
	bound := NewFloat256(0x1p-143)
	check := func(x Float128, got Float256) {
		t.Helper()
		want, _ := x.Float256().Lgamma()
		if want.Abs().Lt(NewFloat256(0x1p-6)) {
			return // the kernels are not used if the result is close to zero.
		}
		if d := got.Sub(want).Abs(); d.Gt(bound) {
			t.Errorf("the kernel of Lgamma(%v) = %v; want %v", x, got, want)
		}
	}
	fix := func(v lgammaFix) Float256 {
		if v.v == (ints.Uint256{}) {
			return Float256{}
		}
		var sign uint64
		if v.neg {
			sign = signMask256[0]
		}
		return fixToFloat256(sign, v.v.Uint512(), false, -192)
	}
	random := func(lo, hi float64) Float128 {
		x := NewFloat128(lo + (hi-lo)*r.Float64())
		x[1] ^= r.Uint64() >> 20 // random low bits
		return x
	}

	// the ranges that include the buckets where log is not accurate, i.e. [2**k, 2**k (1 + 2**-8)).
	for _, rg := range [][2]float64{
		{0.001, 2}, {2, 24}, {24, 2000}, {16, 16.07}, {31.9, 32.2}, {63.5, 64.5}, {127.5, 128.5}, {1023, 1030}, {0.49, 0.51},
	} {
		for range 3000 {
			x := random(rg[0], rg[1])
			_, exp, m := x.normalize()
			check(x, fix(lgamma128Pos(m.Uint256().Lsh(uint(80+exp)))))
		}
	}
	for _, rg := range [][2]float64{{-0.99, -0.01}, {-30, -1}, {-2000, -30}, {-2.5, -2.4}} {
		for range 3000 {
			x := random(rg[0], rg[1])
			if isNegInt128(x) {
				continue
			}
			_, exp, m := x.normalize()
			_, v := lgamma128Neg(exp, m)
			check(x, fix(v))
		}
	}
	for range 3000 {
		// tiny arguments, 2**-100 <= |x| < 2**-40
		x := NewFloat128(math.Ldexp(1+r.Float64(), -100+r.IntN(60)))
		x[1] ^= r.Uint64() >> 20
		if r.IntN(2) == 0 {
			x = x.Neg()
		}
		_, exp, m := x.normalize()
		check(x, fix(lgamma128Tiny(x.Signbit(), exp, m)))
	}

	// Lgamma for 2**15 <= x is calculated in Float256.
	for range 2000 {
		x := NewFloat128(math.Exp2(15 + r.Float64()*50))
		x[1] ^= r.Uint64() >> 20
		want, _ := x.Float256().Lgamma()
		got := lgamma256Stirling(x.Float256())
		if d := got.Sub(want).Abs(); d.Gt(want.Mul(NewFloat256(0x1p-150))) {
			t.Errorf("lgamma256Stirling(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat128_Lgamma(b *testing.B) {
	benchFloat128Lgamma(b)
}

func benchFloat128Lgamma(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float128
	}{
		{"tiny", exact128(1e-300)},          // Lgamma(x) ~ -log(x)
		{"small", exact128(0.37)},           // 0 < x < 1
		{"medium", exact128(1.5)},           // 1 <= x < 3
		{"large", exact128(10.3)},           // 3 <= x < 24
		{"stirling", exact128(100.7)},       // 24 <= x < 2048
		{"big", exact128(5000.5)},           // 2048 <= x < 2**15
		{"huge", exact128(1e10)},            // 2**15 <= x
		{"negative", exact128(-2.5)},        // -2048 < x < 0
		{"verynegative", exact128(-100.5)},  // reflection with large |x|
		{"hugenegative", exact128(-5000.5)}, // -2**15 < x <= -2048
		{"nearzero", exact128(1.0625)},      // the result is close to zero
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				v, s := tt.x.Lgamma()
				runtime.KeepAlive(v)
				runtime.KeepAlive(s)
			}
		})
	}
}
