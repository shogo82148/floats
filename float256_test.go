package floats

import "testing"

func TestFloat256_IsNaN(t *testing.T) {
	tests := []struct {
		name string
		a    Float256
		want bool
	}{
		{
			name: "NaN",
			a: Float256{
				0x7fff_f800_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: true,
		},
		{
			name: "infinity",
			a: Float256{
				0x7fff_f000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
				0x0000_0000_0000_0000,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.IsNaN(); got != tt.want {
				t.Errorf("Float256.IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}
