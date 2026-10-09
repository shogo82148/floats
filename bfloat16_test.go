package floats

import (
	"math"
	"math/big"
	"math/rand/v2"
	"runtime"
	"testing"
)

// bf16FromBig rounds x to the nearest BFloat16, ties to even.
// It is the reference of the conversions and the arithmetic operations,
// and it does not depend on the conversions of this package.
func bf16FromBig(x *big.Float) BFloat16 {
	if x.IsInf() {
		if x.Signbit() {
			return uvneginfbf16
		}
		return uvinfbf16
	}
	var sign uint16
	if x.Signbit() {
		sign = signMaskbf16
	}
	if x.Sign() == 0 {
		return BFloat16(sign)
	}
	ax := new(big.Float).SetPrec(x.Prec()).Abs(x)
	e := ax.MantExp(nil) - 1 // ax = 1.xxx * 2^e

	prec := 8
	if e < -126 {
		prec = e + 134 // subnormal numbers have less bits
	}
	var r *big.Float
	if prec >= 1 {
		r = new(big.Float).SetPrec(uint(prec)).SetMode(big.ToNearestEven).Set(ax)
	} else {
		// less than the smallest subnormal number 2^-133
		half := new(big.Float).SetMantExp(big.NewFloat(1), -134)
		if ax.Cmp(half) > 0 {
			r = new(big.Float).SetMantExp(big.NewFloat(1), -133)
		} else {
			r = new(big.Float)
		}
	}
	if r.Cmp(new(big.Float).SetMantExp(big.NewFloat(1), 128)) >= 0 {
		return BFloat16(sign | uvinfbf16)
	}
	f, _ := r.Float64() // exact
	return BFloat16(sign | uint16(math.Float32bits(float32(f))>>16))
}

// bigFromBF16 converts a finite a to *big.Float exactly.
func bigFromBF16(a BFloat16) *big.Float {
	f := new(big.Float).SetPrec(2000).SetFloat64(float64(math.Float32frombits(uint32(a) << 16)))
	return f
}

func eqbf16(a, b BFloat16) bool {
	if a.IsNaN() && b.IsNaN() {
		return true
	}
	return a == b
}

func TestBFloat16_Constants(t *testing.T) {
	t.Parallel()
	if got := NewBFloat16(1); got != uvonebf16 {
		t.Errorf("NewBFloat16(1) = %#x, want %#x", uint16(got), uvonebf16)
	}
	if got := NewBFloat16(math.Inf(1)); got != uvinfbf16 {
		t.Errorf("NewBFloat16(+Inf) = %#x, want %#x", uint16(got), uvinfbf16)
	}
	if got := NewBFloat16(math.Inf(-1)); got != uvneginfbf16 {
		t.Errorf("NewBFloat16(-Inf) = %#x, want %#x", uint16(got), uvneginfbf16)
	}
	if got := NewBFloat16(math.NaN()); !got.IsNaN() {
		t.Errorf("NewBFloat16(NaN) = %#x, want NaN", uint16(got))
	}
	if got := NewBFloat16NaN(); !got.IsNaN() {
		t.Errorf("NewBFloat16NaN() = %#x, want NaN", uint16(got))
	}
	if got := NewBFloat16Inf(1); got != uvinfbf16 {
		t.Errorf("NewBFloat16Inf(1) = %#x, want %#x", uint16(got), uvinfbf16)
	}
	if got := NewBFloat16Inf(-1); got != uvneginfbf16 {
		t.Errorf("NewBFloat16Inf(-1) = %#x, want %#x", uint16(got), uvneginfbf16)
	}
	if got := NewBFloat16FromBits(0x4049); got.Bits() != 0x4049 {
		t.Errorf("NewBFloat16FromBits(0x4049).Bits() = %#x, want 0x4049", got.Bits())
	}

	// the well-known values
	tests := []struct {
		in   float64
		want BFloat16
	}{
		{0, 0x0000},
		{math.Copysign(0, -1), 0x8000},
		{1, 0x3f80},
		{-2, 0xc000},
		{math.Pi, 0x4049},                   // 3.140625
		{0x1p-133, 0x0001},                  // the smallest subnormal number
		{0x1.fcp-127, 0x007f},               // the largest subnormal number
		{0x1p-126, 0x0080},                  // the smallest normal number
		{0x1.fep+127, 0x7f7f},               // the largest finite number
		{0x1.ffp+127, 0x7f80},               // a tie between the largest finite number and Inf, rounds to Inf
		{0x1p-134, 0x0000},                  // a tie between 0 and the smallest subnormal number
		{0x1.8p-134, 0x0001},                // rounds up
		{0x1.01p+0, 0x3f80},                 // a tie, rounds to even
		{0x1.018p+0, 0x3f81},                // 1 + 2^-7*(1+1/2)... rounds up
		{0x1.02p+0, 0x3f81},                 // exact: 1 + 2^-7
		{0x1.03p+0, 0x3f82},                 // tie, rounds to even
		{0x1.01p+0 + 0x1p-40, 0x3f81},       // not a tie
		{0x1.02p+0 + 0x1p-40, 0x3f81},       // not a tie
		{0x1.03p+0 + 0x1p-40, 0x3f82},       // rounds up
		{0x1.03p+0 - 0x1p-40, 0x3f81},       // rounds down
		{math.MaxFloat32, uvinfbf16},        // the largest finite Float32 overflows
		{0x1.fep+127 + 0x1p+110, 0x7f7f},    // just below the overflow threshold
		{0x1.ffp+127 + 0x1p+110, uvinfbf16}, // just above the overflow threshold
		{1e-45, 0x0000},                     // the smallest subnormal Float32
		{9.183549615799121e-41, 0x0001},     // 2^-133
		{4.591774807899561e-41, 0x0000},     // 2^-134 (tie)
		{4.5917748078995615e-41, 0x0001},    // just above 2^-134
		{math.SmallestNonzeroFloat64, 0x0000},
		{math.MaxFloat64, uvinfbf16},
	}
	for _, tt := range tests {
		if got := NewBFloat16(tt.in); got != tt.want {
			t.Errorf("NewBFloat16(%x) = %#04x, want %#04x", tt.in, uint16(got), uint16(tt.want))
		}
	}
}

func TestBFloat16_Classify(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a := BFloat16(i)
		f := float64(a.Float32())
		if got, want := a.IsNaN(), math.IsNaN(f); got != want {
			t.Errorf("BFloat16(%#04x).IsNaN() = %v, want %v", i, got, want)
		}
		for _, sign := range []int{-1, 0, 1} {
			if got, want := a.IsInf(sign), math.IsInf(f, sign); got != want {
				t.Errorf("BFloat16(%#04x).IsInf(%d) = %v, want %v", i, sign, got, want)
			}
		}
		if got, want := a.Signbit(), math.Signbit(f) && !math.IsNaN(f) || i&0x8000 != 0; got != want {
			t.Errorf("BFloat16(%#04x).Signbit() = %v, want %v", i, got, want)
		}
		if got, want := a.IsZero(), f == 0; got != want {
			t.Errorf("BFloat16(%#04x).IsZero() = %v, want %v", i, got, want)
		}
		if got := a.Neg(); got != BFloat16(i^0x8000) {
			t.Errorf("BFloat16(%#04x).Neg() = %#04x", i, uint16(got))
		}
		if got := a.Abs(); got != BFloat16(i&0x7fff) {
			t.Errorf("BFloat16(%#04x).Abs() = %#04x", i, uint16(got))
		}
		if got := a.Copysign(BFloat16(0x8000)); got != BFloat16(i|0x8000) {
			t.Errorf("BFloat16(%#04x).Copysign(-0) = %#04x", i, uint16(got))
		}
		if got := a.Copysign(0); got != BFloat16(i&0x7fff) {
			t.Errorf("BFloat16(%#04x).Copysign(+0) = %#04x", i, uint16(got))
		}
	}
}

func TestBFloat16_Int(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   float64
		want int64
	}{
		{0, 0}, {1, 1}, {-1, -1}, {1.5, 1}, {-1.5, -1}, {0.5, 0}, {255, 255}, {256, 256}, {-256, -256},
		{0x1p40, 1 << 40}, {-0x1p40, -(1 << 40)}, {0x1p62, 1 << 62},
	}
	for _, tt := range tests {
		a := NewBFloat16(tt.in)
		if got := a.Int64(); got != tt.want {
			t.Errorf("BFloat16(%v).Int64() = %d, want %d", tt.in, got, tt.want)
		}
		if got := a.Int128(); int64(got.Int64()) != tt.want {
			t.Errorf("BFloat16(%v).Int128() = %v, want %d", tt.in, got, tt.want)
		}
		if got := a.Int256(); int64(got.Int64()) != tt.want {
			t.Errorf("BFloat16(%v).Int256() = %v, want %d", tt.in, got, tt.want)
		}
		if tt.want >= 0 {
			if got := a.Uint64(); got != uint64(tt.want) {
				t.Errorf("BFloat16(%v).Uint64() = %d, want %d", tt.in, got, tt.want)
			}
			if got := a.Uint128(); uint64(got.Uint64()) != uint64(tt.want) {
				t.Errorf("BFloat16(%v).Uint128() = %v, want %d", tt.in, got, tt.want)
			}
			if got := a.Uint256(); uint64(got.Uint64()) != uint64(tt.want) {
				t.Errorf("BFloat16(%v).Uint256() = %v, want %d", tt.in, got, tt.want)
			}
		}
	}

	// the values that exceed int64
	a := NewBFloat16(0x1p100)
	if got, want := a.Int128(), Float64(0x1p100).Int128(); got != want {
		t.Errorf("BFloat16(2^100).Int128() = %v, want %v", got, want)
	}
	if got, want := a.Uint128(), Float64(0x1p100).Uint128(); got != want {
		t.Errorf("BFloat16(2^100).Uint128() = %v, want %v", got, want)
	}
	if got, want := a.Int256(), Float64(0x1p100).Int256(); got != want {
		t.Errorf("BFloat16(2^100).Int256() = %v, want %v", got, want)
	}
	if got, want := a.Uint256(), Float64(0x1p100).Uint256(); got != want {
		t.Errorf("BFloat16(2^100).Uint256() = %v, want %v", got, want)
	}
}

// TestBFloat16_Arith compares the arithmetic operations with the exact results of math/big.
func TestBFloat16_Arith(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	rnd := func() BFloat16 {
		b := BFloat16(r.Uint32())
		switch r.IntN(6) {
		case 0:
			b &^= 0x7f80 // subnormal
		case 1:
			b = b&^0x7f80 | BFloat16(r.IntN(30)+112)<<7 // close to 1
		case 2:
			b = b&^0x7f80 | BFloat16(r.IntN(8)+1)<<7 // close to the smallest normal numbers
		case 3:
			b = b&^0x7f80 | BFloat16(r.IntN(8)+240)<<7 // close to the overflow
		}
		return b
	}
	for range 400_000 {
		a, b := rnd(), rnd()
		if r.IntN(3) == 0 {
			b = a ^ BFloat16(r.IntN(3))<<15 ^ BFloat16(r.IntN(2)) // close or opposite
		}
		if r.IntN(3) == 0 {
			n := uint(r.IntN(8))
			a, b = a>>n<<n, b>>n<<n // short fractions
		}
		special := a.IsNaN() || b.IsNaN() || a.IsInf(0) || b.IsInf(0)
		var fa, fb *big.Float
		if !special {
			fa, fb = bigFromBF16(a), bigFromBF16(b)
		}

		check := func(name string, got BFloat16, exact func() *big.Float) {
			t.Helper()
			if special {
				// compared with the built-in float64
				var want BFloat16
				x, y := float64(a.Float32()), float64(b.Float32())
				switch name {
				case "Mul":
					want = NewBFloat16(x * y)
				case "Quo":
					want = NewBFloat16(x / y)
				case "Add":
					want = NewBFloat16(x + y)
				case "Sub":
					want = NewBFloat16(x - y)
				}
				if !eqbf16(got, want) {
					t.Fatalf("BFloat16(%#04x).%s(%#04x) = %#04x, want %#04x", uint16(a), name, uint16(b), uint16(got), uint16(want))
				}
				return
			}
			want := bf16FromBig(exact())
			if exact().Sign() == 0 && name != "Mul" && name != "Quo" {
				// the sign of an exact zero sum
				want = 0
				if a.Signbit() && (name == "Add" && b.Signbit() || name == "Sub" && !b.Signbit()) {
					want = 0x8000
				}
			}
			if got != want {
				t.Fatalf("BFloat16(%#04x).%s(%#04x) = %#04x, want %#04x", uint16(a), name, uint16(b), uint16(got), uint16(want))
			}
		}
		check("Mul", a.Mul(b), func() *big.Float { return new(big.Float).SetPrec(2000).Mul(fa, fb) })
		check("Add", a.Add(b), func() *big.Float { return new(big.Float).SetPrec(2000).Add(fa, fb) })
		check("Sub", a.Sub(b), func() *big.Float { return new(big.Float).SetPrec(2000).Sub(fa, fb) })
		if !b.IsZero() {
			check("Quo", a.Quo(b), func() *big.Float {
				q := new(big.Float).SetPrec(3000).Quo(fa, fb)
				return q
			})
		}
	}
}

func TestBFloat16_Sqrt(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a := BFloat16(i)
		got := a.Sqrt()
		var want BFloat16
		switch {
		case a.IsNaN():
			want = a
		case a.IsZero() || a.IsInf(1):
			want = a
		case a.Signbit():
			want = uvnanbf16
		default:
			// sqrt(a) is rounded correctly: s*s <= a < (s+)^2, where
			// the candidates are the neighbors of the float64 result.
			want = bf16FromBig(sqrtBig(bigFromBF16(a)))
		}
		if !eqbf16(got, want) {
			t.Errorf("BFloat16(%#04x).Sqrt() = %#04x, want %#04x", i, uint16(got), uint16(want))
		}
	}
}

// sqrtBig returns the square root of x with 1000 bits of the precision.
func sqrtBig(x *big.Float) *big.Float {
	return new(big.Float).SetPrec(1000).Sqrt(x)
}

func TestFMABF16(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(3, 4))
	rnd := func() BFloat16 {
		b := BFloat16(r.Uint32())
		switch r.IntN(6) {
		case 0:
			b &^= 0x7f80
		case 1:
			b = b&^0x7f80 | BFloat16(r.IntN(30)+112)<<7
		case 2:
			b = b&^0x7f80 | BFloat16(r.IntN(8)+1)<<7
		case 3:
			b = b&^0x7f80 | BFloat16(r.IntN(8)+240)<<7
		}
		return b & 0xffff
	}
	for range 400_000 {
		x, y, z := rnd(), rnd(), rnd()
		if x.IsNaN() || x.IsInf(0) || y.IsNaN() || y.IsInf(0) {
			// the special cases are checked below
			got := FMABF16(x, y, z)
			want := NewBFloat16(float64(x.Float32())*float64(y.Float32()) + float64(z.Float32()))
			if !eqbf16(got, want) {
				t.Fatalf("FMABF16(%#04x, %#04x, %#04x) = %#04x, want %#04x", uint16(x), uint16(y), uint16(z), uint16(got), uint16(want))
			}
			continue
		}
		p := new(big.Float).SetPrec(2000).Mul(bigFromBF16(x), bigFromBF16(y))
		switch r.IntN(4) {
		case 0:
			// cancellation
			z = bf16FromBig(p).Neg()
		case 1:
			// z is tiny: breaks the tie of the rounding of the product
			z = BFloat16(r.IntN(0x80) + 1)
			if r.IntN(2) == 0 {
				z = z.Neg()
			}
			n := uint(r.IntN(8))
			x, y = x>>n<<n, y>>n<<n
		case 2:
			// the exponent of z is close to the one of the product
			e := int(p.MantExp(nil)) + 126 + r.IntN(20) - 10
			if e > 0 && e < 255 {
				z = z&0x807f | BFloat16(e)<<7
			}
		}
		got := FMABF16(x, y, z)
		special := z.IsNaN() || z.IsInf(0)
		var want BFloat16
		if special {
			want = NewBFloat16(float64(x.Float32())*float64(y.Float32()) + float64(z.Float32()))
		} else {
			p := new(big.Float).SetPrec(2000).Mul(bigFromBF16(x), bigFromBF16(y))
			exact := new(big.Float).SetPrec(2000).Add(p, bigFromBF16(z))
			want = bf16FromBig(exact)
			if exact.Sign() == 0 {
				// the sign of zero
				want = BFloat16(uint16(NewBFloat16(float64(x.Float32())*float64(y.Float32()) + float64(z.Float32()))))
			}
		}
		if !eqbf16(got, want) {
			t.Fatalf("FMABF16(%#04x, %#04x, %#04x) = %#04x, want %#04x", uint16(x), uint16(y), uint16(z), uint16(got), uint16(want))
		}
	}

	// the tie of the product is broken by z
	x, y := BFloat16(0x3f81), BFloat16(0x3fc0) // (1+2^-7) * 1.5 = 1.5 + 1.5*2^-7: not a tie
	for _, tt := range []struct{ x, y, z, want BFloat16 }{
		{0x3f81, 0x3f81, 0x0000, 0x3f82}, // (1+2^-7)^2 = 1 + 2^-6 + 2^-14: rounds down to 1+2^-6
		{0x3f81, 0x3f81, BFloat16(0x0001), BFloat16(0x3f82)},
		{x, y, 0x0000, 0x3fc2},
	} {
		got := FMABF16(tt.x, tt.y, tt.z)
		want := bf16FromBig(new(big.Float).SetPrec(2000).Add(new(big.Float).SetPrec(2000).Mul(bigFromBF16(tt.x), bigFromBF16(tt.y)), bigFromBF16(tt.z)))
		if got != want {
			t.Errorf("FMABF16(%#04x, %#04x, %#04x) = %#04x, want %#04x", uint16(tt.x), uint16(tt.y), uint16(tt.z), uint16(got), uint16(want))
		}
	}
}

func BenchmarkBFloat16_Arith(b *testing.B) {
	r := rand.New(rand.NewPCG(1, 2))
	var xs, ys [1024]BFloat16
	for i := range xs {
		xs[i] = BFloat16(r.Uint32()&0x807f | uint32(120+r.IntN(16))<<7)
		ys[i] = BFloat16(r.Uint32()&0x807f | uint32(120+r.IntN(16))<<7)
	}
	for _, op := range []struct {
		name string
		f    func(a, b BFloat16) BFloat16
	}{
		{"Add", BFloat16.Add},
		{"Mul", BFloat16.Mul},
		{"Quo", BFloat16.Quo},
		{"Sqrt", func(a, b BFloat16) BFloat16 { return a.Abs().Sqrt() }},
		{"FMA", func(a, b BFloat16) BFloat16 { return FMABF16(a, b, a) }},
	} {
		b.Run(op.name, func(b *testing.B) {
			for i := 0; b.Loop(); i++ {
				runtime.KeepAlive(op.f(xs[i%1024], ys[i%1024]))
			}
		})
	}
}

func TestBFloat16_FlushToZero(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a, x := bf16Value(i)
		want := a
		if x != 0 && math.Abs(x) < 0x1p-126 {
			// subnormal
			want = BFloat16(i & 0x8000)
		}
		if got := a.FlushToZero(); got != want {
			t.Errorf("BFloat16(%#04x).FlushToZero() = %#04x, want %#04x", i, uint16(got), uint16(want))
		}
	}

	// the smallest normal number is kept, the largest subnormal number is flushed
	if got := BFloat16(0x0080).FlushToZero(); got != 0x0080 {
		t.Errorf("FlushToZero of the smallest normal number = %#04x", uint16(got))
	}
	if got := BFloat16(0x807f).FlushToZero(); got != 0x8000 {
		t.Errorf("FlushToZero of the largest negative subnormal number = %#04x", uint16(got))
	}

	// emulation of the hardware without subnormal numbers
	a, b := BFloat16(0x0100), BFloat16(0x3f00) // 2^-125 * 0.5 = 2^-126 (normal)
	if got := a.FlushToZero().Mul(b.FlushToZero()).FlushToZero(); got != 0x0080 {
		t.Errorf("flush(2^-125 * 0.5) = %#04x, want 0x0080", uint16(got))
	}
	b = 0x3e80 // 0.25: the product 2^-127 is subnormal
	if got := a.FlushToZero().Mul(b.FlushToZero()).FlushToZero(); got != 0 {
		t.Errorf("flush(2^-125 * 0.25) = %#04x, want 0", uint16(got))
	}
}
