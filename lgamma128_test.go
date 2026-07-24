package floats

import (
	"math"
	"testing"
)

func TestFloat128_Lgamma(t *testing.T) {
	tests := []struct {
		x    Float128
		want string
		sign int
	}{
		{exact128(0.5), "0.5723649429247000870717136756765293558236", 1},
		{exact128(2.5), "0.2846828704729191596324946696827019243201", 1},
		{exact128(-0.5), "1.265512123484645396488945797134705923899", -1},
		{exact128(-2.5), "-0.05624371649767405067259453009765428412294", -1},
		{exact128(100), "359.1342053695753987760440104602869096126", 1},

		// around MaxStirling = 1756, straddling the Gamma()-delegate /
		// log-Stirling boundary
		{exact128(1755), "11352.42723245307564687249568604776301785", 1},
		{exact128(1755.5), "11356.16227329595536842840606592791090684", 1},
		{exact128(1756), "11359.89745658897561305733566853332210121", 1},
		{exact128(1756.5), "11363.63278229156385536179632102015464708", 1},
		{exact128(5000), "37582.6263156853503317465661476968580828", 1},
		{exact128(-1755.5), "-11362.48805240571445518765289366880158837", 1},
		{exact128(-1756.5), "-11369.95913087742033691146413996420582514", -1},
		{exact128(-5000.5), "-37594.25745058162566268740632072478674085", -1},
	}

	for _, tt := range tests {
		got, sign := tt.x.Lgamma()
		if !close128(got, tt.want) || sign != tt.sign {
			t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", tt.x, got, sign, tt.want, tt.sign)
		}
	}

	strictTests := []struct {
		x        Float128
		want     Float128
		wantSign int
	}{
		// special cases
		{exact128(math.Inf(1)), exact128(math.Inf(1)), 1},
		{exact128(math.Inf(-1)), exact128(math.Inf(-1)), 1},
		{exact128(0), exact128(math.Inf(1)), 1},
		{exact128(math.Copysign(0, -1)), exact128(math.Inf(1)), 1},
		{exact128(-1), exact128(math.Inf(1)), 1},
		{exact128(-2), exact128(math.Inf(1)), 1},
		{exact128(math.NaN()), exact128(math.NaN()), 1},
	}

	for _, tt := range strictTests {
		got, sign := tt.x.Lgamma()
		if !eq128(got, tt.want) || sign != tt.wantSign {
			t.Errorf("Lgamma(%v) = (%v, %d); want (%v, %d)", tt.x, got, sign, tt.want, tt.wantSign)
		}
	}
}
