package floats

import (
	"math"
	"testing"
)

func TestFloat16_Text(t *testing.T) {
	tests := []struct {
		x    Float16
		fmt  byte
		prec int
		s    string
	}{
		/****** binary exponent formats ******/
		{0, 'b', -1, "0p-24"},
		{0x8000, 'b', -1, "-0p-24"},
		{exact16(math.Inf(1)), 'x', -1, "+Inf"},
		{exact16(math.Inf(-1)), 'x', -1, "-Inf"},
		{exact16(math.NaN()), 'x', -1, "NaN"},

		{0x0001, 'b', -1, "1p-24"},
		{exact16(1), 'b', -1, "1024p-10"},
		{exact16(1.5), 'b', -1, "1536p-10"},
		{exact16(65504), 'b', -1, "2047p+5"},

		/******* hexadecimal formats *******/
		{0, 'x', -1, "0x0p+00"},
		{0x8000, 'x', -1, "-0x0p+00"},
		{exact16(math.Inf(1)), 'x', -1, "+Inf"},
		{exact16(math.Inf(-1)), 'x', -1, "-Inf"},
		{exact16(math.NaN()), 'x', -1, "NaN"},

		{exact16(0x0.0p0), 'x', 0, "0x0p+00"},
		{exact16(0x1.0p0), 'x', 0, "0x1p+00"},
		{exact16(0x1.7p0), 'x', 0, "0x1p+00"},
		{exact16(0x1.8p0), 'x', 0, "0x1p+01"},

		{exact16(0x0.0p0), 'x', 1, "0x0.0p+00"},
		{exact16(0x1.0p0), 'x', 1, "0x1.0p+00"},
		{exact16(0x1.fp0), 'x', 1, "0x1.fp+00"},
		{exact16(0x1.ffc0p0), 'x', 1, "0x1.0p+01"},
		{exact16(0x1.08p0), 'x', 1, "0x1.0p+00"},
		{exact16(0x1.18p0), 'x', 1, "0x1.2p+00"},

		{exact16(0x0.00p0), 'x', 2, "0x0.00p+00"},
		{exact16(0x1.00p0), 'x', 2, "0x1.00p+00"},
		{exact16(0x1.0fp0), 'x', 2, "0x1.0fp+00"},
		{exact16(0x1.ffc0p0), 'x', 2, "0x1.00p+01"},
		{exact16(0x1.008p0), 'x', 2, "0x1.00p+00"},
		{exact16(0x1.018p0), 'x', 2, "0x1.02p+00"},

		{exact16(0x0.000p0), 'x', 3, "0x0.000p+00"},
		{exact16(0x1.000p0), 'x', 3, "0x1.000p+00"},
		{exact16(0x1.00c0p0), 'x', 3, "0x1.00cp+00"},
		{exact16(0x1.ffc0p0), 'x', 3, "0x1.ffcp+00"},

		{exact16(0x1.ffc0p0), 'x', 4, "0x1.ffc0p+00"},
		{exact16(0x1p-24), 'x', 4, "0x1.0000p-24"},

		{exact16(0x1p0), 'x', -1, "0x1p+00"},
		{exact16(0x1.8p0), 'x', -1, "0x1.8p+00"},
		{exact16(0x1.08p0), 'x', -1, "0x1.08p+00"},
		{exact16(0x1.00cp0), 'x', -1, "0x1.00cp+00"},

		{0, 'X', -1, "0X0P+00"},
		{exact16(0x1.ffc0p0), 'X', 3, "0X1.FFCP+00"},
	}

	for _, tt := range tests {
		got := tt.x.Text(tt.fmt, tt.prec)
		if got != tt.s {
			t.Errorf("%#v: expected %s, got %s", tt, tt.s, got)
		}
	}
}
