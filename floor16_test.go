package floats

import (
	"math"
	"testing"
)

func TestFloat16_Floor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want Float16
	}{
		{exact16(3.0), exact16(3)},
		{exact16(3.5), exact16(3)},
		{exact16(-3.0), exact16(-3)},
		{exact16(-3.5), exact16(-4)},

		// Special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(math.Inf(-1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Floor()
		if !eq16(got, tt.want) {
			t.Errorf("Float16.Floor(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat16_Ceil(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want Float16
	}{
		{exact16(3.0), exact16(3)},
		{exact16(3.5), exact16(4)},
		{exact16(-3.0), exact16(-3)},
		{exact16(-3.5), exact16(-3)},

		// Special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(math.Inf(-1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Ceil()
		if !eq16(got, tt.want) {
			t.Errorf("Float16.Ceil(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat16_Trunc(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want Float16
	}{
		{exact16(3.0), exact16(3)},
		{exact16(3.5), exact16(3)},
		{exact16(-3.0), exact16(-3)},
		{exact16(-3.5), exact16(-3)},

		// Special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(math.Inf(-1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Trunc()
		if !eq16(got, tt.want) {
			t.Errorf("Float16.Trunc(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat16_Round(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want Float16
	}{
		{exact16(3.0), exact16(3)},
		{exact16(3.5), exact16(4)},
		{exact16(4.5), exact16(5)},
		{exact16(-3.0), exact16(-3)},
		{exact16(-3.5), exact16(-4)},
		{exact16(-4.5), exact16(-5)},

		// Special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(math.Inf(-1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.Round()
		if !eq16(got, tt.want) {
			t.Errorf("Float16.Round(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat16_RoundToEven(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float16
		want Float16
	}{
		{exact16(3.0), exact16(3)},
		{exact16(3.5), exact16(4)},
		{exact16(4.5), exact16(4)},
		{exact16(-3.0), exact16(-3)},
		{exact16(-3.5), exact16(-4)},
		{exact16(-4.5), exact16(-4)},

		// Special cases
		{exact16(0), exact16(0)},
		{exact16(math.Copysign(0, -1)), exact16(math.Copysign(0, -1))},
		{exact16(math.Inf(1)), exact16(math.Inf(1))},
		{exact16(math.Inf(-1)), exact16(math.Inf(-1))},
		{exact16(math.NaN()), exact16(math.NaN())},
	}

	for _, tt := range tests {
		got := tt.x.RoundToEven()
		if !eq16(got, tt.want) {
			t.Errorf("Float16.RoundToEven(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat16_Floor_All(t *testing.T) {
	for i := 0; i <= 0xffff; i++ {
		a := Float16(i)
		got := a.Floor()
		want := NewFloat16(math.Floor(a.Float64().BuiltIn()))
		if !eq16(got, want) {
			t.Errorf("Floor(%#04x) = %#04x, want %#04x", i, uint16(got), uint16(want))
		}
	}
}

func BenchmarkFloat16_Floor(b *testing.B) {
	var i uint16
	for b.Loop() {
		Float16(i).Floor()
		i++
	}
}

func TestFloat16_Ceil_All(t *testing.T) {
	for i := 0; i <= 0xffff; i++ {
		a := Float16(i)
		got := a.Ceil()
		want := NewFloat16(math.Ceil(a.Float64().BuiltIn()))
		if !eq16(got, want) {
			t.Errorf("Ceil(%#04x) = %#04x, want %#04x", i, uint16(got), uint16(want))
		}
	}
}

func BenchmarkFloat16_Ceil(b *testing.B) {
	var i uint16
	for b.Loop() {
		Float16(i).Ceil()
		i++
	}
}

func TestFloat16_Trunc_All(t *testing.T) {
	for i := 0; i <= 0xffff; i++ {
		a := Float16(i)
		got := a.Trunc()
		want := NewFloat16(math.Trunc(a.Float64().BuiltIn()))
		if !eq16(got, want) {
			t.Errorf("Trunc(%#04x) = %#04x, want %#04x", i, uint16(got), uint16(want))
		}
	}
}

func BenchmarkFloat16_Trunc(b *testing.B) {
	var i uint16
	for b.Loop() {
		Float16(i).Trunc()
		i++
	}
}

func TestFloat16_Round_All(t *testing.T) {
	for i := 0; i <= 0xffff; i++ {
		a := Float16(i)
		got := a.Round()
		want := NewFloat16(math.Round(a.Float64().BuiltIn()))
		if !eq16(got, want) {
			t.Errorf("Round(%#04x) = %#04x, want %#04x", i, uint16(got), uint16(want))
		}
	}
}

func BenchmarkFloat16_Round(b *testing.B) {
	var i uint16
	for b.Loop() {
		Float16(i).Round()
		i++
	}
}

func TestFloat16_RoundToEven_All(t *testing.T) {
	for i := 0; i <= 0xffff; i++ {
		a := Float16(i)
		got := a.RoundToEven()
		want := NewFloat16(math.RoundToEven(a.Float64().BuiltIn()))
		if !eq16(got, want) {
			t.Errorf("RoundToEven(%#04x) = %#04x, want %#04x", i, uint16(got), uint16(want))
		}
	}
}

func BenchmarkFloat16_RoundToEven(b *testing.B) {
	var i uint16
	for b.Loop() {
		Float16(i).RoundToEven()
		i++
	}
}
