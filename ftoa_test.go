package floats

import (
	"runtime"
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
