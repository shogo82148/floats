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

func ExampleFloat16_Acos() {
	f := floats.NewFloat16(1)
	x := f.Acos()
	fmt.Printf("%.1f\n", x)

	// Output:
	// 0.0
}

func ExampleFloat32_Acos() {
	f := floats.NewFloat32(1)
	x := f.Acos()
	fmt.Printf("%.1f\n", x)

	// Output:
	// 0.0
}

func ExampleFloat64_Acos() {
	f := floats.NewFloat64(1)
	x := f.Acos()
	fmt.Printf("%.1f\n", x)

	// Output:
	// 0.0
}

func ExampleFloat128_Acos() {
	f := floats.NewFloat128(1)
	x := f.Acos()
	fmt.Printf("%.1f\n", x)

	// Output:
	// 0.0
}

func ExampleFloat256_Acos() {
	f := floats.NewFloat256(1)
	x := f.Acos()
	fmt.Printf("%.1f\n", x)

	// Output:
	// 0.0
}
