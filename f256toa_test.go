package floats

import (
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
	}

	for _, tt := range tests {
		if got := tt.x.Text(tt.fmt, tt.prec); got != tt.s {
			t.Errorf("Float256(%x).Text(%q, %d) = %q, want %q", ints.Uint256(tt.x), tt.fmt, tt.prec, got, tt.s)
		}
	}
}
