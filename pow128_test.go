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

func TestFloat128_Pow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float128
		y    Float128
		want string
	}{
		{exact128(2), exact128(3), "8"},
		{exact128(5), exact128(0.5), "2.236067977499789696409173668731276235440618359611525724270897245410520925637804899"},
		{exact128(5), exact128(1.5), "11.1803398874989484820458683436563811772030917980576286213544862270526046281890245"},
	}

	for _, tt := range tests {
		got := tt.x.Pow(tt.y)
		if !close128(got, tt.want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		y    Float128
		want Float128
	}{
		// special cases
		// a.Pow(±0) = 1 for any a
		{exact128(2), exact128(0), exact128(1)},
		{exact128(2), exact128(math.Copysign(0, -1)), exact128(1)},
		{exact128(-2), exact128(0), exact128(1)},
		{exact128(-2), exact128(math.Copysign(0, -1)), exact128(1)},
		{exact128(math.Inf(1)), exact128(0), exact128(1)},
		{exact128(math.Inf(1)), exact128(math.Copysign(0, -1)), exact128(1)},
		{exact128(math.NaN()), exact128(0), exact128(1)},
		{exact128(math.NaN()), exact128(math.Copysign(0, -1)), exact128(1)},

		// 1.Pow(b) = 1 for any b
		{exact128(1), exact128(3), exact128(1)},
		{exact128(1), exact128(-3), exact128(1)},
		{exact128(1), exact128(math.Inf(1)), exact128(1)},
		{exact128(1), exact128(math.Inf(-1)), exact128(1)},
		{exact128(1), exact128(math.NaN()), exact128(1)},

		// a.Pow(1) = a for any a
		{exact128(2), exact128(1), exact128(2)},
		{exact128(-2), exact128(1), exact128(-2)},
		{exact128(math.Inf(1)), exact128(1), exact128(math.Inf(1))},
		{exact128(math.NaN()), exact128(1), exact128(math.NaN())},

		// NaN.Pow(b) = NaN
		{exact128(math.NaN()), exact128(3), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(-3), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.Inf(1)), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.Inf(-1)), exact128(math.NaN())},

		// a.Pow(NaN) = NaN
		{exact128(2), exact128(math.NaN()), exact128(math.NaN())},
		{exact128(-2), exact128(math.NaN()), exact128(math.NaN())},
		{exact128(math.Inf(1)), exact128(math.NaN()), exact128(math.NaN())},
		{exact128(math.Inf(-1)), exact128(math.NaN()), exact128(math.NaN())},

		// ±0.Pow(b) = ±Inf for b an odd integer < 0
		{exact128(0), exact128(-3), exact128(math.Inf(1))},
		{exact128(math.Copysign(0, -1)), exact128(-3), exact128(math.Inf(-1))},

		// ±0.Pow(-Inf) = +Inf
		{exact128(0), exact128(math.Inf(-1)), exact128(math.Inf(1))},
		{exact128(math.Copysign(0, -1)), exact128(math.Inf(-1)), exact128(math.Inf(1))},

		// ±0.Pow(+Inf) = +0
		{exact128(0), exact128(math.Inf(1)), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Inf(1)), exact128(0)},

		// ±0.Pow(b) = +Inf for finite b < 0 and not an odd integer
		{exact128(0), exact128(-2), exact128(math.Inf(1))},
		{exact128(math.Copysign(0, -1)), exact128(-2), exact128(math.Inf(1))},
		{exact128(math.Copysign(0, -1)), exact128(-0.5), exact128(math.Inf(1))},

		// ±0.Pow(b) = ±0 for b an odd integer > 0
		{exact128(0), exact128(3), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(3), exact128(math.Copysign(0, -1))},

		// ±0.Pow(b) = +0 for finite b > 0 and not an odd integer
		{exact128(0), exact128(2), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(2), exact128(0)},

		// -1.Pow(±Inf) = 1
		{exact128(-1), exact128(math.Inf(1)), exact128(1)},
		{exact128(-1), exact128(math.Inf(-1)), exact128(1)},

		// a.Pow(+Inf) = +Inf for |a| > 1
		{exact128(2), exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(-2), exact128(math.Inf(1)), exact128(math.Inf(1))},

		// a.Pow(-Inf) = +0 for |a| > 1
		{exact128(2), exact128(math.Inf(-1)), exact128(0)},
		{exact128(-2), exact128(math.Inf(-1)), exact128(0)},

		// a.Pow(+Inf) = +0 for |a| < 1
		{exact128(0.5), exact128(math.Inf(1)), exact128(0)},
		{exact128(-0.5), exact128(math.Inf(1)), exact128(0)},

		// a.Pow(-Inf) = +Inf for |a| < 1
		{exact128(0.5), exact128(math.Inf(-1)), exact128(math.Inf(1))},
		{exact128(-0.5), exact128(math.Inf(-1)), exact128(math.Inf(1))},

		// +Inf.Pow(b) = +Inf for b > 0
		{exact128(math.Inf(1)), exact128(2), exact128(math.Inf(1))},

		// +Inf.Pow(b) = +0 for b < 0
		{exact128(math.Inf(1)), exact128(-2), exact128(0)},

		// -Inf.Pow(b) = (-0).Pow(-b)
		{exact128(math.Inf(-1)), exact128(3), exact128(math.Inf(-1))},
		{exact128(math.Inf(-1)), exact128(2), exact128(math.Inf(1))},

		// a.Pow(b) = NaN for finite a < 0 and finite non-integer b
		{exact128(-2), exact128(0.5), exact128(math.NaN())},
		{exact128(-2), exact128(-0.5), exact128(math.NaN())},

		// overflow and underflow
		{exact128(3), exact128(1e6), exact128(math.Inf(1))},
		{exact128(3), exact128(-1e6), exact128(0)},
		{exact128(-3), exact128(1000001), exact128(math.Inf(-1))},
		{exact128(-3), exact128(1000000), exact128(math.Inf(1))},
		{exact128(-3), exact128(-1000001), exact128(math.Copysign(0, -1))},
		{exact128(0.5), exact128(1e6), exact128(0)},
		{exact128(2), exact128(16384), exact128(math.Inf(1))},
		{exact128(2), exact128(-16495), exact128(0)},    // a tie of the smallest subnormal number and 0
		{exact128(2), exact128(-16494), Float128{0, 1}}, // the smallest subnormal number
		{exact128(0.5), exact128(0x1p100), exact128(0)}, // powers of two with the huge exponents
		{exact128(0.5), exact128(-0x1p100), exact128(math.Inf(1))},
		{exact128(-2), exact128(0x1p100), exact128(math.Inf(1))},
		{exact128(-0.5), exact128(-0x1p100 - 0x1p-12), exact128(math.Inf(1))},

		// the results of the exact products, which are the midpoints of two adjacent Float128 values
		// (2**57-1)**2 = 2**114 - 2**58 + 1 is rounded to even: 2**114 - 2**58. (exact128 of the constant 2**57-1 is rounded.)
		{exact128(0x1p57).Sub(exact128(1)), exact128(2), exact128(0x1p114).Sub(exact128(0x1p58))},
		{exact128(-0x1p57).Add(exact128(1)), exact128(2), exact128(0x1p114).Sub(exact128(0x1p58))},
		// (2**57+1)**2 = 2**114 + 2**58 + 1 is rounded to even: 2**114 + 2**58.
		{exact128(0x1p57).Add(exact128(1)), exact128(2), exact128(0x1p114).Add(exact128(0x1p58))},

		// exact results
		{exact128(4), exact128(1.5), exact128(8)},
		{exact128(4), exact128(-0.5), exact128(0.5)},
		{exact128(0.25), exact128(-1.5), exact128(8)},
	}

	for _, tt := range strictTests {
		got := tt.x.Pow(tt.y)
		if !eq128(got, tt.want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

// TestFloat128_PowAccuracy requires the correctly rounded result for every vector of the test data.
func TestFloat128_PowAccuracy(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/pow128.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

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
		if len(fields) != 3 || len(fields[0]) != 32 || len(fields[1]) != 32 || len(fields[2]) != 32 {
			t.Fatalf("malformed line: %q", sc.Text())
		}
		x, y, want := parse(fields[0]), parse(fields[1]), parse(fields[2])
		if got := x.Pow(y); !eq128(got, want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", x, y, got, want)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
}

func BenchmarkFloat128_Pow(b *testing.B) {
	for _, tt := range []struct {
		name string
		x, y Float128
	}{
		{"square", exact128(1.5), exact128(2)},    // small integer exponents are exact
		{"integer", exact128(1.5), exact128(100)}, //
		{"negative-integer", exact128(1.5), exact128(-100)},
		{"half", exact128(5), exact128(0.5)},           // Sqrt
		{"fraction", exact128(5), exact128(1.5)},       // exp(y log(x))
		{"large", exact128(1.5), exact128(12345)},      // exp(y log(x))
		{"negative-base", exact128(-1.5), exact128(7)}, // the sign is restored
		{"overflow", exact128(10), exact128(5000)},     // +Inf
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Pow(tt.y))
			}
		})
	}
}
