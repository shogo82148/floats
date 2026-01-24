package floats

// Log1p returns the natural logarithm of 1 plus its argument a.
// It is more accurate than [Log](1 + a) when a is near zero.
//
// Special cases are:
//
//	+Inf.Log1p() = +Inf
//	±0.Log1p() = ±0
//	-1.Log1p() = -Inf
//	(a < -1).Log1p() = NaN
//	NaN.Log1p() = NaN
func (a Float128) Log1p() Float128 {
	var (
		// One = 1.0
		One = Float128(uvone128)

		// Half = 0.5
		Half = Float128{0x3ffe_0000_0000_0000, 0x0000_0000_0000_0000}

		SQRTH = Float128{0x3ffe_6a09_e667_f3bc, 0xc908_ab01_8892_c9e2} // sqrt(0.5)

		C1 = Float128{0x3ffe_62e4_0000_0000, 0x0000_0000_0000_0000} // 0.693145751953125
		C2 = Float128{0x3feb_7f7d_1cf7_9abc, 0x9e3b_39b5_43aa_3474} // 1.4286068203094172321215e-06

		P0 = Float128{0x3ff07bc0962b395c, 0xa37253a16dc85690} // 4.5270000862445199635215e-05
		P1 = Float128{0x3ffdfe818a0fe1a8, 0x339ed3dbe100d392} // 0.49854102823193375972212
		P2 = Float128{0x4001a509f46f4fa5, 0x3284bceec3629e1c} // 6.5787325942061044846969
		P3 = Float128{0x4003de9738b8cb9c, 0x95b9a96a3a1ecd0b} // 29.911919328553073277375
		P4 = Float128{0x4004e798eb86c335, 0x08893f71db809293} // 60.949667980987787057556
		P5 = Float128{0x4004c8e7597479a1, 0x03560bd3408e64d6} // 57.112963590585538103336
		p6 = Float128{0x400340a202d99830, 0x997dfb1c6f1a4f1a} // 20.039553499201281259648

		Q0 = Float128{0x4002e20359e903e3, 0x716e473d88c1a8c5} // 15.062909083469192043167
		Q1 = Float128{0x40054c30b5221349, 0x78610046811b5da5} // 83.047565967967209469434
		Q2 = Float128{0x4006bb86590fcfb5, 0x5dd7f9fe7a261293} // 221.76239823732856465394
		Q3 = Float128{0x4007351945dc908a, 0x57bcf5d45c71f3de} // 309.09872225312059774938
		Q4 = Float128{0x4006b0db13e48e06, 0x623f442ee8bdc9a2} // 216.42788614495947685003
		Q5 = Float128{0x4004e0f304466448, 0xe68de27115cbd005} // 60.118660497603843919306

		R0 = Float128{0x3ff602f6eec7f244, 0xdde7a1fbccfd1e69} // 0.0019757429581415468984296
		R1 = Float128{0xbffe7097bd1e35f2, 0x2bfa1cd8825e6ebd} // -0.71990767473014147232598
		R2 = Float128{0x400258df4a789f1b, 0x172b5d58ee1e58d7} // 10.777257190312272158094
		R3 = Float128{0xc0041dbdd15d69c7, 0x126354d8b67fa0c6} // -35.717684488096787370998

		S0 = Float128{0xc003a3377b8a3f92, 0xf9c82d7b2676fc64} // 26.201045551331104417768
		S1 = Float128{0x4006833ce2de1a20, 0x15e5b3bb2a345bbd} // 193.61891836232102174846
		S2 = Float128{0xc007ac9cba0c1eaa, 0x9af9b63dbdaa4949} // -428.61221385716144629696
	)

	// special cases
	switch {
	case a.Lt(One.Neg()) || a.IsNaN(): // includes -Inf
		return NewFloat128NaN()
	case a.Eq(One.Neg()):
		return NewFloat128Inf(-1)
	case a.IsInf(1):
		return NewFloat128Inf(1)
	}

	// Separate mantissa from exponent.
	// Use frexp so that denormal numbers will be handled properly.
	x := a.Add(One)
	x, exp := x.Frexp()

	// logarithm using log(x) = z + z^3 P(z)/Q(z),
	// where z = 2(x-1)/x+1)
	if exp > 2 || exp < -2 {
		var y, z Float128
		if x.Lt(SQRTH) { // 2(2x-1)/(2x+1)
			exp--
			z = x.Sub(Half)
			y = z.Mul(Half).Add(Half)
		} else { // 2(x-1)/(x+1)
			z = x.Sub(Half)
			z = z.Sub(Half)
			y = z.Mul(Half).Add(Half)
		}
		x = z.Quo(y)
		z = x.Mul(x)
		r := R0
		r = FMA128(r, z, R1)
		r = FMA128(r, z, R2)
		r = FMA128(r, z, R3)
		s := z.Add(S0)
		s = FMA128(s, z, S1)
		s = FMA128(s, z, S2)
		z = x.Mul(z.Mul(r.Quo(s)))
		z = z.Add(NewFloat128(float64(exp)).Mul(C2))
		z = z.Add(x)
		z = z.Add(NewFloat128(float64(exp)).Mul(C1))
		return z
	}

	// logarithm using log(1+x) = x - .5x**2 + x**3 P(x)/Q(x)
	if x.Lt(SQRTH) {
		exp--
		if exp != 0 {
			x = x.Add(x).Sub(One)
		} else {
			x = a
		}
	} else {
		if exp != 0 {
			x = x.Sub(One)
		} else {
			x = a
		}
	}
	z := x.Mul(x)
	p := P0
	p = FMA128(p, x, P1)
	p = FMA128(p, x, P2)
	p = FMA128(p, x, P3)
	p = FMA128(p, x, P4)
	p = FMA128(p, x, P5)
	p = FMA128(p, x, p6)
	q := x.Add(Q0)
	q = FMA128(q, x, Q1)
	q = FMA128(q, x, Q2)
	q = FMA128(q, x, Q3)
	q = FMA128(q, x, Q4)
	q = FMA128(q, x, Q5)
	y := x.Mul(z.Mul(p.Quo(q)))
	y = y.Add(NewFloat128(float64(exp)).Mul(C2))
	z = y.Sub(Half.Mul(z))
	z = z.Add(x)
	z = z.Add(NewFloat128(float64(exp)).Mul(C1))
	return z
}
