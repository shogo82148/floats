package floats

import "math"

// Asin returns the arcsine, in radians, of a.
//
// Special cases are:
//
//	±0.Asin() = ±0
//	x.Asin() = NaN if x < -1 or x > 1
func (a Float32) Asin() Float32 {
	return NewFloat32(math.Asin(a.Float64().BuiltIn()))
}

// Acos returns the arccosine, in radians, of a.
//
// Special case is:
//
//	x.Acos() = NaN if x < -1 or x > 1
func (a Float32) Acos() Float32 {
	return NewFloat32(math.Acos(a.Float64().BuiltIn()))
}

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float32) Atan() Float32 {
	ix := a.Bits() &^ signMask32
	if ix < 0x39800000 { // |a| < 2**-12
		// atan(a) = a - a**3/3 + ... rounds to a.
		return a
	}
	if ix > 0x7f800000 { // NaN
		return NewFloat32NaN()
	}

	// atan(±Inf) = ±Pi/2, because 1/Inf = 0.
	x := float64(math.Float32frombits(ix))
	var r float64
	if x <= 1 {
		r = atanKernel32(x)
	} else {
		// atan(x) = pi/2 - atan(1/x)
		r = math.Pi/2 - atanKernel32(1/x)
	}
	if a < 0 {
		r = -r
	}
	return Float32(r)
}

// atanKernel32 returns atan(x) for 0 <= x <= 1.
func atanKernel32(x float64) float64 {
	// atan(x) = 2 atan(u), where u = x/(1+sqrt(1+x**2)) <= tan(pi/8).
	u := x / (1 + math.Sqrt(1+x*x))
	return 2 * u * atanSeries32(u*u)
}

// atanSeries32 returns atan(sqrt(z))/sqrt(z) for 0 <= z <= tan(pi/8)**2.
// It is a degree-10 minimax polynomial; the relative error is about 1e-16.
func atanSeries32(z float64) float64 {
	return 1.0 + z*(-0.3333333333332843+z*(0.19999999998853618+z*(-0.14285714180856576+z*(0.11111106175584769+z*(-0.09090772961650327+z*(0.07689951902451446+z*(-0.0664022008540203+z*(0.056882767804654844+z*(-0.0434784293457331+z*(0.021132810313338496))))))))))
}

// Atan2 returns the arc tangent of a/b, using
// the signs of the two to determine the quadrant
// of the return value.
//
// Special cases are (in order):
//
//	y.Atan2(NaN) = NaN
//	NaN.Atan2(x) = NaN
//	+0.Atan2(x>=0) = +0
//	-0.Atan2(x>=0) = -0
//	+0.Atan2(x<=-0) = +Pi
//	-0.Atan2(x<=-0) = -Pi
//	y>0.Atan2(0) = +Pi/2
//	y<0.Atan2(0) = -Pi/2
//	+Inf.Atan2(+Inf) = +Pi/4
//	-Inf.Atan2(+Inf) = -Pi/4
//	+Inf.Atan2(-Inf) = 3Pi/4
//	-Inf.Atan2(-Inf) = -3Pi/4
//	y.Atan2(+Inf) = 0
//	(y>0).Atan2(-Inf) = +Pi
//	(y<0).Atan2(-Inf) = -Pi
//	+Inf.Atan2(x) = +Pi/2
//	-Inf.Atan2(x) = -Pi/2
func (a Float32) Atan2(b Float32) Float32 {
	return NewFloat32(math.Atan2(a.Float64().BuiltIn(), b.Float64().BuiltIn()))
}
