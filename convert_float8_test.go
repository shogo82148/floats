package floats

import (
	"math"
	"math/big"
	"math/rand/v2"
	"testing"
)

// float8Spec describes an 8-bit format for the reference implementation.
type float8Spec struct {
	name  string
	mbits int
	bias  int
	// limit is the encoding of the first non-finite magnitude.
	limit uint8
}

var (
	specE4M3 = float8Spec{"Float8E4M3", 3, 7, 0x7f}
	specE5M2 = float8Spec{"Float8E5M2", 2, 15, 0x7c}
)

// value returns the exact magnitude of the encoding enc.
// For enc == limit, it returns the value that the encoding would have
// if the format had no special values.
func (s float8Spec) value(enc uint8) *big.Float {
	e := int(enc >> s.mbits)
	m := int64(enc & (1<<s.mbits - 1))
	f := new(big.Float).SetPrec(300)
	if e == 0 {
		return f.SetMantExp(new(big.Float).SetInt64(m), 1-s.bias-s.mbits)
	}
	return f.SetMantExp(new(big.Float).SetInt64(1<<s.mbits+m), e-s.bias-s.mbits)
}

// round returns the encoding of the magnitude x rounded to nearest even.
// If the result is greater than or equal to s.limit, x is out of the range.
func (s float8Spec) round(x *big.Float) uint8 {
	i := uint8(0)
	for i < s.limit && s.value(i+1).Cmp(x) <= 0 {
		i++
	}
	if i == s.limit || s.value(i).Cmp(x) == 0 {
		return i
	}
	mid := new(big.Float).SetPrec(300).Add(s.value(i), s.value(i+1))
	mid.Quo(mid, big.NewFloat(2))
	switch c := x.Cmp(mid); {
	case c < 0:
		return i
	case c > 0:
		return i + 1
	case i%2 == 0:
		return i
	default:
		return i + 1
	}
}

// isNaN reports whether enc is a NaN encoding.
func (s float8Spec) isNaN(enc uint8) bool {
	enc &= 0x7f
	if s.limit == 0x7f {
		return enc == 0x7f
	}
	return enc > s.limit
}

// convert returns the expected encoding of the exact value x.
// It reports NaN as nan == true.
func (s float8Spec) convert(x Float256) (enc uint8, nan bool) {
	if x.IsNaN() {
		return 0, true
	}
	var sign uint8
	if x.Signbit() {
		sign = 0x80
	}
	if x.IsInf(0) {
		if s.limit == 0x7f {
			return 0, true
		}
		return sign | s.limit, false
	}
	r := s.round(signedBigFromFloat256(x.Abs()))
	if r >= s.limit {
		if s.limit == 0x7f {
			return 0, true
		}
		return sign | s.limit, false
	}
	return sign | r, false
}

type float8Source[T any] interface {
	Float256() Float256
	Float8E4M3() Float8E4M3
	Float8E5M2() Float8E5M2
	Neg() T
	Nextafter(T) T
}

func checkToFloat8[T float8Source[T]](t *testing.T, name string, x T) {
	t.Helper()
	exact := x.Float256()
	if want, nan := specE4M3.convert(exact); nan {
		if got := x.Float8E4M3(); !specE4M3.isNaN(uint8(got)) {
			t.Errorf("%s(%x).Float8E4M3() = %#02x, want NaN", name, exact, uint8(got))
		}
	} else if got := x.Float8E4M3(); uint8(got) != want {
		t.Errorf("%s(%x).Float8E4M3() = %#02x, want %#02x", name, exact, uint8(got), want)
	}
	if want, nan := specE5M2.convert(exact); nan {
		if got := x.Float8E5M2(); !specE5M2.isNaN(uint8(got)) {
			t.Errorf("%s(%x).Float8E5M2() = %#02x, want NaN", name, exact, uint8(got))
		}
	} else if got := x.Float8E5M2(); uint8(got) != want {
		t.Errorf("%s(%x).Float8E5M2() = %#02x, want %#02x", name, exact, uint8(got), want)
	}
}

// float8Candidates returns the values around the boundaries of the rounding.
func float8Candidates() []float64 {
	var ret []float64
	for _, s := range []float8Spec{specE4M3, specE5M2} {
		for i := uint8(0); i <= s.limit; i++ {
			v, _ := s.value(i).Float64()
			ret = append(ret, v)
			if i < s.limit {
				w, _ := s.value(i + 1).Float64()
				ret = append(ret, (v+w)/2) // exact
			}
		}
	}
	ret = append(ret,
		0, math.SmallestNonzeroFloat64, math.MaxFloat64, math.MaxFloat32, math.SmallestNonzeroFloat32,
		0x1p-150, 0x1p-149, 0x1p-126, 0x1p-18, 0x1p-10, 0x1p-11, 0x1p20, 0x1p127, 0x1p128,
	)
	return ret
}

func testToFloat8[T float8Source[T]](t *testing.T, name string, from func(Float256) T) {
	t.Helper()
	inf := from(NewFloat256Inf(1))
	for _, c := range float8Candidates() {
		for _, sign := range []float64{1, -1} {
			x := from(NewFloat256(sign * c))
			checkToFloat8(t, name, x)
			checkToFloat8(t, name, x.Nextafter(inf))
			checkToFloat8(t, name, x.Nextafter(inf.Neg()))
		}
	}
	checkToFloat8(t, name, from(NewFloat256NaN()))
	checkToFloat8(t, name, inf)
	checkToFloat8(t, name, inf.Neg())

	// random values around the range of Float8
	r := rand.New(rand.NewPCG(1, 2))
	for range 1 << 12 {
		f := math.Ldexp(r.Float64()*2-1, r.IntN(60)-30)
		x := from(NewFloat256(f))
		checkToFloat8(t, name, x)
		checkToFloat8(t, name, x.Nextafter(inf))
		checkToFloat8(t, name, x.Nextafter(inf.Neg()))
	}
}

func TestToFloat8(t *testing.T) {
	t.Parallel()
	t.Run("Float32", func(t *testing.T) {
		t.Parallel()
		testToFloat8(t, "Float32", Float256.Float32)
	})
	t.Run("Float64", func(t *testing.T) {
		t.Parallel()
		testToFloat8(t, "Float64", Float256.Float64)
	})
	t.Run("Float128", func(t *testing.T) {
		t.Parallel()
		testToFloat8(t, "Float128", Float256.Float128)
	})
	t.Run("Float256", func(t *testing.T) {
		t.Parallel()
		testToFloat8(t, "Float256", func(a Float256) Float256 { return a })
	})
	t.Run("Float16", func(t *testing.T) {
		t.Parallel()
		for i := range 1 << 16 {
			checkToFloat8(t, "Float16", Float16(i))
		}
	})
	t.Run("BFloat16", func(t *testing.T) {
		t.Parallel()
		for i := range 1 << 16 {
			checkToFloat8(t, "BFloat16", BFloat16(i))
		}
	})
	t.Run("Float32 random bits", func(t *testing.T) {
		t.Parallel()
		r := rand.New(rand.NewPCG(3, 4))
		for range 1 << 16 {
			checkToFloat8(t, "Float32", Float32(math.Float32frombits(r.Uint32())))
		}
		// the boundaries of the exponent
		for e := range uint32(256) {
			for _, frac := range []uint32{0, 1, 0x100000, 0x3fffff, 0x400000, 0x7fffff} {
				checkToFloat8(t, "Float32", Float32(math.Float32frombits(e<<23|frac)))
			}
		}
	})
}

// float8Values calls f for each encoding of the format with the exact value.
func float8Values(s float8Spec, f func(enc uint8, nan bool, neg bool, v *big.Float)) {
	for i := range 256 {
		enc := uint8(i)
		nan := s.isNaN(enc)
		if s.limit == 0x7c && enc&0x7f == 0x7c {
			// infinity
			f(enc, false, enc&0x80 != 0, nil)
			continue
		}
		var v *big.Float
		if !nan {
			v = s.value(enc & 0x7f)
		}
		f(enc, nan, enc&0x80 != 0, v)
	}
}

func checkFromFloat8(t *testing.T, name string, enc uint8, nan, neg bool, v *big.Float, got Float256) {
	t.Helper()
	switch {
	case nan:
		if !got.IsNaN() {
			t.Errorf("%s(%#02x) = %x, want NaN", name, enc, got)
		}
	case v == nil:
		if !got.IsInf(0) || got.Signbit() != neg {
			t.Errorf("%s(%#02x) = %x, want infinity", name, enc, got)
		}
	default:
		if got.IsNaN() || got.IsInf(0) || got.Signbit() != neg || signedBigFromFloat256(got.Abs()).Cmp(v) != 0 {
			t.Errorf("%s(%#02x) = %x, want %v (neg=%v)", name, enc, got, v, neg)
		}
	}
}

func TestFloat8E4M3_ToOthers(t *testing.T) {
	t.Parallel()
	float8Values(specE4M3, func(enc uint8, nan, neg bool, v *big.Float) {
		a := Float8E4M3(enc)
		checkFromFloat8(t, "Float8E4M3.Float16", enc, nan, neg, v, a.Float16().Float256())
		checkFromFloat8(t, "Float8E4M3.BFloat16", enc, nan, neg, v, a.BFloat16().Float256())
		checkFromFloat8(t, "Float8E4M3.Float32", enc, nan, neg, v, a.Float32().Float256())
		checkFromFloat8(t, "Float8E4M3.Float64", enc, nan, neg, v, a.Float64().Float256())
		checkFromFloat8(t, "Float8E4M3.Float128", enc, nan, neg, v, a.Float128().Float256())
		checkFromFloat8(t, "Float8E4M3.Float256", enc, nan, neg, v, a.Float256())
		if got := a.Float8E4M3(); got != a {
			t.Errorf("Float8E4M3(%#02x).Float8E4M3() = %#02x", enc, uint8(got))
		}

		// round trips
		if nan {
			return
		}
		for name, got := range map[string]Float8E4M3{
			"Float16":  a.Float16().Float8E4M3(),
			"BFloat16": a.BFloat16().Float8E4M3(),
			"Float32":  a.Float32().Float8E4M3(),
			"Float64":  a.Float64().Float8E4M3(),
			"Float128": a.Float128().Float8E4M3(),
			"Float256": a.Float256().Float8E4M3(),
		} {
			if got != a {
				t.Errorf("Float8E4M3(%#02x) via %s = %#02x", enc, name, uint8(got))
			}
		}
	})
}

func TestFloat8E5M2_ToOthers(t *testing.T) {
	t.Parallel()
	float8Values(specE5M2, func(enc uint8, nan, neg bool, v *big.Float) {
		a := Float8E5M2(enc)
		checkFromFloat8(t, "Float8E5M2.Float16", enc, nan, neg, v, a.Float16().Float256())
		checkFromFloat8(t, "Float8E5M2.BFloat16", enc, nan, neg, v, a.BFloat16().Float256())
		checkFromFloat8(t, "Float8E5M2.Float32", enc, nan, neg, v, a.Float32().Float256())
		checkFromFloat8(t, "Float8E5M2.Float64", enc, nan, neg, v, a.Float64().Float256())
		checkFromFloat8(t, "Float8E5M2.Float128", enc, nan, neg, v, a.Float128().Float256())
		checkFromFloat8(t, "Float8E5M2.Float256", enc, nan, neg, v, a.Float256())
		if got := a.Float8E5M2(); got != a {
			t.Errorf("Float8E5M2(%#02x).Float8E5M2() = %#02x", enc, uint8(got))
		}

		// round trips
		if nan {
			return
		}
		for name, got := range map[string]Float8E5M2{
			"Float16":  a.Float16().Float8E5M2(),
			"BFloat16": a.BFloat16().Float8E5M2(),
			"Float32":  a.Float32().Float8E5M2(),
			"Float64":  a.Float64().Float8E5M2(),
			"Float128": a.Float128().Float8E5M2(),
			"Float256": a.Float256().Float8E5M2(),
		} {
			if got != a {
				t.Errorf("Float8E5M2(%#02x) via %s = %#02x", enc, name, uint8(got))
			}
		}
	})
}

func TestFloat8_Cross(t *testing.T) {
	t.Parallel()
	for i := range 256 {
		a := Float8E4M3(i)
		if want, nan := specE5M2.convert(a.Float256()); nan {
			if got := a.Float8E5M2(); !specE5M2.isNaN(uint8(got)) {
				t.Errorf("Float8E4M3(%#02x).Float8E5M2() = %#02x, want NaN", i, uint8(got))
			}
		} else if got := a.Float8E5M2(); uint8(got) != want {
			t.Errorf("Float8E4M3(%#02x).Float8E5M2() = %#02x, want %#02x", i, uint8(got), want)
		}

		b := Float8E5M2(i)
		if want, nan := specE4M3.convert(b.Float256()); nan {
			if got := b.Float8E4M3(); !specE4M3.isNaN(uint8(got)) {
				t.Errorf("Float8E5M2(%#02x).Float8E4M3() = %#02x, want NaN", i, uint8(got))
			}
		} else if got := b.Float8E4M3(); uint8(got) != want {
			t.Errorf("Float8E5M2(%#02x).Float8E4M3() = %#02x, want %#02x", i, uint8(got), want)
		}
	}
}

func TestFloat8_Examples(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   float64
		e4m3 uint8
		e5m2 uint8
	}{
		{0, 0x00, 0x00},
		{math.Copysign(0, -1), 0x80, 0x80},
		{1, 0x38, 0x3c},
		{-2, 0xc0, 0xc0},
		{448, 0x7e, 0x5f},
		{464, 0x7e, 0x5f},   // tie to even
		{464.5, 0x7f, 0x5f}, // overflow
		{57344, 0x7f, 0x7b},
		{61440, 0x7f, 0x7c}, // tie to even
		{61439, 0x7f, 0x7b},
		{0x1p-9, 0x01, 0x18},
		{0x1p-10, 0x00, 0x14}, // tie to even
		{0x1p-16, 0x00, 0x01},
		{0x1p-17, 0x00, 0x00}, // tie to even
		{math.Inf(1), 0x7f, 0x7c},
		{math.Inf(-1), 0xff, 0xfc},
	}
	for _, tt := range tests {
		a := NewFloat64(tt.in)
		if got := NewFloat8E4M3(tt.in); uint8(got) != tt.e4m3 {
			t.Errorf("NewFloat8E4M3(%v) = %#02x, want %#02x", tt.in, uint8(got), tt.e4m3)
		}
		if got := NewFloat8E5M2(tt.in); uint8(got) != tt.e5m2 {
			t.Errorf("NewFloat8E5M2(%v) = %#02x, want %#02x", tt.in, uint8(got), tt.e5m2)
		}
		if got := a.Float8E4M3(); uint8(got) != tt.e4m3 {
			t.Errorf("Float8E4M3(%v) = %#02x, want %#02x", tt.in, uint8(got), tt.e4m3)
		}
		if got := a.Float8E5M2(); uint8(got) != tt.e5m2 {
			t.Errorf("Float8E5M2(%v) = %#02x, want %#02x", tt.in, uint8(got), tt.e5m2)
		}
	}
	if got := Float64(math.NaN()).Float8E4M3(); uint8(got)&0x7f != 0x7f {
		t.Errorf("Float8E4M3(NaN) = %#02x", uint8(got))
	}
	if got := Float64(math.NaN()).Float8E5M2(); !specE5M2.isNaN(uint8(got)) {
		t.Errorf("Float8E5M2(NaN) = %#02x", uint8(got))
	}
}
