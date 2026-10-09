package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

type bf16Func struct {
	name string
	f    func(BFloat16) BFloat16 // the BFloat16 implementation
	f256 func(Float256) Float256 // the reference implementation in Float256
	f16  func(Float16) Float16   // for the special values
}

var bf16Funcs = []bf16Func{
	{"Exp", BFloat16.Exp, Float256.Exp, Float16.Exp},
	{"Exp2", BFloat16.Exp2, Float256.Exp2, Float16.Exp2},
	{"Expm1", BFloat16.Expm1, Float256.Expm1, Float16.Expm1},
	{"Log", BFloat16.Log, Float256.Log, Float16.Log},
	{"Log2", BFloat16.Log2, Float256.Log2, Float16.Log2},
	{"Log10", BFloat16.Log10, Float256.Log10, Float16.Log10},
	{"Log1p", BFloat16.Log1p, Float256.Log1p, Float16.Log1p},
	{"Sin", BFloat16.Sin, Float256.Sin, Float16.Sin},
	{"Cos", BFloat16.Cos, Float256.Cos, Float16.Cos},
	{"Tan", BFloat16.Tan, Float256.Tan, Float16.Tan},
	{"Asin", BFloat16.Asin, Float256.Asin, Float16.Asin},
	{"Acos", BFloat16.Acos, Float256.Acos, Float16.Acos},
	{"Atan", BFloat16.Atan, Float256.Atan, Float16.Atan},
	{"Sinh", BFloat16.Sinh, Float256.Sinh, Float16.Sinh},
	{"Cosh", BFloat16.Cosh, Float256.Cosh, Float16.Cosh},
	{"Tanh", BFloat16.Tanh, Float256.Tanh, Float16.Tanh},
	{"Asinh", BFloat16.Asinh, Float256.Asinh, Float16.Asinh},
	{"Acosh", BFloat16.Acosh, Float256.Acosh, Float16.Acosh},
	{"Atanh", BFloat16.Atanh, Float256.Atanh, Float16.Atanh},
	{"Cbrt", BFloat16.Cbrt, Float256.Cbrt, Float16.Cbrt},
	{"Gamma", BFloat16.Gamma, Float256.Gamma, Float16.Gamma},
	{"Erf", BFloat16.Erf, Float256.Erf, Float16.Erf},
	{"Erfc", BFloat16.Erfc, Float256.Erfc, Float16.Erfc},
	{"Erfinv", BFloat16.Erfinv, Float256.Erfinv, Float16.Erfinv},
	{"Erfcinv", BFloat16.Erfcinv, Float256.Erfcinv, Float16.Erfcinv},
	{"J0", BFloat16.J0, Float256.J0, Float16.J0},
	{"J1", BFloat16.J1, Float256.J1, Float16.J1},
	{"Y0", BFloat16.Y0, Float256.Y0, Float16.Y0},
	{"Y1", BFloat16.Y1, Float256.Y1, Float16.Y1},
}

func bf16FuncByName(name string) *bf16Func {
	for i := range bf16Funcs {
		if bf16Funcs[i].name == name {
			return &bf16Funcs[i]
		}
	}
	return nil
}

// TestBFloat16_MathKnown checks the correctly rounded results calculated by mpmath.
// They include the cases where the result in float64 is too close to the midpoint.
func TestBFloat16_MathKnown(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   BFloat16
		want BFloat16
	}{
		{"Exp", 0x3f80, 0x402e},
		{"Exp", 0xbf80, 0x3ebc},
		{"Exp", 0x4120, 0x46ac},
		{"Exp", 0xc2b0, 0x0042},
		{"Exp", 0x42b2, 0x7f80},
		{"Exp", 0x0001, 0x3f80},
		{"Exp2", 0x3f80, 0x4000},
		{"Exp2", 0xc306, 0x0000},
		{"Exp2", 0xc2fc, 0x0080},
		{"Exp2", 0x4300, 0x7f80},
		{"Exp2", 0xbf00, 0x3f35},
		{"Expm1", 0x3f80, 0x3fdc},
		{"Expm1", 0xb000, 0xb000},
		{"Expm1", 0x3000, 0x3000},
		{"Expm1", 0x42b2, 0x7f80},
		{"Log", 0x4000, 0x3f31},
		{"Log", 0x3f00, 0xbf31},
		{"Log", 0x0001, 0xc2b8},
		{"Log", 0x7f7f, 0x42b1},
		{"Log", 0x3f81, 0x3bff},
		{"Log2", 0x4000, 0x3f80},
		{"Log2", 0x4100, 0x4040},
		{"Log2", 0x3f81, 0x3c38},
		{"Log2", 0x0001, 0xc305},
		{"Log10", 0x42c8, 0x4000},
		{"Log10", 0x4120, 0x3f80},
		{"Log10", 0x3f81, 0x3b5d},
		{"Log1p", 0x3f80, 0x3f31},
		{"Log1p", 0xbf00, 0xbf31},
		{"Log1p", 0x3b80, 0x3b80},
		{"Log1p", 0x0001, 0x0001},
		{"Sin", 0x3f80, 0x3f57},
		{"Sin", 0x4049, 0x3a7e},
		{"Sin", 0x7f7f, 0xbf7d},
		{"Sin", 0x5000, 0x3f52},
		{"Sin", 0x3380, 0x3380},
		{"Cos", 0x3f80, 0x3f0a},
		{"Cos", 0x4049, 0xbf80},
		{"Cos", 0x7f7f, 0x3e26},
		{"Cos", 0x5000, 0x3f13},
		{"Cos", 0x0001, 0x3f80},
		{"Tan", 0x3f80, 0x3fc7},
		{"Tan", 0x3fc9, 0x4501},
		{"Tan", 0x4049, 0xba7e},
		{"Tan", 0x7f7f, 0xc0c3},
		{"Asin", 0x3f00, 0x3f06},
		{"Asin", 0xbf00, 0xbf06},
		{"Asin", 0x3f7f, 0x3fbe},
		{"Asin", 0x0001, 0x0001},
		{"Acos", 0x3f00, 0x3f86},
		{"Acos", 0xbf00, 0x4006},
		{"Acos", 0x3f7f, 0x3db5},
		{"Acos", 0x0001, 0x3fc9},
		{"Atan", 0x3f80, 0x3f49},
		{"Atan", 0x7f7f, 0x3fc9},
		{"Atan", 0x0001, 0x0001},
		{"Atan", 0xc040, 0xbfa0},
		{"Sinh", 0x3f80, 0x3f96},
		{"Sinh", 0xc040, 0xc120},
		{"Sinh", 0x4120, 0x462c},
		{"Sinh", 0x4300, 0x7f80},
		{"Cosh", 0x3f80, 0x3fc6},
		{"Cosh", 0xc040, 0x4121},
		{"Cosh", 0x4120, 0x462c},
		{"Cosh", 0x4300, 0x7f80},
		{"Tanh", 0x3f80, 0x3f43},
		{"Tanh", 0xc040, 0xbf7f},
		{"Tanh", 0x3000, 0x3000},
		{"Tanh", 0x4120, 0x3f80},
		{"Asinh", 0x3f80, 0x3f62},
		{"Asinh", 0xc040, 0xbfe9},
		{"Asinh", 0x7f7f, 0x42b3},
		{"Asinh", 0x0001, 0x0001},
		{"Acosh", 0x4000, 0x3fa9},
		{"Acosh", 0x3f81, 0x3e00},
		{"Acosh", 0x7f7f, 0x42b3},
		{"Atanh", 0x3f00, 0x3f0d},
		{"Atanh", 0xbf00, 0xbf0d},
		{"Atanh", 0x3f7f, 0x4048},
		{"Atanh", 0x0001, 0x0001},
		{"Cbrt", 0x4100, 0x4000},
		{"Cbrt", 0x4000, 0x3fa1},
		{"Cbrt", 0x0001, 0x294b},
		{"Cbrt", 0x7f7f, 0x54cb},
		{"Gamma", 0x3f00, 0x3fe3},
		{"Gamma", 0x4040, 0x4000},
		{"Gamma", 0x4100, 0x459e},
		{"Gamma", 0x4110, 0x471e},
		{"Gamma", 0xbf00, 0xc063},
		{"Gamma", 0xc020, 0xbf72},
		{"Gamma", 0x4200, 0x77cb},
		{"Gamma", 0x0001, 0x7f80},
		{"Erf", 0x3f00, 0x3f05},
		{"Erf", 0x3f80, 0x3f58},
		{"Erf", 0xc000, 0xbf7f},
		{"Erf", 0x0001, 0x0001},
		{"Erfc", 0x3f00, 0x3ef6},
		{"Erfc", 0x3f80, 0x3e21},
		{"Erfc", 0x4080, 0x3284},
		{"Erfc", 0xc000, 0x3fff},
		{"Erfc", 0x4140, 0x0000},
		{"Erfinv", 0x3f00, 0x3ef4},
		{"Erfinv", 0xbf00, 0xbef4},
		{"Erfinv", 0x3f7f, 0x4003},
		{"Erfinv", 0x0001, 0x0001},
		{"Erfcinv", 0x3f00, 0x3ef4},
		{"Erfcinv", 0x3f80, 0x0000},
		{"Erfcinv", 0x3fc0, 0xbef4},
		{"Erfcinv", 0x0001, 0x4117},
		{"Erfcinv", 0x3c00, 0x3ff1},
		{"Erfcinv", 0x3bff, 0x3ff1},
		{"J0", 0x3f80, 0x3f44},
		{"J0", 0x4019, 0x3bf2},
		{"J0", 0x4140, 0x3d43},
		{"J0", 0x7f7f, 0x9eef},
		{"J1", 0x3f80, 0x3ee1},
		{"J1", 0xbf80, 0xbee1},
		{"J1", 0x0003, 0x0001},
		{"J1", 0x8003, 0x8001},
		{"J1", 0x00ff, 0x007f},
		{"J1", 0x80ff, 0x807f},
		{"J1", 0x00fd, 0x007e},
		{"J1", 0x4040, 0x3eae},
		{"Y0", 0x3f80, 0x3db5},
		{"Y0", 0x4019, 0x3f03},
		{"Y0", 0x4140, 0xbe67},
		{"Y0", 0x0001, 0xc26b},
		{"Y1", 0x3f80, 0xbf48},
		{"Y1", 0x4059, 0x3ecd},
		{"Y1", 0x4140, 0xbd6a},
		{"Y1", 0x0001, 0xff80},
	}
	for _, tt := range tests {
		f := bf16FuncByName(tt.name)
		if got := f.f(tt.in); got != tt.want {
			t.Errorf("BFloat16(%#04x).%s() = %#04x, want %#04x", uint16(tt.in), tt.name, uint16(got), uint16(tt.want))
		}
	}
}

// TestBFloat16_MathFallback checks that the results of the float64 calculation
// are the same as the ones of Float256 for a part of the inputs.
func TestBFloat16_MathFallback(t *testing.T) {
	t.Parallel()
	for _, f := range bf16Funcs {
		t.Run(f.name, func(t *testing.T) {
			t.Parallel()
			check := func(i int) {
				a := BFloat16(i)
				if f.name == "J1" && a&^signMaskbf16 < 0x0100 {
					// the exact value is a bit smaller than a/2, but the result of Float256 is a/2.
					return
				}
				got := f.f(a)
				want := f.f256(a.Float256()).BFloat16()
				if !eqbf16(got, want) {
					t.Fatalf("BFloat16(%#04x).%s() = %#04x, want %#04x", i, f.name, uint16(got), uint16(want))
				}
			}
			for i := 0; i < 1<<16; i += 13 {
				check(i)
			}
			// the values around the special points
			for _, i := range []int{0, 0x8000, 0x0001, 0x8001, 0x007f, 0x0080, 0x3f80, 0xbf80, 0x3f7f, 0x3f81, 0x7f7f, 0xff7f, 0x7f80, 0xff80, 0x7fc0, 0x4000, 0xc000, 0x3f00, 0xbf00} {
				check(i)
			}
		})
	}
}

// TestBFloat16_MathSpecial checks that the special cases are the same as Float16's.
func TestBFloat16_MathSpecial(t *testing.T) {
	t.Parallel()
	inputs := []float64{math.NaN(), math.Inf(1), math.Inf(-1), 0, math.Copysign(0, -1)}
	for _, f := range bf16Funcs {
		for _, x := range inputs {
			got := f.f(NewBFloat16(x))
			want := f.f16(NewFloat16(x)).BFloat16()
			if !eqbf16(got, want) {
				t.Errorf("%s(%v) = %#04x, want %#04x", f.name, x, uint16(got), uint16(want))
			}
		}
	}
}

func TestBFloat16_MathMisc(t *testing.T) {
	t.Parallel()
	// Pow
	pow := []struct{ x, y, want BFloat16 }{
		{0x3f88, 0x4000, 0x3f90}, // 1.0625**2 = 1.12890625 is the tie of 1.125 and 1.1328125; rounds to even
		{0x4000, 0x4040, 0x4100}, // 2**3
		{0x4040, 0xc000, 0x3de4}, // 3**-2
		{0x0000, 0xbf80, 0x7f80}, // 0**-1
		{0x3f80, 0x7fc0, 0x3f80}, // 1**NaN
		{0xbf80, 0x7f80, 0x3f80}, // (-1)**Inf
	}
	for _, tt := range pow {
		if got := tt.x.Pow(tt.y); got != tt.want {
			t.Errorf("BFloat16(%#04x).Pow(%#04x) = %#04x, want %#04x", uint16(tt.x), uint16(tt.y), uint16(got), uint16(tt.want))
		}
	}
	// Atan2
	if got, want := BFloat16(0x3f80).Atan2(0x3f80), NewBFloat16(math.Pi/4); got != want {
		t.Errorf("Atan2(1, 1) = %#04x, want %#04x", uint16(got), uint16(want))
	}
	if got, want := BFloat16(0x0000).Atan2(0xbf80), NewBFloat16(math.Pi); got != want {
		t.Errorf("Atan2(0, -1) = %#04x, want %#04x", uint16(got), uint16(want))
	}
	// Jn of n = 0, 1, -1
	for _, tt := range []struct {
		n       int
		x, want BFloat16
	}{{0, 0x3f80, BFloat16(0x3f80).J0()}, {1, 0x80ff, 0x807f}, {1, 0x00ff, 0x007f}, {-1, 0x00ff, 0x807f}, {-1, 0x80ff, 0x007f}, {1, 0x3f80, BFloat16(0x3f80).J1()}} {
		if got := tt.x.Jn(tt.n); got != tt.want {
			t.Errorf("BFloat16(%#04x).Jn(%d) = %#04x, want %#04x", uint16(tt.x), tt.n, uint16(got), uint16(tt.want))
		}
	}
	// Atan2 of the tiny ratio: atan(t) is a bit smaller than t, so the tie is rounded toward zero.
	for _, tt := range []struct{ x, y, want BFloat16 }{
		{0x0003, 0x4000, 0x0001}, // 3/2 units
		{0x0007, 0x4000, 0x0003},
		{0x8003, 0x4000, 0x8001},
		{0x0001, 0x4000, 0x0000},
		{0x8001, 0x4000, 0x8000},
		{0x0005, 0x4000, 0x0002},
		{0x000f, 0x4120, 0x0001}, // 15/10 units
	} {
		if got := tt.x.Atan2(tt.y); got != tt.want {
			t.Errorf("BFloat16(%#04x).Atan2(%#04x) = %#04x, want %#04x", uint16(tt.x), uint16(tt.y), uint16(got), uint16(tt.want))
		}
	}
	// Hypot
	if got, want := HypotBF16(0x4040, 0x4080), BFloat16(0x40a0); got != want {
		t.Errorf("HypotBF16(3, 4) = %#04x, want %#04x", uint16(got), uint16(want))
	}
	if got := HypotBF16(0x7f7f, 0x7f7f); got != uvinfbf16 {
		t.Errorf("HypotBF16(max, max) = %#04x, want Inf", uint16(got))
	}
	// Jn, Yn
	if got, want := BFloat16(0x4000).Jn(2), NewBFloat16(math.Jn(2, 2)); got != want {
		t.Errorf("Jn(2, 2) = %#04x, want %#04x", uint16(got), uint16(want))
	}
	if got, want := BFloat16(0x4000).Yn(2), NewBFloat16(math.Yn(2, 2)); got != want {
		t.Errorf("Yn(2, 2) = %#04x, want %#04x", uint16(got), uint16(want))
	}
	// Lgamma
	if v, s := BFloat16(0xc020).Lgamma(); v != NewBFloat16(math.Log(math.Abs(math.Gamma(-2.5)))) || s != -1 {
		t.Errorf("Lgamma(-2.5) = %#04x, %d", uint16(v), s)
	}
	if v, s := BFloat16(0x4000).Lgamma(); v != 0 || s != 1 {
		t.Errorf("Lgamma(2) = %#04x, %d", uint16(v), s)
	}
	// Sincos
	s, c := BFloat16(0x3f80).Sincos()
	if s != BFloat16(0x3f80).Sin() || c != BFloat16(0x3f80).Cos() {
		t.Errorf("Sincos(1) = %#04x, %#04x", uint16(s), uint16(c))
	}
	// Pow10
	for _, tt := range []struct {
		n    int
		want BFloat16
	}{{0, 0x3f80}, {1, 0x4120}, {2, 0x42c8}, {-1, 0x3dcd}, {38, 0x7e96}, {39, uvinfbf16}, {-40, 0x0001}, {-41, 0}, {-100, 0}, {1000, uvinfbf16}} {
		if got := NewBFloat16Pow10(tt.n); got != tt.want {
			t.Errorf("NewBFloat16Pow10(%d) = %#04x, want %#04x", tt.n, uint16(got), uint16(tt.want))
		}
	}
}

func TestBF16Round(t *testing.T) {
	t.Parallel()
	nextUp := func(x float64) float64 { return math.Nextafter(x, math.Inf(1)) }
	nextDown := func(x float64) float64 { return math.Nextafter(x, 0) }
	tests := []struct {
		y      float64
		want   BFloat16
		wantOK bool
	}{
		{1, 0x3f80, true},
		{1.00390625, 0, false}, // the tie of 1 and 1+2^-7
		{nextUp(1.00390625), 0, false},
		{nextDown(1.00390625), 0, false},
		{1.003906251, 0x3f81, true},
		{1.00390624, 0x3f80, true},
		{0x1.ffp+127, 0, false}, // the tie of the largest finite number and Inf
		{0x1.fep+127, 0x7f7f, true},
		{0x1p-134, 0, false}, // the tie of 0 and the smallest subnormal number
		{0x1.1p-134, 0x0001, true},
		{0x1p-135, 0, true},
		{0x1.8p-134, 0x0001, true},
		{0x1p-200, 0, true},
		{0, 0, true},
		{math.Inf(1), uvinfbf16, true},
		{math.Inf(-1), uvneginfbf16, true},
		{1e300, uvinfbf16, true},
		{-1.5, 0xbfc0, true},
	}
	for _, tt := range tests {
		got, ok := bf16Round(tt.y)
		if ok != tt.wantOK || ok && got != tt.want {
			t.Errorf("bf16Round(%x) = %#04x, %v; want %#04x, %v", tt.y, uint16(got), ok, uint16(tt.want), tt.wantOK)
		}
	}
	if got, ok := bf16Round(math.NaN()); !ok || !got.IsNaN() {
		t.Errorf("bf16Round(NaN) = %#04x, %v", uint16(got), ok)
	}
}

// TestBFloat16_MathNonFinite checks that the results are NaN or infinite
// at the same inputs as the math package, where mpmath cannot be used as the reference.
func TestBFloat16_MathNonFinite(t *testing.T) {
	t.Parallel()
	fs := map[string]func(float64) float64{
		"Exp": math.Exp, "Exp2": math.Exp2, "Expm1": math.Expm1, "Log": math.Log, "Log2": math.Log2, "Log10": math.Log10, "Log1p": math.Log1p,
		"Sin": math.Sin, "Cos": math.Cos, "Tan": math.Tan, "Asin": math.Asin, "Acos": math.Acos, "Atan": math.Atan,
		"Sinh": math.Sinh, "Cosh": math.Cosh, "Tanh": math.Tanh, "Asinh": math.Asinh, "Acosh": math.Acosh, "Atanh": math.Atanh,
		"Cbrt": math.Cbrt, "Gamma": math.Gamma, "Erf": math.Erf, "Erfc": math.Erfc, "Erfinv": math.Erfinv, "Erfcinv": math.Erfcinv,
		"J0": math.J0, "J1": math.J1, "Y0": math.Y0, "Y1": math.Y1,
	}
	for _, f := range bf16Funcs {
		m := fs[f.name]
		for i := range 1 << 16 {
			a, x := bf16Value(i)
			want := m(x)
			if f.name == "J1" && math.IsInf(x, 0) {
				want = 0
			}
			if f.name == "Erfcinv" && x != 0 && x < 0x1p-7 {
				// math.Erfcinv calculates Erfinv(1-x), which is wrong for the small x.
				continue
			}
			if !math.IsNaN(want) && !math.IsInf(want, 0) {
				continue
			}
			if got := f.f(a); !eqbf16(got, NewBFloat16(want)) {
				t.Errorf("BFloat16(%#04x).%s() = %#04x, want %#04x", i, f.name, uint16(got), uint16(NewBFloat16(want)))
			}
		}
	}

	// the functions with the additional parameters
	for i := range 1 << 16 {
		a, x := bf16Value(i)
		for _, n := range []int{-3, 0, 2, 5} {
			if want := math.Jn(n, x); math.IsNaN(want) || math.IsInf(want, 0) {
				if got := a.Jn(n); !eqbf16(got, NewBFloat16(want)) {
					t.Errorf("BFloat16(%#04x).Jn(%d) = %#04x, want %#04x", i, n, uint16(got), uint16(NewBFloat16(want)))
				}
			}
			if want := math.Yn(n, x); math.IsNaN(want) || math.IsInf(want, 0) {
				if got := a.Yn(n); !eqbf16(got, NewBFloat16(want)) {
					t.Errorf("BFloat16(%#04x).Yn(%d) = %#04x, want %#04x", i, n, uint16(got), uint16(NewBFloat16(want)))
				}
			}
		}
		lg, sign := math.Lgamma(x)
		if math.IsNaN(lg) || math.IsInf(lg, 0) {
			got, gs := a.Lgamma()
			if !eqbf16(got, NewBFloat16(lg)) || gs != sign {
				t.Errorf("BFloat16(%#04x).Lgamma() = %#04x, %d, want %#04x, %d", i, uint16(got), gs, uint16(NewBFloat16(lg)), sign)
			}
		}
		for _, yi := range []int{0, 0x3f80, 0xbf80, 0x4000, 0x3f00, 0x7f80, 0xff80, 0x8000, 0x7fc0, 0x0001} {
			b, y := bf16Value(yi)
			if want := math.Pow(x, y); math.IsNaN(want) || math.IsInf(want, 0) {
				if got := a.Pow(b); !eqbf16(got, NewBFloat16(want)) {
					t.Errorf("BFloat16(%#04x).Pow(%#04x) = %#04x, want %#04x", i, yi, uint16(got), uint16(NewBFloat16(want)))
				}
			}
			if want := math.Atan2(x, y); math.IsNaN(want) {
				if got := a.Atan2(b); !got.IsNaN() {
					t.Errorf("BFloat16(%#04x).Atan2(%#04x) = %#04x, want NaN", i, yi, uint16(got))
				}
			}
			if want := math.Hypot(x, y); math.IsNaN(want) || math.IsInf(want, 0) {
				if got := HypotBF16(a, b); !eqbf16(got, NewBFloat16(want)) {
					t.Errorf("HypotBF16(%#04x, %#04x) = %#04x, want %#04x", i, yi, uint16(got), uint16(NewBFloat16(want)))
				}
			}
		}
	}
}

func TestBF16Result(t *testing.T) {
	t.Parallel()
	// the slow path is used only if the result is ambiguous.
	slow := func() BFloat16 { return 0x1234 }
	if got := bf16Result(1.00390625, slow); got != 0x1234 {
		t.Errorf("bf16Result(tie) = %#04x, want the slow one", uint16(got))
	}
	if got := bf16Result(1.5, slow); got != 0x3fc0 {
		t.Errorf("bf16Result(1.5) = %#04x, want 0x3fc0", uint16(got))
	}

	// the helpers use the slow path for the ambiguous results.
	tie := func(float64) float64 { return 1.00390625 }
	if got := bf16Unary(1, tie, func(Float256) Float256 { return NewFloat64(2).Float256() }); got != 0x4000 {
		t.Errorf("bf16Unary = %#04x, want 0x4000", uint16(got))
	}
	tie2 := func(x, y float64) float64 { return 1.00390625 }
	if got := bf16Binary(1, 1, tie2, func(x, y Float256) Float256 { return NewFloat64(3).Float256() }); got != 0x4040 {
		t.Errorf("bf16Binary = %#04x, want 0x4040", uint16(got))
	}
}

// TestBFloat16_LgammaSign checks the sign of Gamma.
func TestBFloat16_LgammaSign(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a, x := bf16Value(i)
		if math.IsNaN(x) {
			continue
		}
		_, sign := a.Lgamma()
		g := math.Gamma(x)
		want := 1
		if math.Signbit(g) && !math.IsNaN(g) {
			want = -1
		}
		if x == 0 || math.IsInf(x, 0) || x < 0 && x == math.Floor(x) {
			// the sign of the infinite values follows the math package
			_, want = math.Lgamma(x)
		}
		if sign != want {
			t.Errorf("BFloat16(%#04x).Lgamma() sign = %d, want %d", i, sign, want)
		}
	}
}

func TestBF16Helpers(t *testing.T) {
	t.Parallel()
	tie := func(int, float64) float64 { return 1.00390625 }
	if got := bf16UnaryInt(1, 3, tie, func(Float256, int) Float256 { return NewFloat64(2).Float256() }); got != 0x4000 {
		t.Errorf("bf16UnaryInt = %#04x, want 0x4000", uint16(got))
	}
	if got := BFloat16(0x4000).lgammaSlow(); got != 0 {
		t.Errorf("lgammaSlow(2) = %#04x, want 0", uint16(got))
	}
	if got := BFloat16(0xc020).lgammaSlow(); got != NewBFloat16(math.Log(math.Abs(math.Gamma(-2.5)))) {
		t.Errorf("lgammaSlow(-2.5) = %#04x", uint16(got))
	}

	// bf16RoundTieToZero
	tests := []struct {
		r    Float256
		want BFloat16
	}{
		{NewFloat64(0).Float256(), 0},
		{NewFloat64(math.Inf(1)).Float256(), uvinfbf16},
		{NewFloat64(math.Inf(-1)).Float256(), uvneginfbf16},
		{NewFloat64(1.00390625).Float256(), 0x3f80},  // tie: toward zero
		{NewFloat64(-1.00390625).Float256(), 0xbf80}, // tie: toward zero
		{NewFloat64(1.01171875).Float256(), 0x3f81},  // tie: toward zero (even would be 0x3f82)
		{NewFloat64(1.005).Float256(), 0x3f81},       // not a tie
		{NewFloat64(1.002).Float256(), 0x3f80},       // not a tie
	}
	for _, tt := range tests {
		if got := bf16RoundTieToZero(tt.r); got != tt.want {
			t.Errorf("bf16RoundTieToZero(%x) = %#04x, want %#04x", tt.r, uint16(got), uint16(tt.want))
		}
	}
}

func BenchmarkBFloat16_Math(b *testing.B) {
	r := rand.New(rand.NewPCG(1, 2))
	var xs [1024]BFloat16
	for i := range xs {
		xs[i] = BFloat16(r.Uint32()&0x807f | uint32(120+r.IntN(12))<<7)
	}
	for _, f := range bf16Funcs {
		b.Run(f.name, func(b *testing.B) {
			for i := 0; b.Loop(); i++ {
				runtime.KeepAlive(f.f(xs[i%1024]))
			}
		})
	}
	b.Run("Pow", func(b *testing.B) {
		for i := 0; b.Loop(); i++ {
			runtime.KeepAlive(xs[i%1024].Pow(xs[(i+1)%1024]))
		}
	})
}

// TestBFloat16_MathFallbackN checks the functions with the additional parameters
// with the results of Float256 for a part of the inputs.
func TestBFloat16_MathFallbackN(t *testing.T) {
	t.Parallel()
	ys := []BFloat16{0x3f80, 0xbf80, 0x4000, 0xc000, 0x3f00, 0xbf00, 0x4040, 0x4120, 0xc120, 0x3dcd, 0x42c8, 0x447a, 0x0001, 0x8001, 0x0080, 0x7f7f, 0xff7f, 0x7f80, 0xff80, 0x0000, 0x8000, 0x7fc0}
	for i := 0; i < 1<<16; i += 11 {
		a := BFloat16(i)
		for _, b := range ys {
			if got, want := a.Pow(b), a.Float256().Pow(b.Float256()).BFloat16(); !eqbf16(got, want) {
				t.Fatalf("BFloat16(%#04x).Pow(%#04x) = %#04x, want %#04x", i, uint16(b), uint16(got), uint16(want))
			}
			if got, want := HypotBF16(a, b), Hypot256(a.Float256(), b.Float256()).BFloat16(); !eqbf16(got, want) {
				t.Fatalf("HypotBF16(%#04x, %#04x) = %#04x, want %#04x", i, uint16(b), uint16(got), uint16(want))
			}
			got, want := a.Atan2(b), a.Float256().Atan2(b.Float256()).BFloat16()
			if x, y := a.Float64().BuiltIn(), b.Float64().BuiltIn(); y > 0 && math.Abs(x/y) < 0x1p-100 {
				// the tie is rounded toward zero. It is checked in TestBFloat16_MathMisc.
				want = got
			}
			if !eqbf16(got, want) {
				t.Fatalf("BFloat16(%#04x).Atan2(%#04x) = %#04x, want %#04x", i, uint16(b), uint16(got), uint16(want))
			}
		}
		for _, n := range []int{-2, 0, 1, 2, 3, 5, 10} {
			if got, want := a.Jn(n), a.Float256().Jn(n).BFloat16(); !eqbf16(got, want) && (n != 1 && n != -1 || a&^signMaskbf16 >= 0x0100) {
				t.Fatalf("BFloat16(%#04x).Jn(%d) = %#04x, want %#04x", i, n, uint16(got), uint16(want))
			}
			if got, want := a.Yn(n), a.Float256().Yn(n).BFloat16(); !eqbf16(got, want) {
				t.Fatalf("BFloat16(%#04x).Yn(%d) = %#04x, want %#04x", i, n, uint16(got), uint16(want))
			}
		}
		l256, sign256 := a.Float256().Lgamma()
		if l, sign := a.Lgamma(); !eqbf16(l, l256.BFloat16()) || sign != sign256 {
			t.Fatalf("BFloat16(%#04x).Lgamma() = %#04x, %d, want %#04x, %d", i, uint16(l), sign, uint16(l256.BFloat16()), sign256)
		}
	}
}
