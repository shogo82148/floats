package floats

import "testing"

// TestDecimalBufferExtremes guards the per-type decimal backing buffer sizes
// (decimalDigits16/128/256). The worst-case decimal expansion is produced by
// the largest subnormal value; if a buffer were too small the intermediate
// expansion would be silently truncated and the shortest representation would
// no longer round-trip.
func TestDecimalBufferExtremes(t *testing.T) {
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
