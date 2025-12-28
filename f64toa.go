package floats

import (
	"fmt"
	"strconv"
)

var _ fmt.Stringer = Float64(0)

// String returns the string representation of a.
func (a Float64) String() string {
	return a.Text('g', -1)
}

// Text returns the string representation of a in the given format and precision.
func (a Float64) Text(fmt byte, prec int) string {
	return strconv.FormatFloat(float64(a), fmt, prec, 64)
}

// Append appends the string representation of a in the given format and precision to dst and returns the extended buffer.
func (a Float64) Append(dst []byte, fmt byte, prec int) []byte {
	return strconv.AppendFloat(dst, float64(a), fmt, prec, 64)
}
