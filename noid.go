/*

Implements the noid generator as specified in ....
We aim to be compatible with the ruby noid generator (..link...)
But, no promises are made.
In one respect we are different: our random generation does not rely
on a random number generator. Instead we will always generate the same
sequence of ids---however they will be scattered throughout the idspace.


The noid template string has the following format:

 <slug> '.' <generator code> <bins?> <template> <check?>

Where
     <slug> is any sequence of characters except a period (may be empty).
            Limited to a maximum of 1000 characters
     <generator code> is one of 'r', 'c', 'z'
     <bins?> is an optional sequence of '(' <positive integer> ')'
     <template> is a sequence of 'd' and 'e' characters (limit of 64 characters)
     <check?> is an optional 'k' character

The <bins> element is optional, but can only be present if the generator code is 'r'

Example format strings:

    id.sd
        -- produces id0, id1, id2, ..., id9, id10, id11, ...
    .reeddeeddek
    .zddddk
    .r(500)dek
    .cdek
*/

package main

import (
	"fmt"
	"regexp"
	"strings"
)

// Noid objects are not safe for simultaneous access
type Noid interface {
	Mint() string
	String() string
	Count() (int, int)
}

func NewNoid(template string) {
}

const (
	// XXX: Verify digits match with ruby-noid
	XDigit = "0123456789bcdfghjkmnpqrstvwxz"
	DDigit = "0123456789"
)

var (
	templateRegexp = regexp.MustCompile("^(.*)[.]([rsz])([de]+)(k?)$")
)

type noidState struct {
	template
	position int
	max int
	sizes []int // the base of the i-th digit from the right
	nBins int
	bins []int // only used if generator == 'r', the starting index for each bin
}

func (ns noidState) mint(n int) string {
	s := ns.slug + ns.iton(n)
	if ns.checkDigit {
		s += checksum(s)
	}
	return s
}

func (ns noidState) swizzle(n int) int {
	ln := ns.max / ns.nBins
	bin := n % ln
	offset := n / ln
	return ns.bins[bin] + offset
}

// Given an integer n inside the range of the template,
// return the corresponding id string
func (t *noidState) iton(n int) string {
	var buffer []byte = make([]byte, 0, len(t.sizes))

	for _, size := range t.sizes {
		value := n % size
		n /= size
		buffer = append(buffer, XDigit[value])
	}

	if t.generator == 'z' {
		size := t.sizes[len(t.sizes)-1]
		for n > 0 {
			value := n % size
			n /= size
			buffer = append(buffer, XDigit[value])
		}
	}

	if n > 0 {
		// error, should be 0
	}
	return string(reverse(buffer))
}

// This checksum function comes from the noid spec.
func checksum(s string) string {
	// Noid::XDIGIT[str.split('').map { |x| Noid::XDIGIT.index(x).to_i }.each_with_index.map { |n, idx| n*(idx+1) }.inject { |sum, n| sum += n }  % Noid::XDIGIT.length ]
	var sum int
	for i, c := range s {
		v := strings.IndexRune(XDigit, c)
		if v >= 0 {
			v *= i + 1
			sum += v
		}
	}
	return string(XDigit[sum%len(XDigit)])
}

func generateBins(nBins, maximum int) []int {
	result := make([]int, nBins)
	length := maximum / nBins
	extra := maximum % nBins
	v := 0
	for i := 0; i < nBins; i++ {
		result[i] = v
		v += length
		if i < extra {
			v++
		}
	}
	return result
}

func (ns noidState) maximum() int {
	if ns.generator == 'z' {
		return -1
	}
	var result int = 1
	for _, v := range ns.sizes {
		result *= v
	}
	return result
}

func reverse(z []byte) []byte {
	result := make([]byte, len(z))
	p := len(z) - 1
	for _, b := range z {
		result[p] = b
		p--
	}
	return result
}

type template struct {
	slug       string
	generator  rune
	template   string
	checkDigit bool
	valid      bool
}

func parseTemplate(t string) template {
	var result template

	matches := templateRegexp.FindStringSubmatch(t)

	fmt.Printf("%v\n", matches)

	if len(matches) == 0 {
		// error with match
		return result
	}

	result.slug = matches[1]
	result.generator = rune(matches[2][0])
	result.template = matches[3]
	result.checkDigit = matches[4] == "k"
	result.valid = true

	return result
}

func generateSizes(t string) []int {
	var result []int = make([]int, 0, len(t))

	for i := len(t) - 1; i >= 0; i-- {
		switch t[i] {
		case 'd':
			result = append(result, 10)
		case 'e':
			result = append(result, len(XDigit))
		}
	}

	return result
}
