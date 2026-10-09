// Command bf16dump writes the results of the mathematical functions of BFloat16
// for all the 65536 inputs, one file per function.
// Each line of a file contains the bits of the result in hexadecimal.
//
// It is used by scripts/check_bfloat16_math.py, which compares them with mpmath.
//
// Usage: go run ./internal/cmd/bf16dump DIR
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/shogo82148/floats"
)

type bf16 = floats.BFloat16

// secondArgs are the values of the second argument of Pow, Atan2 and HypotBF16.
// scripts/check_bfloat16_math.py has the same list.
var secondArgs = []float64{
	-3, -2, -1, -0.5, 0.5, 1.0 / 3, 0.25, 1.5, 2, 3, 5, 10, -10, 0.1, 100, 1000, 0.001, 2.5, 7, -0.7,
	0, 1, -1, 0.75, 1e5, 1e-5, 3e38, 1e-40,
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: bf16dump DIR")
		os.Exit(2)
	}
	dir := os.Args[1]
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	unary := map[string]func(bf16) bf16{
		"Exp": bf16.Exp, "Exp2": bf16.Exp2, "Expm1": bf16.Expm1,
		"Log": bf16.Log, "Log2": bf16.Log2, "Log10": bf16.Log10, "Log1p": bf16.Log1p,
		"Sin": bf16.Sin, "Cos": bf16.Cos, "Tan": bf16.Tan,
		"Asin": bf16.Asin, "Acos": bf16.Acos, "Atan": bf16.Atan,
		"Sinh": bf16.Sinh, "Cosh": bf16.Cosh, "Tanh": bf16.Tanh,
		"Asinh": bf16.Asinh, "Acosh": bf16.Acosh, "Atanh": bf16.Atanh,
		"Cbrt": bf16.Cbrt, "Gamma": bf16.Gamma,
		"Erf": bf16.Erf, "Erfc": bf16.Erfc, "Erfinv": bf16.Erfinv, "Erfcinv": bf16.Erfcinv,
		"J0": bf16.J0, "J1": bf16.J1, "Y0": bf16.Y0, "Y1": bf16.Y1,
		"Jn2":    func(a bf16) bf16 { return a.Jn(2) },
		"Jn5":    func(a bf16) bf16 { return a.Jn(5) },
		"Yn2":    func(a bf16) bf16 { return a.Yn(2) },
		"Yn5":    func(a bf16) bf16 { return a.Yn(5) },
		"Jn1":    func(a bf16) bf16 { return a.Jn(1) },
		"Jn-1":   func(a bf16) bf16 { return a.Jn(-1) },
		"Lgamma": func(a bf16) bf16 { r, _ := a.Lgamma(); return r },
		// the sign of Gamma: 0x0001 for +1 and 0xffff for -1
		"LgammaSign": func(a bf16) bf16 { _, s := a.Lgamma(); return bf16(uint16(int16(s))) },
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	failed := false
	sem := make(chan struct{}, 8)
	write := func(name string, f func(i int) bf16) {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			buf := make([]byte, 0, 5*65536)
			for i := range 1 << 16 {
				buf = fmt.Appendf(buf, "%04x\n", uint16(f(i)))
			}
			if err := os.WriteFile(filepath.Join(dir, name+".txt"), buf, 0o644); err != nil {
				mu.Lock()
				defer mu.Unlock()
				fmt.Fprintln(os.Stderr, err)
				failed = true
			}
		}()
	}
	for name, f := range unary {
		write(name, func(i int) bf16 { return f(bf16(i)) })
	}
	for k, y := range secondArgs {
		b := floats.NewBFloat16(y)
		write(fmt.Sprintf("Pow%d", k), func(i int) bf16 { return bf16(i).Pow(b) })
		write(fmt.Sprintf("Atan2_%d", k), func(i int) bf16 { return bf16(i).Atan2(b) })
		write(fmt.Sprintf("Hypot%d", k), func(i int) bf16 { return floats.HypotBF16(bf16(i), b) })
	}
	wg.Wait()
	if failed {
		os.Exit(1)
	}
}
