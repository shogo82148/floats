package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat256_Exp(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-1), "0.3678794411714423215955237701614608674458111310317678345078368016974614957448998033571472743459196437466273252768439952082469757928"},
		{exact256(0), "1"},
		{exact256(1), "2.718281828459045235360287471352662497757247093699959574966967627724076630353547594571382178525166427427466391932003059921817413597"},
		{exact256(2), "7.389056098930650227230427460575007813180315570551847324087127822522573796079057763384312485079121794773753161265478866123884603693"},
		{exact256(3), "20.08553692318766774092852965458171789698790783855415014437893422969884587809197373120449716025301770215360761585194900288181101248"},
		{exact256(4), "54.59815003314423907811026120286087840279073703861406872582659395855366209993586948167698056194473414278411544577108742781175710394"},
	}

	for _, tt := range tests {
		got := tt.x.Exp()
		if !close256(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(math.Inf(-1)), exact256(0)},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Exp()
		if !eq256(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Exp(b *testing.B) {
	x := exact256(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Exp())
	}
}

func TestFloat256_Exp2(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-1), "0.5"},
		{exact256(0), "1"},
		{exact256(1), "2"},
		{exact256(1.5), "2.828427124746190097603377448419396157139343750753896146353359475981464956924214077700775068655283145470027692461824594049849672112"},
	}

	for _, tt := range tests {
		got := tt.x.Exp2()
		if !close256(got, tt.want) {
			t.Errorf("Exp2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		{exact256(262145), exact256(math.Inf(1))},
		{exact256(-262380), exact256(0)},

		// special cases
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(math.Inf(-1)), exact256(0)},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Exp2()
		if !eq256(got, tt.want) {
			t.Errorf("Exp2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
