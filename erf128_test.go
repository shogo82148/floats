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

func TestFloat128_Erf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(-1), "-0.8427007929497148693412206350826092592960669979663029084599378978347172540960108413"},
		{exact128(-0.5), "-0.5204998778130465376827466538919645287364515757579637000588057256471935217168535709"},
		{exact128(0), "0"},
		{exact128(0.5), "0.5204998778130465376827466538919645287364515757579637000588057256471935217168535709"},
		{exact128(1), "0.8427007929497148693412206350826092592960669979663029084599378978347172540960108413"},
		{exact128(2), "0.9953222650189527341620692563672529286108917970400600767383523262004372807199951774"},
		{exact128(3), "0.9999779095030014145586272238704176796201522929126007503427610451570575433163798677"},
		{exact128(4), "0.9999999845827420997199811478403265131159514278547464108088316570950057869589731887"},
		{exact128(11), "0.9999999999999999999999999999999999999999999999999999985591338620563053196601902971"},
	}

	for _, tt := range tests {
		got := tt.x.Erf()
		if !close128(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(math.Inf(1)), exact128(1)},
		{exact128(math.Inf(-1)), exact128(-1)},
		{exact128(math.NaN()), exact128(math.NaN())},
		{exact128(0), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},
		{Float128{0, 1}, Float128{0, 1}},
		{Float128{0x8000_0000_0000_0000, 1}, Float128{0x8000_0000_0000_0000, 1}},
		{exact128(100), exact128(1)},
		{exact128(-100), exact128(-1)},

		// erf(x) is rounded to 1 for |x| >= 8.7334
		{Float128{0x4002_1778_42bc_e674, 0x48bc_471e_ea54_0734}, Float128{0x3ffe_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}},
		{Float128{0x4002_1778_42bc_e674, 0x48bc_471e_ea54_0735}, exact128(1)},
		{Float128{0xc002_1778_42bc_e674, 0x48bc_471e_ea54_0734}, Float128{0xbffe_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}},
		{Float128{0xc002_1778_42bc_e674, 0x48bc_471e_ea54_0735}, exact128(-1)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erf()
		if !eq128(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat128_ErfAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat128_ErfAccuracy(t *testing.T) {
	t.Parallel()
	checkFloat128Testdata(t, "testdata/erf128.txt", "Erf", Float128.Erf)
}

// TestFloat128_ErfinvAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat128_ErfinvAccuracy(t *testing.T) {
	t.Parallel()
	checkFloat128Testdata(t, "testdata/erfinv128.txt", "Erfinv", Float128.Erfinv)
}

// TestFloat128_ErfcinvAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat128_ErfcinvAccuracy(t *testing.T) {
	t.Parallel()
	checkFloat128Testdata(t, "testdata/erfcinv128.txt", "Erfcinv", Float128.Erfcinv)
}

// TestFloat128_ErfcAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat128_ErfcAccuracy(t *testing.T) {
	t.Parallel()
	checkFloat128Testdata(t, "testdata/erfc128.txt", "Erfc", Float128.Erfc)
}

// checkFloat128Testdata requires the correctly rounded result for every vector of the test data.
func checkFloat128Testdata(t *testing.T, name, fn string, f func(Float128) Float128) {
	t.Helper()
	file, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

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
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 2 || len(fields[0]) != 32 || len(fields[1]) != 32 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		x, want := parse(fields[0]), parse(fields[1])
		if got := f(x); !eq128(got, want) {
			t.Errorf("%s(%v) = %v; want %v", fn, x, got, want)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

// TestFloat128_ErfKernel checks the accuracy of the kernel of Erf with the reference values calculated by mpmath.
// The kernel is more accurate than Float128, so that its errors are hidden by the rounding.
func TestFloat128_ErfKernel(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/erf128_kernels.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	parse := func(s string) ints.Uint256 {
		if len(s) != 64 {
			t.Fatalf("malformed number: %q", s)
		}
		var x ints.Uint256
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
		if len(fields) != 3 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		x, want := parse(fields[0]), parse(fields[1])
		got, scaled := erf128Cell(x)
		if scaled != (fields[2] == "1") {
			t.Errorf("erf128Cell(%v): scaled is %v; want %v", fields[0], scaled, fields[2] == "1")
		}
		d := got.Sub(want)
		if got.Cmp(want) < 0 {
			d = want.Sub(got)
		}
		// the relative error is less than 2**-170
		if d.BitLen()+170 > want.BitLen() {
			t.Errorf("erf128Cell(%v) = %v; want %v", fields[0], got, want)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

// TestFloat128_ErfcKernel checks the accuracy of the kernels of Erfc with the reference values calculated by mpmath.
// The kernels are more accurate than Float128, so that their errors are hidden by the rounding.
func TestFloat128_ErfcKernel(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/erfc128_kernels.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	parse := func(s string) ints.Uint256 {
		if len(s) != 64 {
			t.Fatalf("malformed number: %q", s)
		}
		var x ints.Uint256
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
		if len(fields) != 4 || len(fields[1]) != 32 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		var a Float128
		for i := range a {
			a[i], err = strconv.ParseUint(fields[1][16*i:16*(i+1)], 16, 64)
			if err != nil {
				t.Fatal(err)
			}
		}
		want := parse(fields[2])
		wantExp, err := strconv.Atoi(fields[3])
		if err != nil {
			t.Fatal(err)
		}

		_, exp, m := a.normalize()
		x := m.Uint256().Lsh(uint(80 + exp))
		var got ints.Uint256
		var gotExp int
		var bound int
		switch fields[0] {
		case "S":
			got, gotExp = erf128Scaled(x)
			bound = 170
		case "L":
			got, gotExp = erfc128Large(x, exp, m)
			bound = 135
		default:
			t.Fatalf("unknown kernel: %q", fields[0])
		}

		// normalize the result to [1, 2), which is in [2**192, 2**193) in fixed point.
		got, gotExp = gammaNormalize(got, gotExp)
		if gotExp != wantExp {
			t.Errorf("%s(%v): exponent = %d; want %d", fields[0], a, gotExp, wantExp)
			continue
		}
		d := got.Sub(want)
		if got.Cmp(want) < 0 {
			d = want.Sub(got)
		}
		// the relative error is less than 2**-bound
		if d.BitLen()+bound > want.BitLen() {
			t.Errorf("%s(%v) = %v; want %v (the error is 2**%d)", fields[0], a, got, want, d.BitLen()-want.BitLen())
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestFloat128_ErfRound(t *testing.T) {
	t.Parallel()
	tests := []struct {
		v    ints.Uint256
		want uint64
	}{
		{ints.Uint256{5, 0, 0, 0}, 5},
		{ints.Uint256{5, 1<<63 - 1, ^uint64(0), ^uint64(0)}, 5},
		{ints.Uint256{5, 1 << 63, 0, 0}, 6},   // a tie, the even number
		{ints.Uint256{6, 1 << 63, 0, 0}, 6},   // a tie, the even number
		{ints.Uint256{6, 1 << 63, 1, 0}, 7},   // a little larger than a tie
		{ints.Uint256{6, 1 << 63, 0, 1}, 7},   // a little larger than a tie
		{ints.Uint256{6, 1<<63 + 1, 0, 0}, 7}, // a little larger than a tie
		{ints.Uint256{6, ^uint64(0), ^uint64(0), ^uint64(0)}, 7},
		{ints.Uint256{0, 1 << 63, 0, 0}, 0}, // a tie, the even number
		{ints.Uint256{0, 1<<63 - 1, 0, 0}, 0},
		{ints.Uint256{0, 1<<63 + 1, 0, 0}, 1},
	}
	for _, tt := range tests {
		if got := erf128Round(tt.v); got != tt.want {
			t.Errorf("erf128Round(%v) = %d; want %d", tt.v, got, tt.want)
		}
	}
}

func TestFloat128_ErfRoundShift(t *testing.T) {
	t.Parallel()
	tests := []struct {
		v    ints.Uint256
		sh   uint
		want ints.Uint128
	}{
		{ints.Uint256{0, 0, 0, 5 << 3}, 3, ints.Uint128{0, 5}},
		{ints.Uint256{0, 0, 0, 5<<3 | 3}, 3, ints.Uint128{0, 5}},
		{ints.Uint256{0, 0, 0, 5<<3 | 4}, 3, ints.Uint128{0, 6}},         // a tie, the even number
		{ints.Uint256{0, 0, 0, 6<<3 | 4}, 3, ints.Uint128{0, 6}},         // a tie, the even number
		{ints.Uint256{0, 0, 0, 6<<3 | 5}, 3, ints.Uint128{0, 7}},         // a little larger than a tie
		{ints.Uint256{0, 0, 1, 6<<3 | 4}, 3, ints.Uint128{0, 1<<61 + 6}}, // the upper words
		{ints.Uint256{1, 0, 0, 0}, 192, ints.Uint128{0, 1}},
		{ints.Uint256{2, 0, 0, 0}, 194, ints.Uint128{0, 0}}, // a tie, the even number
		{ints.Uint256{2, 0, 0, 1}, 194, ints.Uint128{0, 1}}, // a little larger than a tie
		{ints.Uint256{2, 0, 0, 0}, 193, ints.Uint128{0, 1}},
		{ints.Uint256{3, 0, 0, 0}, 193, ints.Uint128{0, 2}}, // a tie, the even number
		{ints.Uint256{1, 1, 0, 0}, 129, ints.Uint128{0, 1 << 63}},
		{ints.Uint256{0, 1, 0, 0}, 64, ints.Uint128{1, 0}},
	}
	for _, tt := range tests {
		if got := erf128RoundShift(tt.v, tt.sh); got != tt.want {
			t.Errorf("erf128RoundShift(%v, %d) = %v; want %v", tt.v, tt.sh, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Erf(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float128
	}{
		{"tiny", exact128(1e-300)},    // erf(x) ~ 2 x / sqrt(pi)
		{"small", exact128(0.001)},    // |x| < 2**-8
		{"medium", exact128(1.5)},     // 2**-8 <= |x| < 7
		{"large", exact128(5.5)},      // the series has many terms
		{"asymptotic", exact128(7.5)}, // 7 <= |x| < 8.74
		{"negative", exact128(-1.5)},  // the sign is restored
		{"saturated", exact128(10)},   // the result is 1
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erf())
			}
		})
	}
}

func TestFloat128_Erfc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(-4), "1.999999984582742099719981147840326513115951427854746410808831657095005786958973189"},
		{exact128(-3), "1.999977909503001414558627223870417679620152292912600750342761045157057543316379868"},
		{exact128(-2), "1.995322265018952734162069256367252928610891797040060076738352326200437280719995177"},
		{exact128(-1), "1.842700792949714869341220635082609259296066997966302908459937897834717254096010841"},
		{exact128(0), "1"},
		{exact128(1), "0.1572992070502851306587793649173907407039330020336970915400621021652827459039891587"},
		{exact128(2), "0.004677734981047265837930743632747071389108202959939923261647673799562719280004822632"},
		{exact128(3), "2.209049699858544137277612958232037984770708739924965723895484294245668362013226782E-5"},
		{exact128(4), "1.541725790028001885215967348688404857214525358919116834290499421304102681125407277E-8"},
		{exact128(11), "1.440866137943694680339809702856082753964395230973580405075797246278446292140321302E-54"},
		{exact128(100), "6.405961424921732039021339148586394148214414399460338057767107650248902554829505831E-4346"},
		{exact128(1e-300), "1"},
	}

	for _, tt := range tests {
		got := tt.x.Erfc()
		if !close128(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(math.Inf(1)), exact128(0)},
		{exact128(math.Inf(-1)), exact128(2)},
		{exact128(math.NaN()), exact128(math.NaN())},
		{exact128(0), exact128(1)},
		{exact128(math.Copysign(0, -1)), exact128(1)},
		{exact128(0x1p-130), exact128(1)},
		{exact128(-0x1p-130), exact128(1)},

		// erfc(x) is rounded to 2 for x <= -8.6939, and to 0 for x >= 106.9
		{Float128{0xc002_1634_8a58_5939, 0x8f7a_8603_4cb5_b0b2}, Float128{0x3fff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}},
		{Float128{0xc002_1634_8a58_5939, 0x8f7a_8603_4cb5_b0b4}, exact128(2)},
		{exact128(-0x1p100), exact128(2)},
		{Float128{0x4005_ab9c_8393_ddd2, 0x517e_5404_d6dd_3c9c}, Float128{0, 1}},
		{Float128{0x4005_ab9c_8393_ddd2, 0x517e_5404_d6dd_3c9d}, exact128(0)},
		{exact128(0x1p100), exact128(0)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfc()
		if !eq128(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Erfc(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float128
	}{
		{"tiny", exact128(1e-300)},     // erfc(x) = 1
		{"small", exact128(0.001)},     // |x| < 2**-8
		{"medium", exact128(1.5)},      // 2**-8 <= |x| < 4
		{"scaled", exact128(6.5)},      // 4 <= |x| < 16
		{"asymptotic", exact128(30.5)}, // 16 <= |x|
		{"negative", exact128(-1.5)},   // 2 - erfc(-x)
		{"negative-scaled", exact128(-6.5)},
		{"underflow", exact128(110)}, // the result is 0
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erfc())
			}
		})
	}
}

func TestFloat128_Erfinv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(-1).Nextafter(exact128(0)), "-8.693913624339801357999421505949273"},
		{exact128(-0.75), "-0.813419847597618541690289359893421"},
		{exact128(-0.5), "-0.47693627620446987338141835364313055"},
		{exact128(-0.25), "-0.22531205501217810472501401395227754"},
		{exact128(0), "0"},
		{exact128(0.25), "0.22531205501217810472501401395227754"},
		{exact128(0.5), "0.47693627620446987338141835364313055"},
		{exact128(0.75), "0.813419847597618541690289359893421"},
		{exact128(1).Nextafter(exact128(0)), "8.693913624339801357999421505949273"},
		{exact128(0x1p-100), "6.991097746438968341920179842314715E-31"},
	}

	for _, tt := range tests {
		got := tt.x.Erfinv()
		if !close128(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(1), exact128(math.Inf(1))},
		{exact128(-1), exact128(math.Inf(-1))},
		{exact128(2), exact128(math.NaN())},
		{exact128(-2), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.NaN())},
		{exact128(math.Inf(1)), exact128(math.NaN())},
		{exact128(math.Inf(-1)), exact128(math.NaN())},
		{exact128(1).Nextafter(exact128(2)), exact128(math.NaN())},
		{exact128(0), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},
		{Float128{0, 1}, Float128{0, 1}}, // the smallest subnormal number: erfinv(x) = 0.886 x is rounded to x
	}

	for _, tt := range strictTests {
		got := tt.x.Erfinv()
		if !eq128(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestErfcinv64 checks the initial approximation of Erfinv.
func TestErfcinv64(t *testing.T) {
	t.Parallel()
	rnd := rand.New(rand.NewPCG(1, 2))
	for range 10000 {
		// c in [2**-113, 1/2] distributed uniformly in the exponent
		c := math.Ldexp(1+rnd.Float64(), -1-rnd.IntN(112))
		y := erfcinv64(c)
		if got := math.Erfc(y); math.Abs(got-c) > 0x1p-44*c {
			t.Fatalf("Erfc(erfcinv64(%v)) = %v", c, got)
		}
	}
}

func BenchmarkFloat128_Erfinv(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float128
	}{
		{"tiny", exact128(0x1p-100)},                         // |x| < 2**-60
		{"small", exact128(0.001)},                           // |x| <= 1/2
		{"medium", exact128(0.5)},                            // |x| <= 1/2
		{"large", exact128(0.75)},                            // 1/2 < |x| < 1
		{"close-to-one", exact128(1).Nextafter(exact128(0))}, // 1 - 2**-113
		{"negative", exact128(-0.75)},                        // the sign is restored
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erfinv())
			}
		})
	}
}

func TestFloat128_Erfcinv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(0.25), "0.8134198475976185416902893598934208"},
		{exact128(0.5), "0.47693627620446987338141835364313055"},
		{exact128(0.75), "0.22531205501217810472501401395227754"},
		{exact128(1), "0"},
		{exact128(1.25), "-0.22531205501217810472501401395227754"},
		{exact128(1.5), "-0.47693627620446987338141835364313055"},
		{exact128(1.75), "-0.8134198475976185416902893598934208"},
		{exact128(2).Nextafter(exact128(0)), "-8.654218872644272166366043905074009"},
		{exact128(1e-300), "26.20946996051612388552073179045609"},
		{Float128{0, 1}, "106.8996038222388760536653005249475"}, // the smallest subnormal number
	}

	for _, tt := range tests {
		got := tt.x.Erfcinv()
		if !close128(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(0), exact128(math.Inf(1))},
		{exact128(2), exact128(math.Inf(-1))},
		{exact128(-1), exact128(math.NaN())},
		{exact128(3), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.NaN())},
		{exact128(math.Inf(1)), exact128(math.NaN())},
		{exact128(math.Inf(-1)), exact128(math.NaN())},
		{exact128(2).Nextafter(exact128(3)), exact128(math.NaN())},
		{exact128(math.Copysign(0, -1)), exact128(math.Inf(1))},
		{exact128(1), exact128(0)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfcinv()
		if !eq128(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestErfcinv64Log checks the initial approximation of Erfcinv.
func TestErfcinv64Log(t *testing.T) {
	t.Parallel()
	rnd := rand.New(rand.NewPCG(1, 2))
	for range 10000 {
		// ln(erfc(y)) in [-12000, -0.5] distributed uniformly in the exponent
		lc := -math.Ldexp(0.5+rnd.Float64(), rnd.IntN(15))
		if lc < -12000 {
			continue
		}
		y := erfcinv64Log(lc)
		if got := lnErfc64(y); math.Abs(got-lc) > 0x1p-44*math.Abs(lc) {
			t.Fatalf("lnErfc64(erfcinv64Log(%v)) = %v", lc, got)
		}
	}
}

func BenchmarkFloat128_Erfcinv(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float128
	}{
		{"tiny", exact128(1e-300)},                           // x < 2**-113
		{"small", exact128(1e-20)},                           // 2**-113 <= x <= 1/2
		{"medium", exact128(0.25)},                           // 2**-113 <= x <= 1/2
		{"large", exact128(0.75)},                            // 1/2 < x <= 1, erfinv(1-x)
		{"greater-than-one", exact128(1.5)},                  // 1 < x < 2
		{"close-to-two", exact128(2).Nextafter(exact128(0))}, // 2 - 2**-112
		{"subnormal", Float128{0, 1}},                        // the smallest subnormal number
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erfcinv())
			}
		})
	}
}
