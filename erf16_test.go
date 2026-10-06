package floats

import (
	"bufio"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"runtime"
	"testing"
)

func TestFloat16_Erf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Erf(0)},
		{exact16(0x1p-14), math.Erf(0x1p-14)},
		{exact16(1), math.Erf(1)},
		{exact16(2), math.Erf(2)},
		{exact16(3), math.Erf(3)},
		{exact16(4), math.Erf(4)},
		{exact16(11), math.Erf(11)},
	}

	for _, tt := range tests {
		got := tt.x.Erf()
		if !close16(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(1)},
		{exact16(math.Inf(-1)), exact16(-1)},
		{exact16(math.NaN()), exact16(math.NaN())},
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{NewFloat16FromBits(0x0001), NewFloat16FromBits(0x0001)},
		{NewFloat16FromBits(0x8001), NewFloat16FromBits(0x8001)},

		// erf(x) is rounded to 1 for |x| >= 2.5957
		{NewFloat16FromBits(0x4130), NewFloat16FromBits(0x3bff)},
		{NewFloat16FromBits(0x4131), exact16(1)},
		{NewFloat16FromBits(0xc130), NewFloat16FromBits(0xbbff)},
		{NewFloat16FromBits(0xc131), exact16(-1)},
		{NewFloat16FromBits(0x7bff), exact16(1)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erf()
		if !eq16(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat16_ErfAll checks Erf for all bit patterns of Float16.
// For every finite Float16 value, the exact erf is at least 2**-15 ulp away from the midpoint of two adjacent Float16 values
// (checked with mpmath), so math.Erf rounded to Float16 is correctly rounded.
func TestFloat16_ErfAll(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		x := NewFloat16FromBits(uint16(i))
		got, want := x.Erf(), NewFloat16(math.Erf(x.Float64().BuiltIn()))
		if uint16(got) != uint16(want) && !(got.IsNaN() && want.IsNaN()) {
			t.Errorf("Erf(%#04x) = %#04x; want %#04x", i, uint16(got), uint16(want))
		}
	}
}

// TestFloat16_ErfPoly checks the polynomials of Erf, whose errors are hidden by the rounding to Float16, with the math package.
func TestFloat16_ErfPoly(t *testing.T) {
	t.Parallel()
	const bound = 0x1p-34 // the relative errors of the polynomials are less than 2**-35 and 2**-56
	rnd := rand.New(rand.NewPCG(1, 2))
	for range 100000 {
		x := 0.125 + 2.4607*rnd.Float64() // [1/8, 2.5857)
		if got, want := erf16Poly(x), math.Erf(x); math.Abs(got-want) > bound*want {
			t.Errorf("erf16Poly(%v) = %v; want %v", x, got, want)
		}

		x = 0.125 * rnd.Float64() // [0, 1/8)
		c := &erf16SmallCoeffs
		u := x * x
		got := x * (c[0] + u*(c[1]+u*(c[2]+u*(c[3]+u*(c[4]+u*c[5])))))
		if want := math.Erf(x); math.Abs(got-want) > bound*want {
			t.Errorf("erf16SmallCoeffs(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat16_Erf(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float16
	}{
		{"small", NewFloat16(0.1)},  // |x| < 1/8
		{"medium", exact16(1.5)},    // 1/8 <= |x| < 2.6
		{"negative", exact16(-1.5)}, // the sign is restored
		{"saturated", exact16(3)},   // the result is 1
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erf())
			}
		})
	}
}

func TestFloat16_Erfc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0), math.Erfc(0)},
		{exact16(0x1p-14), math.Erfc(0x1p-14)},
		{exact16(1), math.Erfc(1)},
		{exact16(2), math.Erfc(2)},
		{exact16(-1), math.Erfc(-1)},
		{exact16(-2), math.Erfc(-2)},
		{exact16(0.5), math.Erfc(0.5)},
	}

	for _, tt := range tests {
		got := tt.x.Erfc()
		if !close16(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(math.Inf(1)), exact16(0)},
		{exact16(math.Inf(-1)), exact16(2)},
		{exact16(math.NaN()), exact16(math.NaN())},
		{exact16(0), exact16(1)},
		{exact16(math.Copysign(0, -1)), exact16(1)},

		// erfc(x) is rounded to 0 for x >= 3.9199
		{NewFloat16FromBits(0x43d6), NewFloat16FromBits(0x0001)},
		{NewFloat16FromBits(0x43d7), exact16(0)},
		{NewFloat16FromBits(0x7bff), exact16(0)},

		// erfc(x) is rounded to 2 for x <= -2.4668
		{NewFloat16FromBits(0xc0ee), NewFloat16FromBits(0x3fff)},
		{NewFloat16FromBits(0xc0ef), exact16(2)},
		{NewFloat16FromBits(0xfbff), exact16(2)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfc()
		if !eq16(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat16_ErfcAll checks Erfc for all bit patterns of Float16.
// For every finite Float16 value, the exact erfc is at least 2**-30 (relative) away from the midpoint of two adjacent
// Float16 values (checked with mpmath), so math.Erfc rounded to Float16 is correctly rounded.
func TestFloat16_ErfcAll(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		x := NewFloat16FromBits(uint16(i))
		got, want := x.Erfc(), NewFloat16(math.Erfc(x.Float64().BuiltIn()))
		if uint16(got) != uint16(want) && !(got.IsNaN() && want.IsNaN()) {
			t.Errorf("Erfc(%#04x) = %#04x; want %#04x", i, uint16(got), uint16(want))
		}
	}
}

// TestFloat16_ErfcPoly checks the polynomials of Erfc, whose errors are hidden by the rounding to Float16, with the math package.
func TestFloat16_ErfcPoly(t *testing.T) {
	t.Parallel()
	const bound = 0x1p-35 // the relative errors of the polynomials are less than 2**-36
	rnd := rand.New(rand.NewPCG(1, 2))
	for range 100000 {
		x := 0.125 + 3.79*rnd.Float64() // [1/8, 3.915)
		if got, want := erfc16Poly(x), math.Erfc(x); math.Abs(got-want) > bound*want {
			t.Errorf("erfc16Poly(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat16_Erfc(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float16
	}{
		{"small", NewFloat16(0.1)},  // |x| < 1/8
		{"medium", exact16(1.5)},    // 1/8 <= x < 3.92
		{"negative", exact16(-1.5)}, // 2 - erfc(-x)
		{"saturated", exact16(5)},   // the result is 0
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erfc())
			}
		})
	}
}

func TestFloat16_Erfinv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(-0.75), math.Erfinv(-0.75)},
		{exact16(-0.5), math.Erfinv(-0.5)},
		{exact16(-0.25), math.Erfinv(-0.25)},
		{exact16(0), math.Erfinv(0)},
		{exact16(0.25), math.Erfinv(0.25)},
		{exact16(0.5), math.Erfinv(0.5)},
		{exact16(0.75), math.Erfinv(0.75)},
	}

	for _, tt := range tests {
		got := tt.x.Erfinv()
		if !close16(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(1), exact16(math.Inf(1))},
		{exact16(-1), exact16(math.Inf(-1))},
		{exact16(2), exact16(math.NaN())},
		{exact16(-2), exact16(math.NaN())},
		{exact16(math.Inf(1)), exact16(math.NaN())},
		{exact16(math.Inf(-1)), exact16(math.NaN())},
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfinv()
		if !eq16(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat16_ErfinvAll requires the correctly rounded result for all the Float16 values in the test data,
// and checks that Erfinv is an odd function.
func TestFloat16_ErfinvAll(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/erfinv16.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	n := 0
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var x, want uint16
		if _, err := fmt.Sscanf(sc.Text(), "%x %x", &x, &want); err != nil {
			t.Fatalf("malformed line %q: %v", sc.Text(), err)
		}
		if got := NewFloat16FromBits(x).Erfinv(); uint16(got) != want {
			t.Errorf("Erfinv(%#04x) = %#04x; want %#04x", x, uint16(got), want)
		}
		if got := NewFloat16FromBits(x | 1<<15).Erfinv(); uint16(got) != want|1<<15 {
			t.Errorf("Erfinv(%#04x) = %#04x; want %#04x", x|1<<15, uint16(got), want|1<<15)
		}
		n++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if n != 0x3c00 {
		t.Errorf("the number of the test vectors is %d; want %d", n, 0x3c00)
	}

	// |x| > 1 is NaN.
	for i := 0x3c01; i < 0x8000; i++ {
		for _, sign := range []uint16{0, 1 << 15} {
			if got := NewFloat16FromBits(uint16(i) | sign).Erfinv(); !got.IsNaN() {
				t.Errorf("Erfinv(%#04x) = %v; want NaN", uint16(i)|sign, got)
			}
		}
	}
}

// TestFloat16_ErfinvPoly checks the polynomials of Erfinv, whose errors are hidden by the rounding to Float16, with the math package.
func TestFloat16_ErfinvPoly(t *testing.T) {
	t.Parallel()
	const bound = 0x1p-36 // the relative errors of the polynomials are less than 2**-38
	rnd := rand.New(rand.NewPCG(1, 2))
	for range 100000 {
		x := 0.125 + 0.375*rnd.Float64() // [1/8, 1/2)
		if got, want := erfinv16Poly(&erfinv16MidCoeffs[math.Float64bits(x)>>49-8160], x), math.Erfinv(x); math.Abs(got-want) > bound*want {
			t.Errorf("erfinv16Mid(%v) = %v; want %v", x, got, want)
		}

		tt := 0x1p-11 + (0.5-0x1p-11)*rnd.Float64() // t = 1 - x in [2**-11, 1/2)
		if got, want := erfinv16Poly(&erfinv16HiCoeffs[math.Float64bits(tt)>>49-8096], tt), math.Erfinv(1-tt); math.Abs(got-want) > bound*want {
			t.Errorf("erfinv16Hi(%v) = %v; want %v", tt, got, want)
		}

		x = 0.125 * rnd.Float64() // [0, 1/8)
		c := &erfinv16SmallCoeffs
		u := x * x
		got := x * (c[0] + u*(c[1]+u*(c[2]+u*(c[3]+u*(c[4]+u*c[5])))))
		if want := math.Erfinv(x); math.Abs(got-want) > bound*want {
			t.Errorf("erfinv16SmallCoeffs(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat16_Erfinv(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float16
	}{
		{"small", NewFloat16(0.1)},           // |x| < 1/8
		{"medium", exact16(0.25)},            // 1/8 <= |x| < 1/2
		{"large", exact16(0.75)},             // 1/2 <= |x| < 1
		{"close-to-one", NewFloat16(0.9995)}, // the polynomial of the shortest segment
		{"negative", exact16(-0.75)},         // the sign is restored
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erfinv())
			}
		})
	}
}

func TestFloat16_Erfcinv(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want float64
	}{
		{exact16(0.25), math.Erfcinv(0.25)},
		{exact16(0.5), math.Erfcinv(0.5)},
		{exact16(0.75), math.Erfcinv(0.75)},
		{exact16(1), math.Erfcinv(1)},
		{exact16(1.25), math.Erfcinv(1.25)},
		{exact16(1.5), math.Erfcinv(1.5)},
		{exact16(1.75), math.Erfcinv(1.75)},
	}

	for _, tt := range tests {
		got := tt.x.Erfcinv()
		if !close16(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float16
		want Float16
	}{
		// special cases
		{exact16(0), exact16(math.Inf(1))},
		{exact16(2), exact16(math.Inf(-1))},
		{exact16(3), exact16(math.NaN())},
		{exact16(-1), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(math.NaN())},
		{exact16(math.Inf(1)), exact16(math.NaN())},
		{exact16(math.Inf(-1)), exact16(math.NaN())},
		{exact16(math.Copysign(0, -1)), exact16(math.Inf(1))},
		{exact16(1), exact16(0)},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfcinv()
		if !eq16(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat16_ErfcinvAll requires the correctly rounded result for all the Float16 values in the test data,
// and checks Erfcinv(x) = -Erfcinv(2-x).
func TestFloat16_ErfcinvAll(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/erfcinv16.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	n := 0
	wants := make(map[uint16]uint16)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var x, want uint16
		if _, err := fmt.Sscanf(sc.Text(), "%x %x", &x, &want); err != nil {
			t.Fatalf("malformed line %q: %v", sc.Text(), err)
		}
		if got := NewFloat16FromBits(x).Erfcinv(); uint16(got) != want {
			t.Errorf("Erfcinv(%#04x) = %#04x; want %#04x", x, uint16(got), want)
		}
		wants[x] = want
		n++
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	if n != 0x3c00 {
		t.Errorf("the number of the test vectors is %d; want %d", n, 0x3c00)
	}

	// erfcinv(x) = -erfcinv(2-x) for 1 < x < 2, and 2 - x is exactly representable.
	for i := 0x3c01; i < 0x4000; i++ {
		x := NewFloat16FromBits(uint16(i))
		c := NewFloat16(2 - x.Float64().BuiltIn())
		if want := wants[uint16(c.Bits())] | 1<<15; uint16(x.Erfcinv()) != want {
			t.Errorf("Erfcinv(%#04x) = %#04x; want %#04x", i, uint16(x.Erfcinv()), want)
		}
	}

	// x < 0 and x > 2 are NaN.
	for i := 0x4001; i < 0x8000; i++ {
		if got := NewFloat16FromBits(uint16(i)).Erfcinv(); !got.IsNaN() {
			t.Errorf("Erfcinv(%#04x) = %v; want NaN", i, got)
		}
	}
	for i := 0x8001; i < 0x10000; i++ {
		if got := NewFloat16FromBits(uint16(i)).Erfcinv(); !got.IsNaN() {
			t.Errorf("Erfcinv(%#04x) = %v; want NaN", i, got)
		}
	}
}

func BenchmarkFloat16_Erfcinv(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float16
	}{
		{"tiny", NewFloat16(0.0001)},   // x <= 1/2: the polynomials of Float32 Erfinv
		{"small", exact16(0.25)},       // x <= 1/2
		{"medium", exact16(0.75)},      // 1/2 < x < 7/8
		{"near-one", NewFloat16(0.95)}, // 7/8 <= x <= 1
		{"large", exact16(1.5)},        // 1 < x < 2, the sign is restored
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Erfcinv())
			}
		})
	}
}
