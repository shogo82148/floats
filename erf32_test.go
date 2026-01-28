package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat32_Erf(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0), math.Erf(0)},
		{exact32(1), math.Erf(1)},
		{exact32(2), math.Erf(2)},
		{exact32(3), math.Erf(3)},
		{exact32(4), math.Erf(4)},
		{exact32(11), math.Erf(11)},
	}

	for _, tt := range tests {
		got := tt.x.Erf()
		if !close32(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(1)},
		{exact32(math.Inf(-1)), exact32(-1)},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erf()
		if !eq32(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Erf(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erf())
	}
}

func TestFloat32_Erfc(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0), math.Erfc(0)},
		{exact32(0x1p-14), math.Erfc(0x1p-14)},
		{exact32(1), math.Erfc(1)},
		{exact32(2), math.Erfc(2)},
	}

	for _, tt := range tests {
		got := tt.x.Erfc()
		if !close32(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(math.Inf(1)), exact32(0)},
		{exact32(math.Inf(-1)), exact32(2)},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfc()
		if !eq32(got, tt.want) {
			t.Errorf("Erfc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Erfc(b *testing.B) {
	x := exact32(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfc())
	}
}

func TestFloat32_Erfinv(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(-0.75), math.Erfinv(-0.75)},
		{exact32(-0.5), math.Erfinv(-0.5)},
		{exact32(-0.25), math.Erfinv(-0.25)},
		{exact32(0), math.Erfinv(0)},
		{exact32(0.25), math.Erfinv(0.25)},
		{exact32(0.5), math.Erfinv(0.5)},
		{exact32(0.75), math.Erfinv(0.75)},
	}

	for _, tt := range tests {
		got := tt.x.Erfinv()
		if !close32(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(1), exact32(math.Inf(1))},
		{exact32(-1), exact32(math.Inf(-1))},
		{exact32(2), exact32(math.NaN())},
		{exact32(-2), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfinv()
		if !eq32(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Erfinv(b *testing.B) {
	x := exact32(0.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfinv())
	}
}

func TestFloat32_Erfcinv(t *testing.T) {
	tests := []struct {
		x    Float32
		want float64
	}{
		{exact32(0.25), math.Erfcinv(0.25)},
		{exact32(0.5), math.Erfcinv(0.5)},
		{exact32(0.75), math.Erfcinv(0.75)},
		{exact32(1), math.Erfcinv(1)},
		{exact32(1.25), math.Erfcinv(1.25)},
		{exact32(1.5), math.Erfcinv(1.5)},
		{exact32(1.75), math.Erfcinv(1.75)},
	}

	for _, tt := range tests {
		got := tt.x.Erfcinv()
		if !close32(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float32
		want Float32
	}{
		// special cases
		{exact32(0), exact32(math.Inf(1))},
		{exact32(2), exact32(math.Inf(-1))},
		{exact32(3), exact32(math.NaN())},
		{exact32(-1), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfcinv()
		if !eq32(got, tt.want) {
			t.Errorf("Erfcinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat32_Erfcinv(b *testing.B) {
	x := exact32(0.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfcinv())
	}
}
