package floats

import (
	"encoding"
	"encoding/json"
	"fmt"
	"strconv"
)

var _ fmt.Formatter = Float64(0)

// Format implements [fmt.Formatter].
func (a Float64) Format(s fmt.State, verb rune) {
	format(a, s, verb)
}

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

var _ json.Marshaler = Float64(0)

// MarshalJSON implements [json.Marshaler].
func (a Float64) MarshalJSON() ([]byte, error) {
	if a.IsNaN() || a.IsInf(0) {
		return nil, fmt.Errorf("floats: cannot marshal %v to JSON", a)
	}
	return a.Append(nil, 'g', -1), nil
}

var _ encoding.TextMarshaler = Float64(0)

// MarshalText implements [encoding.TextMarshaler].
func (a Float64) MarshalText() ([]byte, error) {
	return a.Append(nil, 'g', -1), nil
}

var _ encoding.TextAppender = Float64(0)

// AppendText implements [encoding.TextAppender].
func (a Float64) AppendText(dst []byte) ([]byte, error) {
	return a.Append(dst, 'g', -1), nil
}
