package floats

import (
	"fmt"
	"strconv"
)

var _ fmt.Formatter = Float32(0)

// Format implements [fmt.Formatter].
func (a Float32) Format(s fmt.State, verb rune) {
	format(a, s, verb)
}

var _ fmt.Stringer = Float32(0)

// String returns the string representation of a.
func (a Float32) String() string {
	return a.Text('g', -1)
}

// Text returns the string representation of a in the given format and precision.
func (a Float32) Text(fmt byte, prec int) string {
	return strconv.FormatFloat(float64(a), fmt, prec, 32)
}

// Append appends the string representation of a in the given format and precision to dst and returns the extended buffer.
func (a Float32) Append(dst []byte, fmt byte, prec int) []byte {
	return strconv.AppendFloat(dst, float64(a), fmt, prec, 32)
}
