package floats

import "testing"

var shifttests = []struct {
	i     uint64
	shift int
	out   string
}{
	{0, -100, "0"},
	{0, 100, "0"},
	{1, 100, "1267650600228229401496703205376"},
	{1, -100,
		"0.00000000000000000000000000000078886090522101180541" +
			"17285652827862296732064351090230047702789306640625",
	},
	{12345678, 8, "3160493568"},
	{12345678, -8, "48225.3046875"},
	{195312, 9, "99999744"},
	{1953125, 9, "1000000000"},
}

func TestDecimalShift(t *testing.T) {
	for _, test := range shifttests {
		d := new(decimal)
		d.AssignUint64(test.i)
		d.Shift(test.shift)
		s := d.String()
		if s != test.out {
			t.Errorf("Decimal %v << %v = %v, want %v",
				test.i, test.shift, s, test.out)
		}
	}
}
