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

func TestFloat256_Asin(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-1), "-1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		{exact256(-0.75), "-0.8480620789814810080529443389984180800733662132631126428607181635702008212284742343"},
		{exact256(-0.5), "-0.5235987755982988730771072305465838140328615665625176368291574320513027343810348331"},
		{exact256(-0.25), "-0.2526802551420786534856574369937109722521937330968381936339237787405750604810212224"},
		{exact256(0.25), "0.2526802551420786534856574369937109722521937330968381936339237787405750604810212224"},
		{exact256(0.5), "0.5235987755982988730771072305465838140328615665625176368291574320513027343810348331"},
		{exact256(0.75), "0.8480620789814810080529443389984180800733662132631126428607181635702008212284742343"},
		{exact256(1), "1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
	}

	for _, tt := range tests {
		got := tt.x.Asin()
		if !close256(got, tt.want) {
			t.Errorf("Asin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
		{exact256(math.NaN()), exact256(math.NaN())},
		{exact256(2), exact256(math.NaN())},
		{exact256(-2), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Asin()
		if !eq256(got, tt.want) {
			t.Errorf("Asin(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat256_Acos(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-1), "3.141592653589793238462643383279502884197169399375105820974944592307816406286208999"},
		{exact256(-0.75), "2.418858405776377627284266030638169522171950912950665553348190459724109024371578734"},
		{exact256(-0.5), "2.094395102393195492308428922186335256131446266250070547316629728205210937524139332"},
		{exact256(-0.25), "1.823476581936975272716979128633462414350778432784391104121396074894483263624125722"},
		{exact256(0.25), "1.318116071652817965745664254646040469846390966590714716853548517413333142662083277"},
		{exact256(0.5), "1.047197551196597746154214461093167628065723133125035273658314864102605468762069666"},
		{exact256(0.75), "0.722734247813415611178377352641333362025218486424440267626754132583707381914630265"},
		{exact256(1), "0"},
	}

	for _, tt := range tests {
		got := tt.x.Acos()
		if !close256(got, tt.want) {
			t.Errorf("Acos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(math.NaN()), exact256(math.NaN())},
		{exact256(2), exact256(math.NaN())},
		{exact256(-2), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Acos()
		if !eq256(got, tt.want) {
			t.Errorf("Acos(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat256_Atan(t *testing.T) {
	t.Parallel()
	tests := []struct {
		x    Float256
		want string
	}{
		{exact256(-0.5), "-0.4636476090008061162142562314612144020285370542861202638109330887201978641657417053"},
		{exact256(-0.25), "-0.2449786631268641541720824812112758109141440983811840671273759146673551195876420966"},
		{exact256(-0.125), "-0.1243549945467614350313548491638710255731701917698040899151141191157222674275667586"},
		{exact256(0.125), "0.1243549945467614350313548491638710255731701917698040899151141191157222674275667586"},
		{exact256(0.25), "0.2449786631268641541720824812112758109141440983811840671273759146673551195876420966"},
		{exact256(0.5), "0.4636476090008061162142562314612144020285370542861202638109330887201978641657417053"},
		{exact256(0.75), "0.6435011087932843868028092287173226380415105911153123828656061187135124748116210887"},
		{exact256(1), "0.7853981633974483096156608458198757210492923498437764552437361480769541015715522497"},
		{exact256(2), "1.107148717794090503017065460178537040070047645401432646676539207433710338977362794"},
		{exact256(math.Inf(-1)), "-1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		{exact256(math.Inf(1)), "1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
	}

	for _, tt := range tests {
		got := tt.x.Atan()
		if !close256(got, tt.want) {
			t.Errorf("Atan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		x    Float256
		want Float256
	}{
		// special cases
		{exact256(0), exact256(0)},
		{exact256(math.Copysign(0, -1)), exact256(math.Copysign(0, -1))},
		{exact256(math.NaN()), exact256(math.NaN())},
	}

	for _, tt := range strictTests {
		got := tt.x.Atan()
		if !eq256(got, tt.want) {
			t.Errorf("Atan(%v) = %v; want %v", tt.x, got, tt.want)
		}
	}
}

func TestFloat256_Atan2(t *testing.T) {
	t.Parallel()
	tests := []struct {
		y, x Float256
		want string
	}{
		{exact256(1), exact256(1), "0.7853981633974483096156608458198757210492923498437764552437361480769541015715522497"},
		{exact256(1), exact256(-1), "2.356194490192344928846982537459627163147877049531329365731208444230862304714656749"},
		{exact256(-1), exact256(-1), "-2.356194490192344928846982537459627163147877049531329365731208444230862304714656749"},
		{exact256(-1), exact256(1), "-.7853981633974483096156608458198757210492923498437764552437361480769541015715522497"},

		// special cases
		// +0.Atan2(x<=-0) = +Pi
		{exact256(0), exact256(-1), "3.141592653589793238462643383279502884197169399375105820974944592307816406286208999"},
		// -0.Atan2(x<=-0) = -Pi
		{exact256(math.Copysign(0, -1)), exact256(-1), "-3.141592653589793238462643383279502884197169399375105820974944592307816406286208999"},
		// y>0.Atan2(0) = +Pi/2
		{exact256(1), exact256(0), "1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		{exact256(1), exact256(math.Copysign(0, -1)), "1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		// y<0.Atan2(0) = -Pi/2
		{exact256(-1), exact256(0), "-1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		{exact256(-1), exact256(math.Copysign(0, -1)), "-1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		// +Inf.Atan2(+Inf) = +Pi/4
		{exact256(math.Inf(1)), exact256(math.Inf(1)), "0.7853981633974483096156608458198757210492923498437764552437361480769541015715522497"},
		// -Inf.Atan2(+Inf) = -Pi/4
		{exact256(math.Inf(-1)), exact256(math.Inf(1)), "-0.7853981633974483096156608458198757210492923498437764552437361480769541015715522497"},
		// +Inf.Atan2(-Inf) = 3*Pi/4
		{exact256(math.Inf(1)), exact256(math.Inf(-1)), "2.356194490192344928846982537459627163147877049531329365731208444230862304714656749"},
		// -Inf.Atan2(-Inf) = -3*Pi/4
		{exact256(math.Inf(-1)), exact256(math.Inf(-1)), "-2.356194490192344928846982537459627163147877049531329365731208444230862304714656749"},
		// y.Atan2(+Inf) = 0
		{exact256(1), exact256(math.Inf(1)), "0"},
		{exact256(-1), exact256(math.Inf(1)), "0"},
		// (y>0).Atan2(-Inf) = +Pi
		{exact256(1), exact256(math.Inf(-1)), "3.141592653589793238462643383279502884197169399375105820974944592307816406286208999"},
		// (y<0).Atan2(-Inf) = -Pi
		{exact256(-1), exact256(math.Inf(-1)), "-3.141592653589793238462643383279502884197169399375105820974944592307816406286208999"},
		// +Inf.Atan2(x) = +Pi/2
		{exact256(math.Inf(1)), exact256(1), "1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
		// -Inf.Atan2(x) = -Pi/2
		{exact256(math.Inf(-1)), exact256(1), "-1.570796326794896619231321691639751442098584699687552910487472296153908203143104499"},
	}

	for _, tt := range tests {
		got := tt.y.Atan2(tt.x)
		if !close256(got, tt.want) {
			t.Errorf("Atan2(%v, %v) = %v; want %v", tt.y, tt.x, got, tt.want)
		}
	}

	strictTests := []struct {
		y, x Float256
		want Float256
	}{
		// special cases
		// y.Atan2(NaN) = NaN
		{exact256(1), exact256(math.NaN()), exact256(math.NaN())},
		{exact256(math.NaN()), exact256(math.NaN()), exact256(math.NaN())},
		// NaN.Atan2(x) = NaN
		{exact256(math.NaN()), exact256(1), exact256(math.NaN())},
		// +0.Atan2(x>=0) = +0
		{exact256(0), exact256(1), exact256(0)},
		// -0.Atan2(x>=0) = -0
		{exact256(math.Copysign(0, -1)), exact256(1), exact256(math.Copysign(0, -1))},
	}

	for _, tt := range strictTests {
		got := tt.y.Atan2(tt.x)
		if !eq256(got, tt.want) {
			t.Errorf("Atan2(%v, %v) = %v; want %v", tt.y, tt.x, got, tt.want)
		}
	}
}

// TestFloat256_Atan2Accuracy checks that the error of Atan2 is less than 1 ulp,
// and that almost all results are correctly rounded.
// The test data is generated by scripts/gen_atan2_testdata.py.
func TestFloat256_Atan2Accuracy(t *testing.T) {
	t.Parallel()
	f, err := os.Open("testdata/atan2_256.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	parse := func(s string) Float256 {
		var x Float256
		for i := range x {
			v, err := strconv.ParseUint(s[16*i:16*(i+1)], 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			x[i] = v
		}
		return x
	}
	within1ulp := func(got, want Float256) bool {
		inf := NewFloat256Inf(1)
		return eq256(got, want) || eq256(got, want.Nextafter(inf)) || eq256(got, want.Nextafter(inf.Neg()))
	}

	var total, misrounded int
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		y, x, want := parse(fields[0]), parse(fields[1]), parse(fields[2])

		got := y.Atan2(x)
		total++
		if !eq256(got, want) {
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
		t.Fatal("testdata/atan2_256.txt has no test cases")
	}
	if misrounded*100 > total {
		t.Errorf("Atan2: %d of %d results are not correctly rounded; want at most 1%%", misrounded, total)
	}
	t.Logf("misrounded: %d of %d", misrounded, total)
}

func BenchmarkFloat256_Atan2(b *testing.B) {
	for _, tt := range []struct {
		name string
		y, x Float256
	}{
		{"small_ratio", exact256(0.001), exact256(3)},
		{"medium", exact256(1.5), exact256(-2.25)},
		{"large_ratio", exact256(1e300), exact256(0.5)},
	} {
		b.Run(tt.name, func(b *testing.B) {
			for b.Loop() {
				runtime.KeepAlive(tt.y.Atan2(tt.x))
			}
		})
	}
}

// TestFloat256_AtanAccuracy checks that the error of Atan is less than 1 ulp,
// and that almost all results are correctly rounded.
// The test data is generated by scripts/gen_atan_testdata.py.
func TestFloat256_AtanAccuracy(t *testing.T) {
	t.Parallel()
	testFloat256LogAccuracy(t, "testdata/atan_256.txt", "Atan", Float256.Atan)
}

func BenchmarkFloat256_Atan(b *testing.B) {
	benchFloat256(b, Float256.Atan, []struct {
		name string
		x    Float256
	}{
		{"tiny", exact256(1e-30)},
		{"small", exact256(0.001)},
		{"medium", exact256(0.75)},
		{"large", exact256(3)},
		{"huge", exact256(1e300)},
	})
}

// TestFloat256_AsinAccuracy checks that the error of Asin is less than 1 ulp,
// and that almost all results are correctly rounded.
// The test data is generated by scripts/gen_asin_testdata.py.
func TestFloat256_AsinAccuracy(t *testing.T) {
	t.Parallel()
	testFloat256LogAccuracy(t, "testdata/asin_256.txt", "Asin", Float256.Asin)
}

func BenchmarkFloat256_Asin(b *testing.B) {
	benchFloat256(b, Float256.Asin, []struct {
		name string
		x    Float256
	}{
		{"tiny", exact256(1e-30)},
		{"small", exact256(0.001)},
		{"medium", exact256(0.5)},
		{"large", exact256(0.9)},
		{"near1", Float256{0x3fffefffffffffff, 0xffffffffffffffff, 0xffffffffffffffff, 0xfffffffffffffff0}},
	})
}
