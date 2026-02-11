package floats_test

import (
	"fmt"

	"github.com/shogo82148/floats"
)

func ExampleFloat16_Abs() {
	f := floats.NewFloat16(-2)
	x := f.Abs()
	fmt.Printf("%.1f\n", x)

	f = floats.NewFloat16(2)
	y := f.Abs()
	fmt.Printf("%.1f\n", y)

	// Output:
	// 2.0
	// 2.0
}

func ExampleFloat32_Abs() {
	f := floats.NewFloat32(-2)
	x := f.Abs()
	fmt.Printf("%.1f\n", x)

	f = floats.NewFloat32(2)
	y := f.Abs()
	fmt.Printf("%.1f\n", y)

	// Output:
	// 2.0
	// 2.0
}

func ExampleFloat64_Abs() {
	f := floats.NewFloat64(-2)
	x := f.Abs()
	fmt.Printf("%.1f\n", x)

	f = floats.NewFloat64(2)
	y := f.Abs()
	fmt.Printf("%.1f\n", y)

	// Output:
	// 2.0
	// 2.0
}

func ExampleFloat128_Abs() {
	f := floats.NewFloat128(-2)
	x := f.Abs()
	fmt.Printf("%.1f\n", x)

	f = floats.NewFloat128(2)
	y := f.Abs()
	fmt.Printf("%.1f\n", y)

	// Output:
	// 2.0
	// 2.0
}

func ExampleFloat256_Abs() {
	f := floats.NewFloat256(-2)
	x := f.Abs()
	fmt.Printf("%.1f\n", x)

	f = floats.NewFloat256(2)
	y := f.Abs()
	fmt.Printf("%.1f\n", y)

	// Output:
	// 2.0
	// 2.0
}
