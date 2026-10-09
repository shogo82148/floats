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

func TestFloat256_Jn(t *testing.T) {
	t.Parallel()
	tests := []struct {
		n    int
		x    Float256
		want string
	}{
		{2, exact256(0.001), "0.00000012499998958333366405833080188822496505714913292743747922799659833350859879945156"},
		{2, exact256(1), "0.11490348493190048046964688133516660534547031423020526041196879444099753274075866"},
		{2, exact256(5), "0.046565116277752215532303284310691057966790394158887521516066010182226476976780157"},
		{2, exact256(-5), "0.046565116277752215532303284310691057966790394158887521516066010182226476976780157"},
		{-2, exact256(5), "0.046565116277752215532303284310691057966790394158887521516066010182226476976780157"},
		{-2, exact256(-5), "0.046565116277752215532303284310691057966790394158887521516066010182226476976780157"},
		{3, exact256(10), "0.058379379305186812342935478410340956290068991381519956394622180663564325173166014"},
		{5, exact256(50), "-0.081400247696569639643974037928222065996025908292219772178875213713583483415564883"},
		{10, exact256(100), "-0.054732176935472014741917456265930408258859303175584501008410660363812634289499813"},

		// straddling the Miller / asymptotic crossover, x = max(500, 2*n**2)
		{2, exact256(499), "0.0097309604002708556761204242700873444634903417751398107705720433039312469615748776"},
		{2, exact256(500), "0.034142447334613487436502928830922426349758657979262673174980459865960700555978892"},
		{2, exact256(500.1), "0.03500367081307590629076917020092519942904854824969269839045352898419557562714"},
		{50, exact256(4999), "0.01108058546927705222865422262387750622585915590246981475131589866231920145830247"},
		{50, exact256(5000), "0.0041868485725039842676492959718957285561741861420531751165905765215767906253347032"},
		{50, exact256(5001), "-0.0065541417556303653627385286709727842994044043354699410556261769663799426877034839"},
		{100, exact256(19999), "0.003443950776193479173640683547073093690996547950246744772079544686622488313280499"},
		{100, exact256(20000), "0.0056211946146883004378826002386642262501972749160663653769841373579743329670924246"},

		// deep Miller descent: large n and x both push the margin up
		{1000, exact256(1999999), "-0.00035464016890324565775080732610689593103577139797851529176764456393972734625650349"},
		{1000, exact256(2000001), "0.00054657660198501083057436487650801541423327155921816735707812906580783640536627447"},
	}

	for _, tt := range tests {
		got := tt.x.Jn(tt.n)
		if !close256(got, tt.want) {
			t.Errorf("Jn(%d, %v) = %v; want %v", tt.n, tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		n    int
		x    Float256
		want Float256
	}{
		// n = 0, 1, -1 are J0 and J1
		{0, exact256(0), exact256(1)},
		{1, exact256(0), exact256(0)},
		{-1, exact256(0), exact256(math.Copysign(0, -1))},
		{-1, exact256(math.Inf(1)), exact256(math.Copysign(0, -1))},

		// underflow of the Taylor series: (x/2)**n/n! is less than the smallest subnormal number
		{40000, exact256(3), exact256(0)},
		{-40001, exact256(-3), exact256(0)},
		{2, exact256(0x1p-200000), exact256(0)},
		{300, exact256(0x1p-1000), exact256(0)},
		{2, exact256(0), exact256(0)},
		{2, exact256(math.Inf(1)), exact256(0)},
		{2, exact256(math.Inf(-1)), exact256(0)},
		{2, exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Jn(tt.n)
		if !eq256(got, tt.want) {
			t.Errorf("Jn(%d, %v) = %v; want %v", tt.n, tt.x, got, tt.want)
		}
	}
}

// TestFloat256_JnAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat256_JnAccuracy(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/jn256.txt")
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
		if len(fields) != 3 || len(fields[1]) != 64 || len(fields[2]) != 64 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		n, err := strconv.Atoi(fields[0])
		if err != nil {
			t.Fatal(err)
		}
		x, want := parse(fields[1]), parse(fields[2])
		if got := x.Jn(n); !eq256(got, want) {
			t.Errorf("Jn(%d, %v) = %v; want %v", n, x, got, want)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkFloat256_Jn(b *testing.B) {
	for _, tt := range []struct {
		name string
		n    int
		x    Float256
	}{
		{"taylor", 10, exact256(2.5)},
		{"miller-small", 10, exact256(20.5)},
		{"miller", 10, exact256(100.5)},
		{"miller-large-n", 1000, exact256(1000.5)},
		{"hankel", 5, exact256(1000.5)},
		{"hankel-huge", 5, exact256(0x1p1000)},
		{"negative", -5, exact256(-100.5)},
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Jn(tt.n))
			}
		})
	}
}

// TestFloat256_JnHankelBreak checks that Hankel's expansion stops where the terms begin to increase
// (the result has the error of the minimum term, which is far larger than Float256's precision for x < 110).
func TestFloat256_JnHankelBreak(t *testing.T) {
	t.Parallel()
	x := exact256(30)
	_, exp, m := x.normalize()
	got, want := jnHankel256(2, x, exp, m), x.Jn(2)
	if d := got.Sub(want).Abs(); d.Gt(exact256(0x1p-60)) {
		t.Errorf("jnHankel256(2, 30) = %v; want %v", got, want)
	}
}
