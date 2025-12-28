package floats

import (
	"math"
	"testing"

	"github.com/shogo82148/ints"
)

func TestFloat128_Text(t *testing.T) {
	tests := []struct {
		x    Float128
		fmt  byte
		prec int
		s    string
	}{
		/****** binary exponent formats ******/
		{exact128(0x0p+0), 'b', -1, "0p-16494"},
		{exact128(1), 'b', -1, "5192296858534827628530496329220096p-112"},
		{exact128(1.5), 'b', -1, "7788445287802241442795744493830144p-112"},

		/******* hexadecimal formats *******/
		{exact128(0), 'x', -1, "0x0p+00"},
		{exact128(math.Inf(1)), 'x', -1, "+Inf"},
		{exact128(math.Inf(-1)), 'x', -1, "-Inf"},
		{exact128(math.NaN()), 'x', -1, "NaN"},

		{exact128(0x0.0p0), 'x', 0, "0x0p+00"},
		{exact128(0x1.0p0), 'x', 0, "0x1p+00"},
		{exact128(0x1.7p0), 'x', 0, "0x1p+00"},
		{exact128(0x1.8p0), 'x', 0, "0x1p+01"},

		{exact128(0x0.0p0), 'x', 1, "0x0.0p+00"},
		{exact128(0x1.0p0), 'x', 1, "0x1.0p+00"},
		{exact128(0x1.fp0), 'x', 1, "0x1.fp+00"},
		{exact128(0x1.08p0), 'x', 1, "0x1.0p+00"},
		{exact128(0x1.18p0), 'x', 1, "0x1.2p+00"},

		{exact128(0x1p0), 'x', -1, "0x1p+00"},
		{exact128(0x1.8p0), 'x', -1, "0x1.8p+00"},
		{exact128(0x1.08p0), 'x', -1, "0x1.08p+00"},
		{exact128(0x1.00cp0), 'x', -1, "0x1.00cp+00"},
	}

	for _, tt := range tests {
		if got := tt.x.Text(tt.fmt, tt.prec); got != tt.s {
			t.Errorf("Float128(%x).Text(%q, %d) = %q, want %q", ints.Uint128(tt.x), tt.fmt, tt.prec, got, tt.s)
		}
	}
}
