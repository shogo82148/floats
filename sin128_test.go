package floats

import (
	"math"
	"testing"
)

func TestFloat128_Sin(t *testing.T) {
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(-2), "-0.9092974268256816953960198659117448427022549714478902683789730115309673015407835446"},
		{exact128(-1), "-0.841470984807896506652502321630298999622563060798371065672751709991910404391239669"},
		{exact128(1), "0.841470984807896506652502321630298999622563060798371065672751709991910404391239669"},
		{exact128(2), "0.9092974268256816953960198659117448427022549714478902683789730115309673015407835446"},
		{exact128(3), "0.1411200080598672221007448028081102798469332642522655841518826412324220099670144719"},
		{exact128(4), "-0.7568024953079282513726390945118290941359128873364725714854167734013104936191794164"},
		{exact128(0x1p35), "-0.6444303510232911324935356442109536997682925839222084839100843345486403626952028341"},
	}

	for _, tt := range tests {
		got := tt.x.Sin()
		if !close128(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(0), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},
		{exact128(math.Inf(1)), exact128(math.NaN())},
		{exact128(math.Inf(-1)), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Sin()
		if !eq128(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
