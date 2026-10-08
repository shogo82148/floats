package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_J1(t *testing.T) {
	t.Parallel()
	tests := []float64{-50, -10, -1, -0.5, 0, 0.5, 1, 2, 5, 10, 50}

	for _, x := range tests {
		want := math.J1(x)
		got := exact32(x).J1()
		if !close32(got, want) {
			t.Errorf("J1(%v) = %v; want %v", x, got, want)
		}
	}
}

func TestFloat32_J1Special(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		x    Float32
		want Float32
	}{
		{exact32(0), exact32(0)},
		{exact32(math.Copysign(0, -1)), exact32(math.Copysign(0, -1))},
		{exact32(math.Inf(1)), exact32(0)},
		{exact32(math.Inf(-1)), exact32(0)},
		{exact32(math.NaN()), exact32(math.NaN())},
	} {
		if got := tt.x.J1(); !eq32(got, tt.want) {
			t.Errorf("J1(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_J1HardCases checks J1 on the inputs in [2, 64) whose results are very close to the midpoint of two
// adjacent Float32 values, or to zero. They are found by checking all the Float32 values in the range, and
// the correctly rounded results were calculated with mpmath.
func TestFloat32_J1HardCases(t *testing.T) {
	t.Parallel()
	// x, and the correctly rounded J1(x)
	tests := [][2]uint32{
		{0x4059bf1f, 0x3e368c9c},
		{0x405b90ee, 0x3e2a640e},
		{0x4069a08b, 0x3d9867e8},
		{0x4074f51b, 0x3ae0448e},
		{0x4075206f, 0x3a291c50},
		{0x407520a0, 0x3a27e06a},
		{0x40752f79, 0x399055e5},
		{0x4075368b, 0x38d4cce8},
		{0x40753a05, 0x3786433f},
		{0x40753a2a, 0x3750ea9b},
		{0x40753a67, 0x36dd48fb},
		{0x40753a97, 0x36054005},
		{0x40753aa3, 0x355fae47},
		{0x40753aa7, 0x34f1260f},
		{0x40753aa9, 0x348a0ad0},
		{0x40753aaa, 0x342cfa61},
		{0x40753aab, 0x338bbe46},
		{0x40753aac, 0xb304f069},
		{0x40753aad, 0xb4085757},
		{0x40753aae, 0xb46f7294},
		{0x40753ab2, 0xb522f7e1},
		{0x40753ac2, 0xb60fd92f},
		{0x40753ac8, 0xb6368363},
		{0x40753b50, 0xb7845d32},
		{0x40753f03, 0xb8dfca3e},
		{0x40754ed6, 0xba01ed6f},
		{0x40756889, 0xba93b990},
		{0x40757f11, 0xbadc417e},
		{0x40758e63, 0xbb06c786},
		{0x4076490c, 0xbbd950f9},
		{0x40765c9b, 0xbbe8ffbd},
		{0x40770af2, 0xbc3a479a},
		{0x4077706d, 0xbc62ccf2},
		{0x4078877f, 0xbca8ed21},
		{0x407f7611, 0xbd80aeba},
		{0x4086abb0, 0xbe10f12c},
		{0x40a9b512, 0xbeb12628},
		{0x40b5e5f2, 0xbea6e1f9},
		{0x40c8e16e, 0xbe5af38f},
		{0x40cc9acf, 0xbe3bb128},
		{0x40de2671, 0xbcb53be7},
		{0x40df5a14, 0xbc30a5b9},
		{0x40e06552, 0xba7d4414},
		{0x40e070d6, 0xba0ea24b},
		{0x40e07ea4, 0xb820a205},
		{0x40e07edb, 0xb7ff3d3e},
		{0x40e07ee9, 0xb7ee6eb8},
		{0x40e07f5f, 0xb7418d95},
		{0x40e07f9f, 0xb61f92b6},
		{0x40e07fa2, 0xb602c30f},
		{0x40e07fad, 0xb4c8f81f},
		{0x40e07fae, 0xb4784776},
		{0x40e07faf, 0xb3bd3d5f},
		{0x40e07fb0, 0x336c2858},
		{0x40e07fb1, 0x3454b2db},
		{0x40e07fb2, 0x34b72dd0},
		{0x40e07fb4, 0x35286b49},
		{0x40e07fb6, 0x35753faa},
		{0x40e07fc4, 0x3643c390},
		{0x40e0808f, 0x380614cb},
		{0x40e080fd, 0x38481aff},
		{0x40e081c8, 0x38a0f974},
		{0x40e081cb, 0x38a1dfee},
		{0x40e0823b, 0x38c37c80},
		{0x40e082b0, 0x38e69923},
		{0x40e08b97, 0x39e49e46},
		{0x40e09e07, 0x3a91a7e8},
		{0x40e0d146, 0x3b43be4a},
		{0x40e10313, 0x3b9d8a41},
		{0x40f2a14c, 0x3e1ecc53},
		{0x40f5b6bc, 0x3e34fb90},
		{0x40fa219f, 0x3e5161aa},
		{0x412217cc, 0x3c2ed634},
		{0x412273c6, 0x3ba57641},
		{0x4122b693, 0x3a7ef792},
		{0x4122c1bc, 0x399924a8},
		{0x4122c4a9, 0x38ee7c2d},
		{0x4122c618, 0x37dcc95e},
		{0x4122c66a, 0x36e3ea9e},
		{0x4122c670, 0x36b3f917},
		{0x4122c686, 0x3405c768},
		{0x4122c687, 0xb3f3d676},
		{0x4122c68a, 0xb55e40c6},
		{0x4122c68f, 0xb6077801},
		{0x4122c697, 0xb683a8a5},
		{0x4122c6c3, 0xb7719ef9},
		{0x4122c722, 0xb81b4aca},
		{0x4122c9fb, 0xb95cd957},
		{0x41232f41, 0xbbd0ecb2},
		{0x41263a59, 0xbd58b64a},
		{0x412de16e, 0xbe1e578b},
		{0x4137855f, 0xbe6849e6},
		{0x4142aa6a, 0xbe567e92},
		{0x41531e45, 0xbce6e2c3},
		{0x41546f31, 0xbc26c0d2},
		{0x41552bba, 0xb8ec8797},
		{0x41552d87, 0xb78cca8f},
		{0x41552d8f, 0xb77da1df},
		{0x41552da6, 0xb72d4695},
		{0x41552dc3, 0xb68fe9d2},
		{0x41552dd6, 0xb4b26971},
		{0x41552dd7, 0xb4053946},
		{0x41552dd8, 0x33b4c0ac},
		{0x41552dd9, 0x349cfcf8},
		{0x41552dda, 0x350664e3},
		{0x41552dfb, 0x36f762f1},
		{0x41552e51, 0x37d413b9},
		{0x41552e56, 0x37dccfb5},
		{0x415534d8, 0x39c3ae78},
		{0x41554960, 0x3ac055f2},
		{0x4160171d, 0x3e0981dd},
		{0x4161b839, 0x3e1978b0},
		{0x418317da, 0x3c875ab7},
		{0x418357b3, 0x3c2a2d89},
		{0x4183a19f, 0x3b574b7d},
		{0x4183bcf8, 0x3a2d129c},
		{0x4183c323, 0x388f91af},
		{0x4183c3d7, 0x358732f9},
		{0x4183c3d8, 0x3529ced3},
		{0x4183c3da, 0xb3fafb2c},
		{0x4183c3e0, 0xb61eba80},
		{0x4183c3f7, 0xb73847c4},
		{0x4183c44f, 0xb838616a},
		{0x4189ebab, 0xbe08d8a7},
		{0x4190a774, 0xbe401c0b},
		{0x419ced36, 0xb6c685a4},
		{0x419ced46, 0xb4e23111},
		{0x419ced47, 0xb3a73a20},
		{0x419ced4b, 0x35adeee4},
		{0x419ced65, 0x372b8ddc},
		{0x419e2661, 0x3cdfc40d},
		{0x41a2de0f, 0x3df4cf49},
		{0x41a40eeb, 0x3e0c5473},
		{0x41b37376, 0x3d5eaec6},
		{0x41b5fbc2, 0x3b0537df},
		{0x41b60b39, 0x3a49d250},
		{0x41b61229, 0x39555f8d},
		{0x41b6146a, 0x37a37e2d},
		{0x41b61478, 0x377c1631},
		{0x41b6148f, 0x370109d9},
		{0x41b614a0, 0x36185c14},
		{0x41b614a7, 0x3323fb7f},
		{0x41b614a8, 0xb496b300},
		{0x41b614be, 0xb6f4d07c},
		{0x41b6155f, 0xb875ef06},
		{0x41b615a7, 0xb8ab1d78},
		{0x41b615ef, 0xb8db435b},
		{0x41b61b84, 0xba12d9c0},
		{0x41b6c6f3, 0xbc6db662},
		{0x41ce7f79, 0xbc6adc08},
		{0x41cf26c9, 0xbac7fe45},
		{0x41cf2bb0, 0xba96ce6d},
		{0x41cf31ad, 0xba357537},
		{0x41cf35fd, 0xb9bddeb1},
		{0x41cf3ab8, 0xb40a7dbd},
		{0x41cf3ab9, 0x34367b4a},
		{0x41cf3ac7, 0x369220c8},
		{0x41cf3b1b, 0x37f72b66},
		{0x41cf5f53, 0x3b378092},
		{0x41e69a2e, 0x3d05ba88},
		{0x41e7a7c0, 0x3c5a199e},
		{0x41e8198a, 0x3ba6b45d},
		{0x41e84ca5, 0x3ab67964},
		{0x41e85e9d, 0x38c3c14a},
		{0x41e85fc7, 0x371a96f5},
		{0x41e85fd7, 0x369d9d9f},
		{0x41e85fdb, 0x366f731e},
		{0x41e85fe7, 0x3441ac0d},
		{0x41e85fe8, 0xb3dae8cb},
		{0x41e86079, 0xb82c1e9f},
		{0x41e8611c, 0xb8b68fc5},
		{0x41e8632c, 0xb9779317},
		{0x41e87cd6, 0xbb08fd78},
		{0x41e893e4, 0xbb7616d9},
		{0x41ed612d, 0xbdafa3dc},
		{0x41f8ca65, 0xbe01f490},
		{0x41fff7cb, 0xbcde6042},
		{0x420074ea, 0xbc2dfb2e},
		{0x4200c23b, 0xb481c002},
		{0x4200c241, 0x3647c06f},
		{0x4200c387, 0x393a77a6},
		{0x4200cd8f, 0x3acbcebc},
		{0x4200ea55, 0x3bb444cf},
		{0x42010a37, 0x3c21a110},
		{0x4207eaea, 0x3e08db95},
		{0x420851fc, 0x3e04eb2d},
		{0x420cebd1, 0x3c6043be},
		{0x420d5412, 0x37e9c427},
		{0x420d543d, 0x36c45d12},
		{0x420d5448, 0x346c9bc8},
		{0x420d5449, 0xb49c8f74},
		{0x421975d8, 0xbc66ffe8},
		{0x4219e4cc, 0xb9339d0f},
		{0x4219e5db, 0xb820c2c4},
		{0x4219e623, 0xb6498055},
		{0x4219e628, 0xb5137ce2},
		{0x4219e629, 0xb37c8cda},
		{0x4219e62c, 0x35bda9b5},
		{0x421b2305, 0x3d1fc8ce},
		{0x4224235b, 0x3d8c54ee},
		{0x4224a916, 0x3d5e7140},
		{0x422673f0, 0x39fb3121},
		{0x42267651, 0x39491c35},
		{0x42267751, 0x3894f12d},
		{0x42267755, 0x3890fc12},
		{0x42267785, 0x3842fdc3},
		{0x422677be, 0x37a46925},
		{0x422677e7, 0x348a1d84},
		{0x422677e8, 0xb4665039},
		{0x42267895, 0xb8ab9adc},
		{0x42267a4a, 0xb996fba7},
		{0x4226a465, 0xbbafeaab},
		{0x422cc943, 0xbdf8a44a},
		{0x4233084a, 0xb9191491},
		{0x4233097d, 0xb6d476dd},
		{0x4233098b, 0x331da15d},
		{0x42330990, 0x361b1a4b},
		{0x423309b1, 0x379150bf},
		{0x42330ab9, 0x39101722},
		{0x423f9791, 0x39d05236},
		{0x423f9b08, 0x36f4628e},
		{0x423f9b14, 0x3606a61c},
		{0x423f9b17, 0x3538797c},
		{0x423f9b19, 0xb44e6c86},
		{0x423f9c10, 0xb8e42e85},
		{0x424c26c0, 0xba26acd8},
		{0x424c2c8c, 0xb672a0ca},
		{0x424c2c94, 0xb45ed442},
		{0x424c2c95, 0x346a92bf},
		{0x424c2c98, 0x35c8d8f6},
		{0x424c2e09, 0x39266474},
		{0x424c4bea, 0x3b5fd6ee},
		{0x42575da2, 0x3d16400b},
		{0x4258baf2, 0x39a9e752},
		{0x4258bdf7, 0x36957783},
		{0x4258bdff, 0x3599ebba},
		{0x4258be02, 0xb3c8f21f},
		{0x4258be09, 0xb648818c},
		{0x4258c126, 0xb9ae53d7},
		{0x4258c4d1, 0xba3ced5d},
		{0x4258caa3, 0xbaaf315c},
		{0x4258d1b7, 0xbb08aca9},
		{0x4259071b, 0xbbfd25e5},
		{0x4264aa52, 0xbc8abe1f},
		{0x42654f38, 0xb7906a90},
		{0x42654f41, 0xb764230c},
		{0x42654f63, 0x3394229e},
		{0x4265514f, 0x394f71e9},
		{0x426551b0, 0x3978541b},
		{0x42655497, 0x3a0c5e97},
		{0x426f1c07, 0x3d86de06},
		{0x4271d0b9, 0x3ad22c54},
		{0x4271d498, 0x3a9f53ee},
		{0x4271e0b9, 0x346a40c2},
		{0x4271e0ba, 0xb439ffbd},
		{0x4271e0bc, 0xb5805016},
		{0x4271e0ff, 0xb7e3fea3},
		{0x4271e104, 0xb7f46923},
		{0x4271ee92, 0xbab5ce18},
		{0x4271f833, 0xbb1a19b0},
		{0x42720d69, 0xbb929cbb},
		{0x427bdf80, 0xbd76f7d3},
		{0x427cea73, 0xbd195bc6},
		{0x427dbcd2, 0xbc9076b1},
		{0x427e54ac, 0xbb3bf84d},
		{0x427e5c10, 0xbb0ca64d},
		{0x427e70a4, 0xb90e3f62},
		{0x427e71ec, 0xb7300016},
		{0x427e7207, 0xb448ecbf},
		{0x427e7208, 0x3450d1b6},
		{0x427e7263, 0x3812776b},
		{0x427e72eb, 0x38b61207},
		{0x427e8044, 0x3ab64384},
	}
	for _, tt := range tests {
		for _, neg := range []bool{false, true} {
			x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
			if neg {
				x, want = -x, -want
			}
			if got := x.J1(); !eq32(got, want) {
				t.Errorf("J1(%v) = %v; want %v", x, got, want)
			}
		}
	}
}

// TestFloat32_J1Random compares J1 with math.J1 on random inputs in [2, 64), where math.J1 rounded to Float32
// is correctly rounded except for the rare cases.
func TestFloat32_J1Random(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	for range 300000 {
		x := NewFloat32(2 + 62*r.Float64())
		if r.IntN(2) == 0 {
			x = -x
		}
		if got, want := x.J1(), NewFloat32(math.J1(float64(x))); !eq32(got, want) {
			t.Fatalf("J1(%v) = %v; want %v", x, got, want)
		}
	}
}

// TestFloat32_J1Boundaries checks J1 on the both sides of the boundaries of the segments.
func TestFloat32_J1Boundaries(t *testing.T) {
	t.Parallel()
	for _, x := range []float32{2, 3, 4, 5, 10, 32, 63, 64} {
		for d := -3; d <= 3; d++ {
			a := NewFloat32FromBits(math.Float32bits(x) + uint32(d))
			if got, want := a.J1(), NewFloat32(math.J1(float64(a))); !eq32(got, want) {
				t.Errorf("J1(%v) = %v; want %v", a, got, want)
			}
		}
	}
}

// TestFloat32_J1Poly checks the absolute error of the polynomials, which are hidden by the rounding to Float32,
// with math.J1.
func TestFloat32_J1Poly(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(3, 4))
	for range 100000 {
		x := 2 + 62*r.Float64()
		i := int(x)
		c := &j132Coeffs[i-2]
		tt := x - float64(i) - 0.5
		got := 0.0
		for k := len(c) - 1; k >= 0; k-- {
			got = got*tt + c[k]
		}
		if want := math.J1(x); math.Abs(got-want) > 0x1p-50 {
			t.Errorf("polynomial(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat32_J1(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float32
	}{
		{"small", exact32(0.5)},      // math.J1
		{"medium", exact32(5)},       // the polynomials
		{"medium2", exact32(50.5)},   // the polynomials
		{"negative", exact32(-10.5)}, // the polynomials
		{"large", exact32(1000.5)},   // math.J1
		{"huge", NewFloat32(1e20)},   // math.J1
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.J1())
			}
		})
	}
}
