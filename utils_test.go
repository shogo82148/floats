package floats

import (
	"cmp"
	"fmt"
)

// exact16 returns the Float16 representation of f.
// It panics if f does not have an exact Float16 representation.
func exact16(f float64) Float16 {
	ret := Float64(f).Float16()
	if cmp.Compare(ret.Float64(), Float64(f)) != 0 {
		panic(fmt.Sprintf("%f doesn't have exact float16 representation", f))
	}
	return ret
}

// exact32 returns the Float32 representation of f.
// It panics if f does not have an exact Float32 representation.
func exact32(f float64) Float32 {
	ret := Float64(f).Float32()
	if cmp.Compare(ret.Float64(), Float64(f)) != 0 {
		panic(fmt.Sprintf("%f doesn't have exact float32 representation", f))
	}
	return ret
}

// exact64 returns the Float64 representation of f.
func exact64(f float64) Float64 {
	return Float64(f)
}

// exact128 returns the Float128 representation of f.
func exact128(f float64) Float128 {
	return Float64(f).Float128()
}

// exact256 returns the Float256 representation of f.
func exact256(f float64) Float256 {
	return Float64(f).Float256()
}

func tolerance(a, b, e float64) bool {
	if a == b {
		return true
	}
	d := a - b
	if d < 0 {
		d = -d
	}

	if b != 0 {
		e = e * b
		if e < 0 {
			e = -e
		}
	}
	return d < e
}

// close16 reports whether a is close to b within tolerance 1e-3.
func close16(a Float16, b float64) bool {
	return tolerance(a.Float64().BuiltIn(), b, 1e-3)
}

// close32 reports whether a is close to b within tolerance 1e-6.
func close32(a Float32, b float64) bool {
	return tolerance(float64(a), b, 1e-6)
}

// close64 reports whether a is close to b within tolerance 1e-15.
func close64(a Float64, b float64) bool {
	return tolerance(float64(a), b, 1e-15)
}

// close128 reports whether a is close to b within tolerance 1e-15.
func close128(a Float128, b float64) bool {
	return tolerance(a.Float64().BuiltIn(), b, 1e-15)
}

// close256 reports whether a is close to b within tolerance 1e-15.
func close256(a Float256, b float64) bool {
	return tolerance(a.Float64().BuiltIn(), b, 1e-15)
}
