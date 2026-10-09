package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_Y1(t *testing.T) {
	t.Parallel()
	tests := []float64{0.5, 1, 2, 5, 10, 50}

	for _, x := range tests {
		want := math.Y1(x)
		got := exact32(x).Y1()
		if !close32(got, want) {
			t.Errorf("Y1(%v) = %v; want %v", x, got, want)
		}
	}
}

func TestFloat32_Y1Special(t *testing.T) {
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
		if got := tt.x.Y1(); !eq32(got, tt.want) {
			t.Errorf("Y1(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_Y1HardCases checks Y1 on the inputs whose results are very close to the midpoint of two adjacent
// Float32 values, which are calculated by the more accurate Float256. They are found by checking all the Float32 values
// in [2, 64), and the correctly rounded results were calculated with mpmath.
func TestFloat32_Y1HardCases(t *testing.T) {
	t.Parallel()
	// x, and the correctly rounded Y1(x)
	tests := [][2]uint32{
		{0x400bfd95, 0xbba76b8a},
		{0x400c82f0, 0xba6147df},
		{0x400c9ddc, 0xb65e1548},
		{0x400c9de5, 0xb61316f6},
		{0x400c9de9, 0xb5e384c4},
		{0x400c9dee, 0xb5903155},
		{0x400c9df1, 0xb53c64f2},
		{0x400c9df5, 0xb45c4ab3},
		{0x400c9df6, 0xb3adf0e0},
		{0x400c9df7, 0x33396749},
		{0x400c9df8, 0x3433ac13},
		{0x400c9dfb, 0x3510e8b4},
		{0x400c9e00, 0x359bc7bf},
		{0x400c9e08, 0x36108cfa},
		{0x400c9e7c, 0x378ae3eb},
		{0x400c9f5d, 0x383a9e75},
		{0x400c9fac, 0x3863c263},
		{0x400ca88c, 0x39b05917},
		{0x400cd075, 0x3ad23921},
		{0x400d4512, 0x3bada618},
		{0x4010a052, 0x3d03b09a},
		{0x4010ef0e, 0x3d0da106},
		{0x401de7cb, 0x3e06859e},
		{0x4023b629, 0x3e2f02ef},
		{0x402404ad, 0x3e311439},
		{0x402eca9f, 0x3e74ea52},
		{0x40891b05, 0x3eb406d1},
		{0x40a9309f, 0x3d48a08e},
		{0x40ad0b23, 0x3bf69d49},
		{0x40ad2178, 0x3bd81ab7},
		{0x40ad96ed, 0x3adf7780},
		{0x40ada16d, 0x3aa63f76},
		{0x40ada21c, 0x3aa285ee},
		{0x40adb170, 0x3a1e09a4},
		{0x40adbf50, 0x37dd2551},
		{0x40adbf96, 0x377bb618},
		{0x40adbfaf, 0x3737a5b0},
		{0x40adbfef, 0x35167624},
		{0x40adbff0, 0x34d5cd30},
		{0x40adbff1, 0x347d5c34},
		{0x40adbff2, 0x339e3c11},
		{0x40adbff3, 0xb3be4045},
		{0x40adc011, 0xb6a6533b},
		{0x40adc014, 0xb6b6a90b},
		{0x40adc044, 0xb75e02fa},
		{0x40b256d1, 0xbd44aaa3},
		{0x410491f6, 0xbdacf956},
		{0x41063529, 0xbd687d0c},
		{0x41072fe9, 0xbd241075},
		{0x41097de8, 0xba44f030},
		{0x41098398, 0xb9c434d9},
		{0x410985dd, 0xb96aacb0},
		{0x4109893d, 0xb3f9d6c7},
		{0x4109893e, 0x34190e30},
		{0x41098940, 0x35314054},
		{0x41098947, 0x3625ed41},
		{0x41098948, 0x36374cd8},
		{0x41098958, 0x36e6a323},
		{0x410989f2, 0x38440c0a},
		{0x41098ad2, 0x38dba232},
		{0x41098b29, 0x39056ec1},
		{0x41098b90, 0x3921642c},
		{0x410ae5ef, 0x3cbc282d},
		{0x413a3abb, 0x3cd1b0d0},
		{0x413ae01b, 0x3c8486a9},
		{0x413b9207, 0x3bc64662},
		{0x413bd1c2, 0x3b1f2f49},
		{0x413bf758, 0x399a91bf},
		{0x413bf853, 0x397ac8fb},
		{0x413bf889, 0x396e3b28},
		{0x413bfc85, 0x358e1b86},
		{0x413bfc89, 0x3438b252},
		{0x413bfc8a, 0xb355608f},
		{0x413bfc8f, 0xb59b718b},
		{0x413bfcd3, 0xb7882c8d},
		{0x413c2ea9, 0xbb3a524a},
		{0x413d257d, 0xbc8983e5},
		{0x41442248, 0xbde3442d},
		{0x4149c85d, 0xbe2e933d},
		{0x414b731b, 0xbe3c7183},
		{0x4166180b, 0xbdd4ad8f},
		{0x416d4eb4, 0xbc5ec23a},
		{0x416de4f8, 0xbbc4b8b3},
		{0x416e58d0, 0xb9247912},
		{0x416e5b14, 0xb832afda},
		{0x416e5bea, 0xb4f0d098},
		{0x416e5bec, 0xb36a7c04},
		{0x416e5bed, 0x3418e215},
		{0x416e5bf2, 0x35974cef},
		{0x416e5bf4, 0x35cc2d32},
		{0x416e5bfb, 0x36429f0e},
		{0x416e5bfd, 0x365d0f2f},
		{0x416e5c02, 0x368f93c1},
		{0x416e5c08, 0x36b73bf1},
		{0x416e5c0e, 0x36dee420},
		{0x416e5c46, 0x3794416a},
		{0x416e5c5b, 0x37b6f489},
		{0x416e6081, 0x3972369f},
		{0x416e618e, 0x3994e28b},
		{0x416e705b, 0x3a870693},
		{0x416e872a, 0x3b0ed913},
		{0x416e9629, 0x3b405c83},
		{0x4170a841, 0x3cf10b7a},
		{0x417faa30, 0x3e34638c},
		{0x418bb25e, 0x3dd6a240},
		{0x419057a2, 0x38f144a6},
		{0x419058b1, 0x37971037},
		{0x419058cf, 0x36f3cfe7},
		{0x419058e3, 0x33e13130},
		{0x419058fb, 0xb70e6a9d},
		{0x41906350, 0xba7a75e9},
		{0x4191d761, 0xbd0e0b6e},
		{0x4194f92f, 0xbdcee114},
		{0x419be346, 0xbe379f68},
		{0x41a362c1, 0xbdfa4f5b},
		{0x41a71484, 0xbd5555ab},
		{0x41a98105, 0xb74ebabc},
		{0x41a98125, 0xb5ea6ff4},
		{0x41a9812a, 0xb3ca81f6},
		{0x41a9812b, 0x347d9852},
		{0x41a9812d, 0x3570d2ba},
		{0x41a98192, 0x380fc2da},
		{0x41a9838e, 0x3953f90d},
		{0x41a9fc45, 0x3c2a4b63},
		{0x41c18b0e, 0x3cb81f26},
		{0x41c1e517, 0x3c7c0471},
		{0x41c26441, 0x3baee43b},
		{0x41c2a2be, 0x39d22492},
		{0x41c2a6f2, 0x38909124},
		{0x41c2a7ca, 0x361b601c},
		{0x41c2a7d1, 0x3427d9fa},
		{0x41c2a7d2, 0xb423503f},
		{0x41c2a7ff, 0xb76b66d5},
		{0x41c2af2a, 0xba180722},
		{0x41c2b1fe, 0xba528d94},
		{0x41c2b279, 0xba5c7ed8},
		{0x41c5b2f9, 0xbd740d9d},
		{0x41d3cb7b, 0xbe05b0c8},
		{0x41d86c1b, 0xbd80cc18},
		{0x41daafcf, 0xbcadba6d},
		{0x41db285a, 0xbc490563},
		{0x41dbc547, 0xba1e506a},
		{0x41dbccaa, 0xb866a735},
		{0x41dbcd2c, 0xb790c43f},
		{0x41dbcd30, 0xb78706e6},
		{0x41dbcd64, 0xb5869709},
		{0x41dbcd66, 0xb4e2b154},
		{0x41dbcd67, 0xb40db7d7},
		{0x41dbcd68, 0x3429f2f7},
		{0x41dbcd74, 0x36745f44},
		{0x41dbcf33, 0x390bddba},
		{0x41f475b5, 0x3c0fb1ae},
		{0x41f4c4e7, 0x3b5143d0},
		{0x41f4ebaa, 0x39f31a76},
		{0x41f4ef79, 0x394d1012},
		{0x41f4f239, 0x3604a4f6},
		{0x41f4f240, 0x335e2eb7},
		{0x41f4f241, 0xb46fb4d4},
		{0x41f4f252, 0xb6a457e6},
		{0x41f4f26d, 0xb74ebb1a},
		{0x41f4f383, 0xb8ba26f1},
		{0x41f4f437, 0xb910f977},
		{0x41f86d46, 0xbd772370},
		{0x4206f75f, 0xbb2effa3},
		{0x42070a87, 0xb8d44330},
		{0x42070afd, 0xb8254dea},
		{0x42070aff, 0xb820e92b},
		{0x42070b48, 0xb40c1c3d},
		{0x42070b4a, 0x35762815},
		{0x42070cdf, 0x395f609d},
		{0x4213809a, 0x3b70c3d6},
		{0x421386d5, 0x3b3c6157},
		{0x42138cb2, 0x3b0b1590},
		{0x42139c33, 0x390c0c8d},
		{0x42139d3e, 0xb4549542},
		{0x4213a674, 0xba9ad634},
		{0x4213a8c3, 0xbac1a16a},
		{0x4214b122, 0xbd0ea8ee},
		{0x421e8b42, 0xbd4ee00b},
		{0x42202d14, 0xb97e37eb},
		{0x42202e17, 0xb8f73639},
		{0x42202f0c, 0xb366e158},
		{0x42202f0f, 0x35ba6d7f},
		{0x42203362, 0x3a0bea4e},
		{0x4221671d, 0x3d1a5bba},
		{0x422c0a66, 0x3cb08391},
		{0x422c89fe, 0x3bd4b1b8},
		{0x422cbfc8, 0x38ed35e4},
		{0x422cc05b, 0x383ce321},
		{0x422cc0bc, 0x33f3934c},
		{0x422cd49f, 0xbb1a754f},
		{0x422cd6ee, 0xbb2c61f0},
		{0x4238c6f1, 0xbc827daa},
		{0x42394d22, 0xba1beb86},
		{0x42394d66, 0xba13f2eb},
		{0x42394feb, 0xb990af45},
		{0x4239511b, 0xb912d5d2},
		{0x42395132, 0xb9080d32},
		{0x42395254, 0xb3adf6ce},
		{0x42395255, 0x34c48eb9},
		{0x42395256, 0x355a4d93},
		{0x4239525b, 0x364c9b25},
		{0x42395299, 0x38010f9f},
		{0x42395543, 0x39b000ca},
		{0x423957ab, 0x3a20325e},
		{0x42399efb, 0x3c0f8038},
		{0x4239d0f4, 0x3c6c8c70},
		{0x4245dd90, 0x3a3676dd},
		{0x4245e3d8, 0x3475c1cf},
		{0x4245e3d9, 0xb45ad953},
		{0x4245e3da, 0xb52add1d},
		{0x4245e60a, 0xb97ebe3d},
		{0x4245e9c5, 0xba2c0046},
		{0x4245ef4d, 0xbaa64aaa},
		{0x4246388d, 0xbc196cc2},
		{0x42467d45, 0xbc8a7dfd},
		{0x424e8ef9, 0xbdbc3383},
		{0x42505051, 0xbd675011},
		{0x42520662, 0xbc43049a},
		{0x42527009, 0xba144093},
		{0x42527341, 0xb96674bf},
		{0x42527502, 0xb803a19a},
		{0x42527541, 0xb6a61429},
		{0x4252754b, 0xb54a51ef},
		{0x4252754d, 0x33b788d8},
		{0x4252754e, 0x3507929f},
		{0x425275c4, 0x3851c7dc},
		{0x4252e233, 0x3c3f17e0},
		{0x425e8348, 0x3c60508a},
		{0x425eaa7b, 0x3c1d92b4},
		{0x425ef0f0, 0x3b14d929},
		{0x425f06b2, 0x353a170a},
		{0x425f06b4, 0xb402f01d},
		{0x425f0715, 0xb82656be},
		{0x4260f637, 0xbd4abbc0},
		{0x4262e73a, 0xbdb2df7d},
		{0x426b7d64, 0xbb31764b},
		{0x426b980f, 0xb421f750},
		{0x426b99af, 0x392cd403},
		{0x4274467a, 0x3dacabb3},
		{0x427497bb, 0x3da2ac47},
		{0x42777ff7, 0x3c88d7e4},
		{0x4277fcaf, 0x3b90e309},
		{0x42781bfc, 0x3aadb55a},
		{0x42782821, 0x3901e02d},
		{0x42782957, 0x3688a7e1},
		{0x42782961, 0x3460005f},
		{0x42782962, 0xb43ee590},
		{0x427839ef, 0xbad69388},
	}
	for _, tt := range tests {
		x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
		if got := x.Y1(); !eq32(got, want) {
			t.Errorf("Y1(%v) = %v; want %v", x, got, want)
		}
	}
}

// TestFloat32_Y1Boundaries checks Y1 on the both sides of the boundaries of the segments of the calculation.
func TestFloat32_Y1Boundaries(t *testing.T) {
	t.Parallel()
	for _, x := range []float32{0.5, 1, 2, 2.25, 2.75, 3, 3.5, 4.5, 5, 6, 32, 63, 64, 65} {
		for d := -3; d <= 3; d++ {
			a := NewFloat32FromBits(math.Float32bits(x) + uint32(d))
			if got, want := a.Y1(), NewFloat256(float64(a)).Y1().Float32(); !eq32(got, want) {
				t.Errorf("Y1(%v) = %v; want %v", a, got, want)
			}
		}
	}
}

// TestFloat32_Y1Random compares Y1 with Float256 on random inputs in [2, 64), and the zeros of Y1.
func TestFloat32_Y1Random(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	for range 20000 {
		x := NewFloat32(2 + 62*r.Float64())
		if got, want := x.Y1(), NewFloat256(float64(x)).Y1().Float32(); !eq32(got, want) {
			t.Fatalf("Y1(%v) = %v; want %v", x, got, want)
		}
	}
	// near the zeros, where the relative error is large
	for _, z := range []float64{2.197141326031017, 5.429681040794135, 8.596005868331169, 11.749154830839881, 14.897442128336726, 18.043402276727857, 21.188068934142212, 24.33194257135691, 27.475294980449224, 30.618286491641115, 33.76101779610933, 36.90355531614295, 40.045944640266875, 43.18821809739321, 46.33039925070169, 49.4725056799241, 52.61455076717296, 55.756544879208136, 58.89849617143305, 62.040411147670696} {
		for d := -20; d <= 20; d++ {
			x := NewFloat32FromBits(math.Float32bits(float32(z)) + uint32(d))
			if got, want := x.Y1(), NewFloat256(float64(x)).Y1().Float32(); !eq32(got, want) {
				t.Errorf("Y1(%v) = %v; want %v", x, got, want)
			}
		}
	}
}

// TestFloat32_Y1Poly checks the absolute error of the polynomials, which are hidden by the rounding to Float32,
// with Float256.
func TestFloat32_Y1Poly(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(3, 4))
	for range 5000 {
		x := 2 + 62*r.Float64()
		var i int
		switch {
		case x < 3:
			i = int(x*4) - 8
		case x < 5:
			i = 4 + int((x-3)*2)
		default:
			i = 8 + int(x) - 5
		}
		c := &y132Coeffs[i]
		tt := x - y132Centers[i]
		got := 0.0
		for k := len(c) - 1; k >= 0; k-- {
			got = got*tt + c[k]
		}
		want := NewFloat256(x).Y1().Float64().BuiltIn()
		if math.Abs(got-want) > 0x1p-50 {
			t.Errorf("polynomial(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat32_Y1(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float32
	}{
		{"small", exact32(0.5)},                  // math.Y1
		{"medium", exact32(5)},                   // the polynomials
		{"medium2", exact32(50.5)},               // the polynomials
		{"hard", NewFloat32FromBits(0x400bfd95)}, // the result is too close to the midpoint: Float256.Y1
		{"large", exact32(1000.5)},               // math.Y1
		{"huge", NewFloat32(1e20)},               // math.Y1
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.Y1())
			}
		})
	}
}
