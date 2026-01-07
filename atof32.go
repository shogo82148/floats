package floats

import "strconv"

// ParseFloat32 parses s as a Float32.
func ParseFloat32(s string) (Float32, error) {
	f, err := strconv.ParseFloat(s, 32)
	return Float32(f), err
}
