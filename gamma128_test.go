package floats

import (
	"bufio"
	"math"
	"math/rand/v2"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/shogo82148/ints"
)

func TestFloat128_Gamma(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(-0.5), "-3.544907701811032054596334966682290365595098912244774256427615579705822569182064363"},
		{exact128(0.25), "3.625609908221908311930685155867672002995167682880065467433377999569919243538729122"},
		{exact128(0.5), "1.772453850905516027298167483341145182797549456122387128213807789852911284591032181"},
		{exact128(0.75), "1.225416702465177645129098303362890526851239248108070611230118938289822888426798357"},
		{exact128(1), "1"},
		{exact128(1.25), "0.9064024770554770779826712889669180007487919207200163668583444998924798108846822804"},
		{exact128(1.5), "0.8862269254527580136490837416705725913987747280611935641069038949264556422955160907"},
		{exact128(1.75), "0.9190625268488832338468237275221678951384294360810529584225892037173671663200987679"},
		{exact128(2), "1"},
		{exact128(2.5), "1.329340388179137020473625612505858887098162092091790346160355842389683463443274136"},
		{exact128(3), "2"},
		{exact128(53), "80658175170943878571660636856403766975289505440883277824000000000000"},
		{exact128(55), "230843697339241380472092742683027581083278564571807941132288000000000000"},
		{exact128(60), "1.3868311854568983573793901972038940634590287677268743254082129494016E+80"},
		{exact128(100), "9.3326215443944152681699238856266700490715968264381621468592963895217599993E+155"},
		{exact128(1755), "1.979261890105010055381794327532605804610806878373860932441926508881930448E+4930"},

		{exact128(-55.125), "3.9131911034762969245238588487381041440846027762491865228553788345525176E-73"},
	}

	for _, tt := range tests {
		got := tt.x.Gamma()
		if !close128(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// overflow and underflow
		{exact128(1756), exact128(math.Inf(1))},
		{exact128(2048), exact128(math.Inf(1))},
		{exact128(0x1p100), exact128(math.Inf(1))},
		{exact128(-2048.5), exact128(math.Copysign(0, -1))},
		{exact128(-2049.5), exact128(0)},
		{exact128(-0x1.8p100), exact128(math.NaN())},                // an integer
		{Float128{0x0000_0000_0000_0000, 1}, exact128(math.Inf(1))}, // the smallest subnormal
		{Float128{0x8000_0000_0000_0000, 1}, exact128(math.Inf(-1))},

		// special cases
		{exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(0), exact128(math.Inf(1))},
		{exact128(math.Copysign(0, -1)), exact128(math.Inf(-1))},
		{exact128(-1), exact128(math.NaN())},
		{exact128(-2), exact128(math.NaN())},
		{exact128(math.Inf(-1)), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Gamma()
		if !eq128(got, tt.want) {
			t.Errorf("Gamma(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat128_GammaAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat128_GammaAccuracy(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/gamma128.txt")
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
		x, want := parse(fields[0]), parse(fields[1])
		if got := x.Gamma(); !eq128(got, want) {
			t.Errorf("Gamma(%v) = %v; want %v", x, got, want)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

// gammaFixToFloat256 converts v in fixed point with 192 fractional bits to Float256 exactly.
func gammaFixToFloat256(v ints.Uint256) Float256 {
	return fixToFloat256(0, v.Uint512(), false, -192)
}

// TestFloat128_GammaKernels compares the kernels of Gamma, whose errors are hidden by the rounding to Float128,
// with Float256.
func TestFloat128_GammaKernels(t *testing.T) {
	t.Parallel()
	pi := Float256{
		0x4000_0921_fb54_442d, 0x1846_9898_cc51_701b,
		0x839a_2520_49c1_114c, 0xf98e_8041_77d4_c762,
	}
	bound := exact256(0x1p-144)
	rnd := rand.New(rand.NewPCG(1, 2))
	for i := range 1000 {
		// e**s = 2**k × m
		s := ints.Uint256{uint64(rnd.IntN(12000)), rnd.Uint64(), rnd.Uint64(), rnd.Uint64()}
		m, k := gammaExp192(s)
		got := fixToFloat256(0, m.Uint512(), false, k-192)
		want := gammaFixToFloat256(s).Exp()
		if got.Sub(want).Abs().Quo(want).Gt(bound) {
			t.Fatalf("exp(%v) = %v; want %v", gammaFixToFloat256(s), got, want)
		}

		// log(Gamma(z)), whose absolute error is less than 2**-144
		z := ints.Uint256{uint64(24 + rnd.IntN(2000)), rnd.Uint64(), rnd.Uint64(), rnd.Uint64()}
		if !gammaLogBucket0(z) {
			got := gammaFixToFloat256(gammaLogGamma192(z))
			want, _ := gammaFixToFloat256(z).Lgamma()
			if got.Sub(want).Abs().Gt(bound) {
				t.Fatalf("lgamma(%v) = %v; want %v", gammaFixToFloat256(z), got, want)
			}
		}

		// sin(pi r)/(pi r) for 0 < r < 1/2
		r := ints.Uint256{0, rnd.Uint64() &^ (1 << 63), rnd.Uint64(), rnd.Uint64()}
		if i%3 == 0 {
			r = r.Rsh(uint(rnd.IntN(120)))
		}
		if !r.IsZero() {
			x := pi.Mul(gammaFixToFloat256(r))
			got := gammaFixToFloat256(gammaSinc192(r))
			want := x.Sin().Quo(x)
			if got.Sub(want).Abs().Quo(want).Gt(bound) {
				t.Fatalf("sinc(%v) = %v; want %v", gammaFixToFloat256(r), got, want)
			}
		}
	}
}

func BenchmarkFloat128_Gamma(b *testing.B) {
	benchFloat128(b, Float128.Gamma, []struct {
		name string
		x    Float128
	}{
		{"tiny", exact128(1e-300)},     // Gamma(x) ~ 1/x
		{"small", exact128(0.37)},      // 0 < x < 1
		{"medium", exact128(1.5)},      // 1 <= x < 3
		{"recurrence", exact128(10.3)}, // 3 <= x < 55
		{"stirling", exact128(100.7)},  // 55 <= x
		{"huge", exact128(1700.5)},     // close to the overflow
		{"negative", exact128(-2.5)},   // reflection
		{"negative-large", exact128(-100.5)},
		{"overflow", exact128(2000.5)}, // +Inf
	})
}
