package floats

import (
	"bufio"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestFloat128_Asin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(-1), "-1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		{exact128(-0.75), "-0.8480620789814810080529443389984180800733662132631126428607181635702008212284742343"},
		{exact128(-0.5), "-0.5235987755982988730771072305465838140328615665625176368291574320513027343810348331"},
		{exact128(-0.25), "-0.2526802551420786534856574369937109722521937330968381936339237787405750604810212224"},
		{exact128(0.25), "0.2526802551420786534856574369937109722521937330968381936339237787405750604810212224"},
		{exact128(0.5), "0.5235987755982988730771072305465838140328615665625176368291574320513027343810348331"},
		{exact128(0.75), "0.8480620789814810080529443389984180800733662132631126428607181635702008212284742343"},
		{exact128(1), "1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
	}

	for _, tt := range tests {
		got := tt.x.Asin()
		if !close128(got, tt.want) {
			t.Errorf("Asin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(0), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},
		{exact128(math.NaN()), exact128(math.NaN())},
		{exact128(2), exact128(math.NaN())},
		{exact128(-2), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Asin()
		if !eq128(got, tt.want) {
			t.Errorf("Asin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat128_Acos(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(-1), "3.141592653589793238462643383279502884197169399375105820974944592307816406286208999"},
		{exact128(-0.75), "2.418858405776377627284266030638169522171950912950665553348190459724109024371578734"},
		{exact128(-0.5), "2.094395102393195492308428922186335256131446266250070547316629728205210937524139332"},
		{exact128(-0.25), "1.823476581936975272716979128633462414350778432784391104121396074894483263624125722"},
		{exact128(0.25), "1.318116071652817965745664254646040469846390966590714716853548517413333142662083277"},
		{exact128(0.5), "1.047197551196597746154214461093167628065723133125035273658314864102605468762069666"},
		{exact128(0.75), "0.722734247813415611178377352641333362025218486424440267626754132583707381914630265"},
		{exact128(1), "0"},
	}

	for _, tt := range tests {
		got := tt.x.Acos()
		if !close128(got, tt.want) {
			t.Errorf("Acos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(math.NaN()), exact128(math.NaN())},
		{exact128(2), exact128(math.NaN())},
		{exact128(-2), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Acos()
		if !eq128(got, tt.want) {
			t.Errorf("Acos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat128_Atan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float128
		want string
	}{
		{exact128(-0.5), "-0.4636476090008061162142562314612144020285370542861202638109330887201978641657417053"},
		{exact128(-0.25), "-0.2449786631268641541720824812112758109141440983811840671273759146673551195876420966"},
		{exact128(-0.125), "-0.1243549945467614350313548491638710255731701917698040899151141191157222674275667586"},
		{exact128(0.125), "0.1243549945467614350313548491638710255731701917698040899151141191157222674275667586"},
		{exact128(0.25), "0.2449786631268641541720824812112758109141440983811840671273759146673551195876420966"},
		{exact128(0.5), "0.4636476090008061162142562314612144020285370542861202638109330887201978641657417053"},
		{exact128(0.75), "0.6435011087932843868028092287173226380415105911153123828656061187135124748116210887"},
		{exact128(1), "0.7853981633974483096156608458198757210492923498437764552437361480769541015715522497"},
		{exact128(2), "1.107148717794090503017065460178537040070047645401432646676539207433710338977362794"},
		{exact128(math.Inf(-1)), "-1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		{exact128(math.Inf(1)), "1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
	}

	for _, tt := range tests {
		got := tt.x.Atan()
		if !close128(got, tt.want) {
			t.Errorf("Atan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float128
		want Float128
	}{
		// special cases
		{exact128(0), exact128(0)},
		{exact128(math.Copysign(0, -1)), exact128(math.Copysign(0, -1))},
		{exact128(math.NaN()), exact128(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Atan()
		if !eq128(got, tt.want) {
			t.Errorf("Atan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat128_Atan2(t *testing.T) {
	t.Parallel()
	tests := []struct {
		y, x Float128
		want string
	}{
		{exact128(1), exact128(1), "0.7853981633974483096156608458198757210492923498437764552437361480769541015715522497"},
		{exact128(1), exact128(-1), "2.356194490192344928846982537459627163147877049531329365731208444230862304714656749"},
		{exact128(-1), exact128(-1), "-2.356194490192344928846982537459627163147877049531329365731208444230862304714656749"},
		{exact128(-1), exact128(1), "-.7853981633974483096156608458198757210492923498437764552437361480769541015715522497"},

		// special cases
		// +0.Atan2(x<=-0) = +Pi
		{exact128(0), exact128(-1), "3.141592653589793238462643383279502884197169399375105820974944592307816406286208999"},
		// -0.Atan2(x<=-0) = -Pi
		{exact128(math.Copysign(0, -1)), exact128(-1), "-3.141592653589793238462643383279502884197169399375105820974944592307816406286208999"},
		// y>0.Atan2(0) = +Pi/2
		{exact128(1), exact128(0), "1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		{exact128(1), exact128(math.Copysign(0, -1)), "1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		// y<0.Atan2(0) = -Pi/2
		{exact128(-1), exact128(0), "-1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		{exact128(-1), exact128(math.Copysign(0, -1)), "-1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		// +Inf.Atan2(+Inf) = +Pi/4
		{exact128(math.Inf(1)), exact128(math.Inf(1)), "0.7853981633974483096156608458198757210492923498437764552437361480769541015715522497"},
		// -Inf.Atan2(+Inf) = -Pi/4
		{exact128(math.Inf(-1)), exact128(math.Inf(1)), "-0.7853981633974483096156608458198757210492923498437764552437361480769541015715522497"},
		// +Inf.Atan2(-Inf) = 3*Pi/4
		{exact128(math.Inf(1)), exact128(math.Inf(-1)), "2.356194490192344928846982537459627163147877049531329365731208444230862304714656749"},
		// -Inf.Atan2(-Inf) = -3*Pi/4
		{exact128(math.Inf(-1)), exact128(math.Inf(-1)), "-2.356194490192344928846982537459627163147877049531329365731208444230862304714656749"},
		// y.Atan2(+Inf) = 0
		{exact128(1), exact128(math.Inf(1)), "0"},
		{exact128(-1), exact128(math.Inf(1)), "0"},
		// (y>0).Atan2(-Inf) = +Pi
		{exact128(1), exact128(math.Inf(-1)), "3.141592653589793238462643383279502884197169399375105820974944592307816406286208999"},
		// (y<0).Atan2(-Inf) = -Pi
		{exact128(-1), exact128(math.Inf(-1)), "-3.141592653589793238462643383279502884197169399375105820974944592307816406286208999"},
		// +Inf.Atan2(x) = +Pi/2
		{exact128(math.Inf(1)), exact128(1), "1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		// -Inf.Atan2(x) = -Pi/2
		{exact128(math.Inf(-1)), exact128(1), "-1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
	}

	for _, tt := range tests {
		got := tt.y.Atan2(tt.x)
		if !close128(got, tt.want) {
			t.Errorf("Atan2(%v, %v) = %v; want %v", tt.y, tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		y, x Float128
		want Float128
	}{
		// special cases
		// y.Atan2(NaN) = NaN
		{exact128(1), exact128(math.NaN()), exact128(math.NaN())},
		{exact128(math.NaN()), exact128(math.NaN()), exact128(math.NaN())},
		// NaN.Atan2(x) = NaN
		{exact128(math.NaN()), exact128(1), exact128(math.NaN())},
		// +0.Atan2(x>=0) = +0
		{exact128(0), exact128(1), exact128(0)},
		// -0.Atan2(x>=0) = -0
		{exact128(math.Copysign(0, -1)), exact128(1), exact128(math.Copysign(0, -1))},
	}

	for _, tt := range strictTests {
		got := tt.y.Atan2(tt.x)
		if !eq128(got, tt.want) {
			t.Errorf("Atan2(%v, %v) = %v; want %v", tt.y, tt.x, got, tt.want)
		}
	}
}

func BenchmarkFloat128_Atan2(b *testing.B) {
	for _, tt := range []struct {
		name string
		y, x Float128
	}{
		{"small_ratio", exact128(0.001), exact128(3)},
		{"medium", exact128(1.5), exact128(-2.25)},
		{"large_ratio", exact128(1e300), exact128(0.5)},
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.y.Atan2(tt.x))
			}
		})
	}
}

// TestFloat128_Atan2Accuracy checks that the error of Atan2 is less than 1 ulp,
// and that almost all results are correctly rounded.
// The test data is generated by scripts/gen_atan2_testdata.py.
func TestFloat128_Atan2Accuracy(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/atan2_128.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	parse := func(s string) Float128 {
		var x Float128
		for i := range x {
			v, err := strconv.ParseUint(s[16*i:16*(i+1)], 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			x[i] = v
		}
		return x
	}
	within1ulp := func(got, want Float128) bool {
		inf := NewFloat128Inf(1)
		return eq128(got, want) || eq128(got, want.Nextafter(inf)) || eq128(got, want.Nextafter(inf.Neg()))
	}

	var total, misrounded int
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		y, x, want := parse(fields[0]), parse(fields[1]), parse(fields[2])

		got := y.Atan2(x)
		total++
		if !eq128(got, want) {
			misrounded++
		}
		if !within1ulp(got, want) {
			t.Errorf("Atan2(%v, %v) = %v; want %v", y, x, got, want)
		}
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	if total == 0 {
		t.Fatal("testdata/atan2_128.txt has no test cases")
	}
	if misrounded*100 > total {
		t.Errorf("Atan2: %d of %d results are not correctly rounded; want at most 1%%", misrounded, total)
	}
	t.Logf("misrounded: %d of %d", misrounded, total)
}

// TestFloat128_AtanAccuracy checks that the error of Atan is less than 1 ulp,
// and that almost all results are correctly rounded.
// The test data is generated by scripts/gen_atan_testdata.py.
func TestFloat128_AtanAccuracy(t *testing.T) {
	t.Parallel()
	testFloat128LogAccuracy(t, "testdata/atan_128.txt", "Atan", Float128.Atan)
}

func BenchmarkFloat128_Atan(b *testing.B) {
	benchFloat128(b, Float128.Atan, []struct {
		name string
		x    Float128
	}{
		{"tiny", exact128(1e-30)},
		{"small", exact128(0.001)},
		{"medium", exact128(0.75)},
		{"large", exact128(3)},
		{"huge", exact128(1e300)},
	})
}

// TestFloat128_AsinAccuracy checks that the error of Asin is less than 1 ulp,
// and that almost all results are correctly rounded.
// The test data is generated by scripts/gen_asin_testdata.py.
func TestFloat128_AsinAccuracy(t *testing.T) {
	t.Parallel()
	testFloat128LogAccuracy(t, "testdata/asin_128.txt", "Asin", Float128.Asin)
}

func BenchmarkFloat128_Asin(b *testing.B) {
	benchFloat128(b, Float128.Asin, []struct {
		name string
		x    Float128
	}{
		{"tiny", exact128(1e-30)},
		{"small", exact128(0.001)},
		{"medium", exact128(0.5)},
		{"large", exact128(0.9)},
		{"near1", Float128{0x3ffeffffffffffff, 0xfffffffffffffff0}},
	})
}

// TestFloat128_AcosAccuracy checks that the error of Acos is less than 1 ulp,
// and that almost all results are correctly rounded.
// The test data is generated by scripts/gen_acos_testdata.py.
func TestFloat128_AcosAccuracy(t *testing.T) {
	t.Parallel()
	testFloat128LogAccuracy(t, "testdata/acos_128.txt", "Acos", Float128.Acos)
}

func BenchmarkFloat128_Acos(b *testing.B) {
	benchFloat128(b, Float128.Acos, []struct {
		name string
		x    Float128
	}{
		{"tiny", exact128(1e-30)},
		{"small", exact128(0.001)},
		{"medium", exact128(0.5)},
		{"negative", exact128(-0.5)},
		{"near1", Float128{0x3ffeffffffffffff, 0xfffffffffffffff0}},
		{"near-1", Float128{0xbffeffffffffffff, 0xfffffffffffffff0}},
	})
}
