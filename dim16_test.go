package floats

import (
	"math"
	"testing"
)

func TestFloat16_Dim(t *testing.T) {
	tests := []struct {
		a    Float16
		b    Float16
		want Float16
	}{
		{exact16(5.0), exact16(3.0), exact16(2.0)},
		{exact16(3.0), exact16(5.0), exact16(0.0)},
		{exact16(3.0), exact16(3.0), exact16(0.0)},

		// Special cases
		{exact16(math.Inf(1)), exact16(math.Inf(1)), exact16(math.NaN())},
		{exact16(math.Inf(-1)), exact16(math.Inf(-1)), exact16(math.NaN())},
		{exact16(math.NaN()), exact16(1.0), exact16(math.NaN())},
		{exact16(1.0), exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.a.Dim(tt.b)
		if !eq16(got, tt.want) {
			t.Errorf("Float16.Dim(%v, %v) = %v; want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
