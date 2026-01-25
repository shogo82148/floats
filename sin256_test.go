package floats

import (
	"math"
	"testing"
)

func TestFloat256_Sin(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-2), "-0.9092974268256816953960198659117448427022549714478902683789730115309673015407835446"},
		{exact256(-1), "-0.841470984807896506652502321630298999622563060798371065672751709991910404391239669"},
		{exact256(1), "0.841470984807896506652502321630298999622563060798371065672751709991910404391239669"},
		{exact256(2), "0.9092974268256816953960198659117448427022549714478902683789730115309673015407835446"},
		{exact256(3), "0.1411200080598672221007448028081102798469332642522655841518826412324220099670144719"},
		{exact256(4), "-0.7568024953079282513726390945118290941359128873364725714854167734013104936191794164"},
		{exact256(0x1p35), "-0.6444303510232911324935356442109536997682925839222084839100843345486403626952028341"},
	}

	for _, tt := range tests {
		got := tt.x.Sin()
		if !close256(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
		{exact256(math.Inf(1)), exact256(math.NaN())},
		{exact256(math.Inf(-1)), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Sin()
		if !eq256(got, tt.want) {
			t.Errorf("Sin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat256_Cos(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-2), "-0.4161468365471423869975682295007621897660007710755448907551499737819649361240791691"},
		{exact256(-1), "0.5403023058681397174009366074429766037323104206179222276700972553811003947744717645"},
		{exact256(1), "0.5403023058681397174009366074429766037323104206179222276700972553811003947744717645"},
		{exact256(2), "-0.4161468365471423869975682295007621897660007710755448907551499737819649361240791691"},
		{exact256(3), "-0.9899924966004454572715727947312613023936790966155883288140859329283291975131332204"},
		{exact256(4), "-0.6536436208636119146391681830977503814241335966462182470070102838527376558106033799"},
		{exact256(0x1p35), "-0.7646630124963530453707219570852124444583931450059469568478191202614839150819875904"},
	}

	for _, tt := range tests {
		got := tt.x.Cos()
		if !close256(got, tt.want) {
			t.Errorf("Cos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(math.Inf(1)), exact256(math.NaN())},
		{exact256(math.Inf(-1)), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Cos()
		if !eq256(got, tt.want) {
			t.Errorf("Cos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat256_Sincos(t *testing.T) {
	tests := []struct {
		x   Float256
		sin string
		cos string
	}{
		{
			exact256(1),
			"0.841470984807896506652502321630298999622563060798371065672751709991910404391239669",
			"0.5403023058681397174009366074429766037323104206179222276700972553811003947744717645",
		},
		{
			exact256(2),
			"0.9092974268256816953960198659117448427022549714478902683789730115309673015407835446",
			"-0.4161468365471423869975682295007621897660007710755448907551499737819649361240791691",
		},
		{
			exact256(3),
			"0.1411200080598672221007448028081102798469332642522655841518826412324220099670144719",
			"-0.9899924966004454572715727947312613023936790966155883288140859329283291975131332204",
		},
	}

	for _, tt := range tests {
		sin, cos := tt.x.Sincos()
		if !close256(sin, tt.sin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", tt.x, sin, tt.sin)
		}
		if !close256(cos, tt.cos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", tt.x, cos, tt.cos)
		}
	}

	strictTests := []struct {
		x   Float256
		sin Float256
		cos Float256
	}{
		// special cases
		{exact256(0), exact256(0), exact256(1)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1)), exact256(1)},
		{exact256(math.Inf(1)), exact256(math.NaN()), exact256(math.NaN())},
		{exact256(math.Inf(-1)), exact256(math.NaN()), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		sin, cos := tt.x.Sincos()
		if !eq256(sin, tt.sin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", tt.x, sin, tt.sin)
		}
		if !eq256(cos, tt.cos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", tt.x, cos, tt.cos)
		}
	}
}
