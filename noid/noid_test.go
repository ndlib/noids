package noid

import (
	"fmt"
	"testing"
)

func TestSwizzle(t *testing.T) {
	expected := []struct {
		count int
		perm  []int
	}{{1, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
		{2, []int{0, 6, 1, 7, 2, 8, 3, 9, 4, 5}},
		{3, []int{0, 4, 8, 1, 5, 9, 2, 6, 3, 7}},
		{4, []int{0, 3, 6, 9, 1, 4, 7, 2, 5, 8}},
		{5, []int{0, 3, 6, 9, 1, 4, 7, 2, 5, 8}},
		{6, []int{0, 2, 4, 6, 8, 1, 3, 5, 7, 9}},
		{7, []int{0, 2, 4, 6, 8, 1, 3, 5, 7, 9}},
		{8, []int{0, 2, 4, 6, 8, 1, 3, 5, 7, 9}},
		{9, []int{0, 2, 4, 6, 8, 1, 3, 5, 7, 9}},
		{10, []int{0, 2, 4, 6, 8, 1, 3, 5, 7, 9}},
		{11, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
	}
	for _, e := range expected {
		rs := newRandomState(e.count, 10)
		for i := 0; i < 10; i++ {
			m := rs.swizzle(i)
			if m != e.perm[i] {
				t.Errorf("%v bins, swizzle(%v) = %v != %v\n", e.count, i, m, e.perm[i])
			}
			j := rs.invSwizzle(m)
			if j != i {
				t.Errorf("%v bins, invSwizzle(%v) = %v != %v\n", e.count, m, j, i)
			}
		}
	}
}

func TestParseTemplate(t *testing.T) {
	tests := []struct {
		format   string
		expected template
		valid    bool
	}{{"no period", template{}, false},
		{"id.reeddk", template{slug: "id", generator: 'r', template: "eedd", checkDigit: true}, true},
	}

	for _, test := range tests {
		p, valid := parseTemplate(test.format)
		if p != test.expected || valid != test.valid {
			t.Errorf("expected (%v,%v), got (%v,%v)\n", test.expected, test.valid, p, valid)
		}
	}
}

func TestGenerateSizes(t *testing.T) {
	r := generateSizes("eeedd")
	ln := len(XDigit)
	if !(len(r) == 5 && r[0] == 10 && r[1] == 10 && r[2] == ln && r[3] == ln && r[4] == ln) {
		t.Errorf("Got %v\n", r)
	}
}

func TestIton(t *testing.T) {
	// TODO: copy ruby noid tests into here
	ns := noidState{template: template{generator: 'z', template: "ed"}, sizes: generateSizes("ed")}
	fmt.Println(ns.iton(2901))
}

func TestChecksum(t *testing.T) {
	// TODO: add better test here using the expected checksums from ruby noid
	fmt.Println(checksum("abcdefg"))
}

func TestNoid(t *testing.T) {
	n, e := NewNoid(".r2dk")
	if e != nil {
		t.Errorf("Got error %v\n", e)
	}
	ids := []string{
		"00", "66", "11", "77", "22", "88", "33", "99", "44", "55",
	}
	for _, s := range ids {
		z := n.Mint()
		if z != s {
			t.Errorf("%v != %v\n", z, s)
		}
		fmt.Printf("%v\n", n)
	}
}
