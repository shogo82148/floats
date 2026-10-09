package floats

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"
)

func TestBFloat16_Text(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a, x := bf16Value(i)
		if a.IsNaN() {
			if got := a.String(); got != "NaN" {
				t.Errorf("BFloat16(%#04x).String() = %q, want NaN", i, got)
			}
			continue
		}
		if a.IsInf(0) {
			want := "+Inf"
			if i&0x8000 != 0 {
				want = "-Inf"
			}
			if got := a.String(); got != want {
				t.Errorf("BFloat16(%#04x).String() = %q, want %q", i, got, want)
			}
			continue
		}

		// the exact value of a is representable in float64,
		// so the fixed precision formats are the same as the ones of float64.
		for _, f := range []byte{'e', 'E', 'f', 'g', 'G', 'x', 'X'} {
			for _, prec := range []int{0, 1, 2, 3, 5, 8, 20} {
				if got, want := a.Text(f, prec), strconv.FormatFloat(x, f, prec, 64); got != want {
					t.Errorf("BFloat16(%#04x).Text(%q, %d) = %q, want %q", i, f, prec, got, want)
				}
			}
		}
		if got, want := a.Text('x', -1), strconv.FormatFloat(x, 'x', -1, 64); got != want {
			t.Errorf("BFloat16(%#04x).Text('x', -1) = %q, want %q", i, got, want)
		}

		// the shortest representation
		for _, f := range []byte{'e', 'f', 'g'} {
			s := a.Text(f, -1)
			b, err := ParseBFloat16(s)
			if err != nil || b != a {
				t.Errorf("ParseBFloat16(%q) = %#04x, %v, want %#04x", s, uint16(b), err, i)
				continue
			}
		}
		s := a.Text('e', -1)
		mant, _, _ := strings.Cut(s, "e")
		digits := len(strings.NewReplacer("-", "", ".", "").Replace(mant))
		if x == 0 {
			if s != "0e+00" && s != "-0e+00" {
				t.Errorf("BFloat16(%#04x).Text('e', -1) = %q", i, s)
			}
			continue
		}
		if digits > 1 {
			// no shorter representation exists
			short := strconv.FormatFloat(x, 'e', digits-2, 64)
			if b, err := ParseBFloat16(short); err == nil && b == a {
				t.Errorf("BFloat16(%#04x).Text('e', -1) = %q, but %q is shorter", i, s, short)
			}
		}
		// the shortest representation is the nearest one of the same length,
		// except for the one-digit representations of the smallest subnormal numbers
		// whose rounding interval is so wide that the digit above can be chosen.
		if want := strconv.FormatFloat(x, 'e', digits-1, 64); want != s && digits > 1 {
			if b, err := ParseBFloat16(want); err == nil && b == a {
				t.Errorf("BFloat16(%#04x).Text('e', -1) = %q, want %q", i, s, want)
			}
		}
	}

	bin := []struct {
		in   BFloat16
		want string
	}{
		{0, "0p-133"},
		{0x3f80, "128p-7"},
		{0xbf80, "-128p-7"},
		{0x4049, "201p-6"},
		{0x0001, "1p-133"},
		{0x0080, "128p-133"},
		{0x7f7f, "255p+120"},
	}
	for _, tt := range bin {
		if got := tt.in.Text('b', -1); got != tt.want {
			t.Errorf("BFloat16(%#04x).Text('b', -1) = %q, want %q", uint16(tt.in), got, tt.want)
		}
	}
	if got := BFloat16(0x7f80).Text('b', -1); got != "+Inf" {
		t.Errorf("Text('b') of +Inf = %q", got)
	}
	if got := BFloat16(0x4049).Text('z', -1); got != "%z" {
		t.Errorf("Text('z') = %q", got)
	}
}

func TestBFloat16_Format(t *testing.T) {
	t.Parallel()
	pi := BFloat16(0x4049) // 3.140625
	tests := []struct {
		format string
		x      BFloat16
		want   string
	}{
		{"%v", pi, "3.14"},
		{"%g", pi, "3.14"},
		{"%.3g", pi, "3.14"},
		{"%8.2f", pi, "    3.14"},
		{"%-8.2f|", pi, "3.14    |"},
		{"%+.2f", pi, "+3.14"},
		{"% .2f", pi, " 3.14"},
		{"%e", pi, "3.14e+00"},
		{"%b", pi, "201p-6"},
		{"%x", pi, "0x1.92p+01"},
		{"%X", pi, "0X1.92P+01"},
		{"%v", BFloat16(0x7fc0), "NaN"},
		{"%v", BFloat16(0x7f80), "+Inf"},
		{"%v", BFloat16(0xff80), "-Inf"},
		{"%v", BFloat16(0x7f7f), "3.39e+38"},
		{"%v", BFloat16(0x0001), "1e-40"},
	}
	for _, tt := range tests {
		if got := fmt.Sprintf(tt.format, tt.x); got != tt.want {
			t.Errorf("Sprintf(%q, %#04x) = %q, want %q", tt.format, uint16(tt.x), got, tt.want)
		}
	}
}

func TestBFloat16_AppendText(t *testing.T) {
	t.Parallel()
	a := BFloat16(0x4049)
	got, err := a.AppendText([]byte("x="))
	if err != nil || string(got) != "x=3.14" {
		t.Errorf("AppendText = %q, %v", got, err)
	}
	got, err = a.MarshalText()
	if err != nil || string(got) != "3.14" {
		t.Errorf("MarshalText = %q, %v", got, err)
	}
	var b BFloat16
	if err := b.UnmarshalText([]byte("3.14")); err != nil || b != a {
		t.Errorf("UnmarshalText = %#04x, %v", uint16(b), err)
	}
	if err := b.UnmarshalText([]byte("x")); err == nil {
		t.Error("UnmarshalText: want a syntax error")
	}
}

func TestBFloat16_JSON(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a := BFloat16(i)
		if a.IsNaN() || a.IsInf(0) {
			if _, err := json.Marshal(a); err == nil {
				t.Errorf("json.Marshal(%#04x): want an error", i)
			}
			continue
		}
		data, err := json.Marshal(a)
		if err != nil {
			t.Fatalf("json.Marshal(%#04x): %v", i, err)
		}
		var b BFloat16
		if err := json.Unmarshal(data, &b); err != nil {
			t.Fatalf("json.Unmarshal(%s): %v", data, err)
		}
		if b != a && (!a.IsZero() || !b.IsZero() || string(data) != "0") {
			t.Fatalf("json round trip of %#04x = %#04x (%s)", i, uint16(b), data)
		}
	}
	var b BFloat16
	if err := json.Unmarshal([]byte(`"x"`), &b); err == nil {
		t.Error("json.Unmarshal: want an error")
	}
}

func TestParseBFloat16(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want BFloat16
		err  error
	}{
		{"0", 0, nil},
		{"-0", 0x8000, nil},
		{"1", 0x3f80, nil},
		{"+1", 0x3f80, nil},
		{"-2.5", 0xc020, nil},
		{"3.14", 0x4049, nil},
		{"1e0", 0x3f80, nil},
		{".5", 0x3f00, nil},
		{"5.", 0x40a0, nil},
		{"inf", 0x7f80, nil},
		{"+Inf", 0x7f80, nil},
		{"-infinity", 0xff80, nil},
		{"NaN", 0x7fc0, nil},
		{"0x1p0", 0x3f80, nil},
		{"0x1.8p1", 0x4040, nil},
		{"-0x1p-133", 0x8001, nil},
		{"0x1p-134", 0x0000, nil},
		{"0x1.1p-134", 0x0001, nil},
		{"0x1p127", 0x7f00, nil},
		{"0x1.01p0", 0x3f80, nil},
		{"0x1.010000000000000001p0", 0x3f81, nil}, // the tie is broken by the truncated digits
		{"0x1.03p0", 0x3f82, nil},
		{"0x1.030000000000000001p0", 0x3f82, nil},
		{"0x1.02fffffffffffffffp0", 0x3f81, nil},
		{"0x1.fep127", 0x7f7f, nil},
		{"0x1.ffp127", 0x7f80, strconv.ErrRange},
		{"0x1p128", 0x7f80, strconv.ErrRange},
		{"-0x1p128", 0xff80, strconv.ErrRange},
		{"1e39", 0x7f80, strconv.ErrRange},
		{"-1e39", 0xff80, strconv.ErrRange},
		{"3.39e38", 0x7f7f, nil},
		{"5e-41", 0x0001, nil},
		{"4e-41", 0x0000, nil},
		{"1e-50", 0x0000, nil},
		{"-1e-50", 0x8000, nil},
		{"1e-1000000", 0x0000, nil},
		{"1e1000000", 0x7f80, strconv.ErrRange},
		{"", 0, strconv.ErrSyntax},
		{"x", 0, strconv.ErrSyntax},
		{"1x", 0, strconv.ErrSyntax},
		{"1e", 0, strconv.ErrSyntax},
		{"0x", 0, strconv.ErrSyntax},
		{"--1", 0, strconv.ErrSyntax},
	}
	for _, tt := range tests {
		got, err := ParseBFloat16(tt.in)
		if got != tt.want || !errors.Is(err, tt.err) {
			t.Errorf("ParseBFloat16(%q) = %#04x, %v; want %#04x, %v", tt.in, uint16(got), err, uint16(tt.want), tt.err)
		}
	}
}

// TestParseBFloat16_Exact compares the results with the exact ones of math/big.
func TestParseBFloat16_Exact(t *testing.T) {
	t.Parallel()
	check := func(s string) {
		t.Helper()
		f, _, err := big.ParseFloat(s, 10, 4000, big.ToNearestEven)
		if err != nil {
			t.Fatalf("big.ParseFloat(%q): %v", s, err)
		}
		want := bf16FromBig(f)
		got, err := ParseBFloat16(s)
		if got != want {
			t.Fatalf("ParseBFloat16(%q) = %#04x, %v; want %#04x", s, uint16(got), err, uint16(want))
		}
		if got.IsInf(0) != (err != nil) {
			t.Fatalf("ParseBFloat16(%q): the error is %v, the result is %#04x", s, err, uint16(got))
		}
	}

	// the shortest representations and the values around the ties
	for i := range 1 << 16 {
		a, x := bf16Value(i)
		if a.IsNaN() || a.IsInf(0) {
			continue
		}
		check(a.String())
		if i&0x7fff == 0x7f7f {
			continue
		}
		// the tie between a and the next value: the exact decimal expansion of the midpoint
		mid := math.Float32frombits(uint32(i)<<16 | 0x8000)
		ms := strconv.FormatFloat(float64(mid), 'f', -1, 64)
		check(ms)
		check(ms + "1") // just above the tie
		if strings.Contains(ms, ".") {
			// just below the tie: the last digit of the fraction is 5.
			check(ms[:len(ms)-1] + "4" + strings.Repeat("9", 20))
		}
		// the neighbors in float64
		check(strconv.FormatFloat(math.Nextafter(float64(mid), 0), 'f', -1, 64))
		check(strconv.FormatFloat(math.Nextafter(float64(mid), math.Inf(1)), 'f', -1, 64))
		check(strconv.FormatFloat(x, 'e', 20, 64))
	}

	r := rand.New(rand.NewPCG(1, 2))
	for range 200_000 {
		digits := make([]byte, r.IntN(30)+1)
		for i := range digits {
			digits[i] = byte('0' + r.IntN(10))
		}
		s := string(digits)
		if len(s) > 1 && r.IntN(2) == 0 {
			p := r.IntN(len(s))
			s = s[:p] + "." + s[p:]
		}
		s += "e" + strconv.Itoa(r.IntN(100)-60)
		check(s)
	}
}

func TestParseBFloat16_Hex(t *testing.T) {
	t.Parallel()
	for i := range 1 << 16 {
		a, x := bf16Value(i)
		if a.IsNaN() || a.IsInf(0) {
			continue
		}
		s := strconv.FormatFloat(x, 'x', -1, 64)
		got, err := ParseBFloat16(s)
		if err != nil || got != a {
			t.Fatalf("ParseBFloat16(%q) = %#04x, %v; want %#04x", s, uint16(got), err, i)
		}
		// the ties and the values around them
		mid := float64(math.Float32frombits(uint32(i)<<16 | 0x8000))
		hm := strconv.FormatFloat(mid, 'x', -1, 64)
		p := strings.IndexByte(hm, 'p')
		for _, s := range []string{
			hm,
			hm[:p] + "01" + hm[p:], // just above the tie
			strconv.FormatFloat(math.Nextafter(mid, 0), 'x', -1, 64), // just below the tie
		} {
			f, _, err := big.ParseFloat(s, 0, 400, big.ToNearestEven)
			if err != nil {
				continue
			}
			got, _ := ParseBFloat16(s)
			if want := bf16FromBig(f); got != want {
				t.Fatalf("ParseBFloat16(%q) = %#04x, want %#04x", s, uint16(got), uint16(want))
			}
		}
	}
}
