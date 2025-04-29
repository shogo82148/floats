package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/shogo82148/floats"
)

const (
	inexact   = 0x01
	underflow = 0x02
	overflow  = 0x04
	infinite  = 0x08
	invalid   = 0x10
)

var count atomic.Int64

func showProgress() {
	start := time.Now()
	ticker := time.NewTicker(3 * time.Second)
	for range ticker.C {
		log.Printf("%s: %d", time.Since(start), count.Load())
	}
}

func main() {
	start := time.Now()
	go showProgress()

	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <test-name>", filepath.Base(os.Args[0]))
	}

	switch os.Args[1] {
	case "f16_to_f32":
		if err := f16_to_f32(); err != nil {
			log.Fatal(err)
		}
	case "f16_to_f64":
		if err := f16_to_f64(); err != nil {
			log.Fatal(err)
		}
	case "f16_to_f128":
		if err := f16_to_f128(); err != nil {
			log.Fatal(err)
		}
	case "f32_to_f16":
		if err := f32_to_f16(); err != nil {
			log.Fatal(err)
		}
	case "f32_to_f64":
		if err := f32_to_f64(); err != nil {
			log.Fatal(err)
		}
	case "f32_to_f128":
		if err := f32_to_f128(); err != nil {
			log.Fatal(err)
		}
	case "f64_to_f16":
		if err := f64_to_f16(); err != nil {
			log.Fatal(err)
		}
	case "f64_to_f32":
		if err := f64_to_f32(); err != nil {
			log.Fatal(err)
		}
	case "f64_to_f128":
		if err := f64_to_f128(); err != nil {
			log.Fatal(err)
		}
	case "f128_to_f16":
		if err := f128_to_f16(); err != nil {
			log.Fatal(err)
		}
	case "f128_to_f32":
		if err := f128_to_f32(); err != nil {
			log.Fatal(err)
		}
	case "f128_to_f64":
		if err := f128_to_f64(); err != nil {
			log.Fatal(err)
		}

	case "f16_to_i64":
		if err := f16_to_i64(); err != nil {
			log.Fatal(err)
		}
	case "f32_to_i64":
		if err := f32_to_i64(); err != nil {
			log.Fatal(err)
		}
	case "f64_to_i64":
		if err := f64_to_i64(); err != nil {
			log.Fatal(err)
		}
	case "f128_to_i64":
		if err := f128_to_i64(); err != nil {
			log.Fatal(err)
		}

	case "f16_mul":
		if err := f16_mul(); err != nil {
			log.Fatal(err)
		}
	case "f16_div":
		if err := f16_div(); err != nil {
			log.Fatal(err)
		}

	case "f32_mul":
		if err := f32_mul(); err != nil {
			log.Fatal(err)
		}
	case "f32_div":
		if err := f32_div(); err != nil {
			log.Fatal(err)
		}
	case "f32_add":
		if err := f32_add(); err != nil {
			log.Fatal(err)
		}

	case "f64_mul":
		if err := f64_mul(); err != nil {
			log.Fatal(err)
		}
	case "f64_div":
		if err := f64_div(); err != nil {
			log.Fatal(err)
		}

	case "f128_mul":
		if err := f128_mul(); err != nil {
			log.Fatal(err)
		}
	case "f128_div":
		if err := f128_div(); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown test name: %q", os.Args[1])
	}
	log.Printf("%s: %d", time.Since(start), count.Load())
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
		if !eq32(got, f32) {
			log.Printf("f16: %s, f32: %s", s16, s32)
			log.Printf("got: %x, want: %x", got, f32)
			return fmt.Errorf("f16(%x).Float32() = %x, want %x", f16, got, f32)
		}
		count.Add(1)
	}
	return nil
}

func f16_to_f64() error {
	for {
		var s16, s64, flag string
		if _, err := fmt.Scanf("%s %s %s", &s16, &s64, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f16, err := parseFloat16(s16)
		if err != nil {
			return err
		}

		f64, err := parseFloat64(s64)
		if err != nil {
			return err
		}

		got := f16.Float64()
		if !eq64(got, f64) {
			log.Printf("f16: %s, f64: %s", s16, s64)
			log.Printf("got: %x, want: %x", got, f64)
			return fmt.Errorf("f16(%x).Float64() = %x, want %x", f16, got, f64)
		}
		count.Add(1)
	}
	return nil
}

func f16_to_f128() error {
	for {
		var s16, s128, flag string
		if _, err := fmt.Scanf("%s %s %s", &s16, &s128, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f16, err := parseFloat16(s16)
		if err != nil {
			return err
		}

		f128, err := parseFloat128(s128)
		if err != nil {
			return err
		}

		got := f16.Float128()
		if !eq128(got, f128) {
			log.Printf("f16: %s, f64: %s", s16, s128)
			log.Printf("got: %x, want: %x", got, f128)
			return fmt.Errorf("f16(%x).Float128() = %x, want %x", f16, got, f128)
		}
		count.Add(1)
	}
	return nil
}

func f32_to_f16() error {
	for {
		var s32, s16, flag string
		if _, err := fmt.Scanf("%s %s %s", &s32, &s16, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f32, err := parseFloat32(s32)
		if err != nil {
			return err
		}

		f16, err := parseFloat16(s16)
		if err != nil {
			return err
		}

		got := f32.Float16()
		if !eq16(got, f16) {
			log.Printf("f32: %s, f16: %s", s32, s16)
			log.Printf("got: %x, want: %x", got, f16)
			return fmt.Errorf("f32(%x).Float16() = %x, want %x", f32, got, f16)
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
		if !eq64(got, f64) {
			log.Printf("f32: %s, f64: %s", s32, s64)
			log.Printf("got: %x, want: %x", got, f64)
			return fmt.Errorf("f32(%x).Float64() = %x, want %x", f32, got, f64)
		}
		count.Add(1)
	}
	return nil
}

func f32_to_f128() error {
	for {
		var s32, s128, flag string
		if _, err := fmt.Scanf("%s %s %s", &s32, &s128, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f32, err := parseFloat32(s32)
		if err != nil {
			return err
		}

		f128, err := parseFloat128(s128)
		if err != nil {
			return err
		}

		got := f32.Float128()
		if !eq128(got, f128) {
			log.Printf("f32: %s, f128: %s", s32, s128)
			log.Printf("got: %x, want: %x", got, f128)
			return fmt.Errorf("f32(%x).Float128() = %x, want %x", f32, got, f128)
		}
		count.Add(1)
	}
	return nil
}

func f64_to_f16() error {
	for {
		var s64, s16, flag string
		if _, err := fmt.Scanf("%s %s %s", &s64, &s16, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f64, err := parseFloat64(s64)
		if err != nil {
			return err
		}

		f16, err := parseFloat16(s16)
		if err != nil {
			return err
		}

		got := f64.Float16()
		if !eq16(got, f16) {
			log.Printf("f64: %s, f16: %s", s64, s16)
			log.Printf("got: %x, want: %x", got, f16)
			return fmt.Errorf("f64(%x).Float16() = %x, want %x", f64, got, f16)
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
		if !eq32(got, f32) {
			log.Printf("f64: %s, f32: %s", s64, s32)
			log.Printf("got: %x, want: %x", got, f32)
			return fmt.Errorf("f64(%x).Float32() = %x, want %x", f64, got, f32)
		}
		count.Add(1)
	}
	return nil
}

func f64_to_f128() error {
	for {
		var s64, s128, flag string
		if _, err := fmt.Scanf("%s %s %s", &s64, &s128, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f64, err := parseFloat64(s64)
		if err != nil {
			return err
		}

		f128, err := parseFloat128(s128)
		if err != nil {
			return err
		}

		got := f64.Float128()
		if !eq128(got, f128) {
			log.Printf("f64: %s, f128: %s", s64, s128)
			log.Printf("got: %x, want: %x", got, f128)
			return fmt.Errorf("f64(%x).Float128() = %x, want %x", f64, got, f128)
		}
		count.Add(1)
	}
	return nil
}

func f128_to_f16() error {
	for {
		var s128, s16, flag string
		if _, err := fmt.Scanf("%s %s %s", &s128, &s16, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f128, err := parseFloat128(s128)
		if err != nil {
			return err
		}

		f16, err := parseFloat16(s16)
		if err != nil {
			return err
		}

		got := f128.Float16()
		if !eq16(got, f16) {
			log.Printf("f128: %s, f16: %s", s128, s16)
			log.Printf("got: %x, want: %x", got, f16)
			return fmt.Errorf("f128(%x).Float16() = %x, want %x", f128, got, f16)
		}
		count.Add(1)
	}
	return nil
}

func f128_to_f32() error {
	for {
		var s128, s32, flag string
		if _, err := fmt.Scanf("%s %s %s", &s128, &s32, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f128, err := parseFloat128(s128)
		if err != nil {
			return err
		}

		f32, err := parseFloat32(s32)
		if err != nil {
			return err
		}

		got := f128.Float32()
		if !eq32(got, f32) {
			log.Printf("f128: %s, f32: %s", s128, s32)
			log.Printf("got: %x, want: %x", got, f32)
			return fmt.Errorf("f128(%x).Float32() = %x, want %x", f128, got, f32)
		}
		count.Add(1)
	}
	return nil
}

func f128_to_f64() error {
	for {
		var s128, s64, flag string
		if _, err := fmt.Scanf("%s %s %s", &s128, &s64, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f128, err := parseFloat128(s128)
		if err != nil {
			return err
		}

		f64, err := parseFloat64(s64)
		if err != nil {
			return err
		}

		got := f128.Float64()
		if !eq64(got, f64) {
			log.Printf("f128: %s, f64: %s", s128, s64)
			log.Printf("got: %x, want: %x", got, f64)
			return fmt.Errorf("f128(%x).Float64() = %x, want %x", f128, got, f64)
		}
		count.Add(1)
	}
	return nil
}

func f16_to_i64() error {
	for {
		var s16, i64, flag string
		if _, err := fmt.Scanf("%s %s %s", &s16, &i64, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f, err := parseFlag(flag)
		if err != nil {
			return err
		}
		if f&invalid != 0 {
			// The behavior when the conversion is invalid is undefined.
			continue
		}

		f16, err := parseFloat16(s16)
		if err != nil {
			return err
		}

		u64, err := strconv.ParseUint(i64, 16, 64)
		if err != nil {
			return err
		}
		i64v := int64(u64)

		got := f16.Int64()
		if got != i64v {
			log.Printf("f16: %s, i64: %s", s16, i64)
			log.Printf("got: %x, want: %x", got, i64v)
			return fmt.Errorf("f16(%x).Int64() = %x, want %x", f16, got, i64v)
		}
		count.Add(1)
	}
	return nil
}

func f32_to_i64() error {
	for {
		var s32, i64, flag string
		if _, err := fmt.Scanf("%s %s %s", &s32, &i64, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f, err := parseFlag(flag)
		if err != nil {
			return err
		}
		if f&invalid != 0 {
			// The behavior when the conversion is invalid is undefined.
			continue
		}

		f32, err := parseFloat32(s32)
		if err != nil {
			return err
		}

		u64, err := strconv.ParseUint(i64, 16, 64)
		if err != nil {
			return err
		}
		i64v := int64(u64)

		got := f32.Int64()
		if got != i64v {
			log.Printf("f32: %s, i64: %s", s32, i64)
			log.Printf("got: %x, want: %x", got, i64v)
			return fmt.Errorf("f32(%x).Int64() = %x, want %x", f32, got, i64v)
		}
		count.Add(1)
	}
	return nil
}

func f64_to_i64() error {
	for {
		var s64, i64, flag string
		if _, err := fmt.Scanf("%s %s %s", &s64, &i64, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f, err := parseFlag(flag)
		if err != nil {
			return err
		}
		if f&invalid != 0 {
			// The behavior when the conversion is invalid is undefined.
			continue
		}

		f64, err := parseFloat64(s64)
		if err != nil {
			return err
		}

		u64, err := strconv.ParseUint(i64, 16, 64)
		if err != nil {
			return err
		}
		i64v := int64(u64)

		got := f64.Int64()
		if got != i64v {
			log.Printf("f64: %s, i64: %s", s64, i64)
			log.Printf("got: %x, want: %x", got, i64v)
			return fmt.Errorf("f64(%x).Int64() = %x, want %x", f64, got, i64v)
		}
		count.Add(1)
	}
	return nil
}

func f128_to_i64() error {
	for {
		var s128, i64, flag string
		if _, err := fmt.Scanf("%s %s %s", &s128, &i64, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f, err := parseFlag(flag)
		if err != nil {
			return err
		}
		if f&invalid != 0 {
			// The behavior when the conversion is invalid is undefined.
			continue
		}

		f128, err := parseFloat128(s128)
		if err != nil {
			return err
		}

		u64, err := strconv.ParseUint(i64, 16, 64)
		if err != nil {
			return err
		}
		i64v := int64(u64)

		got := f128.Int64()
		if got != i64v {
			log.Printf("f128: %s, i64: %s", s128, i64)
			log.Printf("got: %x, want: %x", got, i64v)
			return fmt.Errorf("f128(%x).Int64() = %x, want %x", f128, got, i64v)
		}
		count.Add(1)
	}
	return nil
}

func f16_mul() error {
	for {
		var a, b, want, flag string
		if _, err := fmt.Scanf("%s %s %s %s", &a, &b, &want, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f16a, err := parseFloat16(a)
		if err != nil {
			return err
		}
		f16b, err := parseFloat16(b)
		if err != nil {
			return err
		}
		wantf, err := parseFloat16(want)
		if err != nil {
			return err
		}
		got := f16a.Mul(f16b)
		if !eq16(got, wantf) {
			log.Printf("a: %s, b: %s, want: %s", a, b, want)
			log.Printf("got: %x, want: %x", got, wantf)
			return fmt.Errorf("Float16(%x).Mul(%x) = %x, want %x", f16a, f16b, got, wantf)
		}
		count.Add(1)
	}
	return nil
}

func f16_div() error {
	for {
		var a, b, want, flag string
		if _, err := fmt.Scanf("%s %s %s %s", &a, &b, &want, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f16a, err := parseFloat16(a)
		if err != nil {
			return err
		}
		f16b, err := parseFloat16(b)
		if err != nil {
			return err
		}
		wantf, err := parseFloat16(want)
		if err != nil {
			return err
		}
		got := f16a.Quo(f16b)
		if !eq16(got, wantf) {
			log.Printf("a: %s, b: %s, want: %s", a, b, want)
			log.Printf("got: %x, want: %x", got, wantf)
			return fmt.Errorf("Float16(%x).Div(%x) = %x, want %x", f16a, f16b, got, wantf)
		}
		count.Add(1)
	}
	return nil
}

func f32_mul() error {
	for {
		var a, b, want, flag string
		if _, err := fmt.Scanf("%s %s %s %s", &a, &b, &want, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f32a, err := parseFloat32(a)
		if err != nil {
			return err
		}
		f32b, err := parseFloat32(b)
		if err != nil {
			return err
		}
		wantf, err := parseFloat32(want)
		if err != nil {
			return err
		}
		got := f32a.Mul(f32b)
		if !eq32(got, wantf) {
			log.Printf("a: %s, b: %s, want: %s", a, b, want)
			log.Printf("got: %x, want: %x", got, wantf)
			return fmt.Errorf("Float32(%x).Mul(%x) = %x, want %x", f32a, f32b, got, wantf)
		}
		count.Add(1)
	}
	return nil
}

func f32_div() error {
	for {
		var a, b, want, flag string
		if _, err := fmt.Scanf("%s %s %s %s", &a, &b, &want, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f32a, err := parseFloat32(a)
		if err != nil {
			return err
		}
		f32b, err := parseFloat32(b)
		if err != nil {
			return err
		}
		wantf, err := parseFloat32(want)
		if err != nil {
			return err
		}
		got := f32a.Quo(f32b)
		if !eq32(got, wantf) {
			log.Printf("a: %s, b: %s, want: %s", a, b, want)
			log.Printf("got: %x, want: %x", got, wantf)
			return fmt.Errorf("Float32(%x).Quo(%x) = %x, want %x", f32a, f32b, got, wantf)
		}
		count.Add(1)
	}
	return nil
}

func f32_add() error {
	for {
		var a, b, want, flag string
		if _, err := fmt.Scanf("%s %s %s %s", &a, &b, &want, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f32a, err := parseFloat32(a)
		if err != nil {
			return err
		}
		f32b, err := parseFloat32(b)
		if err != nil {
			return err
		}
		wantf, err := parseFloat32(want)
		if err != nil {
			return err
		}
		got := f32a.Add(f32b)
		if !eq32(got, wantf) {
			log.Printf("a: %s, b: %s, want: %s", a, b, want)
			log.Printf("got: %x, want: %x", got, wantf)
			return fmt.Errorf("Float32(%x).Add(%x) = %x, want %x", f32a, f32b, got, wantf)
		}
		count.Add(1)
	}
	return nil
}

func f64_mul() error {
	for {
		var a, b, want, flag string
		if _, err := fmt.Scanf("%s %s %s %s", &a, &b, &want, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f64a, err := parseFloat64(a)
		if err != nil {
			return err
		}
		f64b, err := parseFloat64(b)
		if err != nil {
			return err
		}
		wantf, err := parseFloat64(want)
		if err != nil {
			return err
		}
		got := f64a.Mul(f64b)
		if !eq64(got, wantf) {
			log.Printf("a: %s, b: %s, want: %s", a, b, want)
			log.Printf("got: %x, want: %x", got, wantf)
			return fmt.Errorf("Float64(%x).Mul(%x) = %x, want %x", f64a, f64b, got, wantf)
		}
		count.Add(1)
	}
	return nil
}

func f64_div() error {
	for {
		var a, b, want, flag string
		if _, err := fmt.Scanf("%s %s %s %s", &a, &b, &want, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f64a, err := parseFloat64(a)
		if err != nil {
			return err
		}
		f64b, err := parseFloat64(b)
		if err != nil {
			return err
		}
		wantf, err := parseFloat64(want)
		if err != nil {
			return err
		}
		got := f64a.Quo(f64b)
		if !eq64(got, wantf) {
			log.Printf("a: %s, b: %s, want: %s", a, b, want)
			log.Printf("got: %x, want: %x", got, wantf)
			return fmt.Errorf("Float64(%x).Quo(%x) = %x, want %x", f64a, f64b, got, wantf)
		}
		count.Add(1)
	}
	return nil
}

func f128_mul() error {
	for {
		var a, b, want, flag string
		if _, err := fmt.Scanf("%s %s %s %s", &a, &b, &want, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f128a, err := parseFloat128(a)
		if err != nil {
			return err
		}
		f128b, err := parseFloat128(b)
		if err != nil {
			return err
		}
		wantf, err := parseFloat128(want)
		if err != nil {
			return err
		}
		got := f128a.Mul(f128b)
		if !eq128(got, wantf) {
			log.Printf("a: %s, b: %s, want: %s", a, b, want)
			log.Printf("got: %x, want: %x", got, wantf)
			return fmt.Errorf("Float128(%x).Mul(%x) = %x, want %x", f128a, f128b, got, wantf)
		}
		count.Add(1)
	}
	return nil
}

func f128_div() error {
	for {
		var a, b, want, flag string
		if _, err := fmt.Scanf("%s %s %s %s", &a, &b, &want, &flag); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		f128a, err := parseFloat128(a)
		if err != nil {
			return err
		}
		f128b, err := parseFloat128(b)
		if err != nil {
			return err
		}
		wantf, err := parseFloat128(want)
		if err != nil {
			return err
		}
		got := f128a.Quo(f128b)
		if !eq128(got, wantf) {
			log.Printf("a: %s, b: %s, want: %s", a, b, want)
			log.Printf("got: %x, want: %x", got, wantf)
			return fmt.Errorf("Float128(%x).Quo(%x) = %x, want %x", f128a, f128b, got, wantf)
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

func parseFloat128(s string) (floats.Float128, error) {
	a0, err := strconv.ParseUint(s[:16], 16, 64)
	if err != nil {
		return floats.Float128{}, err
	}
	a1, err := strconv.ParseUint(s[16:], 16, 64)
	if err != nil {
		return floats.Float128{}, err
	}
	return floats.Float128{a0, a1}, nil
}

func parseFlag(s string) (byte, error) {
	b, err := strconv.ParseUint(s, 16, 8)
	if err != nil {
		return 0, err
	}
	return byte(b), nil
}

// eq16 reports whether a and b are equal.
// It returns true if both a and b are NaN.
// It distinguishes between +0 and -0.
func eq16(a, b floats.Float16) bool {
	if a.IsNaN() && b.IsNaN() {
		return true
	}
	return a == b
}

// eq32 reports whether a and b are equal.
// It returns true if both a and b are NaN.
// It distinguishes between +0 and -0.
func eq32(a, b floats.Float32) bool {
	if a.IsNaN() && b.IsNaN() {
		return true
	}
	return math.Float32bits(float32(a)) == math.Float32bits(float32(b))
}

// eq64 reports whether a and b are equal.
// It returns true if both a and b are NaN.
// It distinguishes between +0 and -0.
func eq64(a, b floats.Float64) bool {
	if a.IsNaN() && b.IsNaN() {
		return true
	}
	return math.Float64bits(float64(a)) == math.Float64bits(float64(b))
}

// eq128 reports whether a and b are equal.
// It returns true if both a and b are NaN.
// It distinguishes between +0 and -0.
func eq128(a, b floats.Float128) bool {
	if a.IsNaN() && b.IsNaN() {
		return true
	}
	return a[0] == b[0] && a[1] == b[1]
}
