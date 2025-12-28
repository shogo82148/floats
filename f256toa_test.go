package floats

import (
	"math"
	"testing"

	"github.com/shogo82148/ints"
)

func TestFloat256_Text(t *testing.T) {
	tests := []struct {
		x    Float256
		fmt  byte
		prec int
		s    string
	}{
		/****** binary exponent formats ******/
		{exact256(0.0), 'b', -1, "0p-262378"},
		{exact256(1.0), 'b', -1, "110427941548649020598956093796432407239217743554726184882600387580788736p-236"},

		/******* hexadecimal formats *******/
		{exact256(0), 'x', -1, "0x0p+00"},
		{exact256(math.Inf(1)), 'x', -1, "+Inf"},
		{exact256(math.Inf(-1)), 'x', -1, "-Inf"},
		{exact256(math.NaN()), 'x', -1, "NaN"},

		{exact256(0x0.0p0), 'x', 0, "0x0p+00"},
		{exact256(0x1.0p0), 'x', 0, "0x1p+00"},
		{exact256(0x1.7p0), 'x', 0, "0x1p+00"},
		{exact256(0x1.8p0), 'x', 0, "0x1p+01"},

		{exact256(0x0.0p0), 'x', 1, "0x0.0p+00"},
		{exact256(0x1.0p0), 'x', 1, "0x1.0p+00"},
		{exact256(0x1.fp0), 'x', 1, "0x1.fp+00"},
		{exact256(0x1.08p0), 'x', 1, "0x1.0p+00"},
		{exact256(0x1.18p0), 'x', 1, "0x1.2p+00"},

		{exact256(0x1p0), 'x', -1, "0x1p+00"},
		{exact256(0x1.8p0), 'x', -1, "0x1.8p+00"},
		{exact256(0x1.08p0), 'x', -1, "0x1.08p+00"},
		{exact256(0x1.00cp0), 'x', -1, "0x1.00cp+00"},
	}

	for _, tt := range tests {
		if got := tt.x.Text(tt.fmt, tt.prec); got != tt.s {
			t.Errorf("Float256(%x).Text(%q, %d) = %q, want %q", ints.Uint256(tt.x), tt.fmt, tt.prec, got, tt.s)
		}
	}
}
