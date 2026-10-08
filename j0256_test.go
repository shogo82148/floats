package floats

import (
	"math"
	"runtime"
	"testing"
)

func TestFloat256_J0(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(0.001), "0.99999975000001562499955556388944908789228085295344841856146202193800421"},
		{exact256(0.5), "0.938469807240812904228404673599712625568926797096821576554705168024483425"},
		{exact256(1), "0.765197686557966551449717526102663220909274289755325241861547549119278915"},
		{exact256(-1), "0.765197686557966551449717526102663220909274289755325241861547549119278915"},
		{exact256(2), "0.223890779141235668051827454649948625825154482218607603128349706010853957"},
		{exact256(5), "-0.177596771314338304347397013074758711071130356008509128990658268208176604"},
		{exact256(10), "-0.245935764451348335197760862485328753829600072826566569699158393641165344"},
		{exact256(50), "0.055812327669251815004750478529433968176592671045578136196613253154920232"},
		{exact256(100), "0.0199858503042231224242283909508489906806335788590279295586421144472257617"},
		{exact256(300), "-0.033298554876305668007483094399845506597811583512877179452888853667873014"},

		// straddling the Miller / asymptotic crossover (500)
		{exact256(499), "-0.00959309963497892125727745113269906270369147536559013074215007087539142256"},
		{exact256(499.9), "-0.0328845634278467585544905323264667216501471458296806103915102259460048454"},
		{exact256(500), "-0.0341005568807319982651250604518945581314427915316838704826030197473365708"},
		{exact256(500.1), "-0.034975619734642624401914548092351975052855910052023691515126910695090064"},
		{exact256(501), "-0.0272384637476335846730432521095317102031836011425477990413031118487782047"},

		{exact256(1000), "0.0247866861524201745613307311156937087861664471332465484148069500135488624"},
		{exact256(5000), "-0.00664898425144834789358688489684694028828180428391891374649020864164147328"},
		{exact256(50000), "-0.00256784217783323990184372699495243756532821944730674169631089453607131253"},
	}

	for _, tt := range tests {
		got := tt.x.J0()
		if !close256(got, tt.want) {
			t.Errorf("J0(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		{exact256(0), exact256(1)},
		{exact256(math.Copysign(0, -1)), exact256(1)},
		{exact256(math.Inf(1)), exact256(0)},
		{exact256(math.Inf(-1)), exact256(0)},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.J0()
		if !eq256(got, tt.want) {
			t.Errorf("J0(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat256_J0Accuracy requires the correctly rounded result for every vector of the test data.
func TestFloat256_J0Accuracy(t *testing.T) {
	t.Parallel()
	checkFloat256Testdata(t, "testdata/j0256.txt", "J0", Float256.J0)
}

func BenchmarkFloat256_J0(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float256
	}{
		{"one", exact256(1)},                       // 1 - x**2/4 + ... is rounded to 1 for tiny x
		{"tiny", exact256(0x1p-110)},               // exact rounding of 1 - x**2/4
		{"taylor", exact256(1.5)},                  // x < 16
		{"miller", exact256(50.5)},                 // 16 <= x < 110
		{"hankel-small", exact256(200.5)},          // 110 <= x
		{"hankel-large", exact256(1e5)},            // the argument reduction
		{"hankel-huge", exact256(0x1p1000)},        // Payne-Hanek
		{"near-zero", exact256(2.404825557695773)}, // the Taylor series around a zero
		{"negative", exact256(-50.5)},              // J0 is even
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.J0())
			}
		})
	}
}

func TestFloat256_J0Internals(t *testing.T) {
	t.Parallel()

	// j0Result returns 0 for 0
	if got := j0Result(lgammaFix256{}, 0); !eq256(got, Float256{}) {
		t.Errorf("j0Result(0) = %v", got)
	}

	// j0Near256 handles only the values close to the zeros of the table.
	for _, x := range []Float256{exact256(3), exact256(2.4), exact256(1000), exact256(0.5)} {
		_, exp, m := x.normalize()
		if got, ok := j0Near256(exp, m); ok {
			t.Errorf("j0Near256(%v) = %v, true; want false", x, got)
		}
	}

	// float256ToFix
	if got := float256ToFix(Float256{}); got != (lgammaFix256{}) {
		t.Errorf("float256ToFix(0) = %v", got)
	}
	if got, want := float256ToFix(exact256(-2.5)), (lgammaFix256{true, gammaFix256{2, 1 << 63}}); got != want {
		t.Errorf("float256ToFix(-2.5) = %v; want %v", got, want)
	}

	// rotateOctant returns cos(k pi/4) and sin(k pi/4) for z = 0.
	one, zero := lgammaFix256{false, gammaOne256}, lgammaFix256{}
	h := gammaFix256(j0256HalfSqrt2)
	for k, want := range [8][2]lgammaFix256{
		{{false, gammaOne256}, {}},
		{{false, h}, {false, h}},
		{{}, {false, gammaOne256}},
		{{true, h}, {false, h}},
		{{true, gammaOne256}, {}},
		{{true, h}, {true, h}},
		{{}, {true, gammaOne256}},
		{{false, h}, {true, h}},
	} {
		c, s := rotateOctant(k, one, zero)
		if c.v != want[0].v || s.v != want[1].v || (c.v != gammaFix256{} && c.neg != want[0].neg) || (s.v != gammaFix256{} && s.neg != want[1].neg) {
			t.Errorf("rotateOctant(%d) = %v, %v; want %v, %v", k, c, s, want[0], want[1])
		}
	}
}
