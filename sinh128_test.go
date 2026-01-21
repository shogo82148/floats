package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat128_Sinh(t *testing.T) {
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(0x1p-100), "7.888609052210118054117285652827862296732064351090230047702790124822869E-31"},
		{exact128(0.25), "0.2526123168081683079141251505420579055197542874276608074880949653019811"},
		{exact128(0.5), "0.5210953054937473616224256264114915591059289826114805279460935764528023"},
		{exact128(1), "1.175201193643801456882381850595600815155717981334095870229565413013308"},
		{exact128(2), "3.62686040784701876766821398280126170488634201232113572130948447493425"},
		{exact128(3), "10.01787492740990189897459361946582806017810412318286346440565325104639"},
		{exact128(4), "27.28991719712775244890827159079381858028941248553029656552849701555721"},
		{exact128(43), "2363919734114673280.737228781372140185304119429310915061152028929774621"},
	}

	for _, tt := range tests {
		got := tt.x.Sinh()
		if !close128(got, tt.want) {
			t.Errorf("Sinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		{exact128(0), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},

		// special cases
		{exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Sinh()
		if !eq128(got, tt.want) {
			t.Errorf("Sinh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Sinh(b *testing.B) {
	x := exact128(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Sinh())
	}
}

func TestFloat128_Cosh(t *testing.T) {
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(0), "1"},
		{exact128(1), "1.543080634815243778477905620757061682601529112365863704737402214710769"},
		{exact128(2), "3.762195691083631459562213477773746108293973558230711602777643347588324"},
		{exact128(3), "10.06766199577776584195393603511588983680980371537128667997328097865245"},
		{exact128(4), "27.30823283601648662920198961206705982250132455308377216029809694299646"},
		{exact128(43), "2363919734114673280.737228781372140185515632533070023109815169030476886"},
	}

	for _, tt := range tests {
		got := tt.x.Cosh()
		if !close128(got, tt.want) {
			t.Errorf("Cosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(0), exact128(1)},
		{exact128(math.Copysign(0, -1)), exact128(1)},
		{exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(math.Inf(-1)), exact128(math.Inf(1))},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Cosh()
		if !eq128(got, tt.want) {
			t.Errorf("Cosh(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Cosh(b *testing.B) {
	x := exact128(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Cosh())
	}
}
