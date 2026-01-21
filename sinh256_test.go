package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat256_Sinh(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(0.25), "0.2526123168081683079141251505420579055197542874276608074880949653019810769068593791"},
		{exact256(0.5), "0.5210953054937473616224256264114915591059289826114805279460935764528023"},
		{exact256(1), "1.175201193643801456882381850595600815155717981334095870229565413013308"},
		{exact256(2), "3.62686040784701876766821398280126170488634201232113572130948447493425"},
		{exact256(3), "10.01787492740990189897459361946582806017810412318286346440565325104639"},
		{exact256(4), "27.28991719712775244890827159079381858028941248553029656552849701555721"},
		{exact256(85), "4111506357311456755152164008203887347.743143094238137019761966924872080517377059248"},
	}

	for _, tt := range tests {
		got := tt.x.Sinh()
		if !close256(got, tt.want) {
			t.Errorf("Sinh(%v) = %v; want %v", tt.x, got, tt.want)
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
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Sinh()
		if !eq256(got, tt.want) {
			t.Errorf("Sinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Sinh(b *testing.B) {
	x := exact256(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Sinh())
	}
}

func TestFloat256_Cosh(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(0), "1"},
		{exact256(1), "1.543080634815243778477905620757061682601529112365863704737402214710769"},
		{exact256(2), "3.762195691083631459562213477773746108293973558230711602777643347588324"},
		{exact256(3), "10.06766199577776584195393603511588983680980371537128667997328097865245"},
		{exact256(4), "27.30823283601648662920198961206705982250132455308377216029809694299646"},
		{exact256(43), "2363919734114673280.737228781372140185515632533070023109815169030476886"},
	}

	for _, tt := range tests {
		got := tt.x.Cosh()
		if !close256(got, tt.want) {
			t.Errorf("Cosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(0), exact256(1)},
		{exact256(math.Copysign(0, -1)), exact256(1)},
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(math.Inf(-1)), exact256(math.Inf(1))},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Cosh()
		if !eq256(got, tt.want) {
			t.Errorf("Cosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Cosh(b *testing.B) {
	x := exact256(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Cosh())
	}
}

func TestFloat256_Tanh(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(0), "0"},
		{exact256(0x1p-20), "9.536743164059608794206706373323853753492884138265436942109400747002263410862969192E-7"},
		{exact256(0.5), "0.4621171572600097585023184836436725487302892803301130385527318158380809061404092788"},
		{exact256(1), "0.7615941559557648881194582826047935904127685972579365515968105001219532445766384835"},
		{exact256(2), "0.9640275800758168839464137241009231502550299762409347760482632174131079463176102026"},
		{exact256(3), "0.9950547536867304513318801852554884750978138547002824918238788151306647027825591767"},
		{exact256(4), "0.9993292997390670437922433443417249620053398528944096480078068964566788861183201802"},
		{exact256(43), "0.9999999999999999999999999999999999999105244138763775853074468832166051200464931478"},
		{exact256(100000), "1"},
	}

	for _, tt := range tests {
		got := tt.x.Tanh()
		if !close256(got, tt.want) {
			t.Errorf("Tanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
		{exact256(math.Inf(1)), exact256(1)},
		{exact256(math.Inf(-1)), exact256(-1)},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Tanh()
		if !eq256(got, tt.want) {
			t.Errorf("Tanh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Tanh(b *testing.B) {
	x := exact256(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Tanh())
	}
}
