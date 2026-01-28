package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat256_Erf(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-1), "-0.8427007929497148693412206350826092592960669979663029084599378978347172540960108413"},
		{exact256(-0.5), "-0.5204998778130465376827466538919645287364515757579637000588057256471935217168535709"},
		{exact256(0), "0"},
		{exact256(0.5), "0.5204998778130465376827466538919645287364515757579637000588057256471935217168535709"},
		{exact256(1), "0.8427007929497148693412206350826092592960669979663029084599378978347172540960108413"},
		{exact256(2), "0.9953222650189527341620692563672529286108917970400600767383523262004372807199951774"},
		{exact256(3), "0.9999779095030014145586272238704176796201522929126007503427610451570575433163798677"},
		{exact256(4), "0.9999999845827420997199811478403265131159514278547464108088316570950057869589731887"},
		{exact256(11), "0.9999999999999999999999999999999999999999999999999999985591338620563053196601902971"},
	}

	for _, tt := range tests {
		got := tt.x.Erf()
		if !close256(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(math.Inf(1)), exact256(1)},
		{exact256(math.Inf(-1)), exact256(-1)},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erf()
		if !eq256(got, tt.want) {
			t.Errorf("Erf(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Erf(b *testing.B) {
	x := exact256(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erf())
	}
}

func TestFloat256_Erfinv(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-1).Nextafter(exact256(0)), "-12.6789204758043082417334375024209655173495726943819012599854669839592728"},
		{exact256(-0.75), "-0.813419847597618541690289359893421085324724835957501548147510003331798062"},
		{exact256(-0.5), "-0.476936276204469873381418353643130559808969749059470644703882695919383452"},
		{exact256(-0.25), "-0.225312055012178104725014013952277554782118447807246757600782894957738222"},
		{exact256(0), "0"},
		{exact256(0.25), "0.225312055012178104725014013952277554782118447807246757600782894957738222"},
		{exact256(0.5), "0.476936276204469873381418353643130559808969749059470644703882695919383452"},
		{exact256(0.75), "0.813419847597618541690289359893421085324724835957501548147510003331798062"},
		{exact256(1).Nextafter(exact256(0)), "12.6789204758043082417334375024209655173495726943819012599854669839592728"},
	}

	for _, tt := range tests {
		got := tt.x.Erfinv()
		if !close256(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(1), exact256(math.Inf(1))},
		{exact256(-1), exact256(math.Inf(-1))},
		{exact256(2), exact256(math.NaN())},
		{exact256(-2), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Erfinv()
		if !eq256(got, tt.want) {
			t.Errorf("Erfinv(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Erfinv(b *testing.B) {
	x := exact256(0.5)
	for b.Loop() {
		runtime.KeepAlive(x.Erfinv())
	}
}
