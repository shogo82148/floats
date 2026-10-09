package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Y0(t *testing.T) {
	t.Parallel()
	tests := []float64{0.5, 1, 2, 5, 10, 50}

	for _, x := range tests {
		want := math.Y0(x)
		got := exact32(x).Y0()
		if !close32(got, want) {
			t.Errorf("Y0(%v) = %v; want %v", x, got, want)
		}
	}
}

func TestFloat32_Y0Special(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		x    Float32
		want Float32
	}{
		{exact32(0), exact32(math.Inf(-1))},
		{exact32(math.Copysign(0, -1)), exact32(math.Inf(-1))},
		{exact32(-1), exact32(math.NaN())},
		{exact32(-5), exact32(math.NaN())},
		{exact32(math.Inf(1)), exact32(0)},
		{exact32(math.Inf(-1)), exact32(math.NaN())},
		{exact32(math.NaN()), exact32(math.NaN())},
	} {
		if got := tt.x.Y0(); !eq32(got, tt.want) {
			t.Errorf("Y0(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_Y0HardCases checks Y0 on the inputs whose results are very close to the midpoint of two adjacent
// Float32 values, which are calculated by the more accurate Float256. They are found by checking all the Float32 values
// in [2, 64), and the correctly rounded results were calculated with mpmath.
func TestFloat32_Y0HardCases(t *testing.T) {
	t.Parallel()
	// x, and the correctly rounded Y0(x)
	tests := [][2]uint32{
		{0x4026f6b5, 0x3ef5956e},
		{0x404e9a38, 0x3e97d3e9},
		{0x40599fdc, 0x3e6af81e},
		{0x407195b5, 0x3d9988a6},
		{0x40741399, 0x3d70fdcf},
		{0x407b7d7b, 0x3c3a4299},
		{0x407c4938, 0x3bcf9eca},
		{0x407cf0b3, 0x3b10dc4b},
		{0x407cfc52, 0x3afc3fc0},
		{0x407d32af, 0x3a1a164c},
		{0x407d3c6b, 0x39b6bed3},
		{0x407d3e02, 0x39a24306},
		{0x407d4502, 0x391028c4},
		{0x407d471a, 0x38b46df3},
		{0x407d49c0, 0x37afda10},
		{0x407d49d3, 0x37a08e13},
		{0x407d4a2d, 0x37303179},
		{0x407d4a57, 0x36d921a8},
		{0x407d4a90, 0x3586494b},
		{0x407d4a92, 0x35590c0e},
		{0x407d4a97, 0x34b07780},
		{0x407d4a98, 0x3479e1f5},
		{0x407d4a99, 0x3412d4ea},
		{0x407d4a9a, 0x332f1f83},
		{0x407d4a9b, 0xb36d14a3},
		{0x407d4a9c, 0xb4225232},
		{0x407d4aa4, 0xb576ae9b},
		{0x407d4aa6, 0xb5951a8f},
		{0x407d4abc, 0xb6583f28},
		{0x407d4ac3, 0xb682aa6a},
		{0x407d4ac7, 0xb68f8c0a},
		{0x407d4ad9, 0xb6c98353},
		{0x407d4b24, 0xb75d84c0},
		{0x407d4b45, 0xb78953ac},
		{0x407d4c1d, 0xb81b9c77},
		{0x407d61ca, 0xba154e4f},
		{0x407d8dce, 0xbad84c8b},
		{0x407da3d6, 0xbb0f9496},
		{0x407dc7c6, 0xbb495963},
		{0x407f3e7a, 0xbc486ae9},
		{0x40807cad, 0xbcbc4733},
		{0x40822da4, 0xbd33292b},
		{0x408277e1, 0xbd41959f},
		{0x408bd2ba, 0xbe1cf476},
		{0x40b1fb8d, 0xbeacbb9b},
		{0x40c3d60f, 0xbe87e4a1},
		{0x40db0783, 0xbd957a93},
		{0x40e23105, 0xbbacf467},
		{0x40e2a773, 0xba74c04c},
		{0x40e2ba78, 0xb97837ae},
		{0x40e2bb4b, 0xb9588dce},
		{0x40e2bbe7, 0xb94124e8},
		{0x40e2bf14, 0xb88e4b65},
		{0x40e2c0d5, 0xb6718cb4},
		{0x40e2c0ea, 0xb51f890c},
		{0x40e2c0ed, 0xb43130fa},
		{0x40e2c0ee, 0xb2bc54a2},
		{0x40e2c0ef, 0x34021bd0},
		{0x40e2c0f0, 0x348de11a},
		{0x40e2c0f4, 0x356096f0},
		{0x40e2c104, 0x3651cc18},
		{0x40e2c107, 0x366e9b47},
		{0x40e2c115, 0x36ba8666},
		{0x40e2c11d, 0x36e0eff8},
		{0x40e2c15b, 0x3782a87d},
		{0x40e2c179, 0x37a6ab69},
		{0x40e2c213, 0x382fc375},
		{0x40e2f13f, 0x3ae7e4bc},
		{0x40e4bbca, 0x3c9759ab},
		{0x40e675e8, 0x3d0cecf8},
		{0x410aadd6, 0x3e8aa24c},
		{0x411b599e, 0x3e00be36},
		{0x411dcd2c, 0x3db74681},
		{0x41226cd2, 0x3c91295b},
		{0x41231215, 0x3bf9524c},
		{0x41231f9d, 0x3bde39eb},
		{0x412372b9, 0x3adfd398},
		{0x412382ea, 0x3a3cc7c2},
		{0x41238ad9, 0x3977e24a},
		{0x41238df5, 0x3844716e},
		{0x41238ea4, 0x36ad341e},
		{0x41238eb3, 0x35d56325},
		{0x41238eb4, 0x35b56cee},
		{0x41238eb8, 0x34d65052},
		{0x41238eb9, 0x342ceef3},
		{0x41238eba, 0xb3a58577},
		{0x41238ebb, 0xb4a93a34},
		{0x41238ec4, 0xb624fb34},
		{0x41238ee4, 0xb729178f},
		{0x41238f15, 0xb7b66da5},
		{0x41238f59, 0xb81f21d0},
		{0x4123faf4, 0xbbd7e496},
		{0x41240cc4, 0xbbfb5c8d},
		{0x4124e15c, 0xbca83e20},
		{0x4126281b, 0xbd241dab},
		{0x4131f304, 0xbe3fe21c},
		{0x4152c650, 0xbd280f8f},
		{0x41556403, 0xbbad27dc},
		{0x41556cc2, 0xbb9ddae5},
		{0x4155af28, 0xbaa7076c},
		{0x4155c66a, 0xb80f5ece},
		{0x4155c6fc, 0xb67d926d},
		{0x4155c704, 0xb60dc5bd},
		{0x4155c70d, 0xb47ff7c8},
		{0x4155c70e, 0xb30179d7},
		{0x4155c70f, 0x343f3adb},
		{0x4155c710, 0x34cf6a16},
		{0x4155c711, 0x351f9b5e},
		{0x4155c713, 0x3587b403},
		{0x4155c716, 0x35db8d7e},
		{0x4155ca3f, 0x39325ca6},
		{0x4157ef46, 0x3cef3a5f},
		{0x4158a865, 0x3d1f111c},
		{0x415a8d36, 0x3d820430},
		{0x415b5a7d, 0x3d96c0bb},
		{0x4183b93e, 0x3be48f28},
		{0x4183f058, 0x3adcaa92},
		{0x4184018a, 0x380cddba},
		{0x418401c3, 0x374d1c24},
		{0x418401e0, 0x35b65219},
		{0x418401e3, 0x347b7c18},
		{0x418401e4, 0xb416e01e},
		{0x41840273, 0xb86157e7},
		{0x4184027c, 0xb86f7d16},
		{0x41840515, 0xb9a093f0},
		{0x41841046, 0xbab4d30d},
		{0x4184641f, 0xbc1a1c28},
		{0x41947411, 0xbe27b15b},
		{0x419d20f9, 0xb81e6da4},
		{0x419d2162, 0xb5e59486},
		{0x419d2167, 0x31e69825},
		{0x419d276b, 0x3a0aa3cc},
		{0x419d2a76, 0x3a50c2c6},
		{0x419d679b, 0x3bca0a53},
		{0x419e9445, 0x3d0438c3},
		{0x41b54049, 0x3cac182f},
		{0x41b63bbf, 0x39fa46fd},
		{0x41b64186, 0x36bfbff9},
		{0x41b64193, 0x35d29be7},
		{0x41b64196, 0x35246c2c},
		{0x41b64198, 0xb2d8c7e2},
		{0x41b641f7, 0xb7fe54c5},
		{0x41b642a8, 0xb8b5f295},
		{0x41b6578e, 0xbaeaebdd},
		{0x41b65c30, 0xbb0e3b2c},
		{0x41b662b5, 0xbb31157f},
		{0x41b7c81e, 0xbd013fbe},
		{0x41bb734f, 0xbdcc1e85},
		{0x41cf614c, 0xb893b673},
		{0x41cf61d5, 0xb7f74e34},
		{0x41cf622d, 0xb6551654},
		{0x41cf6234, 0xb59152bf},
		{0x41cf6237, 0xb447aaf2},
		{0x41cf6238, 0x33f29c1c},
		{0x41cf6276, 0x379c6b20},
		{0x41cf6303, 0x387efe43},
		{0x41e5af76, 0x3d531428},
		{0x41e76b33, 0x3ca59ba0},
		{0x41e86656, 0x3b086d17},
		{0x41e8802d, 0x39602e58},
		{0x41e88315, 0x367bd821},
		{0x41e88321, 0x34c3fe35},
		{0x41e88322, 0x33b1b7f7},
		{0x41e88323, 0xb4564473},
		{0x41e88335, 0xb6b1345a},
		{0x41e8833e, 0xb7033ab9},
		{0x41e88398, 0xb80b5ff3},
		{0x41e884d3, 0xb90016c8},
		{0x41e8860c, 0xb95cbd82},
		{0x41e8d1a0, 0xbbb9b679},
		{0x4200d1f8, 0xb7b899d0},
		{0x4200d221, 0xb2809c36},
		{0x4200d227, 0x3656f733},
		{0x4200d24f, 0x37ced875},
		{0x4200d435, 0x39959785},
		{0x420d4616, 0x3b766c5f},
		{0x420d5d1f, 0x3a42008c},
		{0x420d5dbf, 0x3a2c8685},
		{0x420d6297, 0x37c2f3f2},
		{0x420d62c4, 0x345818f6},
		{0x420d62c5, 0xb4a6d0db},
		{0x420d62cf, 0xb6b6375c},
		{0x420d62d0, 0xb6c76531},
		{0x420d62e6, 0xb7905552},
		{0x4212a581, 0xbe028d30},
		{0x4219bbc4, 0xbbe548a1},
		{0x4219ea1f, 0xba99cee2},
		{0x4219f313, 0xb84cb7b6},
		{0x4219f370, 0xb6555c9c},
		{0x4219f373, 0xb5e52b15},
		{0x4219f376, 0xb47ce7af},
		{0x4219f3a6, 0x37c3942f},
		{0x421a004b, 0x3ad32e98},
		{0x421a31e6, 0x3c004d8d},
		{0x421a7138, 0x3c80dd2a},
		{0x42267a12, 0x3aa06950},
		{0x42268433, 0x34f3715a},
		{0x42268434, 0xb29d4460},
		{0x42268435, 0xb5038cf2},
		{0x4226844a, 0xb72e6e74},
		{0x42268455, 0xb782bf2a},
		{0x4226e1a2, 0xbc3868f7},
		{0x42275b61, 0xbcd2c9c5},
		{0x423313c0, 0xb915fd7b},
		{0x42331454, 0xb89ec926},
		{0x423314fa, 0xb4577c32},
		{0x42331631, 0x39142263},
		{0x42331dbb, 0x3a859427},
		{0x42331e0a, 0x3a8a49ba},
		{0x423324db, 0x3af24c1a},
		{0x42395bc2, 0x3df009d2},
		{0x423fa4c9, 0x38eb4d01},
		{0x423fa5c8, 0x338f784f},
		{0x423fa5c9, 0xb4c83689},
		{0x423fa5ce, 0xb62c93b1},
		{0x423fa674, 0xb89e79b4},
		{0x423fa697, 0xb8bec065},
		{0x423fbc1a, 0xbb249d16},
		{0x4244258d, 0xbdd285d7},
		{0x424c09d3, 0xbba00d4f},
		{0x424c35dd, 0xb8aa6b59},
		{0x424c3692, 0xb68b8219},
		{0x424c3695, 0xb64140e0},
		{0x424c369c, 0x33db8676},
		{0x424c3b2d, 0x3a02903f},
		{0x424c9591, 0x3c29446e},
		{0x424dda24, 0x3d3578b8},
		{0x42581acb, 0x3c953c40},
		{0x42587c6b, 0x3c0218b8},
		{0x4258c75c, 0x3728268e},
		{0x4258c773, 0x3509b740},
		{0x4258c774, 0x33d5d56e},
		{0x4258c776, 0xb5433e75},
		{0x4258c7ae, 0xb7c853d7},
		{0x4264a26a, 0xbc98c5f2},
		{0x4264b797, 0xbc871eb8},
		{0x4265584e, 0xb596be00},
		{0x42655850, 0xb4ab5b47},
		{0x42655851, 0x33b1cc4f},
		{0x42655852, 0x350220b7},
		{0x4265587a, 0x378af1f2},
		{0x42655930, 0x38bc28d7},
		{0x42655936, 0x38c137a7},
		{0x4265594f, 0x38d64ab7},
		{0x426559ab, 0x3911ec3c},
		{0x4265718a, 0x3b2a0d40},
		{0x4271ddb4, 0x3a96dfbc},
		{0x4271e446, 0x3a012820},
		{0x4271e926, 0x368da4dd},
		{0x4271e930, 0x34a50b22},
		{0x4271e931, 0xb3b45476},
		{0x4271e933, 0xb568aacd},
		{0x4271e936, 0xb608f6c9},
		{0x4271effa, 0xba323a67},
		{0x4272b87e, 0xbca8b647},
		{0x42779945, 0xbdcd6585},
		{0x427c9ad9, 0xbd3984c6},
		{0x427ce14c, 0xbd1fc2e8},
		{0x427e5b9a, 0xbb4320c7},
		{0x427e75aa, 0xb9e20710},
		{0x427e788c, 0xb91cbd41},
		{0x427e7a12, 0xb52ec386},
		{0x427e7a14, 0x33f0dda8},
		{0x427e7a3b, 0x377b91b7},
		{0x427e7a48, 0x37a7662b},
		{0x427e7fce, 0x3a12ac78},
		{0x427e969f, 0x3b36aed1},
	}
	for _, tt := range tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Y0(); !eq32(got, want) {
			t.Errorf("Y0(%v) = %v; want %v", x, got, want)
		}
	}
}

// TestFloat32_Y0Boundaries checks Y0 on the both sides of the boundaries of the segments of the calculation.
func TestFloat32_Y0Boundaries(t *testing.T) {
	t.Parallel()
	for _, x := range []float32{0.5, 1, 2, 2.25, 3.75, 4, 4.5, 5.5, 6, 7, 32, 63, 64, 65} {
		for d := -3; d <= 3; d++ {
			a := NewFloat32FromBits(math.Float32bits(x) + uint32(d))
			if got, want := a.Y0(), NewFloat256(float64(a)).Y0().Float32(); !eq32(got, want) {
				t.Errorf("Y0(%v) = %v; want %v", a, got, want)
			}
		}
	}
}

// TestFloat32_Y0Random compares Y0 with Float256 on random inputs in [2, 64), and the zeros of Y0.
func TestFloat32_Y0Random(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	for range 20000 {
		x := NewFloat32(2 + 62*r.Float64())
		if got, want := x.Y0(), NewFloat256(float64(x)).Y0().Float32(); !eq32(got, want) {
			t.Fatalf("Y0(%v) = %v; want %v", x, got, want)
		}
	}
	// near the zeros, where the relative error is large
	for _, z := range []float64{3.957678419314858, 7.086051060301773, 10.222345043496418, 13.361097473872764, 16.50092244152809, 19.64130970088794, 22.782028047291558, 25.922957653180923, 29.064030252728397, 32.20520411649328, 35.34645230521432, 38.48775665308154, 41.62910446621381, 44.77048660722199, 47.91189633151648, 51.05332855236236, 54.19477936108706, 57.33624570476628, 60.47772516422348, 63.61921579772038} {
		for d := -20; d <= 20; d++ {
			x := NewFloat32FromBits(math.Float32bits(float32(z)) + uint32(d))
			if got, want := x.Y0(), NewFloat256(float64(x)).Y0().Float32(); !eq32(got, want) {
				t.Errorf("Y0(%v) = %v; want %v", x, got, want)
			}
		}
	}
}

// TestFloat32_Y0Poly checks the absolute error of the polynomials, which are hidden by the rounding to Float32,
// with Float256.
func TestFloat32_Y0Poly(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(3, 4))
	for range 5000 {
		x := 2 + 62*r.Float64()
		var i int
		switch {
		case x < 4:
			i = int(x*4) - 8
		case x < 6:
			i = 8 + int((x-4)*2)
		default:
			i = 12 + int(x) - 6
		}
		c := &y032Coeffs[i]
		tt := x - y032Centers[i]
		got := 0.0
		for k := len(c) - 1; k >= 0; k-- {
			got = got*tt + c[k]
		}
		want := NewFloat256(x).Y0().Float64().BuiltIn()
		if math.Abs(got-want) > 0x1p-50 {
			t.Errorf("polynomial(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat32_Y0(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float32
	}{
		{"small", exact32(0.5)},                  // math.Y0
		{"medium", exact32(5)},                   // the polynomials
		{"medium2", exact32(50.5)},               // the polynomials
		{"hard", NewFloat32FromBits(0x4026f6b5)}, // the result is too close to the midpoint: Float256.Y0
		{"large", exact32(1000.5)},               // math.Y0
		{"huge", NewFloat32(1e20)},               // math.Y0
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Y0())
			}
		})
	}
}
