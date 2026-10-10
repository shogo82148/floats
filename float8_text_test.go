package floats

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"testing"
)

type float8Value interface {
	Text(fmt byte, prec int) string
	String() string
	Float64() Float64
	IsNaN() bool
}

func forEachFloat8(t *testing.T, f func(t *testing.T, enc uint8, a, b float8Value)) {
	t.Helper()
	for i := range 256 {
		f(t, uint8(i), Float8E4M3(i), Float8E5M2(i))
	}
}

func TestFloat8_Text(t *testing.T) {
	t.Parallel()
	forEachFloat8(t, func(t *testing.T, enc uint8, a, b float8Value) {
		for name, v := range map[string]float8Value{"Float8E4M3": a, "Float8E5M2": b} {
			// the exact digits must be the same as Float64
			for _, f := range []byte{'e', 'E', 'f', 'g', 'G'} {
				for prec := 0; prec <= 24; prec++ {
					if got, want := v.Text(f, prec), v.Float64().Text(f, prec); got != want {
						t.Errorf("%s(%#02x).Text(%q, %d) = %q, want %q", name, enc, f, prec, got, want)
					}
				}
			}
			if got, want := v.Text('x', -1), v.Float64().Text('x', -1); got != want {
				t.Errorf("%s(%#02x).Text('x', -1) = %q, want %q", name, enc, got, want)
			}
			if got, want := v.Text('x', 3), v.Float64().Text('x', 3); got != want {
				t.Errorf("%s(%#02x).Text('x', 3) = %q, want %q", name, enc, got, want)
			}
			if got, want := v.String(), v.Text('g', -1); got != want {
				t.Errorf("%s(%#02x).String() = %q, want %q", name, enc, got, want)
			}
		}
	})
}

func TestFloat8_TextShortest(t *testing.T) {
	t.Parallel()
	// the shortest representation must be parsed as the original value,
	// and any shorter representation must not.
	forEachFloat8(t, func(t *testing.T, enc uint8, a, b float8Value) {
		check := func(name string, v float8Value, parse func(string) (uint8, error)) {
			if v.IsNaN() || math.IsInf(float64(v.Float64()), 0) {
				return
			}
			for _, f := range []byte{'e', 'g', 'f'} {
				s := v.Text(f, -1)
				got, err := parse(s)
				if err != nil || got != enc {
					t.Errorf("%s(%#02x).Text(%q, -1) = %q, parsed as %#02x (%v)", name, enc, f, s, got, err)
				}
			}
			digits := len(strings.Replace(strings.TrimLeft(strings.SplitN(v.Text('e', -1), "e", 2)[0], "-"), ".", "", 1))
			if v.Float64() == 0 {
				return
			}
			if digits > 1 {
				s := v.Float64().Text('e', digits-2)
				if got, err := parse(s); err == nil && got == enc {
					t.Errorf("%s(%#02x).Text('e', -1) has %d digits, but %q is parsed as the same value", name, enc, digits, s)
				}
			}
		}
		check("Float8E4M3", a, func(s string) (uint8, error) { v, err := ParseFloat8E4M3(s); return uint8(v), err })
		check("Float8E5M2", b, func(s string) (uint8, error) { v, err := ParseFloat8E5M2(s); return uint8(v), err })
	})
}

func TestFloat8_TextExamples(t *testing.T) {
	t.Parallel()
	if got := Float8E4M3(0x38).String(); got != "1" {
		t.Errorf("got %q", got)
	}
	if got := Float8E4M3(0x7e).String(); got != "450" {
		t.Errorf("got %q", got)
	}
	if got := Float8E4M3(0xff).String(); got != "NaN" {
		t.Errorf("got %q", got)
	}
	if got := Float8E4M3(0x01).Text('g', -1); got != "0.002" {
		t.Errorf("got %q", got)
	}
	if got := Float8E4M3(0xb8).Text('b', 0); got != "-8p-3" {
		t.Errorf("got %q", got)
	}
	if got := Float8E5M2(0x7b).String(); got != "60000" {
		t.Errorf("got %q", got)
	}
	if got := Float8E5M2(0x7c).String(); got != "+Inf" {
		t.Errorf("got %q", got)
	}
	if got := Float8E5M2(0xfc).String(); got != "-Inf" {
		t.Errorf("got %q", got)
	}
	if got := Float8E5M2(0x01).Text('b', 0); got != "1p-16" {
		t.Errorf("got %q", got)
	}
	if got := Float8E5M2(0x01).Text('q', 0); got != "%q" {
		t.Errorf("got %q", got)
	}
	if got := Float8E5M2(0x7e).Text('g', 3); got != "NaN" {
		t.Errorf("got %q", got)
	}
}

// refParse returns the expected result of parsing the decimal or hexadecimal string s.
func refParse(t *testing.T, s float8Spec, str string) (enc uint8, overflow bool) {
	t.Helper()
	x, _, err := big.ParseFloat(str, 0, 600, big.ToNearestEven)
	if err != nil {
		t.Fatalf("big.ParseFloat(%q): %v", str, err)
	}
	var sign uint8
	if x.Signbit() {
		sign = 0x80
		x.Neg(x)
	}
	r := s.round(x)
	if r >= s.limit {
		return sign | s.limit, true
	}
	return sign | r, false
}

func TestParseFloat8(t *testing.T) {
	t.Parallel()
	var strs []string
	for _, c := range float8Candidates() {
		if c > 1e300 {
			continue
		}
		for _, v := range []float64{c, math.Nextafter(c, math.Inf(1)), math.Nextafter(c, 0)} {
			// exact decimal expansions of the candidates are short.
			e := strconv.FormatFloat(v, 'e', 80, 64)
			strs = append(strs, e, "-"+e)
			if v == c {
				// slightly larger than c
				m, x, _ := strings.Cut(e, "e")
				strs = append(strs, m+"1e"+x, "-"+m+"1e"+x)
			}
			h := strconv.FormatFloat(v, 'x', -1, 64)
			strs = append(strs, h, "-"+h)
			if v == c {
				m, x, _ := strings.Cut(h, "p")
				if !strings.Contains(m, ".") {
					m += "."
				}
				strs = append(strs, m+"000000000000000000001p"+x)
			}
		}
	}
	strs = append(strs, "0", "-0", "1e-400", "-1e-400", "0x1p-100000", "1e400", "-1e400", "0x1p100000", "1.5", "  ")
	for _, str := range strs {
		if strings.TrimSpace(str) == "" {
			continue
		}
		// Go's strconv does not accept the hexadecimal float without the exponent, but ours does not need it.
		for _, tt := range []struct {
			spec  float8Spec
			parse func(string) (uint8, error)
		}{
			{specE4M3, func(s string) (uint8, error) { v, err := ParseFloat8E4M3(s); return uint8(v), err }},
			{specE5M2, func(s string) (uint8, error) { v, err := ParseFloat8E5M2(s); return uint8(v), err }},
		} {
			want, overflow := refParse(t, tt.spec, str)
			got, err := tt.parse(str)
			if overflow {
				if !errors.Is(err, strconv.ErrRange) {
					t.Errorf("%s: Parse(%q) error = %v, want ErrRange", tt.spec.name, str, err)
				}
				if tt.spec.limit == 0x7f {
					if !tt.spec.isNaN(got) {
						t.Errorf("%s: Parse(%q) = %#02x, want NaN", tt.spec.name, str, got)
					}
					continue
				}
			} else if err != nil {
				t.Errorf("%s: Parse(%q) error = %v", tt.spec.name, str, err)
			}
			if got != want {
				t.Errorf("%s: Parse(%q) = %#02x, want %#02x", tt.spec.name, str, got, want)
			}
		}
	}
}

func TestParseFloat8_Special(t *testing.T) {
	t.Parallel()
	for _, s := range []string{"NaN", "nan"} {
		if v, err := ParseFloat8E4M3(s); err != nil || !v.IsNaN() {
			t.Errorf("ParseFloat8E4M3(%q) = %v, %v", s, v, err)
		}
		if v, err := ParseFloat8E5M2(s); err != nil || !v.IsNaN() {
			t.Errorf("ParseFloat8E5M2(%q) = %v, %v", s, v, err)
		}
	}
	for _, s := range []string{"Inf", "+inf", "infinity"} {
		if v, err := ParseFloat8E5M2(s); err != nil || !v.IsInf(1) {
			t.Errorf("ParseFloat8E5M2(%q) = %v, %v", s, v, err)
		}
		if v, err := ParseFloat8E4M3(s); err != nil || !v.IsNaN() {
			t.Errorf("ParseFloat8E4M3(%q) = %v, %v", s, v, err)
		}
	}
	if v, err := ParseFloat8E5M2("-Inf"); err != nil || !v.IsInf(-1) {
		t.Errorf("ParseFloat8E5M2(-Inf) = %v, %v", v, err)
	}
	for _, s := range []string{"", "abc", "1.5x", "1e", "0x", "--1"} {
		if _, err := ParseFloat8E4M3(s); !errors.Is(err, strconv.ErrSyntax) {
			t.Errorf("ParseFloat8E4M3(%q) error = %v, want ErrSyntax", s, err)
		}
		if _, err := ParseFloat8E5M2(s); !errors.Is(err, strconv.ErrSyntax) {
			t.Errorf("ParseFloat8E5M2(%q) error = %v, want ErrSyntax", s, err)
		}
	}
}

func TestFloat8_Marshal(t *testing.T) {
	t.Parallel()
	type S struct {
		A Float8E4M3
		B Float8E5M2
		C []Float8E4M3
	}
	in := S{A: 0x7e, B: 0xfb, C: []Float8E4M3{0, 0x01, 0xb8}}
	data, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"A":450,"B":-60000,"C":[0,0.002,-1]}`; string(data) != want {
		t.Errorf("json.Marshal = %s, want %s", data, want)
	}
	var out S
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out.A != in.A || out.B != in.B || len(out.C) != 3 || out.C[0] != 0 || out.C[1] != 1 || out.C[2] != 0xb8 {
		t.Errorf("json round trip: got %+v", out)
	}

	for _, v := range []json.Marshaler{Float8E4M3(0x7f), Float8E4M3(0xff), Float8E5M2(0x7e), Float8E5M2(0x7c), Float8E5M2(0xfc)} {
		if _, err := v.MarshalJSON(); err == nil {
			t.Errorf("MarshalJSON(%v) should fail", v)
		}
	}

	// all the finite values are round trip through the text and JSON.
	for i := range 256 {
		a := Float8E4M3(i)
		if a.IsNaN() {
			continue
		}
		text, err := a.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		var b Float8E4M3
		if err := b.UnmarshalText(text); err != nil || b != a {
			t.Errorf("Float8E4M3(%#02x): text %q -> %#02x (%v)", i, text, uint8(b), err)
		}
		js, err := a.MarshalJSON()
		if err != nil {
			t.Fatal(err)
		}
		var c Float8E4M3
		if err := c.UnmarshalJSON(js); err != nil || c != a {
			t.Errorf("Float8E4M3(%#02x): json %q -> %#02x (%v)", i, js, uint8(c), err)
		}

		d := Float8E5M2(i)
		if d.IsNaN() {
			continue
		}
		text, err = d.MarshalText()
		if err != nil {
			t.Fatal(err)
		}
		var e Float8E5M2
		if err := e.UnmarshalText(text); err != nil || e != d {
			t.Errorf("Float8E5M2(%#02x): text %q -> %#02x (%v)", i, text, uint8(e), err)
		}
		if !d.IsInf(0) {
			js, err = d.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			var f Float8E5M2
			if err := f.UnmarshalJSON(js); err != nil || f != d {
				t.Errorf("Float8E5M2(%#02x): json %q -> %#02x (%v)", i, js, uint8(f), err)
			}
		}
	}
	var a Float8E4M3
	if err := a.UnmarshalText([]byte("x")); err == nil {
		t.Error("UnmarshalText should fail")
	}
	if err := a.UnmarshalJSON([]byte("1e999")); err == nil {
		t.Error("UnmarshalJSON should fail")
	}
}

func TestFloat8_Append(t *testing.T) {
	t.Parallel()
	forEachFloat8(t, func(t *testing.T, enc uint8, a, b float8Value) {
		x, y := a.(Float8E4M3), b.(Float8E5M2)
		for _, f := range []byte{'b', 'e', 'f', 'g', 'x', 'q'} {
			for _, prec := range []int{-1, 0, 3} {
				if got, want := string(x.Append([]byte("pre:"), f, prec)), "pre:"+x.Text(f, prec); got != want {
					t.Errorf("Float8E4M3(%#02x).Append(%q, %d) = %q, want %q", enc, f, prec, got, want)
				}
				if got, want := string(y.Append([]byte("pre:"), f, prec)), "pre:"+y.Text(f, prec); got != want {
					t.Errorf("Float8E5M2(%#02x).Append(%q, %d) = %q, want %q", enc, f, prec, got, want)
				}
			}
		}
		if got, err := x.AppendText([]byte("pre:")); err != nil || string(got) != "pre:"+x.String() {
			t.Errorf("Float8E4M3(%#02x).AppendText = %q, %v", enc, got, err)
		}
		if got, err := y.AppendText([]byte("pre:")); err != nil || string(got) != "pre:"+y.String() {
			t.Errorf("Float8E5M2(%#02x).AppendText = %q, %v", enc, got, err)
		}
	})
}

func TestFloat8_Format(t *testing.T) {
	t.Parallel()
	// Float64 prints the exact value with the formats that have an explicit precision.
	exact := []string{"%.3f", "%.0e", "%10.2f", "%-10.2f|", "%+.2e", "%x", "%.2x", "%.3g", "%12.4E", "%010.3f", "%+012.3f", "%-+10.1e|", "% .2f"}
	// Float64 prints the shortest representation of Float64, so the shortest representations
	// of the Float8 values are compared with the Float64 values parsed from them.
	shortest := []string{"%v", "%g", "%G", "%8v", "%-8v|", "% v", "%+v", "%+8v", "%08v", "%-08v|"}
	forEachFloat8(t, func(t *testing.T, enc uint8, a, b float8Value) {
		for name, v := range map[string]float8Value{"Float8E4M3": a, "Float8E5M2": b} {
			for _, f := range exact {
				if got, want := fmt.Sprintf(f, v), fmt.Sprintf(f, float64(v.Float64())); got != want {
					t.Errorf("Sprintf(%q, %s(%#02x)) = %q, want %q", f, name, enc, got, want)
				}
			}
			p, err := strconv.ParseFloat(v.String(), 64)
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range shortest {
				if got, want := fmt.Sprintf(f, v), fmt.Sprintf(f, p); got != want {
					t.Errorf("Sprintf(%q, %s(%#02x)) = %q, want %q", f, name, enc, got, want)
				}
			}
		}
	})
}
