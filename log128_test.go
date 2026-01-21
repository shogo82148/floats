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
