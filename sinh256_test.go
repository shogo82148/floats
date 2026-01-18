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
		{exact256(0.5), "0.5210953054937473616224256264114915591059289826114805279460935764528023"},
		{exact256(1), "1.175201193643801456882381850595600815155717981334095870229565413013308"},
		{exact256(2), "3.62686040784701876766821398280126170488634201232113572130948447493425"},
		{exact256(3), "10.01787492740990189897459361946582806017810412318286346440565325104639"},
		{exact256(4), "27.28991719712775244890827159079381858028941248553029656552849701555721"},
		{exact256(41), "319921746765027474.6113317017577854086582272160309332556820384242679222"},
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
