package floats

import (
	"math"
	"testing"
)

func TestFloat32_IsNaN(t *testing.T) {
	tests := []struct {
		name string
		a    Float32
		want bool
	}{
		{
			name: "NaN",
			a:    Float32(math.NaN()),
			want: true,
		},
		{
			name: "Inf",
			a:    Float32(math.Inf(1)),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.IsNaN(); got != tt.want {
				t.Errorf("Float32.IsNaN() = %v, want %v", got, tt.want)
			}
		})
	}
}
