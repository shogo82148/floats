package floats

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"testing"
)

func BenchmarkString(b *testing.B) {
	b.Run("float16", func(b *testing.B) {
		f16 := exact16(0x1p-24)
		for b.Loop() {
			runtime.KeepAlive(f16.String())
		}
	})
	b.Run("float32", func(b *testing.B) {
		f32 := exact32(0x1p-24)
		for b.Loop() {
			runtime.KeepAlive(f32.String())
		}
	})
	b.Run("float64", func(b *testing.B) {
		f64 := exact64(0x1p-24)
		for b.Loop() {
			runtime.KeepAlive(f64.String())
		}
	})
	b.Run("float128", func(b *testing.B) {
		f128 := exact128(0x1p-24)
		for b.Loop() {
			runtime.KeepAlive(f128.String())
		}
	})
	b.Run("float256", func(b *testing.B) {
		f256 := exact256(0x1p-24)
		for b.Loop() {
			runtime.KeepAlive(f256.String())
		}
	})
}

// formatCases returns the formats to be compared with the fmt package.
func formatCases() []string {
	var ret []string
	for _, verb := range []string{"v", "g", "G", "e", "E", "f", "F", "x", "X", "b"} {
		for _, flags := range []string{"", "+", " ", "-", "0", "+0", "-+", " 0", "- ", "+ "} {
			for _, width := range []string{"", "1", "12"} {
				for _, prec := range []string{"", ".0", ".3"} {
					if prec == "" && strings.Contains("eEfF", verb) {
						// Without the precision, these verbs print the shortest representation.
						continue
					}
					ret = append(ret, "%"+flags+width+prec+verb)
				}
			}
		}
	}
	return ret
}

func TestFormat_Float64(t *testing.T) {
	t.Parallel()
	values := []float64{
		0, math.Copysign(0, -1), 1, -1, 1.5, -2.5, 123456789, 1e-7, -1e21, 0.1, 3.14159265358979,
		math.MaxFloat64, math.SmallestNonzeroFloat64, math.Inf(1), math.Inf(-1), math.NaN(),
	}
	for _, f := range formatCases() {
		for _, v := range values {
			if got, want := fmt.Sprintf(f, Float64(v)), fmt.Sprintf(f, v); got != want {
				t.Errorf("Sprintf(%q, %v) = %q, want %q", f, v, got, want)
			}
		}
	}
}

func TestFormat_Float32(t *testing.T) {
	t.Parallel()
	values := []float32{
		0, float32(math.Copysign(0, -1)), 1, -1, 1.5, -2.5, 123456789, 1e-7, -1e21, 0.1, 3.14159265,
		math.MaxFloat32, math.SmallestNonzeroFloat32, float32(math.Inf(1)), float32(math.Inf(-1)), float32(math.NaN()),
	}
	for _, f := range formatCases() {
		for _, v := range values {
			if got, want := fmt.Sprintf(f, Float32(v)), fmt.Sprintf(f, v); got != want {
				t.Errorf("Sprintf(%q, %v) = %q, want %q", f, v, got, want)
			}
		}
	}
}

func TestFormat_BadVerb(t *testing.T) {
	t.Parallel()
	got := fmt.Sprintf("%d", Float16(0x3c00))
	if want := "%!d(floats.Float16=1)"; got != want {
		t.Errorf("Sprintf(%%d) = %q, want %q", got, want)
	}
	if got := fmt.Sprintf("%s", Float64(1.5)); !strings.HasPrefix(got, "%!s(") {
		t.Errorf("Sprintf(%%s) = %q", got)
	}
}
