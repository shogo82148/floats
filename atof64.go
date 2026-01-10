package floats

import (
	"encoding"
	"encoding/json"
	"strconv"
)

// ParseFloat64 parses s as a Float64.
func ParseFloat64(s string) (Float64, error) {
	f, err := strconv.ParseFloat(s, 64)
	return Float64(f), err
}

var _ json.Unmarshaler = (*Float64)(nil)

// UnmarshalJSON implements [json.Unmarshaler].
func (a *Float64) UnmarshalJSON(data []byte) error {
	ret, err := ParseFloat64(string(data))
	if err != nil {
		return err
	}
	*a = ret
	return nil
}

var _ encoding.TextUnmarshaler = (*Float64)(nil)

// UnmarshalText implements [encoding.TextUnmarshaler].
func (a *Float64) UnmarshalText(data []byte) error {
	ret, err := ParseFloat64(string(data))
	if err != nil {
		return err
	}
	*a = ret
	return nil
}
