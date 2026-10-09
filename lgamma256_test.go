package floats

import (
	"bufio"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestFloat256_Lgamma(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		want string
		sign int
	}{
		{exact256(0.5), "0.57236494292470008707171367567652935582364740645765578575681153573606888494241304", 1},
		{exact256(2.5), "0.28468287047291915963249466968270192432013769555989472925014585038677593422163258", 1},
		{exact256(-0.5), "1.2655121234846453964889457971347059238991475408179110398774915452294625069121078", -1},
		{exact256(-2.5), "-0.056243716497674050672594530097654284122944102552845625528490660895423530074769545", -1},
		{exact256(100), "359.13420536957539877604401046028690961262171808562972877561279307484079922862431", 1},

		// around MaxStirling = 20368, straddling the Gamma()-delegate /
		// log-Stirling boundary
		{exact256(20367), "181703.63590523834174280751352918755997758359121047630155185622699622838323420021", 1},
		{exact256(20367.5), "181708.59673471244752817104195584628995334483761435211985208415109379313033586806", 1},
		{exact256(20368), "181713.55757646131149956532529284338592877424088336729924039604880182529499417916", 1},
		{exact256(20368.5), "181718.51843048463232501089711244531517037564471862872066667524177041744702011456", 1},
		{exact256(50000), "490984.42327157182173146714600978963141888862691236523762240771773984560780884109", 1},
		{exact256(-20367.5), "-181717.37370059878292483675368509396211166399742381580535510372814734597488234468", 1},
		{exact256(-20368.5), "-181727.29544546758988393471922581961633712999915826144861372897233316958932699747", -1},
		{exact256(-50000.5), "-490999.50821661253775633400917018122366021815437791950962903797213601934970537546", -1},
	}

	for _, tt := range tests {
		got, sign := tt.x.Lgamma()
		if !close256(got, tt.want) || sign != tt.sign {
			t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", tt.x, got, sign, tt.want, tt.sign)
		}
	}

	strictTests := []struct {
		x        Float256
		want     Float256
		wantSign int
	}{
		// special cases
		{exact256(math.Inf(1)), exact256(math.Inf(1)), 1},
		{exact256(math.Inf(-1)), exact256(math.Inf(-1)), 1},
		{exact256(0), exact256(math.Inf(1)), 1},
		{exact256(math.Copysign(0, -1)), exact256(math.Inf(1)), 1},
		{exact256(-1), exact256(math.Inf(1)), 1},
		{exact256(-2), exact256(math.Inf(1)), 1},
		{exact256(math.NaN()), exact256(math.NaN()), 1},
		{exact256(1), exact256(0), 1},
		{exact256(2), exact256(0), 1},
	}

	for _, tt := range strictTests {
		got, sign := tt.x.Lgamma()
		if !eq256(got, tt.want) || sign != tt.wantSign {
			t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", tt.x, got, sign, tt.want, tt.wantSign)
		}
	}
}

// TestFloat256_LgammaAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat256_LgammaAccuracy(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/lgamma256.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

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
		if len(fields) != 3 || len(fields[0]) != 64 || len(fields[1]) != 64 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		x, want := parse(fields[0]), parse(fields[1])
		wantSign, err := strconv.Atoi(fields[2])
		if err != nil {
			t.Fatal(err)
		}
		if got, sign := x.Lgamma(); !eq256(got, want) || sign != wantSign {
			t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", x, got, sign, want, wantSign)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

// TestFloat256_LgammaKernels checks the accuracy of the kernels of Lgamma with the reference values calculated by mpmath.
// The kernels are more accurate than Float256, so that their errors are hidden by the rounding.
func TestFloat256_LgammaKernels(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/lgamma256_kernels.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

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
	// the signed number
	parseSigned := func(s string) lgammaFix256 {
		return lgammaFix256{s[0] == '-', parse(s[1:])}
	}
	diff := func(a, b gammaFix256) int {
		d := a.sub(b)
		if a.cmp(b) < 0 {
			d = b.sub(a)
		}
		return d.bitLen() - 320
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "pos", "neg":
			x, want := parse(fields[1]), parseSigned(fields[2])
			var got lgammaFix256
			if fields[0] == "pos" {
				got = lgamma320Pos(x)
			} else {
				_, got = lgamma320NegFix(x)
			}
			// the absolute error is less than 2**-288
			if got.neg != want.neg || diff(got.v, want.v) > -288 {
				t.Errorf("%s(%v): the error is 2**%d; want less than 2**-288", fields[0], fields[1], diff(got.v, want.v))
			}
		case "stirling":
			if len(fields) != 5 {
				t.Fatalf("malformed line: %q", sc.Text())
			}
			ze, err := strconv.Atoi(fields[2])
			if err != nil {
				t.Fatal(err)
			}
			ye, err := strconv.Atoi(fields[4])
			if err != nil {
				t.Fatal(err)
			}
			got, e := lgamma320Stirling(parse(fields[1]), ze)
			want := parse(fields[3])
			if e != ye || diff(got, want) > -285 {
				t.Errorf("stirling(%v, %d) = (%v, %d): the error is 2**%d; want (%v, %d) less than 2**-285", fields[1], ze, got, e, diff(got, want), want, ye)
			}
		default:
			t.Fatalf("malformed line: %q", sc.Text())
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkFloat256_Lgamma(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float256
	}{
		{"tiny", exact256(1e-300)},           // Lgamma(x) ~ -log(x)
		{"small", exact256(0.37)},            // 0 < x < 1
		{"medium", exact256(1.5)},            // 1 <= x < 3
		{"large", exact256(10.3)},            // 3 <= x < 48
		{"stirling", exact256(100.7)},        // 48 <= x < 2048
		{"big", exact256(5000.5)},            // 2048 <= x < 2**15
		{"huge", exact256(1e10)},             // 2**15 <= x
		{"negative", exact256(-2.5)},         // -2**15 < x < 0
		{"verynegative", exact256(-100.5)},   // reflection with large |x|
		{"hugenegative", exact256(-50000.5)}, // -2**15 > x
		{"nearzero", exact256(1.0625)},       // the result is close to zero
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
