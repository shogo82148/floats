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
	"sync/atomic"
	"time"

	"github.com/shogo82148/floats"
)

var count atomic.Int64

func showProgress() {
	start := time.Now()
	for {
		time.Sleep(3 * time.Second)
		log.Printf("%s: %d", time.Since(start), count.Load())
	}
}

func main() {
	go showProgress()

	switch os.Args[1] {
	case "f16_to_f32":
		if err := f16_to_f32(); err != nil {
			log.Fatal(err)
		}
	case "f32_to_f64":
		if err := f32_to_f64(); err != nil {
			log.Fatal(err)
		}
	case "f64_to_f32":
		if err := f64_to_f32(); err != nil {
			log.Fatal(err)
		}
	}
}

func f16_to_f32() error {
	for {
		var s16, s32, flag string
		if _, err := fmt.Scanf("%s %s %s", &s16, &s32, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f16, err := parseFloat16(s16)
		if err != nil {
			return err
		}

		f32, err := parseFloat32(s32)
		if err != nil {
			return err
		}

		got := f16.Float32()
		if cmp.Compare(got, f32) != 0 {
			log.Printf("f16: %s, f32: %s", s16, s32)
			log.Printf("got: %x, want: %x", got, f32)
			return fmt.Errorf("f16(%x).Float32() = %x, want %x", f16, got, f32)
		}
		count.Add(1)
	}
	return nil
}

func f32_to_f64() error {
	for {
		var s32, s64, flag string
		if _, err := fmt.Scanf("%s %s %s", &s32, &s64, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
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
			return fmt.Errorf("f32(%x).Float64() = %x, want %x", f32, got, f64)
		}
		count.Add(1)
	}
	return nil
}

func f64_to_f32() error {
	for {
		var s64, s32, flag string
		if _, err := fmt.Scanf("%s %s %s", &s64, &s32, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f64, err := parseFloat64(s64)
		if err != nil {
			return err
		}

		f32, err := parseFloat32(s32)
		if err != nil {
			return err
		}

		got := f64.Float32()
		if cmp.Compare(got, f32) != 0 {
			log.Printf("f64: %s, f32: %s", s64, s32)
			log.Printf("got: %x, want: %x", got, f32)
			return fmt.Errorf("f64(%x).Float32() = %x, want %x", f64, got, f32)
		}
		count.Add(1)
	}
	return nil
}

func parseFloat16(s string) (floats.Float16, error) {
	bits, err := strconv.ParseUint(s, 16, 16)
	if err != nil {
		return 0, err
	}
	return floats.Float16(bits), nil
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
