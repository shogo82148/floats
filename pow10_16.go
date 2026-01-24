package floats

import "math"

func NewFloat16Pow10(n int) Float16 {
	return NewFloat16(math.Pow10(n))
}
