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

func TestFloat32_Pow(t *testing.T) {
	tests := []struct {
		x    Float32
		y    Float32
		want float64
	}{
		{exact32(2), exact32(3), math.Pow(2, 3)},
		{exact32(5), exact32(0.5), math.Pow(5, 0.5)},
		{exact32(5), exact32(1.5), math.Pow(5, 1.5)},
		{exact32(5), exact32(-1.5), math.Pow(5, -1.5)},
	}

	for _, tt := range tests {
		got := tt.x.Pow(tt.y)
		if !close32(got, tt.want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		y    Float32
		want Float32
	}{
		{exact32(2), exact32(1 << 64), exact32(math.Inf(1))},  // overflow
		{exact32(-2), exact32(1 << 64), exact32(math.Inf(1))}, // overflow
		{exact32(0.5), exact32(1 << 64), exact32(0)},          // underflow
		{exact32(-0.5), exact32(1 << 64), exact32(0)},         // underflow
		{exact32(-1), exact32(1 << 64), exact32(1)},

		// special cases
		// a.Pow(±0) = 1 for any a
		{exact32(2), exact32(0), exact32(1)},
		{exact32(2), exact32(math.Copysign(0, -1)), exact32(1)},
		{exact32(-2), exact32(0), exact32(1)},
		{exact32(-2), exact32(math.Copysign(0, -1)), exact32(1)},
		{exact32(math.Inf(1)), exact32(0), exact32(1)},
		{exact32(math.Inf(1)), exact32(math.Copysign(0, -1)), exact32(1)},
		{exact32(math.NaN()), exact32(0), exact32(1)},
		{exact32(math.NaN()), exact32(math.Copysign(0, -1)), exact32(1)},

		// 1.Pow(b) = 1 for any b
		{exact32(1), exact32(3), exact32(1)},
		{exact32(1), exact32(-3), exact32(1)},
		{exact32(1), exact32(math.Inf(1)), exact32(1)},
		{exact32(1), exact32(math.Inf(-1)), exact32(1)},
		{exact32(1), exact32(math.NaN()), exact32(1)},

		// a.Pow(1) = a for any a
		{exact32(2), exact32(1), exact32(2)},
		{exact32(-2), exact32(1), exact32(-2)},
		{exact32(math.Inf(1)), exact32(1), exact32(math.Inf(1))},
		{exact32(math.NaN()), exact32(1), exact32(math.NaN())},

		// NaN.Pow(b) = NaN
		{exact32(math.NaN()), exact32(3), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(-3), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.Inf(1)), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.Inf(-1)), exact32(math.NaN())},

		// a.Pow(NaN) = NaN
		{exact32(2), exact32(math.NaN()), exact32(math.NaN())},
		{exact32(-2), exact32(math.NaN()), exact32(math.NaN())},
		{exact32(math.Inf(1)), exact32(math.NaN()), exact32(math.NaN())},
		{exact32(math.Inf(-1)), exact32(math.NaN()), exact32(math.NaN())},

		// ±0.Pow(b) = ±Inf for b an odd integer < 0
		{exact32(0), exact32(-3), exact32(math.Inf(1))},
		{exact32(math.Copysign(0, -1)), exact32(-3), exact32(math.Inf(-1))},

		// ±0.Pow(-Inf) = +Inf
		{exact32(0), exact32(math.Inf(-1)), exact32(math.Inf(1))},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(-1)), exact32(math.Inf(1))},

		// ±0.Pow(+Inf) = +0
		{exact32(0), exact32(math.Inf(1)), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(1)), exact32(0)},

		// ±0.Pow(b) = +Inf for finite b < 0 and not an odd integer
		{exact32(0), exact32(-2), exact32(math.Inf(1))},
		{exact32(math.Copysign(0, -1)), exact32(-2), exact32(math.Inf(1))},
		{exact32(math.Copysign(0, -1)), exact32(-0.5), exact32(math.Inf(1))},

		// ±0.Pow(b) = ±0 for b an odd integer > 0
		{exact32(0), exact32(3), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(3), exact32(math.Copysign(0, -1))},

		// ±0.Pow(b) = +0 for finite b > 0 and not an odd integer
		{exact32(0), exact32(2), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(2), exact32(0)},

		// -1.Pow(±Inf) = 1
		{exact32(-1), exact32(math.Inf(1)), exact32(1)},
		{exact32(-1), exact32(math.Inf(-1)), exact32(1)},

		// a.Pow(+Inf) = +Inf for |a| > 1
		{exact32(2), exact32(math.Inf(1)), exact32(math.Inf(1))},
		{exact32(-2), exact32(math.Inf(1)), exact32(math.Inf(1))},

		// a.Pow(-Inf) = +0 for |a| > 1
		{exact32(2), exact32(math.Inf(-1)), exact32(0)},
		{exact32(-2), exact32(math.Inf(-1)), exact32(0)},

		// a.Pow(+Inf) = +0 for |a| < 1
		{exact32(0.5), exact32(math.Inf(1)), exact32(0)},
		{exact32(-0.5), exact32(math.Inf(1)), exact32(0)},

		// a.Pow(-Inf) = +Inf for |a| < 1
		{exact32(0.5), exact32(math.Inf(-1)), exact32(math.Inf(1))},
		{exact32(-0.5), exact32(math.Inf(-1)), exact32(math.Inf(1))},

		// +Inf.Pow(b) = +Inf for b > 0
		{exact32(math.Inf(1)), exact32(2), exact32(math.Inf(1))},

		// +Inf.Pow(b) = +0 for b < 0
		{exact32(math.Inf(1)), exact32(-2), exact32(0)},

		// -Inf.Pow(b) = (-0).Pow(-b)
		{exact32(math.Inf(-1)), exact32(3), exact32(math.Inf(-1))},
		{exact32(math.Inf(-1)), exact32(2), exact32(math.Inf(1))},

		// a.Pow(b) = NaN for finite a < 0 and finite non-integer b
		{exact32(-2), exact32(0.5), exact32(math.NaN())},
		{exact32(-2), exact32(-0.5), exact32(math.NaN())},

		// exact results on the midpoint of two adjacent Float32 values are rounded to even.
		{NewFloat32FromBits(0x41300000), NewFloat32FromBits(0x40e00000), NewFloat32FromBits(0x4b94ace2)}, // 11**7 = 19487171
		{NewFloat32FromBits(0xc1300000), NewFloat32FromBits(0x40e00000), NewFloat32FromBits(0xcb94ace2)}, // (-11)**7
		{NewFloat32FromBits(0x3f800800), NewFloat32FromBits(0x40000000), NewFloat32FromBits(0x3f801000)}, // (1 + 2**-12)**2 = 1 + 2**-11 + 2**-24
		{NewFloat32FromBits(0x27400000), NewFloat32FromBits(0x40400000), NewFloat32FromBits(0x0000000e)}, // (3 * 2**-50)**3 = 13.5 * 2**-149

		// exact results on the midpoint with non-integer y
		{NewFloat32FromBits(0x47810080), NewFloat32FromBits(0x3fc00000), NewFloat32FromBits(0x4b818180)}, // (257**2)**1.5 = 257**3
		{NewFloat32FromBits(0x4664c400), NewFloat32FromBits(0x3fe00000), NewFloat32FromBits(0x4b94ace2)}, // (11**4)**(7/4) = 11**7
		{NewFloat32FromBits(0x4264c400), NewFloat32FromBits(0x3fe00000), NewFloat32FromBits(0x4494ace2)}, // (11**4 * 2**-8)**(7/4) = 11**7 * 2**-14
		{NewFloat32FromBits(0x0f100000), NewFloat32FromBits(0x3fc00000), NewFloat32FromBits(0x0000000e)}, // (9 * 2**-100)**1.5 = 13.5 * 2**-149

		// 2**-150 is the midpoint of 0 and the smallest subnormal.
		{exact32(2), exact32(-150), exact32(0)},
		{exact32(0.5), exact32(150), exact32(0)},
		{exact32(4), exact32(-75), exact32(0)},
		{exact32(0x1p-100), exact32(1.5), exact32(0)},
		{exact32(2), exact32(-149.5), NewFloat32FromBits(0x00000001)},
		{exact32(2), exact32(-149), NewFloat32FromBits(0x00000001)},
	}

	for _, tt := range strictTests {
		got := tt.x.Pow(tt.y)
		if !eq32(got, tt.want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

// TestFloat32_PowAccuracy checks that the error of Pow is less than 1 ulp,
// and that almost all results are correctly rounded.
// The test data is generated by scripts/gen_pow_testdata.py.
func TestFloat32_PowAccuracy(t *testing.T) {
	f, err := os.Open("testdata/pow32.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	parse := func(s string) Float32 {
		v, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			t.Fatal(err)
		}
		return NewFloat32FromBits(uint32(v))
	}

	var total, misrounded int
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		x, y, want := parse(fields[0]), parse(fields[1]), parse(fields[2])
		got := x.Pow(y)
		total++
		if !eq32(got, want) {
			misrounded++
		}
		if !within1ulp32(got, want) {
			t.Errorf("Pow(%v, %v) = %v; want %v", x, y, got, want)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d of %d results are not correctly rounded", misrounded, total)

	// The relative error is less than 2**-40,
	// so the result may be misrounded if the exact value is very close to a midpoint.
	if misrounded*100 > total {
		t.Errorf("%d of %d results are not correctly rounded; want at most 1%%", misrounded, total)
	}
}

func BenchmarkFloat32_Pow(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Pow(x))
	}
}
