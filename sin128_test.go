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

func TestFloat128_Cos(t *testing.T) {
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(-2), "-0.4161468365471423869975682295007621897660007710755448907551499737819649361240791691"},
		{exact128(-1), "0.5403023058681397174009366074429766037323104206179222276700972553811003947744717645"},
		{exact128(1), "0.5403023058681397174009366074429766037323104206179222276700972553811003947744717645"},
		{exact128(2), "-0.4161468365471423869975682295007621897660007710755448907551499737819649361240791691"},
		{exact128(3), "-0.9899924966004454572715727947312613023936790966155883288140859329283291975131332204"},
		{exact128(4), "-0.6536436208636119146391681830977503814241335966462182470070102838527376558106033799"},
		{exact128(0x1p35), "-0.7646630124963530453707219570852124444583931450059469568478191202614839150819875904"},
	}

	for _, tt := range tests {
		got := tt.x.Cos()
		if !close128(got, tt.want) {
			t.Errorf("Cos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(math.Inf(1)), exact128(math.NaN())},
		{exact128(math.Inf(-1)), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.NaN())},
	}
	for _, tt := range strictTests {
		got := tt.x.Cos()
		if !eq128(got, tt.want) {
			t.Errorf("Cos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat128_Sincos(t *testing.T) {
	tests := []struct {
		x   Float128
		sin string
		cos string
	}{
		{
			exact128(1),
			"0.841470984807896506652502321630298999622563060798371065672751709991910404391239669",
			"0.5403023058681397174009366074429766037323104206179222276700972553811003947744717645",
		},
		{
			exact128(2),
			"0.9092974268256816953960198659117448427022549714478902683789730115309673015407835446",
			"-0.4161468365471423869975682295007621897660007710755448907551499737819649361240791691",
		},
		{
			exact128(3),
			"0.1411200080598672221007448028081102798469332642522655841518826412324220099670144719",
			"-0.9899924966004454572715727947312613023936790966155883288140859329283291975131332204",
		},
	}

	for _, tt := range tests {
		sin, cos := tt.x.Sincos()
		if !close128(sin, tt.sin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", tt.x, sin, tt.sin)
		}
		if !close128(cos, tt.cos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", tt.x, cos, tt.cos)
		}
	}

	strictTests := []struct {
		x   Float128
		sin Float128
		cos Float128
	}{
		// special cases
		{exact128(0), exact128(0), exact128(1)},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1)), exact128(1)},
		{exact128(math.Inf(1)), exact128(math.NaN()), exact128(math.NaN())},
		{exact128(math.Inf(-1)), exact128(math.NaN()), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		sin, cos := tt.x.Sincos()
		if !eq128(sin, tt.sin) {
			t.Errorf("Sincos(%v) sin = %v; want %v", tt.x, sin, tt.sin)
		}
		if !eq128(cos, tt.cos) {
			t.Errorf("Sincos(%v) cos = %v; want %v", tt.x, cos, tt.cos)
		}
	}
}

func TestFloat128_Tan(t *testing.T) {
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(-2), "2.185039863261518991643306102313682543432017746227663164562955869966773747209194182"},
		{exact128(-1), "-1.55740772465490223050697480745836017308725077238152003838394660569886139715172729"},
		{exact128(1), "1.55740772465490223050697480745836017308725077238152003838394660569886139715172729"},
		{exact128(2), "-2.185039863261518991643306102313682543432017746227663164562955869966773747209194182"},
		{exact128(3), "-0.1425465430742778052956354105339134932260922849018046476332389766888585952215385381"},
		{exact128(4), "1.157821282349577583137342418267323923119762767367142130084857189358985762063503791"},
	}

	for _, tt := range tests {
		got := tt.x.Tan()
		if !close128(got, tt.want) {
			t.Errorf("Tan(%v) = %v; want %v", tt.x, got, tt.want)
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
		got := tt.x.Tan()
		if !eq128(got, tt.want) {
			t.Errorf("Tan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
