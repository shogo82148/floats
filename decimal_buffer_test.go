package floats

import "testing"

// TestDecimalBufferExtremes guards the per-type decimal backing buffer sizes
// (decimalDigits16/BF16/128/256). The worst-case decimal expansion is produced by
// the largest subnormal value; if a buffer were too small the intermediate
// expansion would be silently truncated and the shortest representation would
// no longer round-trip.
func TestDecimalBufferExtremes(t *testing.T) {
	t.Parallel()
	// largest and smallest subnormal BFloat16
	for _, x := range []BFloat16{0x007f, 0x0001, 0x807f, 0x8001} {
		if y, err := ParseBFloat16(x.String()); err != nil || y != x {
			t.Errorf("BFloat16 %#04x round-trip failed: s=%s err=%v", uint16(x), x.String(), err)
		}
		if y, err := ParseBFloat16(x.Text('e', 100)); err != nil || y != x {
			t.Errorf("BFloat16 %#04x round-trip of the exact value failed: err=%v", uint16(x), err)
		}
	}

	// largest subnormal Float128
	x128 := Float128{0x0000_0fff_ffff_ffff, 0xffff_ffff_ffff_ffff}
	if y, err := ParseFloat128(x128.String()); err != nil || y != x128 {
		t.Errorf("Float128 largest subnormal round-trip failed: s=%s err=%v", x128.String(), err)
	}

	// smallest subnormal Float128
	min128 := Float128{0x0000_0000_0000_0000, 0x0000_0000_0000_0001}
	if y, err := ParseFloat128(min128.String()); err != nil || y != min128 {
		t.Errorf("Float128 smallest subnormal round-trip failed: err=%v", err)
	}

	// largest subnormal Float256
	x256 := Float256{0x0000_0fff_ffff_ffff, 0xffff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff, 0xffff_ffff_ffff_ffff}
	if y, err := ParseFloat256(x256.String()); err != nil || y != x256 {
		t.Errorf("Float256 largest subnormal round-trip failed: err=%v", err)
	}

	// smallest subnormal Float256
	min256 := Float256{0x0000_0000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0000, 0x0000_0000_0000_0001}
	if y, err := ParseFloat256(min256.String()); err != nil || y != min256 {
		t.Errorf("Float256 smallest subnormal round-trip failed: err=%v", err)
	}
}
