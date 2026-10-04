package floats

import "math"

// Log returns the natural logarithm of a.
//
// Special cases are:
//
//	+Inf.Log() = +Inf
//	0.Log() = -Inf
//	(x < 0).Log() = NaN
//	NaN.Log() = NaN
func (a Float32) Log() Float32 {
	ix := a.Bits()
	if ix-0x00000001 >= 0x7f7fffff { // a <= 0, +Inf, or NaN
		return log32Special(a)
	}
	k, l := logKernel32(ix)
	return Float32(k*math.Ln2 + l)
}

// Log10 returns the decimal logarithm of a.
// The special cases are the same as for [Log].
func (a Float32) Log10() Float32 {
	ix := a.Bits()
	if ix-0x00000001 >= 0x7f7fffff { // a <= 0, +Inf, or NaN
		return log32Special(a)
	}
	k, l := logKernel32(ix)
	return Float32(k*(math.Ln2/math.Ln10) + l*(1/math.Ln10))
}

// Log2 returns the binary logarithm of a.
// The special cases are the same as for [Log].
func (a Float32) Log2() Float32 {
	ix := a.Bits()
	if ix-0x00000001 >= 0x7f7fffff { // a <= 0, +Inf, or NaN
		return log32Special(a)
	}
	// l is exactly zero for powers of two, so they give an exact answer.
	k, l := logKernel32(ix)
	return Float32(k + l*(1/math.Ln2))
}

// log32Special handles the special cases of Log, Log10, and Log2.
func log32Special(a Float32) Float32 {
	switch {
	case a.IsNaN() || a.IsInf(1):
		return a
	case a == 0:
		return NewFloat32Inf(-1)
	default:
		// a < 0
		return NewFloat32NaN()
	}
}

// logKernel32 returns k and l such that log(x) = k*ln(2) + l,
// where ix is the bits of a positive finite Float32 value x.
// l is exactly zero if x is a power of two.
// The absolute error of l is less than 2**-50 for |l| < 2**-6,
// and the relative error is less than 2**-46 otherwise.
func logKernel32(ix uint32) (k, l float64) {
	var ke int32
	if ix < 0x00800000 {
		// x is subnormal; normalize it.
		ix = math.Float32bits(math.Float32frombits(ix) * 0x1p23)
		ke = -23
	}

	// x = 2**k * z, where 0.75 <= z < 1.5.
	tmp := ix - 0x3f400000
	ke += int32(tmp) >> 23
	iz := ix - tmp&0xff800000
	z := float64(math.Float32frombits(iz))

	// log(z) = log(c) + log(1+r), where r = z/c - 1.
	// c is the endpoint of the subinterval that is closer to 1,
	// so log(c) and log(1+r) have the same sign, and c = 1 near 1.
	i := (iz >> 17) & 63
	r := z*log32InvC[i] - 1
	return float64(ke), log32LogC[i] + log1pKernel(r)
}

// logKernel64 is the same as logKernel32, but x is a positive normal float64 value.
func logKernel64(x float64) (k, l float64) {
	// x = 2**k * z, where 0.75 <= z < 1.5.
	ix := math.Float64bits(x)
	tmp := ix - 0x3fe8000000000000
	ke := int64(tmp) >> 52
	iz := ix - tmp&0xfff0000000000000
	z := math.Float64frombits(iz)

	// log(z) = log(c) + log(1+r), where r = z/c - 1.
	i := (iz >> 46) & 63
	r := z*log32InvC[i] - 1
	return float64(ke), log32LogC[i] + log1pKernel(r)
}

// log1pKernel returns log(1+r) for -1/96 < r < 1/64.
// The relative error is less than 2**-50.
func log1pKernel(r float64) float64 {
	const (
		c3 = 0x1.555555554a969p-2
		c4 = -0x1.ffffffeb1d4abp-3
		c5 = 0x1.9999b20fad7bep-3
		c6 = -0x1.555ec177487afp-3
		c7 = 0x1.214efe25e612ep-3
	)
	r2 := r * r
	return r - 0.5*r2 + r2*r*(c3+r*(c4+r*(c5+r*(c6+r*c7))))
}

// log32InvC[i] = 1/c and log32LogC[i] = log(c), where
// c = 1 + i/64 for i < 32, and c = (1 + (i+1)/64)/2 for i >= 32.
var log32InvC = [64]float64{
	0x1.0000000000000p+0, 0x1.f81f81f81f820p-1, 0x1.f07c1f07c1f08p-1, 0x1.e9131abf0b767p-1,
	0x1.e1e1e1e1e1e1ep-1, 0x1.dae6076b981dbp-1, 0x1.d41d41d41d41dp-1, 0x1.cd85689039b0bp-1,
	0x1.c71c71c71c71cp-1, 0x1.c0e070381c0e0p-1, 0x1.bacf914c1bad0p-1, 0x1.b4e81b4e81b4fp-1,
	0x1.af286bca1af28p-1, 0x1.a98ef606a63bep-1, 0x1.a41a41a41a41ap-1, 0x1.9ec8e951033d9p-1,
	0x1.999999999999ap-1, 0x1.948b0fcd6e9e0p-1, 0x1.8f9c18f9c18fap-1, 0x1.8acb90f6bf3aap-1,
	0x1.8618618618618p-1, 0x1.8181818181818p-1, 0x1.7d05f417d05f4p-1, 0x1.78a4c8178a4c8p-1,
	0x1.745d1745d1746p-1, 0x1.702e05c0b8170p-1, 0x1.6c16c16c16c17p-1, 0x1.6816816816817p-1,
	0x1.642c8590b2164p-1, 0x1.6058160581606p-1, 0x1.5c9882b931057p-1, 0x1.58ed2308158edp-1,
	0x1.51d07eae2f815p+0, 0x1.4e5e0a72f0539p+0, 0x1.4afd6a052bf5bp+0, 0x1.47ae147ae147bp+0,
	0x1.446f86562d9fbp+0, 0x1.4141414141414p+0, 0x1.3e22cbce4a902p+0, 0x1.3b13b13b13b14p+0,
	0x1.3813813813814p+0, 0x1.3521cfb2b78c1p+0, 0x1.323e34a2b10bfp+0, 0x1.2f684bda12f68p+0,
	0x1.2c9fb4d812ca0p+0, 0x1.29e4129e4129ep+0, 0x1.27350b8812735p+0, 0x1.2492492492492p+0,
	0x1.21fb78121fb78p+0, 0x1.1f7047dc11f70p+0, 0x1.1cf06ada2811dp+0, 0x1.1a7b9611a7b96p+0,
	0x1.1811811811812p+0, 0x1.15b1e5f75270dp+0, 0x1.135c81135c811p+0, 0x1.1111111111111p+0,
	0x1.0ecf56be69c90p+0, 0x1.0c9714fbcda3bp+0, 0x1.0a6810a6810a7p+0, 0x1.0842108421084p+0,
	0x1.0624dd2f1a9fcp+0, 0x1.0410410410410p+0, 0x1.0204081020408p+0, 0x1.0000000000000p+0,
}

var log32LogC = [64]float64{
	0x0.0p+0, 0x1.fc0a8b0fc03e4p-7, 0x1.f829b0e783300p-6, 0x1.77458f632dcfcp-5,
	0x1.f0a30c01162a6p-5, 0x1.341d7961bd1d1p-4, 0x1.6f0d28ae56b4cp-4, 0x1.a926d3a4ad563p-4,
	0x1.e27076e2af2e6p-4, 0x1.0d77e7cd08e59p-3, 0x1.29552f81ff523p-3, 0x1.44d2b6ccb7d1ep-3,
	0x1.5ff3070a793d4p-3, 0x1.7ab890210d909p-3, 0x1.9525a9cf456b4p-3, 0x1.af3c94e80bff3p-3,
	0x1.c8ff7c79a9a22p-3, 0x1.e27076e2af2e6p-3, 0x1.fb9186d5e3e2bp-3, 0x1.0a324e27390e3p-2,
	0x1.1675cababa60ep-2, 0x1.22941fbcf7966p-2, 0x1.2e8e2bae11d31p-2, 0x1.3a64c556945eap-2,
	0x1.4618bc21c5ec2p-2, 0x1.51aad872df82dp-2, 0x1.5d1bdbf5809cap-2, 0x1.686c81e9b14afp-2,
	0x1.739d7f6bbd007p-2, 0x1.7eaf83b82afc3p-2, 0x1.89a3386c1425bp-2, 0x1.947941c2116fbp-2,
	-0x1.1bf99635a6b95p-2, -0x1.1178e8227e47cp-2, -0x1.07138604d5862p-2, -0x1.f991c6cb3b379p-3,
	-0x1.e530effe71012p-3, -0x1.d1037f2655e7bp-3, -0x1.bd087383bd8adp-3, -0x1.a93ed3c8ad9e3p-3,
	-0x1.95a5adcf7017fp-3, -0x1.823c16551a3c2p-3, -0x1.6f0128b756abcp-3, -0x1.5bf406b543db2p-3,
	-0x1.4913d8333b561p-3, -0x1.365fcb0159016p-3, -0x1.23d712a49c202p-3, -0x1.1178e8227e47cp-3,
	-0x1.fe89139dbd566p-4, -0x1.da727638446a2p-4, -0x1.b6ac88dad5b1cp-4, -0x1.9335e5d594989p-4,
	-0x1.700d30aeac0e1p-4, -0x1.4d3115d207eacp-4, -0x1.2aa04a44717a5p-4, -0x1.08598b59e3a07p-4,
	-0x1.ccb73cdddb2ccp-5, -0x1.894aa149fb343p-5, -0x1.466aed42de3eap-5, -0x1.0415d89e74444p-5,
	-0x1.8492528c8cabfp-6, -0x1.0205658935847p-6, -0x1.010157588de71p-7, 0x0.0p+0,
}
