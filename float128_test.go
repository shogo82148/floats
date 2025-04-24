package floats

import "testing"

func TestFloat128_IsNaN(t *testing.T) {
	tests := []struct {
		name string
		a    Float128
		want bool
	}{
		{
			name: "NaN",
			a:    Float128{0x7fff_8000_0000_0000, 0x0000_0000_0000_0000},
			want: true,
		},
		{
			name: "infinity",
			a:    Float128{0x7fff_0000_0000_0000, 0x0000_0000_0000_0000},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.IsNaN(); got != tt.want {
				t.Errorf("Float128.IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}
