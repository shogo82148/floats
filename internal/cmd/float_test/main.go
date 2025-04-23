package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"

	"github.com/shogo82148/floats"
)

func main() {
	switch os.Args[1] {
	case "f32_to_f64":
		if err := f32_to_f64(); err != nil {
			log.Fatal(err)
		}
	}
}

func f32_to_f64() error {
	for {
		var s32, s64, flag string
		if _, err := fmt.Scanf("%s %s %s", &s32, &s64, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			os.Exit(1)
		}

		f32, err := parseFloat32(s32)
		if err != nil {
			return err
		}

		f64, err := parseFloat64(s64)
		if err != nil {
			return err
		}

		got := f32.Float64()
		if cmp.Compare(got, f64) != 0 {
			log.Printf("f32: %s, f64: %s", s32, s64)
			log.Printf("got: %x, want: %x", got, f64)
			return fmt.Errorf("f32(%x).Float64() = %x, want %x", s32, got, f64)
		}
	}
	return nil
}

func parseFloat32(s string) (floats.Float32, error) {
	bits, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return 0, err
	}
	return floats.Float32(math.Float32frombits(uint32(bits))), nil
}

func parseFloat64(s string) (floats.Float64, error) {
	bits, err := strconv.ParseUint(s, 16, 64)
	if err != nil {
		return 0, err
	}
	return floats.Float64(math.Float64frombits(bits)), nil
}
