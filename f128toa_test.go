package floats

import (
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
	}

	for _, tt := range tests {
		if got := tt.x.Text(tt.fmt, tt.prec); got != tt.s {
			t.Errorf("Float128(%x).Text(%q, %d) = %q, want %q", ints.Uint128(tt.x), tt.fmt, tt.prec, got, tt.s)
		}
	}
}
