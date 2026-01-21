package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat256_Log(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(0.25), "-1.386294361119890618834464242916353136151000268720510508241360018986787243939389431211726653992837375084002962041141371467371040472"},
		{exact256(0.5), "-0.6931471805599453094172321214581765680755001343602552541206800094933936219696947156058633269964186875420014810205706857336855202358"},
		{exact256(1), "0"},
		{exact256(2), "0.6931471805599453094172321214581765680755001343602552541206800094933936219696947156058633269964186875420014810205706857336855202358"},
		{exact256(3), "1.098612288668109691395245236922525704647490557822749451734694333637494293218608966873615754813732088787970029065957865742368004226"},
		{exact256(4), "1.386294361119890618834464242916353136151000268720510508241360018986787243939389431211726653992837375084002962041141371467371040472"},
		{exact256(100), "4.60517018598809136803598290936872841520220297725754595206665580193514521935470496047199441017919659668393556808457249726681905093"},
	}

	for _, tt := range tests {
		got := tt.x.Log()
		if !close256(got, tt.want) {
			t.Errorf("Log(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(0), exact256(math.Inf(-1))},
		{exact256(math.Copysign(0, -1)), exact256(math.Inf(-1))},
		{exact256(-1), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log()
		if !eq256(got, tt.want) {
			t.Errorf("Log(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Log(b *testing.B) {
	x := exact256(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log())
	}
}

func TestFloat256_Log10(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(0.25), "-0.602059991327962390427477789448986053536379762924217082620854922254216378548849019"},
		{exact256(0.5), "-0.3010299956639811952137388947244930267681898814621085413104274611271081892744245095"},
		{exact256(1), "0"},
		{exact256(2), "0.3010299956639811952137388947244930267681898814621085413104274611271081892744245095"},
		{exact256(3), "0.477121254719662437295027903255115309200128864190695864829865640305229152783661123"},
		{exact256(4), "0.602059991327962390427477789448986053536379762924217082620854922254216378548849019"},
		{exact256(100), "2"},
	}

	for _, tt := range tests {
		got := tt.x.Log10()
		if !close256(got, tt.want) {
			t.Errorf("Log10(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(0), exact256(math.Inf(-1))},
		{exact256(math.Copysign(0, -1)), exact256(math.Inf(-1))},
		{exact256(-1), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log10()
		if !eq256(got, tt.want) {
			t.Errorf("Log10(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Log10(b *testing.B) {
	x := exact256(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log10())
	}
}

func TestFloat256_Log2(t *testing.T) {
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(0.25), "-2"},
		{exact256(0.5), "-1"},
		{exact256(1), "0"},
		{exact256(2), "1"},
		{exact256(3), "1.584962500721156181453738943947816508759814407692481060455752654541098227794358563"},
		{exact256(4), "2"},
		{exact256(100), "6.643856189774724695740638858978780351729662786049161224109512791631869553217250432"},
	}

	for _, tt := range tests {
		got := tt.x.Log2()
		if !close256(got, tt.want) {
			t.Errorf("Log2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(math.Inf(1)), exact256(math.Inf(1))},
		{exact256(0), exact256(math.Inf(-1))},
		{exact256(math.Copysign(0, -1)), exact256(math.Inf(-1))},
		{exact256(-1), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Log2()
		if !eq256(got, tt.want) {
			t.Errorf("Log2(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat256_Log2(b *testing.B) {
	x := exact256(1.5)
	for b.Loop() {
		runtime.KeepAlive(x.Log2())
	}
}
