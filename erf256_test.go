package floats

import (
	"bufio"
	"math"
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
	f, err := os.Open("testdata/erf256.txt")
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
		if got := x.Erf(); !eq256(got, want) {
			t.Errorf("Erf(%v) = %v; want %v", x, got, want)
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

func TestFloat256_Erfinv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-1).Nextafter(exact256(0)), "-12.6789204758043082417334375024209655173495726943819012599854669839592728"},
		{exact256(-0.75), "-0.813419847597618541690289359893421085324724835957501548147510003331798062"},
		{exact256(-0.5), "-0.476936276204469873381418353643130559808969749059470644703882695919383452"},
		{exact256(-0.25), "-0.225312055012178104725014013952277554782118447807246757600782894957738222"},
		{exact256(0), "0"},
		{exact256(0.25), "0.225312055012178104725014013952277554782118447807246757600782894957738222"},
		{exact256(0.5), "0.476936276204469873381418353643130559808969749059470644703882695919383452"},
		{exact256(0.75), "0.813419847597618541690289359893421085324724835957501548147510003331798062"},
		{exact256(1).Nextafter(exact256(0)), "12.6789204758043082417334375024209655173495726943819012599854669839592728"},
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
	}

	for _, tt := range strictTests {
		got := tt.x.Erfinv()
		if !eq256(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Erfinv(b *testing.B) {
	x := exact256(0.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfinv())
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
		{exact256(2).Nextafter(exact256(0)), "-12.65882204301587198669990229531130561057683528766947257145199350963364816"},
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
	}

	for _, tt := range strictTests {
		got := tt.x.Erfcinv()
		if !eq256(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Erfcinv(b *testing.B) {
	x := exact256(0.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfcinv())
	}
}
