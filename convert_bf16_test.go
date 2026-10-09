package floats

import (
	"math"
	"math/big"
	"math/rand/v2"
	"runtime"
	"testing"
)

// signedBigFromFloat256 converts a finite Float256 to an exact big.Float.
func signedBigFromFloat256(a Float256) *big.Float {
	f := bigFromFloat256(a)
	if a[0]>>63 != 0 {
		f.Neg(f)
	}
	return f
}

func TestBFloat16_ToOthers(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a := BFloat16(i)
		if a.IsNaN() {
			if !a.Float16().IsNaN() || !a.Float32().IsNaN() || !a.Float64().IsNaN() || !a.Float128().IsNaN() || !a.Float256().IsNaN() {
				t.Errorf("BFloat16(%#04x) is NaN, but the converted value is not NaN", i)
			}
			continue
		}
		want := math.Float32frombits(uint32(i) << 16)
		if got := a.Float32(); math.Float32bits(float32(got)) != math.Float32bits(want) {
			t.Errorf("BFloat16(%#04x).Float32() = %x, want %x", i, got, want)
		}
		if got := a.Float64(); math.Float64bits(float64(got)) != math.Float64bits(float64(want)) {
			t.Errorf("BFloat16(%#04x).Float64() = %x, want %x", i, got, want)
		}
		if got, want := a.Float128(), Float64(float64(want)).Float128(); got != want {
			t.Errorf("BFloat16(%#04x).Float128() = %x, want %x", i, got, want)
		}
		if got, want := a.Float256(), Float64(float64(want)).Float256(); got != want {
			t.Errorf("BFloat16(%#04x).Float256() = %x, want %x", i, got, want)
		}
		if got, want := a.Float16(), Float64(float64(want)).Float16(); got != want {
			t.Errorf("BFloat16(%#04x).Float16() = %x, want %x", i, got, want)
		}
		if got := a.BFloat16(); got != a {
			t.Errorf("BFloat16(%#04x).BFloat16() = %x", i, got)
		}

		// round trips
		if got := a.Float32().BFloat16(); got != a {
			t.Errorf("BFloat16(%#04x).Float32().BFloat16() = %#04x", i, uint16(got))
		}
		if got := a.Float64().BFloat16(); got != a {
			t.Errorf("BFloat16(%#04x).Float64().BFloat16() = %#04x", i, uint16(got))
		}
		if got := a.Float128().BFloat16(); got != a {
			t.Errorf("BFloat16(%#04x).Float128().BFloat16() = %#04x", i, uint16(got))
		}
		if got := a.Float256().BFloat16(); got != a {
			t.Errorf("BFloat16(%#04x).Float256().BFloat16() = %#04x", i, uint16(got))
		}
	}
}

func TestFloat16_BFloat16(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a := Float16(i)
		got := a.BFloat16()
		if a.IsNaN() {
			if !got.IsNaN() {
				t.Errorf("Float16(%#04x).BFloat16() = %#04x, want NaN", i, uint16(got))
			}
			continue
		}
		var want BFloat16
		if a.IsInf(0) {
			want = BFloat16(uint16(i&0x8000) | uvinfbf16)
		} else {
			want = bf16FromBig(new(big.Float).SetPrec(100).SetFloat64(float64(a.Float32())))
			if i == 0x8000 {
				want = 0x8000
			}
		}
		if got != want {
			t.Errorf("Float16(%#04x).BFloat16() = %#04x, want %#04x", i, uint16(got), uint16(want))
		}
	}
}

func TestFloat32_BFloat16(t *testing.T) {
	t.Parallel()
	check := func(b uint32) {
		t.Helper()
		a := Float32(math.Float32frombits(b))
		got := a.BFloat16()
		var want BFloat16
		switch {
		case a.IsNaN():
			if !got.IsNaN() {
				t.Fatalf("Float32(%#08x).BFloat16() = %#04x, want NaN", b, uint16(got))
			}
			return
		case a.IsInf(0):
			want = BFloat16(uint16(b>>16)&0x8000 | uvinfbf16)
		default:
			want = bf16FromBig(new(big.Float).SetPrec(100).SetFloat64(float64(a)))
			if a == 0 {
				want = BFloat16(b >> 16 & 0x8000)
			}
		}
		if got != want {
			t.Fatalf("Float32(%#08x).BFloat16() = %#04x, want %#04x", b, uint16(got), uint16(want))
		}
	}

	// ties, and the values around them
	for hi := range 1 << 16 {
		for _, lo := range []uint32{0, 1, 0x7fff, 0x8000, 0x8001, 0xffff} {
			check(uint32(hi)<<16 | lo)
		}
	}

	r := rand.New(rand.NewPCG(1, 2))
	for range 300_000 {
		check(r.Uint32())
		check(r.Uint32() & 0x807f_ffff) // subnormal
	}
}

// perturb returns the values around the midpoint of a and the next BFloat16 of a.
func midpointsFloat32(r *rand.Rand) (Float32, BFloat16) {
	a := BFloat16(r.Uint32())
	return Float32(math.Float32frombits(uint32(a)<<16 | 0x8000)), a
}

func TestFloat64_BFloat16(t *testing.T) {
	t.Parallel()
	check := func(f float64) {
		t.Helper()
		a := Float64(f)
		got := a.BFloat16()
		var want BFloat16
		switch {
		case math.IsNaN(f):
			if !got.IsNaN() {
				t.Fatalf("Float64(%x).BFloat16() = %#04x, want NaN", f, uint16(got))
			}
			return
		case math.IsInf(f, 0):
			want = BFloat16(uint16(math.Float64bits(f)>>48)&0x8000 | uvinfbf16)
		case f == 0:
			want = BFloat16(math.Float64bits(f) >> 48 & 0x8000)
		default:
			want = bf16FromBig(new(big.Float).SetPrec(100).SetFloat64(f))
		}
		if got != want {
			t.Fatalf("Float64(%x).BFloat16() = %#04x, want %#04x", f, uint16(got), uint16(want))
		}
	}

	r := rand.New(rand.NewPCG(3, 4))
	for range 300_000 {
		check(math.Float64frombits(r.Uint64()))
		// the values around the ties
		mid, _ := midpointsFloat32(r)
		if mid.IsNaN() || mid.IsInf(0) {
			continue
		}
		b := math.Float64bits(float64(mid))
		for _, d := range []uint64{0, 1, 2, 1 << 20, 1 << 29} {
			check(math.Float64frombits(b + d))
			check(math.Float64frombits(b - d))
		}
	}
	for _, f := range []float64{math.Inf(1), math.Inf(-1), math.NaN(), 0, math.Copysign(0, -1), math.MaxFloat64, math.SmallestNonzeroFloat64, -math.MaxFloat64, -math.SmallestNonzeroFloat64, math.MaxFloat32, 0x1p-134, 0x1p-134 + 0x1p-1074, 0x1p-134 - 0x1p-1074} {
		check(f)
	}
}

func TestFloat128_BFloat16(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(5, 6))
	check := func(a Float128) {
		t.Helper()
		got := a.BFloat16()
		var want BFloat16
		switch {
		case a.IsNaN():
			if !got.IsNaN() {
				t.Fatalf("Float128(%x).BFloat16() = %#04x, want NaN", a, uint16(got))
			}
			return
		case a.IsInf(0):
			want = BFloat16(uint16(a[0]>>48)&0x8000 | uvinfbf16)
		case a.IsZero():
			want = BFloat16(a[0] >> 48 & 0x8000)
		default:
			want = bf16FromBig(signedBigFromFloat256(a.Float256()))
		}
		if got != want {
			t.Fatalf("Float128(%x).BFloat16() = %#04x, want %#04x", a, uint16(got), uint16(want))
		}
	}
	for range 100_000 {
		check(Float128{r.Uint64(), r.Uint64()})
		mid, _ := midpointsFloat32(r)
		if mid.IsNaN() || mid.IsInf(0) {
			continue
		}
		m := mid.Float128()
		for _, d := range []uint64{1, 2, 1 << 20, 1 << 40, 1 << 62} {
			up := Float128(ints128Add(m, d, false))
			down := Float128(ints128Add(m, d, true))
			check(up)
			check(down)
		}
		check(m)
	}
	// around the overflow threshold
	max := BFloat16(0x7f7f).Float32()
	for _, d := range []uint64{0, 1, 1 << 30} {
		m := Float32(math.Float32frombits(math.Float32bits(float32(max)) + 0x8000)).Float128()
		check(Float128(ints128Add(m, d, false)))
		check(Float128(ints128Add(m, d, true)))
	}
}

// ints128Add adds d to (or subtracts d from) the magnitude of the finite Float128 value a.
func ints128Add(a Float128, d uint64, sub bool) [2]uint64 {
	lo, hi := a[1], a[0]
	if sub {
		if lo < d {
			hi--
		}
		lo -= d
	} else {
		lo += d
		if lo < d {
			hi++
		}
	}
	return [2]uint64{hi, lo}
}

func TestFloat256_BFloat16(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(7, 8))
	check := func(a Float256) {
		t.Helper()
		got := a.BFloat16()
		var want BFloat16
		switch {
		case a.IsNaN():
			if !got.IsNaN() {
				t.Fatalf("Float256(%x).BFloat16() = %#04x, want NaN", a, uint16(got))
			}
			return
		case a.IsInf(0):
			want = BFloat16(uint16(a[0]>>48)&0x8000 | uvinfbf16)
		case a.IsZero():
			want = BFloat16(a[0] >> 48 & 0x8000)
		default:
			want = bf16FromBig(signedBigFromFloat256(a))
		}
		if got != want {
			t.Fatalf("Float256(%x).BFloat16() = %#04x, want %#04x", a, uint16(got), uint16(want))
		}
	}
	for range 100_000 {
		check(Float256{r.Uint64(), r.Uint64(), r.Uint64(), r.Uint64()})
		mid, _ := midpointsFloat32(r)
		if mid.IsNaN() || mid.IsInf(0) {
			continue
		}
		m := mid.Float256()
		for _, d := range []uint64{1, 2, 1 << 20, 1 << 40, 1 << 62} {
			for _, sub := range []bool{false, true} {
				v := m
				lo := v[3]
				if sub {
					if lo < d {
						// borrow
						v[2]--
						if v[2] == math.MaxUint64 {
							v[1]--
							if v[1] == math.MaxUint64 {
								v[0]--
							}
						}
					}
					v[3] = lo - d
				} else {
					v[3] = lo + d
					if v[3] < d {
						v[2]++
						if v[2] == 0 {
							v[1]++
							if v[1] == 0 {
								v[0]++
							}
						}
					}
				}
				check(v)
			}
		}
		check(m)
	}
}

func BenchmarkFloat32_BFloat16(b *testing.B) {
	for i := 0; b.Loop(); i++ {
		bits := math.Float32bits(float32(BFloat16(i).Float32())) | uint32((uint64(i)*0x9e3779b97f4a7c15)>>48)
		runtime.KeepAlive(Float32(math.Float32frombits(bits)).BFloat16())
	}
}
