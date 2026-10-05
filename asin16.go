package floats

import "math"

// Asin returns the arcsine, in radians, of a.
//
// Special cases are:
//
//	±0.Asin() = ±0
//	x.Asin() = NaN if x < -1 or x > 1
func (a Float16) Asin() Float16 {
	return NewFloat16(math.Asin(a.Float64().BuiltIn()))
}

// Acos returns the arccosine, in radians, of a.
//
// Special case is:
//
//	x.Acos() = NaN if x < -1 or x > 1
func (a Float16) Acos() Float16 {
	return NewFloat16(math.Acos(a.Float64().BuiltIn()))
}

// Atan returns the arctangent, in radians, of a.
//
// Special cases are:
//
//	±0.Atan() = ±0
//	±Inf.Atan() = ±Pi/2
func (a Float16) Atan() Float16 {
	ix := a &^ signMask16
	if ix < 0x2400 { // |a| < 2**-6
		// atan(a) = a - a**3/3 + ... rounds to a.
		return a
	}
	if ix >= uvinf16 {
		if ix > uvinf16 {
			// atan(NaN) = NaN
			return NewFloat16NaN()
		}
		// atan(±Inf) = ±Pi/2
		return a&signMask16 | NewFloat16(math.Pi/2)
	}

	// a is a normal Float16 value, so the conversion is exact.
	x := normal16ToFloat64(ix)
	var r float64
	if x <= 1 {
		r = atanKernel16(x)
	} else {
		// atan(x) = pi/2 - atan(1/x)
		r = math.Pi/2 - atanKernel16(1/x)
	}
	if a&signMask16 != 0 {
		r = -r
	}
	return NewFloat16(r)
}

// atanKernel16 returns atan(x) for 0 <= x <= 1.
func atanKernel16(x float64) float64 {
	// atan(x) = 2 atan(u), where u = x/(1+sqrt(1+x**2)) <= tan(pi/8).
	u := x / (1 + math.Sqrt(1+x*x))
	return 2 * u * atanSeries16(u*u)
}

// atanSeries16 returns atan(sqrt(z))/sqrt(z) for 0 <= z <= tan(pi/8)**2.
// It is a degree-6 minimax polynomial; the relative error is about 2e-11,
// far below the Float16 precision.
func atanSeries16(z float64) float64 {
	return 0.9999999999791293 + z*(-0.3333333212787915+z*(0.19999885898675399+z*(-0.1428163925942391+z*(0.11041054162292493+z*(-0.08459109187133435+z*(0.04712997339340858))))))
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
func (a Float16) Atan2(b Float16) Float16 {
	return NewFloat16(math.Atan2(a.Float64().BuiltIn(), b.Float64().BuiltIn()))
}
