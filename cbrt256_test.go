package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat256_Cbrt(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(3375), "15"},
		{exact256(27), "3"},
		{exact256(8), "2"},
		{exact256(2), "1.259921049894873164767210607278228350570251464701507980081975112155299676513959484"},
		{exact256(1), "1"},
		{exact256(0.5), "0.7937005259840997373758528196361541301957466639499265049041428809126082528121095866"},
		{exact256(0), "0"},
		{exact256(-0.5), "-0.7937005259840997373758528196361541301957466639499265049041428809126082528121095866"},
		{exact256(-1), "-1"},
		{exact256(-2), "-1.259921049894873164767210607278228350570251464701507980081975112155299676513959484"},
		{exact256(-8), "-2"},
		{exact256(-27), "-3"},
		{exact256(-3375), "-15"},
	}

	for _, test := range tests {
		got := test.x.Cbrt()
		if !close256(got, test.want) {
			t.Errorf("Cbrt(%v) = %v; want %v", test.x, got, test.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// Special cases
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(math.Inf(-1)), exact256(math.Inf(-1))},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Cbrt()
		if !eq256(got, tt.want) {
			t.Errorf("Float256.Cbrt(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Cbrt(b *testing.B) {
	x := NewFloat256(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Cbrt())
	}
}
