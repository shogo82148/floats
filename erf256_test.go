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

func TestFloat256_Erf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-1), "-0.8427007929497148693412206350826092592960669979663029084599378978347172540960108413"},
		{exact256(-0.5), "-0.5204998778130465376827466538919645287364515757579637000588057256471935217168535709"},
		{exact256(0), "0"},
		{exact256(0.5), "0.5204998778130465376827466538919645287364515757579637000588057256471935217168535709"},
		{exact256(1), "0.8427007929497148693412206350826092592960669979663029084599378978347172540960108413"},
		{exact256(2), "0.9953222650189527341620692563672529286108917970400600767383523262004372807199951774"},
		{exact256(3), "0.9999779095030014145586272238704176796201522929126007503427610451570575433163798677"},
		{exact256(4), "0.9999999845827420997199811478403265131159514278547464108088316570950057869589731887"},
		{exact256(11), "0.9999999999999999999999999999999999999999999999999999985591338620563053196601902971"},
	}

	for _, tt := range tests {
		got := tt.x.Erf()
		if !close256(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(math.Inf(1)), exact256(1)},
		{exact256(math.Inf(-1)), exact256(-1)},
		{exact256(math.NaN()), exact256(math.NaN())},
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
		{exact256(100), exact256(1)},
		{exact256(-100), exact256(-1)},

		// erf(x) is rounded to 1 for |x| >= 12.722
		{Float256{0x4000_2971_af06_3fb1, 0x3a98_f6a7_2e5f_b545, 0x635d_1d72_2ec8_07a0, 0xaca2_ae63_3fc9_898b}, exact256(1).Nextafter(exact256(0))},
		{Float256{0x4000_2971_af06_3fb1, 0x3a98_f6a7_2e5f_b545, 0x635d_1d72_2ec8_07a0, 0xaca2_ae63_3fc9_898c}, exact256(1)},
		{Float256{0xc000_2971_af06_3fb1, 0x3a98_f6a7_2e5f_b545, 0x635d_1d72_2ec8_07a0, 0xaca2_ae63_3fc9_898b}, exact256(-1).Nextafter(exact256(0))},
		{Float256{0xc000_2971_af06_3fb1, 0x3a98_f6a7_2e5f_b545, 0x635d_1d72_2ec8_07a0, 0xaca2_ae63_3fc9_898c}, exact256(-1)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erf()
		if !eq256(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat256_ErfAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat256_ErfAccuracy(t *testing.T) {
	t.Parallel()
	checkFloat256Testdata(t, "testdata/erf256.txt", "Erf", Float256.Erf)
}

// TestFloat256_ErfinvAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat256_ErfinvAccuracy(t *testing.T) {
	t.Parallel()
	checkFloat256Testdata(t, "testdata/erfinv256.txt", "Erfinv", Float256.Erfinv)
}

// TestFloat256_ErfcinvAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat256_ErfcinvAccuracy(t *testing.T) {
	t.Parallel()
	checkFloat256Testdata(t, "testdata/erfcinv256.txt", "Erfcinv", Float256.Erfcinv)
}

// TestFloat256_ErfcAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat256_ErfcAccuracy(t *testing.T) {
	t.Parallel()
	checkFloat256Testdata(t, "testdata/erfc256.txt", "Erfc", Float256.Erfc)
}

// checkFloat256Testdata requires the correctly rounded result for every vector of the test data.
func checkFloat256Testdata(t *testing.T, name, fn string, f func(Float256) Float256) {
	t.Helper()
	file, err := os.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

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
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 2 || len(fields[0]) != 64 || len(fields[1]) != 64 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		x, want := parse(fields[0]), parse(fields[1])
		if got := f(x); !eq256(got, want) {
			t.Errorf("%s(%v) = %v; want %v", fn, x, got, want)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

// TestFloat256_ErfKernel checks the accuracy of the kernel of Erf with the reference values calculated by mpmath.
// The kernel is more accurate than Float256, so that its errors are hidden by the rounding.
func TestFloat256_ErfKernel(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/erf256_kernels.txt")
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
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 4 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		x, want := parse(fields[0]), parse(fields[1])
		wantExp, err := strconv.Atoi(fields[3])
		if err != nil {
			t.Fatal(err)
		}
		got, e, scaled := erf256Cell(x)
		if scaled != (fields[2] == "1") || e != wantExp {
			t.Errorf("erf256Cell(%v): (%d, %v); want (%d, %v)", fields[0], e, scaled, wantExp, fields[2] == "1")
			continue
		}
		d := got.sub(want)
		if got.cmp(want) < 0 {
			d = want.sub(got)
		}
		// the relative error is less than 2**-310
		if d.bitLen()+310 > want.bitLen() {
			t.Errorf("erf256Cell(%v) = %v; want %v", fields[0], got, want)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

// TestFloat256_ErfcKernel checks the accuracy of the kernels of Erfc with the reference values calculated by mpmath.
// The kernels are more accurate than Float256, so that their errors are hidden by the rounding.
func TestFloat256_ErfcKernel(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/erfc256_kernels.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	parse := func(s string, n int) []uint64 {
		if len(s) != 16*n {
			t.Fatalf("malformed number: %q", s)
		}
		x := make([]uint64, n)
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
		if len(fields) != 4 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		a := Float256(parse(fields[1], 4))
		want := gammaFix256(parse(fields[2], 6))
		wantExp, err := strconv.Atoi(fields[3])
		if err != nil {
			t.Fatal(err)
		}

		_, exp, m := a.normalize()
		x := gammaFix256FromUint256(m, uint(exp+84))
		var got gammaFix256
		var gotExp int
		var bound int
		switch fields[0] {
		case "S":
			got, gotExp = erf256Scaled(x)
			bound = 310
		case "L":
			got, gotExp = erfc256Large(x, exp, m)
			bound = 275
		default:
			t.Fatalf("unknown kernel: %q", fields[0])
		}

		if gotExp != wantExp {
			t.Errorf("%s(%v): exponent = %d; want %d", fields[0], a, gotExp, wantExp)
			continue
		}
		d := got.sub(want)
		if got.cmp(want) < 0 {
			d = want.sub(got)
		}
		// the relative error is less than 2**-bound
		if d.bitLen()+bound > want.bitLen() {
			t.Errorf("%s(%v) = %v; want %v (the error is 2**%d)", fields[0], a, got, want, d.bitLen()-want.bitLen())
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

func TestFloat256_ErfRound(t *testing.T) {
	t.Parallel()
	// v × 2**(e-320) with v = (m: the fraction part × 2**320)
	half := ints.Uint512{3: 1 << 63}                                     // 1/2 in fixed point with 320 fractional bits
	integer := func(n uint64) ints.Uint512 { return ints.Uint512{2: n} } // n × 2**320
	tests := []struct {
		v    ints.Uint512
		e    int
		want uint64
	}{
		{integer(5), 0, 5},
		{integer(5).Add(half).Sub(ints.Uint512{7: 1}), 0, 5},
		{integer(5).Add(half), 0, 6},                         // a tie, the even number
		{integer(6).Add(half), 0, 6},                         // a tie, the even number
		{integer(6).Add(half).Add(ints.Uint512{7: 1}), 0, 7}, // a little larger than a tie
		{integer(6).Add(half).Add(ints.Uint512{3: 1}), 0, 7},
		{integer(3), 1, 6},
		{integer(3).Add(half), 1, 7}, // 7 exactly
		{integer(3).Add(half).Add(half), 2, 16},
		{half.Add(ints.Uint512{7: 1}), 1, 1},        // slightly more than 1/2 × 2 = 1
		{integer(1).Add(half), -1, 1},               // 3/4
		{integer(1), -1, 0},                         // 1/2, a tie, the even number
		{integer(1).Add(ints.Uint512{7: 1}), -1, 1}, // a little larger than a tie
	}
	for _, tt := range tests {
		got := erf256Round(tt.v, tt.e)
		if got != (ints.Uint256{3: tt.want}) {
			t.Errorf("erf256Round(%v, %d) = %v; want %d", tt.v, tt.e, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Erf(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float256
	}{
		{"tiny", exact256(1e-300)},     // erf(x) ~ 2 x / sqrt(pi)
		{"small", exact256(0.001)},     // |x| < 2**-8
		{"medium", exact256(1.5)},      // 2**-8 <= |x| < 7
		{"large", exact256(5.5)},       // the series has many terms
		{"asymptotic", exact256(10.5)}, // 7 <= |x| < 12.72
		{"negative", exact256(-1.5)},   // the sign is restored
		{"saturated", exact256(20)},    // the result is 1
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erf())
			}
		})
	}
}

func TestFloat256_Erfc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-1), "1.842700792949714869341220635082609259296066997966302908459937897834717254096010841"},
		{exact256(-0.5), "1.520499877813046537682746653891964528736451575757963700058805725647193521716853571"},
		{exact256(0), "1"},
		{exact256(0.5), "0.4795001221869534623172533461080354712635484242420362999411942743528064782831464291"},
		{exact256(1), "0.1572992070502851306587793649173907407039330020336970915400621021652827459039891587"},
		{exact256(2), "0.004677734981047265837930743632747071389108202959939923261647673799562719280004822632"},
		{exact256(4), "1.541725790028001885215967348688404857214525358919116834290499421304102681125407277E-8"},
		{exact256(11), "1.440866137943694680339809702856082753964395230973580405075797246278446292140321302E-54"},
		{exact256(100), "6.405961424921732039021339148586394148214414399460338057767107650248902554829505831E-4346"},
		{exact256(400), "1.077107859467669326180798128257033507159126340810440750431127736133942555552880804E-69490"},
	}

	for _, tt := range tests {
		got := tt.x.Erfc()
		if !close256(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(math.Inf(1)), exact256(0)},
		{exact256(math.Inf(-1)), exact256(2)},
		{exact256(math.NaN()), exact256(math.NaN())},
		{exact256(0), exact256(1)},
		{exact256(math.Copysign(0, -1)), exact256(1)},
		{exact256(0x1p-250), exact256(1)},
		{exact256(-0x1p-250), exact256(1)},
		{exact256(1e-300), exact256(1)},

		// erfc(x) is rounded to 2 for x <= -12.695, and to 0 for x >= 426.45
		{Float256{0xc000_2963_c382_2856, 0xda94_b711_9017_42fd, 0x8025_b5e1_20c3_3205, 0xfccc_8b78_050a_bc4f}, exact256(2).Nextafter(exact256(0))},
		{Float256{0xc000_2963_c382_2856, 0xda94_b711_9017_42fd, 0x8025_b5e1_20c3_3205, 0xfccc_8b78_050a_bc50}, exact256(2)},
		{exact256(-0x1p100), exact256(2)},
		{Float256{0x4000_7aa7_382a_14ed, 0x2d77_5059_d7cf_77a2, 0xca42_44d8_9adb_8ce1, 0x0325_ae1c_e38f_78c3}, Float256{0, 0, 0, 1}},
		{Float256{0x4000_7aa7_382a_14ed, 0x2d77_5059_d7cf_77a2, 0xca42_44d8_9adb_8ce1, 0x0325_ae1c_e38f_78c4}, exact256(0)},
		{exact256(0x1p100), exact256(0)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfc()
		if !eq256(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Erfc(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float256
	}{
		{"tiny", exact256(1e-300)},     // erfc(x) = 1
		{"small", exact256(0.001)},     // |x| < 2**-8
		{"medium", exact256(1.5)},      // 2**-8 <= |x| < 4
		{"scaled", exact256(6.5)},      // 4 <= |x| < 16
		{"asymptotic", exact256(30.5)}, // 16 <= |x|
		{"negative", exact256(-1.5)},   // 2 - erfc(-x)
		{"negative-scaled", exact256(-6.5)},
		{"underflow", exact256(500)}, // the result is 0
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erfc())
			}
		})
	}
}

func TestFloat256_Erfinv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-1).Nextafter(exact256(0)), "-12.69485098593161357335362500394592023522571345295539758597965430708524"},
		{exact256(-0.75), "-0.813419847597618541690289359893421085324724835957501548147510003331798062"},
		{exact256(-0.5), "-0.476936276204469873381418353643130559808969749059470644703882695919383452"},
		{exact256(-0.25), "-0.225312055012178104725014013952277554782118447807246757600782894957738222"},
		{exact256(0), "0"},
		{exact256(0.25), "0.225312055012178104725014013952277554782118447807246757600782894957738222"},
		{exact256(0.5), "0.476936276204469873381418353643130559808969749059470644703882695919383452"},
		{exact256(0.75), "0.813419847597618541690289359893421085324724835957501548147510003331798062"},
		{exact256(1).Nextafter(exact256(0)), "12.69485098593161357335362500394592023522571345295539758597965430708524"},
		{exact256(0x1p-200), "5.515003696744420228223845868649816296584181034421833816578678785356795E-61"},
	}

	for _, tt := range tests {
		got := tt.x.Erfinv()
		if !close256(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(1), exact256(math.Inf(1))},
		{exact256(-1), exact256(math.Inf(-1))},
		{exact256(2), exact256(math.NaN())},
		{exact256(-2), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN())},
		{exact256(math.Inf(1)), exact256(math.NaN())},
		{exact256(math.Inf(-1)), exact256(math.NaN())},
		{exact256(1).Nextafter(exact256(2)), exact256(math.NaN())},
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfinv()
		if !eq256(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat256_ErfDefect checks erf256Defect, which calculates erf(y) - t and erfc(y) - t more accurately than Float256.
func TestFloat256_ErfDefect(t *testing.T) {
	t.Parallel()
	rnd := rand.New(rand.NewPCG(1, 2))
	for range 200 {
		for _, complement := range []bool{false, true} {
			// y in [2**-125, 0.48) for erf, and [0.47, 12.7) for erfc
			var y Float256
			if complement {
				y = NewFloat256(0.47 + 12.2*rnd.Float64()).Add(NewFloat256(rnd.Float64()).Mul(exact256(0x1p-60)))
			} else {
				y = NewFloat256(math.Ldexp(1+rnd.Float64(), -2-rnd.IntN(120))).Add(NewFloat256(rnd.Float64()).Mul(exact256(0x1p-180)))
			}
			f := y.Erf()
			if complement {
				f = y.Erfc()
			}
			// the rounding error of f is at most half an ulp of f.
			if d := erf256Defect(y, f, complement); d.Abs().Gt(f.Mul(exact256(0x1p-237))) {
				t.Errorf("erf256Defect(%v, %v, %v) = %v", y, f, complement, d)
			}
			// f - target is accurate if target is close to f.
			delta := f.Mul(exact256(0x1p-100))
			target := f.Add(delta)
			if d := erf256Defect(y, target, complement); d.Add(delta).Abs().Gt(delta.Mul(exact256(0x1p-130))) {
				t.Errorf("erf256Defect(%v, %v, %v) = %v; want %v", y, target, complement, d, delta.Neg())
			}
		}
	}
}

func TestFloat256_ErfFixDiff(t *testing.T) {
	t.Parallel()
	a, b := gammaFix256{1, 5}, gammaFix256{1, 3}
	if got := erf256FixDiff(a, a, 0); !eq256(got, Float256{}) {
		t.Errorf("erf256FixDiff(a, a) = %v; want 0", got)
	}
	// 2 × 2**-64 × 2**4
	if got, want := erf256FixDiff(a, b, 4), exact256(0x1p-59); !eq256(got, want) {
		t.Errorf("erf256FixDiff(a, b) = %v; want %v", got, want)
	}
	if got, want := erf256FixDiff(b, a, 4), exact256(-0x1p-59); !eq256(got, want) {
		t.Errorf("erf256FixDiff(b, a) = %v; want %v", got, want)
	}
}

func BenchmarkFloat256_Erfinv(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float256
	}{
		{"tiny", exact256(0x1p-200)},                         // |x| < 2**-125
		{"small", exact256(0.001)},                           // |x| <= 1/2
		{"medium", exact256(0.5)},                            // |x| <= 1/2
		{"large", exact256(0.75)},                            // 1/2 < |x| < 1
		{"close-to-one", exact256(1).Nextafter(exact256(0))}, // 1 - 2**-237
		{"negative", exact256(-0.75)},                        // the sign is restored
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erfinv())
			}
		})
	}
}

func TestFloat256_Erfcinv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(0.25), "0.813419847597618541690289359893421085324724835957501548147510003331798053"},
		{exact256(0.5), "0.476936276204469873381418353643130559808969749059470644703882695919383447"},
		{exact256(0.75), "0.225312055012178104725014013952277554782118447807246757600782894957738225"},
		{exact256(1), "0"},
		{exact256(1.25), "-0.225312055012178104725014013952277554782118447807246757600782894957738225"},
		{exact256(1.5), "-0.476936276204469873381418353643130559808969749059470644703882695919383447"},
		{exact256(1.75), "-0.813419847597618541690289359893421085324724835957501548147510003331798053"},
		{exact256(2).Nextafter(exact256(0)), "-12.66760552376903039787483533474977284586485812791479406737157556230541"},
		{exact256(1e-300), "26.20946996051612388552073179045608917320124038287459055671374005636769"},
		{exact256(1e-20), "6.601580622355142565624345890770360211174777232005676641310482551233419"},
		{Float256{0, 0, 0, 1}, "426.4503993164995476498937507289818753430127517185428612511297259712896"}, // the smallest subnormal number
	}

	for _, tt := range tests {
		got := tt.x.Erfcinv()
		if !close256(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(0), exact256(math.Inf(1))},
		{exact256(2), exact256(math.Inf(-1))},
		{exact256(-1), exact256(math.NaN())},
		{exact256(3), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN())},
		{exact256(math.Inf(1)), exact256(math.NaN())},
		{exact256(math.Inf(-1)), exact256(math.NaN())},
		{exact256(2).Nextafter(exact256(3)), exact256(math.NaN())},
		{exact256(math.Copysign(0, -1)), exact256(math.Inf(1))},
		{exact256(1), exact256(0)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfcinv()
		if !eq256(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat256_ErfFixRatio(t *testing.T) {
	t.Parallel()
	a, b := gammaFix256{1, 5}, gammaFix256{1, 3}
	if got := erf256FixRatio(a, a); !eq256(got, Float256{}) {
		t.Errorf("erf256FixRatio(a, a) = %v; want 0", got)
	}
	// (a - b)/b = 2**-63/(1 + 3 × 2**-64)
	want := exact256(0x1p-63).Quo(NewFloat256(1).Add(exact256(3 * 0x1p-64)))
	if got := erf256FixRatio(a, b); !eq256(got, want) {
		t.Errorf("erf256FixRatio(a, b) = %v; want %v", got, want)
	}
	if got := erf256FixRatio(b, a); got.Signbit() == false || got.Abs().Gt(want) {
		t.Errorf("erf256FixRatio(b, a) = %v; want negative and not larger than %v", got, want)
	}
}

func BenchmarkFloat256_Erfcinv(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float256
	}{
		{"tiny", exact256(1e-300)},                           // x < 2**-113
		{"small", exact256(1e-20)},                           // 2**-113 <= x <= 1/2
		{"medium", exact256(0.25)},                           // 2**-113 <= x <= 1/2
		{"large", exact256(0.75)},                            // 1/2 < x <= 1, erfinv(1-x)
		{"greater-than-one", exact256(1.5)},                  // 1 < x < 2
		{"close-to-two", exact256(2).Nextafter(exact256(0))}, // 2 - 2**-236
		{"subnormal", Float256{0, 0, 0, 1}},                  // the smallest subnormal number
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erfcinv())
			}
		})
	}
}
