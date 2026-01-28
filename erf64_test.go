package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat64_Erf(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(0), math.Erf(0)},
		{exact64(1), math.Erf(1)},
		{exact64(2), math.Erf(2)},
		{exact64(3), math.Erf(3)},
		{exact64(4), math.Erf(4)},
		{exact64(11), math.Erf(11)},
	}

	for _, tt := range tests {
		got := tt.x.Erf()
		if !close64(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(math.Inf(1)), exact64(1)},
		{exact64(math.Inf(-1)), exact64(-1)},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erf()
		if !eq64(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Erf(b *testing.B) {
	x := exact64(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erf())
	}
}

func TestFloat64_Erfc(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(0), math.Erfc(0)},
		{exact64(0x1p-14), math.Erfc(0x1p-14)},
		{exact64(1), math.Erfc(1)},
		{exact64(2), math.Erfc(2)},
	}

	for _, tt := range tests {
		got := tt.x.Erfc()
		if !close64(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(math.Inf(1)), exact64(0)},
		{exact64(math.Inf(-1)), exact64(2)},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfc()
		if !eq64(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Erfc(b *testing.B) {
	x := exact64(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfc())
	}
}

func TestFloat64_Erfinv(t *testing.T) {
	tests := []struct {
		x    Float64
		want float64
	}{
		{exact64(-0.75), math.Erfinv(-0.75)},
		{exact64(-0.5), math.Erfinv(-0.5)},
		{exact64(-0.25), math.Erfinv(-0.25)},
		{exact64(0), math.Erfinv(0)},
		{exact64(0.25), math.Erfinv(0.25)},
		{exact64(0.5), math.Erfinv(0.5)},
		{exact64(0.75), math.Erfinv(0.75)},
	}

	for _, tt := range tests {
		got := tt.x.Erfinv()
		if !close64(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float64
		want Float64
	}{
		// special cases
		{exact64(1), exact64(math.Inf(1))},
		{exact64(-1), exact64(math.Inf(-1))},
		{exact64(2), exact64(math.NaN())},
		{exact64(-2), exact64(math.NaN())},
		{exact64(math.NaN()), exact64(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfinv()
		if !eq64(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat64_Erfinv(b *testing.B) {
	x := exact64(0.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfinv())
	}
}
