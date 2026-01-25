package floats

import (
	"math"
	"testing"
)

func TestFloat256_Asinh(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(0), "0"},
		{exact256(0x1p-171), "3.340955887615244557675670583939352348518996100131310872452421865200498301233912948E-52"},
		{exact256(0.25), "0.247466461547263452944781549788359289253766903098567696469117357944365179443666365"},
		{exact256(0.5), "0.4812118250596034474977589134243684231351843343856605196610181688401638676082217744"},
		{exact256(1), "0.8813735870195430252326093249797923090281603282616354107532956086533771842220260878"},
		{exact256(21), "3.738236030261542702926345171286353186644971465768736202615459010981959402821114935"},
		{exact256(0x1p171), "119.2213150563105932197639248908063697089860231099639037087569616328637029787874911"},

		{exact256(-0), "-0"},
		{exact256(-0.25), "-0.247466461547263452944781549788359289253766903098567696469117357944365179443666365"},
		{exact256(-0.5), "-0.4812118250596034474977589134243684231351843343856605196610181688401638676082217744"},
		{exact256(-1), "-0.8813735870195430252326093249797923090281603282616354107532956086533771842220260878"},
		{exact256(-21), "-3.738236030261542702926345171286353186644971465768736202615459010981959402821114935"},
	}

	for _, tt := range tests {
		got := tt.x.Asinh()
		if !close256(got, tt.want) {
			t.Errorf("Asinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},

		// special cases
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(math.Inf(-1)), exact256(math.Inf(-1))},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Asinh()
		if !eq256(got, tt.want) {
			t.Errorf("Asinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat256_Acosh(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(1), "0"},
		{exact256(1.5), "0.9624236501192068949955178268487368462703686687713210393220363376803277352164435488"},
		{exact256(21), "3.737102242198923900976805054113159386884204741143073889422198983717266311856623625"},
		{exact256(0x1p59), "41.58883083359671856503392728749059408377769167708905124214080918578138008352876707"},
	}

	for _, tt := range tests {
		got := tt.x.Acosh()
		if !close256(got, tt.want) {
			t.Errorf("Acosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(math.Inf(-1)), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Acosh()
		if !eq256(got, tt.want) {
			t.Errorf("Acosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat256_Atanh(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(0x1p-171), "3.340955887615244557675670583939352348518996100131310872452421865200498301233912948E-52"},
		{exact256(0.25), "0.2554128118829953416027570481518309674390553982228841350889767789183423472445243989"},
		{exact256(0.5), "0.5493061443340548456976226184612628523237452789113747258673471668187471466093044834"},
		{exact256(0.75), "0.9729550745276566525526763717215898648185423647909305942296950749687899313760346339"},

		{exact256(-0x1p-171), "-3.340955887615244557675670583939352348518996100131310872452421865200498301233912948E-52"},
		{exact256(-0.25), "-0.2554128118829953416027570481518309674390553982228841350889767789183423472445243989"},
		{exact256(-0.5), "-0.5493061443340548456976226184612628523237452789113747258673471668187471466093044834"},
		{exact256(-0.75), "-0.9729550745276566525526763717215898648185423647909305942296950749687899313760346339"},
	}

	for _, tt := range tests {
		got := tt.x.Atanh()
		if !close256(got, tt.want) {
			t.Errorf("Atanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(2), exact256(math.NaN())},
		{exact256(1), exact256(math.Inf(1))},
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
		{exact256(-1), exact256(math.Inf(-1))},
		{exact256(-2), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Atanh()
		if !eq256(got, tt.want) {
			t.Errorf("Atanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
