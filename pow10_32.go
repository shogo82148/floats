package floats

import "math"

func NewFloat32Pow10(n int) Float32 {
	return NewFloat32(math.Pow10(n))
}
