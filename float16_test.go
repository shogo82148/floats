package floats

import "testing"

func TestFloat16_IsNaN(t *testing.T) {
	tests := []struct {
		name string
		a    Float16
		want bool
	}{
		{
			name: "NaN",
			a:    0x7e00,
			want: true,
		},
		{
			name: "not NaN",
			a:    0x7c00,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.IsNaN(); got != tt.want {
				t.Errorf("Float16.IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}
