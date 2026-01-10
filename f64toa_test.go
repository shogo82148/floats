package floats

import (
	"fmt"
	"math"
	"testing"
)

func TestFloat64_Format(t *testing.T) {
	tests := []struct {
		format string
		x      Float64
		want   string
	}{
		// verb "%b"
		{"%b", exact64(0), "0p-1074"},

		// verb "%f"
		{"%f", exact64(0.5), "0.5"},
		{"%f", exact64(-0.5), "-0.5"},
		{"%+f", exact64(0.5), "+0.5"},
		{"%+f", exact64(-0.5), "-0.5"},
		{"% f", exact64(0.5), " 0.5"},
		{"% f", exact64(-0.5), "-0.5"},
		{"%8f", exact64(0.5), "     0.5"},
		{"%-8f", exact64(0.5), "0.5     "},
		{"%.2f", exact64(0.5), "0.50"},

		// verb "%e"
		{"%.6e", exact64(0.5), "5.000000e-01"},

		// verb "%g"
		{"%g", exact64(0.5), "0.5"},
		{"%.1g", exact64(0.25), "0.2"},

		// verb "%x"
		{"%x", exact64(0.5), "0x1p-01"},
		{"%#x", exact64(0.5), "0x1p-01"},
		{"%.1x", exact64(0.5), "0x1.0p-01"},

		// verb "%X"
		{"%X", exact64(0.5), "0X1P-01"},
		{"%#X", exact64(0.5), "0X1P-01"},
		{"%.1X", exact64(0.5), "0X1.0P-01"},

		// verb "%v"
		{"%v", exact64(0.5), "0.5"},
		{"%v", exact64(math.NaN()), "NaN"},
	}

	for _, tt := range tests {
		got := fmt.Sprintf(tt.format, tt.x)
		if got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

func TestFloat64_Text(t *testing.T) {
	tests := []struct {
		x    Float64
		fmt  byte
		prec int
		s    string
	}{
		{Float64(0), 'e', 6, "0.000000e+00"},
		{Float64(1.5), 'f', 2, "1.50"},
		{Float64(-2.75), 'g', -1, "-2.75"},
		{Float64(3.1415927), 'g', 5, "3.1416"},
		{Float64(math.Inf(1)), 'g', -1, "+Inf"},
		{Float64(math.Inf(-1)), 'g', -1, "-Inf"},
		{Float64(math.NaN()), 'g', -1, "NaN"},
	}

	for _, tt := range tests {
		if got := tt.x.Text(tt.fmt, tt.prec); got != tt.s {
			t.Errorf("Float64(%g).Text(%q, %d) = %q, want %q", float64(tt.x), tt.fmt, tt.prec, got, tt.s)
		}
	}
}

func TestFloat64_MarshalJSON(t *testing.T) {
	tests := []struct {
		x    Float64
		want string
	}{
		{exact64(0), "0"},
		{exact64(-0), "0"},
		{exact64(1), "1"},
		{exact64(-1), "-1"},
		{exact64(0.5), "0.5"},
	}

	for _, tt := range tests {
		got, err := tt.x.MarshalJSON()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if string(got) != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}

	// JSON does not support NaN and Inf values.
	var err error
	_, err = exact64(math.NaN()).MarshalJSON()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	_, err = exact64(math.Inf(1)).MarshalJSON()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	_, err = exact64(math.Inf(-1)).MarshalJSON()
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestFloat64_MarshalText(t *testing.T) {
	tests := []struct {
		x    Float64
		want string
	}{
		{exact64(0), "0"},
		{exact64(-0), "0"},
		{exact64(1), "1"},
		{exact64(-1), "-1"},
		{exact64(0.5), "0.5"},

		// special values
		{exact64(math.Inf(1)), "+Inf"},
		{exact64(math.Inf(-1)), "-Inf"},
		{exact64(math.NaN()), "NaN"},
	}

	for _, tt := range tests {
		got, err := tt.x.MarshalText()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if string(got) != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

func TestFloat64_AppendText(t *testing.T) {
	tests := []struct {
		x    Float64
		want string
	}{
		{exact64(0), "0"},
		{exact64(-0), "0"},
		{exact64(1), "1"},
		{exact64(-1), "-1"},
		{exact64(0.5), "0.5"},

		// special values
		{exact64(math.Inf(1)), "+Inf"},
		{exact64(math.Inf(-1)), "-Inf"},
		{exact64(math.NaN()), "NaN"},
	}

	for _, tt := range tests {
		got, err := tt.x.AppendText(nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			continue
		}
		if string(got) != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}
