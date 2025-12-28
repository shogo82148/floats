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

// exact128 returns the Float128 representation of f.
func exact128(f float64) Float128 {
	return Float64(f).Float128()
}

// exact256 returns the Float256 representation of f.
func exact256(f float64) Float256 {
	return Float64(f).Float256()
}
