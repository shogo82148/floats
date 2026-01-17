package floats

import (
	"math"
	"testing"
)

func TestFloat128_Exp(t *testing.T) {
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(-1), "0.3678794411714423215955237701614608674458111310317678345078368016974614957449"},
		{exact128(0), "1"},
		{exact128(1), "2.71828182845904523536028747135266249775724709369995957496696762772407663035355"},
		{exact128(2), "7.38905609893065022723042746057500781318031557055184732408712782252257379607906"},
		{exact128(3), "20.085536923187667740928529654581717896987907838554150144378934229698845878092"},
		{exact128(4), "54.5981500331442390781102612028608784027907370386140687258265939585536620999359"},
	}

	for _, tt := range tests {
		got := tt.x.Exp()
		if !close128(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(math.Inf(1)), exact128(math.Inf(1))},
		{exact128(math.Inf(-1)), exact128(0)},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Exp()
		if !eq128(got, tt.want) {
			t.Errorf("Exp(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}
