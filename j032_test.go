package floats

import (
	"math"
	"math/rand/v2"
	"runtime"
	"testing"
)

func TestFloat32_J0(t *testing.T) {
	t.Parallel()
	tests := []float64{-10, -1, -0.5, 0, 0.5, 1, 2, 5, 10, 50}

	for _, x := range tests {
		want := math.J0(x)
		got := exact32(x).J0()
		if !close32(got, want) {
			t.Errorf("J0(%v) = %v; want %v", x, got, want)
		}
	}
}

func TestFloat32_J0Special(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		x    Float32
		want Float32
	}{
		{exact32(0), exact32(1)},
		{exact32(math.Copysign(0, -1)), exact32(1)},
		{exact32(math.Inf(1)), exact32(0)},
		{exact32(math.Inf(-1)), exact32(0)},
		{exact32(math.NaN()), exact32(math.NaN())},
	} {
		if got := tt.x.J0(); !eq32(got, tt.want) {
			t.Errorf("J0(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

// TestFloat32_J0HardCases checks J0 on the inputs whose results are very close to the midpoint of two adjacent
// Float32 values, which are calculated by the more accurate Float256. They are found by checking all the Float32 values
// in [2, 64), and the correctly rounded results were calculated with mpmath.
func TestFloat32_J0HardCases(t *testing.T) {
	t.Parallel()
	// x, and the correctly rounded J0(x)
	tests := [][2]uint32{
		{0x400db34c, 0x3dd2085d},
		{0x4015ebbd, 0x3d06299c},
		{0x401889f9, 0x3c36dc4e},
		{0x4018cc80, 0x3c140c90},
		{0x4019025d, 0x3befcfc5},
		{0x4019d8f2, 0x3a0292f0},
		{0x4019e781, 0x3819e819},
		{0x4019e7f9, 0x37b73705},
		{0x4019e86c, 0x36ff3d06},
		{0x4019e895, 0x3629ea25},
		{0x4019e8a3, 0x3556806f},
		{0x4019e8a8, 0x34417f6c},
		{0x4019e8a9, 0x33726247},
		{0x4019e8aa, 0xb3909c90},
		{0x4019e8ab, 0xb44d3521},
		{0x4019e8ae, 0xb516fa69},
		{0x4019e8bb, 0xb611ba23},
		{0x4019e8c1, 0xb64390ae},
		{0x4019e8c8, 0xb67db5a4},
		{0x4019e8d9, 0xb6c57563},
		{0x4019e921, 0xb7783e15},
		{0x4019ef01, 0xb952b2a0},
		{0x4019f1d4, 0xb99841f1},
		{0x401a01d6, 0xba510996},
		{0x401a0ce6, 0xba966d6c},
		{0x401a42e8, 0xbb3b2f69},
		{0x401a7321, 0xbb8f840c},
		{0x401a9284, 0xbbaff9dd},
		{0x401b014e, 0xbc112be5},
		{0x402107b3, 0xbd66d198},
		{0x4029806a, 0xbdf44e74},
		{0x402f2224, 0xbe2224f4},
		{0x40532b0f, 0xbeb03939},
		{0x40a3eb32, 0xbe0c06b4},
		{0x40a83bae, 0xbdb97b19},
		{0x40b00b6f, 0xbbd0a56e},
		{0x40b02a84, 0xbba637fc},
		{0x40b08a41, 0xba8ed254},
		{0x40b0a229, 0xb8ca1236},
		{0x40b0a3b1, 0xb8095d92},
		{0x40b0a44b, 0xb7024056},
		{0x40b0a478, 0xb4f82ebf},
		{0x40b0a47a, 0xb413ef19},
		{0x40b0a47b, 0x32d240be},
		{0x40b0a47c, 0x34487f47},
		{0x40b0a826, 0x391fc605},
		{0x40b0bbad, 0x3a7c80bd},
		{0x40b10f51, 0x3b913c38},
		{0x40b54b31, 0x3d47410e},
		{0x40ba0e96, 0x3dc4df69},
		{0x40bc0bbe, 0x3deb952d},
		{0x40c05bf6, 0x3e1d6db2},
		{0x40caf630, 0x3e6e02a6},
		{0x40ebf250, 0x3e901709},
		{0x4105ba86, 0x3da4d4e2},
		{0x410a416e, 0x3b630d1d},
		{0x410a5a2e, 0x3aeee173},
		{0x410a5ed9, 0x3ac64def},
		{0x410a74dc, 0x3861330d},
		{0x410a74f7, 0x3843e1c7},
		{0x410a74f8, 0x3842cbcd},
		{0x410a759e, 0x3668cc3d},
		{0x410a75a7, 0x3598e1dc},
		{0x410a75ab, 0x33de60e5},
		{0x410a75ac, 0xb426c725},
		{0x410a75ba, 0xb67da509},
		{0x410a75be, 0xb6a19174},
		{0x410a773a, 0xb8d865a3},
		{0x410a78c6, 0xb957afc1},
		{0x410aaf0c, 0xbb78fe67},
		{0x410ac725, 0xbbb0b7ca},
		{0x410b3cbc, 0xbc5774f8},
		{0x413c71de, 0xbb515bc7},
		{0x413c7270, 0xbb4f3c2c},
		{0x413ca430, 0xb9b0aea0},
		{0x413ca6d6, 0xb943bdd5},
		{0x413caa01, 0xb6e6b809},
		{0x413caa1f, 0xb471d452},
		{0x413caa20, 0xb172976f},
		{0x413caa21, 0x346a3f95},
		{0x413caa23, 0x353194de},
		{0x413caa32, 0x3685c73f},
		{0x413caa4e, 0x372b07e8},
		{0x413caa54, 0x374158d2},
		{0x413cac5b, 0x3904ba51},
		{0x413cad22, 0x3932fc1b},
		{0x413cb37b, 0x3a0b2bba},
		{0x413cc82a, 0x3adf614d},
		{0x413ccedc, 0x3b089303},
		{0x413cd126, 0x3b1114c1},
		{0x413cf031, 0x3b82333b},
		{0x416cef15, 0x3ccfb090},
		{0x416ec6f0, 0x3ac70090},
		{0x416ee4c1, 0x37712358},
		{0x416ee4db, 0x371b36e3},
		{0x416ee4eb, 0x36ccad42},
		{0x416ee507, 0x351ce347},
		{0x416ee509, 0x344c8b4a},
		{0x416ee50a, 0xb1deb3ec},
		{0x416ee50b, 0xb45a7687},
		{0x416ee50e, 0xb5553e4f},
		{0x416ee514, 0xb6049fe8},
		{0x416ee51c, 0xb66e6057},
		{0x416ee524, 0xb6ac1061},
		{0x416ee675, 0xb895f6ed},
		{0x416ee6ab, 0xb8ac4561},
		{0x416ee74e, 0xb8ef9a5f},
		{0x416eea6b, 0xb98e344e},
		{0x416eec27, 0xb9bc0d8b},
		{0x416ef391, 0xba4003da},
		{0x416efcf3, 0xba9e00bf},
		{0x41703248, 0xbc89228b},
		{0x4174a823, 0xbd934c10},
		{0x418ff94e, 0xbc64e406},
		{0x4190918a, 0x31b1104e},
		{0x4190918f, 0x35f0fc06},
		{0x4197b151, 0x3e11e042},
		{0x4199c174, 0x3e29fd5b},
		{0x41a2a2e9, 0x3e0be9ba},
		{0x41a9a77f, 0x3a5c6111},
		{0x41a9b0d1, 0x385a6127},
		{0x41a9b13c, 0x378c2002},
		{0x41a9b16d, 0x350918a7},
		{0x41a9b16e, 0x3441895d},
		{0x41a9b16f, 0xb4214fdd},
		{0x41a9b170, 0xb5010a45},
		{0x41a9b177, 0xb63b8197},
		{0x41a9b199, 0xb76b63b0},
		{0x41a9b32b, 0xb91a03b9},
		{0x41a9b80a, 0xba128638},
		{0x41aac70c, 0xbcbf34b5},
		{0x41b9b4e0, 0xbe1a0663},
		{0x41c05fa0, 0xbd48a5dc},
		{0x41c175e7, 0xbce0c419},
		{0x41c2817a, 0xbbd016b8},
		{0x41c2c7bf, 0xba51641d},
		{0x41c2d1dc, 0xb444aee8},
		{0x41c2d1dd, 0x34067b47},
		{0x41c2d1ef, 0x36be7b8c},
		{0x41c2d229, 0x37c5addb},
		{0x41c2d23a, 0x37f1a963},
		{0x41c2d2ab, 0x388580ef},
		{0x41c2d5be, 0x39a0a104},
		{0x41c3002e, 0x3b6f89da},
		{0x41cf7807, 0x3e206a1e},
		{0x41d776a6, 0x3da768df},
		{0x41dbf2a4, 0x34b9009b},
		{0x41dbf2a5, 0x336959ae},
		{0x41dbf2a6, 0xb47d545d},
		{0x41dbf2af, 0xb63f2552},
		{0x41dbf41d, 0xb8e4c37b},
		{0x41dbf423, 0xb8e86a75},
		{0x41dc2f47, 0xbb938833},
		{0x41dedd93, 0xbd5cde71},
		{0x41f5137e, 0xb756c738},
		{0x41f5139a, 0xb6ab35de},
		{0x41f513ac, 0xb4243225},
		{0x41f513ad, 0x34030e59},
		{0x41f513af, 0x353463d4},
		{0x41f513b3, 0x35edd227},
		{0x41f51cfa, 0x3a2ba5f6},
		{0x41f58c50, 0x3c0aebf6},
		{0x41f61fb5, 0x3c99cbb9},
		{0x4206ef60, 0x3bbd4490},
		{0x4207194c, 0x3920b3e4},
		{0x42071a6d, 0x35fe0dfd},
		{0x42071a70, 0x34acaa79},
		{0x42071a71, 0xb4590955},
		{0x42071a81, 0xb70ffbaf},
		{0x420fa5da, 0xbde634f3},
		{0x4212fd7b, 0xbcb5f56c},
		{0x4213aafa, 0xb78e2ba6},
		{0x4213ab1b, 0xb4df73c2},
		{0x4213ab1c, 0x33b600ad},
		{0x4213ab1d, 0x351d3a0b},
		{0x4213ab21, 0x362dc876},
		{0x42151fc4, 0x3d3e86ae},
		{0x421d62a3, 0x3daa39c2},
		{0x4220284a, 0x3b1dad3f},
		{0x42203a11, 0x39636a73},
		{0x42203bd2, 0x357d7f2e},
		{0x42203bd4, 0xb2963192},
		{0x42203bd7, 0xb5c3fd50},
		{0x42203bd9, 0xb6228ad5},
		{0x422c985a, 0xbbcaf3e3},
		{0x422ccc62, 0xb7cac7c1},
		{0x422ccc95, 0xb5150865},
		{0x422ccc96, 0xb3c5c3c9},
		{0x422cd10a, 0x3a0a5ce7},
		{0x422d717b, 0x3c9f27c5},
		{0x422db7e4, 0x3ce1eb1d},
		{0x42395d60, 0x34689b83},
		{0x42395d61, 0xb4777d56},
		{0x4239600f, 0xb9a0ee0c},
		{0x42396aac, 0xbac76e79},
		{0x423973ec, 0xbb291157},
		{0x4245eda3, 0xb8811e54},
		{0x4245ee26, 0xb6a3e74d},
		{0x4245ee30, 0xb515b674},
		{0x4245ee31, 0xb4063eaf},
		{0x4245ee34, 0x359d7255},
		{0x4245f57d, 0x3a53d5fe},
		{0x4245f651, 0x3a6be0ea},
		{0x42527ea4, 0x382f02de},
		{0x42527f07, 0x3448dc3a},
		{0x42527f08, 0xb479a9d9},
		{0x42527f14, 0xb6b0bf90},
		{0x42527f87, 0xb86079e9},
		{0x4252811a, 0xb9696b9f},
		{0x42528db5, 0xbace9d16},
		{0x4253f9b4, 0xbd22432c},
		{0x425f0fe2, 0xb2b3d5c6},
		{0x425f0ff6, 0x370869fa},
		{0x425f2dcb, 0x3b4c7317},
		{0x425fda37, 0x3cab85e0},
		{0x425fdf48, 0x3cafc082},
		{0x426ba028, 0x387d7c9f},
		{0x426ba0c0, 0x3427e87a},
		{0x426ba0c5, 0xb5f525e1},
		{0x426ba0db, 0xb73104b8},
		{0x426ba10b, 0xb7f830cb},
		{0x426c7a45, 0xbcb33969},
		{0x426c9433, 0xbcc82a07},
		{0x42744966, 0xbdad4819},
		{0x427830ec, 0xb89368bc},
		{0x427831a2, 0x3319ac67},
		{0x427831ab, 0x366bc803},
		{0x4278324b, 0x388905ec},
		{0x42789405, 0x3c1f1652},
		{0x427fd7bd, 0x3dc08562},
	}
	for _, tt := range tests {
		for _, neg := range []bool{false, true} {
			x, want := NewFloat32FromBits(tt[0]), NewFloat32FromBits(tt[1])
			if neg {
				x = -x
			}
			if got := x.J0(); !eq32(got, want) {
				t.Errorf("J0(%v) = %v; want %v", x, got, want)
			}
		}
	}
}

// TestFloat32_J0Boundaries checks J0 on the both sides of the boundaries of the segments of the calculation.
func TestFloat32_J0Boundaries(t *testing.T) {
	t.Parallel()
	for _, x := range []float32{2, 3, 4, 5, 10, 32, 63, 64} {
		for d := -3; d <= 3; d++ {
			a := NewFloat32FromBits(math.Float32bits(x) + uint32(d))
			if got, want := a.J0(), NewFloat256(float64(a)).J0().Float32(); !eq32(got, want) {
				t.Errorf("J0(%v) = %v; want %v", a, got, want)
			}
		}
	}
}

// TestFloat32_J0Random compares J0 with Float256 on random inputs in [2, 64), and the zeros of J0.
func TestFloat32_J0Random(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(1, 2))
	for range 20000 {
		x := NewFloat32(2 + 62*r.Float64())
		if r.IntN(2) == 0 {
			x = -x
		}
		if got, want := x.J0(), NewFloat256(float64(x)).J0().Float32(); !eq32(got, want) {
			t.Fatalf("J0(%v) = %v; want %v", x, got, want)
		}
	}
	// near the zeros, where the relative error is large
	for _, z := range []float64{2.404825557695773, 5.520078110286311, 8.653727912911013, 11.791534439014281, 14.930917708487787, 18.071063967910924, 21.21163662987926, 24.352471530749302, 27.493479132040253, 30.634606468431976, 62.048469190227166} {
		for d := -20; d <= 20; d++ {
			x := NewFloat32FromBits(math.Float32bits(float32(z)) + uint32(d))
			if got, want := x.J0(), NewFloat256(float64(x)).J0().Float32(); !eq32(got, want) {
				t.Errorf("J0(%v) = %v; want %v", x, got, want)
			}
		}
	}
}

// TestFloat32_J0Poly checks the absolute error of the polynomials, which are hidden by the rounding to Float32,
// with Float256.
func TestFloat32_J0Poly(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewPCG(3, 4))
	for range 5000 {
		x := 2 + 62*r.Float64()
		i := int(x)
		c := &j032Coeffs[i-2]
		tt := x - float64(i) - 0.5
		got := 0.0
		for k := len(c) - 1; k >= 0; k-- {
			got = got*tt + c[k]
		}
		want := NewFloat256(x).J0().Float64().BuiltIn()
		if math.Abs(got-want) > 0x1p-50 {
			t.Errorf("polynomial(%v) = %v; want %v", x, got, want)
		}
	}
}

func BenchmarkFloat32_J0(b *testing.B) {
	for _, tt := range []struct {
		name string
		x    Float32
	}{
		{"small", exact32(0.5)},    // math.J0
		{"medium", exact32(5)},     // the polynomials
		{"medium2", exact32(50.5)}, // the polynomials
		{"negative", exact32(-10.5)},
		{"large", exact32(1000.5)}, // math.J0
		{"huge", NewFloat32(1e20)}, // math.J0
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.x.J0())
			}
		})
	}
}
