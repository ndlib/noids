package main

import (
	"fmt"
	"testing"
)

func TestParseTemplate(t *testing.T) {
	tests := []struct {
		format   string
		expected template
	}{{"no period", template{}},
		{"id.reeddk", template{slug: "id", generator: 'r', template: "eedd", checkDigit: true, valid: true}},
	}

	for _, test := range tests {
		p := parseTemplate(test.format)
		if p != test.expected {
			t.Errorf("expected %v, got %v\n", test.expected, p)
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

func TestGenerateBins(t *testing.T) {
	fmt.Printf("%v\n", generateBins(3, 10))
	fmt.Printf("%v\n", generateBins(3, 11))
	fmt.Printf("%v\n", generateBins(261, 290))
}

func TestSwizzle(t *testing.T) {
	ns := noidState{max: 10, nBins: 3, bins: generateBins(3, 10)}
	fmt.Println(ns.swizzle(8))
}
