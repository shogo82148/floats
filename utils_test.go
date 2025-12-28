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
