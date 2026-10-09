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

func TestFloat256_Pow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		y    Float256
		want string
	}{
		{exact256(2), exact256(3), "8"},
		{exact256(5), exact256(0.5), "2.236067977499789696409173668731276235440618359611525724270897245410520925637804899414414408378782274969508176150773783504253267724"},
		{exact256(5), exact256(1.5), "11.18033988749894848204586834365638117720309179805762862135448622705260462818902449707207204189391137484754088075386891752126633862"},
	}

	for _, tt := range tests {
		got := tt.x.Pow(tt.y)
		if !close256(got, tt.want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		y    Float256
		want Float256
	}{
		// special cases
		// a.Pow(±0) = 1 for any a
		{exact256(2), exact256(0), exact256(1)},
		{exact256(2), exact256(math.Copysign(0, -1)), exact256(1)},
		{exact256(-2), exact256(0), exact256(1)},
		{exact256(-2), exact256(math.Copysign(0, -1)), exact256(1)},
		{exact256(math.Inf(1)), exact256(0), exact256(1)},
		{exact256(math.Inf(1)), exact256(math.Copysign(0, -1)), exact256(1)},
		{exact256(math.NaN()), exact256(0), exact256(1)},
		{exact256(math.NaN()), exact256(math.Copysign(0, -1)), exact256(1)},

		// 1.Pow(b) = 1 for any b
		{exact256(1), exact256(3), exact256(1)},
		{exact256(1), exact256(-3), exact256(1)},
		{exact256(1), exact256(math.Inf(1)), exact256(1)},
		{exact256(1), exact256(math.Inf(-1)), exact256(1)},
		{exact256(1), exact256(math.NaN()), exact256(1)},

		// a.Pow(1) = a for any a
		{exact256(2), exact256(1), exact256(2)},
		{exact256(-2), exact256(1), exact256(-2)},
		{exact256(math.Inf(1)), exact256(1), exact256(math.Inf(1))},
		{exact256(math.NaN()), exact256(1), exact256(math.NaN())},

		// NaN.Pow(b) = NaN
		{exact256(math.NaN()), exact256(3), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(-3), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.Inf(1)), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.Inf(-1)), exact256(math.NaN())},

		// a.Pow(NaN) = NaN
		{exact256(2), exact256(math.NaN()), exact256(math.NaN())},
		{exact256(-2), exact256(math.NaN()), exact256(math.NaN())},
		{exact256(math.Inf(1)), exact256(math.NaN()), exact256(math.NaN())},
		{exact256(math.Inf(-1)), exact256(math.NaN()), exact256(math.NaN())},

		// ±0.Pow(b) = ±Inf for b an odd integer < 0
		{exact256(0), exact256(-3), exact256(math.Inf(1))},
		{exact256(math.Copysign(0, -1)), exact256(-3), exact256(math.Inf(-1))},

		// ±0.Pow(-Inf) = +Inf
		{exact256(0), exact256(math.Inf(-1)), exact256(math.Inf(1))},
		{exact256(math.Copysign(0, -1)), exact256(math.Inf(-1)), exact256(math.Inf(1))},

		// ±0.Pow(+Inf) = +0
		{exact256(0), exact256(math.Inf(1)), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Inf(1)), exact256(0)},

		// ±0.Pow(b) = +Inf for finite b < 0 and not an odd integer
		{exact256(0), exact256(-2), exact256(math.Inf(1))},
		{exact256(math.Copysign(0, -1)), exact256(-2), exact256(math.Inf(1))},
		{exact256(math.Copysign(0, -1)), exact256(-0.5), exact256(math.Inf(1))},

		// ±0.Pow(b) = ±0 for b an odd integer > 0
		{exact256(0), exact256(3), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(3), exact256(math.Copysign(0, -1))},

		// ±0.Pow(b) = +0 for finite b > 0 and not an odd integer
		{exact256(0), exact256(2), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(2), exact256(0)},

		// -1.Pow(±Inf) = 1
		{exact256(-1), exact256(math.Inf(1)), exact256(1)},
		{exact256(-1), exact256(math.Inf(-1)), exact256(1)},

		// a.Pow(+Inf) = +Inf for |a| > 1
		{exact256(2), exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(-2), exact256(math.Inf(1)), exact256(math.Inf(1))},

		// a.Pow(-Inf) = +0 for |a| > 1
		{exact256(2), exact256(math.Inf(-1)), exact256(0)},
		{exact256(-2), exact256(math.Inf(-1)), exact256(0)},

		// a.Pow(+Inf) = +0 for |a| < 1
		{exact256(0.5), exact256(math.Inf(1)), exact256(0)},
		{exact256(-0.5), exact256(math.Inf(1)), exact256(0)},

		// a.Pow(-Inf) = +Inf for |a| < 1
		{exact256(0.5), exact256(math.Inf(-1)), exact256(math.Inf(1))},
		{exact256(-0.5), exact256(math.Inf(-1)), exact256(math.Inf(1))},

		// +Inf.Pow(b) = +Inf for b > 0
		{exact256(math.Inf(1)), exact256(2), exact256(math.Inf(1))},

		// +Inf.Pow(b) = +0 for b < 0
		{exact256(math.Inf(1)), exact256(-2), exact256(0)},

		// -Inf.Pow(b) = (-0).Pow(-b)
		{exact256(math.Inf(-1)), exact256(3), exact256(math.Inf(-1))},
		{exact256(math.Inf(-1)), exact256(2), exact256(math.Inf(1))},

		// a.Pow(b) = NaN for finite a < 0 and finite non-integer b
		{exact256(-2), exact256(0.5), exact256(math.NaN())},
		{exact256(-2), exact256(-0.5), exact256(math.NaN())},

		// overflow and underflow
		{exact256(3), exact256(1e7), exact256(math.Inf(1))},
		{exact256(3), exact256(-1e7), exact256(0)},
		{exact256(-3), exact256(10000001), exact256(math.Inf(-1))},
		{exact256(-3), exact256(10000000), exact256(math.Inf(1))},
		{exact256(-3), exact256(-10000001), exact256(math.Copysign(0, -1))},
		{exact256(0.5), exact256(1e7), exact256(0)},
		{exact256(2), exact256(262144), exact256(math.Inf(1))},
		{exact256(2), exact256(-262379), exact256(0)},          // a tie of the smallest subnormal number and 0
		{exact256(2), exact256(-262378), Float256{0, 0, 0, 1}}, // the smallest subnormal number
		{Float256{0, 0, 0, 1}, exact256(0x1p25), exact256(0)},  // powers of two with the huge exponents
		{Float256{0, 0, 0, 1}, exact256(-0x1p25), exact256(math.Inf(1))},
		{exact256(0.5), exact256(0x1p100), exact256(0)},
		{exact256(-2), exact256(0x1p100), exact256(math.Inf(1))},

		// exact results
		{exact256(4), exact256(1.5), exact256(8)},
		{exact256(4), exact256(-0.5), exact256(0.5)},
		{exact256(0.25), exact256(-1.5), exact256(8)},
	}

	for _, tt := range strictTests {
		got := tt.x.Pow(tt.y)
		if !eq256(got, tt.want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

// TestFloat256_PowAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat256_PowAccuracy(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/pow256.txt")
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
		if len(fields) != 3 || len(fields[0]) != 64 || len(fields[1]) != 64 || len(fields[2]) != 64 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		x, y, want := parse(fields[0]), parse(fields[1]), parse(fields[2])
		if got := x.Pow(y); !eq256(got, want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", x, y, got, want)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkFloat256_Pow(b *testing.B) {
	for _, tt := range []struct {
		name string
		x, y Float256
	}{
		{"square", exact256(1.5), exact256(2)},         // exact products
		{"integer", exact256(1.5), exact256(100)},      // exp(y log(x))
		{"half", exact256(5), exact256(0.5)},           // Sqrt
		{"fraction", exact256(5), exact256(1.5)},       // exp(y log(x))
		{"large", exact256(1.5), exact256(12345)},      // exp(y log(x))
		{"negative-base", exact256(-1.5), exact256(7)}, // the sign is restored
		{"overflow", exact256(10), exact256(500000)},   // +Inf
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Pow(tt.y))
			}
		})
	}
}
