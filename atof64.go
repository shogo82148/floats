package floats

import "strconv"

// ParseFloat64 parses s as a Float64.
func ParseFloat64(s string) (Float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	return Float64(f), err
}
