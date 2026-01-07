package floats

import (
	"fmt"
	"math"
	"testing"

	"github.com/shogo82148/ints"
)

func TestFloat256_Format(t *testing.T) {
	tests := []struct {
		format string
		x      Float256
		want   string
	}{
		// verb "%b"
		{"%b", exact256(0), "0p-262378"},

		// verb "%f"
		{"%f", exact256(0.5), "0.5"},
		{"%f", exact256(-0.5), "-0.5"},
		{"%+f", exact256(0.5), "+0.5"},
		{"%+f", exact256(-0.5), "-0.5"},
		{"% f", exact256(0.5), " 0.5"},
		{"% f", exact256(-0.5), "-0.5"},
		{"%8f", exact256(0.5), "     0.5"},
		{"%-8f", exact256(0.5), "0.5     "},
		{"%.2f", exact256(0.5), "0.50"},

		// verb "%e"
		{"%.6e", exact256(0.5), "5.000000e-01"},

		// verb "%g"
		{"%g", exact256(0.5), "0.5"},
		{"%.1g", exact256(0.25), "0.2"},

		// verb "%x"
		{"%x", exact256(0.5), "0x1p-01"},
		{"%#x", exact256(0.5), "0x1p-01"},
		{"%.1x", exact256(0.5), "0x1.0p-01"},

		// verb "%X"
		{"%X", exact256(0.5), "0X1P-01"},
		{"%#X", exact256(0.5), "0X1P-01"},
		{"%.1X", exact256(0.5), "0X1.0P-01"},

		// verb "%v"
		{"%v", exact256(0.5), "0.5"},
		{"%v", exact256(math.NaN()), "NaN"},
	}

	for _, tt := range tests {
		got := fmt.Sprintf(tt.format, tt.x)
		if got != tt.want {
			t.Errorf("expected %s, got %s", tt.want, got)
		}
	}
}

func TestFloat256_Text(t *testing.T) {
	tests := []struct {
		x    Float256
		fmt  byte
		prec int
		s    string
	}{
		/****** decimal formats ******/
		{exact256(0x1p-24), 'f', 24, "0.000000059604644775390625"},
		{exact256(0x2p-24), 'f', 24, "0.000000119209289550781250"},
		{exact256(0x3p-24), 'f', 24, "0.000000178813934326171875"},
		{exact256(0x4p-24), 'f', 24, "0.000000238418579101562500"},
		{exact256(0x5p-24), 'f', 24, "0.000000298023223876953125"},
		{exact256(0x6p-24), 'f', 24, "0.000000357627868652343750"},
		{exact256(0x7p-24), 'f', 24, "0.000000417232513427734375"},
		{exact256(0x8p-24), 'f', 24, "0.000000476837158203125000"},
		{exact256(0x9p-24), 'f', 24, "0.000000536441802978515625"},
		{exact256(0xap-24), 'f', 24, "0.000000596046447753906250"},

		/****** binary exponent formats ******/
		{exact256(0.0), 'b', -1, "0p-262378"},
		{exact256(1.0), 'b', -1, "110427941548649020598956093796432407239217743554726184882600387580788736p-236"},

		/******* alternate formats *******/
		{
			Float256{
				0x7fff_efff_ffff_ffff, 0xffff_ffff_ffff_ffff,
				0xffff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff,
			}, 'g', -1,
			"1.61132571748576047361957211845200501064402387454966951747637125049607183e+78913",
		},
		{
			Float256{
				0x0000_0000_0000_0000, 0x0000_0000_0000_0000,
				0x0000_0000_0000_0000, 0x0000_0000_0000_0001,
			}, 'g', -1, "2e-78984",
		},

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
