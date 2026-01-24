package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat128_Cbrt(t *testing.T) {
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(3375), "15"},
		{exact128(27), "3"},
		{exact128(8), "2"},
		{exact128(2), "1.259921049894873164767210607278228350570251464701507980081975112155299676513959484"},
		{exact128(1), "1"},
		{exact128(0.5), "0.7937005259840997373758528196361541301957466639499265049041428809126082528121095866"},
		{exact128(0), "0"},
		{exact128(-0.5), "-0.7937005259840997373758528196361541301957466639499265049041428809126082528121095866"},
		{exact128(-1), "-1"},
		{exact128(-2), "-1.259921049894873164767210607278228350570251464701507980081975112155299676513959484"},
		{exact128(-8), "-2"},
		{exact128(-27), "-3"},
		{exact128(-3375), "-15"},
	}

	for _, test := range tests {
		got := test.x.Cbrt()
		if !close128(got, test.want) {
			t.Errorf("Cbrt(%v) = %v; want %v", test.x, got, test.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// Special cases
		{exact128(0), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},
		{exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(math.Inf(-1)), exact128(math.Inf(-1))},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Cbrt()
		if !eq128(got, tt.want) {
			t.Errorf("Float128.Cbrt(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Cbrt(b *testing.B) {
	x := NewFloat128(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Cbrt())
	}
}
