package floats

import (
	"encoding"
	"encoding/json"
	"strconv"
)

// ParseFloat32 parses s as a Float32.
func ParseFloat32(s string) (Float32, error) {
	f, err := strconv.ParseFloat(s, 32)
	return Float32(f), err
}

var _ json.Unmarshaler = (*Float32)(nil)

// UnmarshalJSON implements [json.Unmarshaler].
func (a *Float32) UnmarshalJSON(data []byte) error {
	ret, err := ParseFloat32(string(data))
	if err != nil {
		return err
	}
	*a = ret
	return nil
}

var _ encoding.TextUnmarshaler = (*Float32)(nil)

// UnmarshalText implements [encoding.TextUnmarshaler].
func (a *Float32) UnmarshalText(data []byte) error {
	ret, err := ParseFloat32(string(data))
	if err != nil {
		return err
	}
	*a = ret
	return nil
}
