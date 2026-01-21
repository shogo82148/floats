package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat128_Log(t *testing.T) {
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(0.25), "-1.3862943611198906188344642429163531361510002687205"},
		{exact128(0.5), "-0.69314718055994530941723212145817656807550013436026"},
		{exact128(1), "0"},
		{exact128(2), "0.69314718055994530941723212145817656807550013436026"},
		{exact128(3), "1.0986122886681096913952452369225257046474905578228"},
		{exact128(4), "1.3862943611198906188344642429163531361510002687205"},
		{exact128(100), "4.6051701859880913680359829093687284152022029772575"},
	}

	for _, tt := range tests {
		got := tt.x.Log()
		if !close128(got, tt.want) {
			t.Errorf("Log(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(0), exact128(math.Inf(-1))},
		{exact128(math.Copysign(0, -1)), exact128(math.Inf(-1))},
		{exact128(-1), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log()
		if !eq128(got, tt.want) {
			t.Errorf("Log(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Log(b *testing.B) {
	x := exact128(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log())
	}
}

func TestFloat128_Log10(t *testing.T) {
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(0.25), "-0.602059991327962390427477789448986053536379762924217082620854922254216378548849019"},
		{exact128(0.5), "-0.3010299956639811952137388947244930267681898814621085413104274611271081892744245095"},
		{exact128(1), "0"},
		{exact128(2), "0.3010299956639811952137388947244930267681898814621085413104274611271081892744245095"},
		{exact128(3), "0.477121254719662437295027903255115309200128864190695864829865640305229152783661123"},
		{exact128(4), "0.602059991327962390427477789448986053536379762924217082620854922254216378548849019"},
		{exact128(100), "2"},
	}

	for _, tt := range tests {
		got := tt.x.Log10()
		if !close128(got, tt.want) {
			t.Errorf("Log10(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(0), exact128(math.Inf(-1))},
		{exact128(math.Copysign(0, -1)), exact128(math.Inf(-1))},
		{exact128(-1), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log10()
		if !eq128(got, tt.want) {
			t.Errorf("Log10(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Log10(b *testing.B) {
	x := exact128(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log10())
	}
}

func TestFloat128_Log2(t *testing.T) {
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(0.25), "-2"},
		{exact128(0.5), "-1"},
		{exact128(1), "0"},
		{exact128(2), "1"},
		{exact128(3), "1.584962500721156181453738943947816508759814407692481060455752654541098227794358563"},
		{exact128(4), "2"},
		{exact128(100), "6.643856189774724695740638858978780351729662786049161224109512791631869553217250432"},
	}

	for _, tt := range tests {
		got := tt.x.Log2()
		if !close128(got, tt.want) {
			t.Errorf("Log2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(0), exact128(math.Inf(-1))},
		{exact128(math.Copysign(0, -1)), exact128(math.Inf(-1))},
		{exact128(-1), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log2()
		if !eq128(got, tt.want) {
			t.Errorf("Log2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Log2(b *testing.B) {
	x := exact128(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log2())
	}
}
