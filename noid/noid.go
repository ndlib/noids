/*

Implements the noid generator as specified in ....
We aim to be compatible with the ruby noid generator (..link...)
But, no promises are made.
In one respect we are different: our random generation does not rely
on a random number generator. Instead we will always generate the same
sequence of ids---however they will be scattered throughout the idspace.


The noid template string has the following format:

 <slug> '.' <generator> <bins?> <template> <check?>

Where
     <slug> is any sequence of characters (may be empty).
     <generator> is one of 'r', 'c', 'z'
     <bins?> is an optional sequence of decimal digits
     <digits> is a sequence of 'd' and 'e' characters
     <check?> is an optional 'k' character

The <bins> element is optional, but can only be present if the generator is 'r'

Example format strings:

    id.sd
        -- produces id0, id1, id2, ..., id9, id10, id11, ...
    .reeddeeddek
    .zddddk
    .r(500)dek
    .cdek

We extend the template string with state information to completely describe
a noid minter as a string. The format is
    <noid template> '+' <number ids minted>

    <id count> is a decimal integer >= 0. it is taken to be 0 if omitted.
*/

package noid

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	TemplateError = errors.New("Bad Template String")
)

const (
	XDigit = "0123456789bcdfghjkmnpqrstvwxz"
)

// Noid objects are not safe for simultaneous access
type Noid interface {
	Mint() string
	String() string
	Count() (int, int)
}

func NewNoid(template string) (Noid, error) {
	result := &noidState{}
	t, ok := parseTemplate(template)
	if !ok {
		return result, TemplateError
	}

	result.template = t
	result.sizes = generateSizes(t.template)
	result.max = result.maximum()
	if result.max == -1 || (result.pos <= result.max) {
		result.position = result.pos
	}
	if result.generator == 'r' {
		var bincount int = 293
		if result.binCount > 0 && result.binCount < 10000 {
			bincount = result.binCount
		}
		result.r = newRandomState(bincount, result.max)
	}
	return result, nil
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

func (ns *noidState) Mint() string {
	if ns.max >= 0 && ns.position >= ns.max {
		// pool exhausted
		return ""
	}
	var id int = ns.position
	ns.position++
	if ns.generator == 'r' {
		id = ns.r.swizzle(id)
	}
	return ns.mint(id)
}

// returns the number of ids minted so far
// and the maximum number of ids in the pool
// (the maximum is == -1 iff the pool is infinite)
func (ns *noidState) Count() (int, int) {
	return ns.position, ns.max
}

// returns the noid state as an "extended" template:
// the template string followed by the current position (i.e. the next id to be used
func (ns *noidState) String() string {
	return ns.template.String() + fmt.Sprintf("+%d", ns.position)
}

type noidState struct {
	template
	position int
	max      int
	sizes    []int        // the base of the i-th digit from the right
	r        *randomState // non-nil iff generator == 'r'
}

func (ns noidState) mint(n int) string {
	s := ns.slug + ns.iton(n)
	if ns.checkDigit {
		s += checksum(s)
	}
	return s
}

// This is complicated since we want to use the same binning method as the ruby
// noid gem. Yet the ruby noid object makes too many bins too large:
// 1. it computes the size of the bins by using the floor function and adding 1
//    this will usually be too large
// 2. it makes every bin except the last one have this size. Hence, the bins
//    are all too large (Except for the last bin, which is too small), and
//    there may be fewer bins than asked for due to bins being systematically
//    too large
// We assign numbers from each bin in sequence, but since the last bin is so
// much smaller, it is exhausted first. The `cutoff` value tells us when the
// smaller bin has been exhausted so we do not assign any more numbers to it.
type randomState struct {
	binList   []int // list of bin starting positions
	nBins     int   // actual number of bins
	cutoff    int   // number which exhausts the smaller last bin
	smallSize int   // the size of the small bin
	askedFor  int   // the number of bins asked for
}

func newRandomState(binCount, countermax int) *randomState {
	binSize := (countermax / binCount) + 1
	smallBinSize := countermax % binSize
	nBins := (countermax / binSize) + 1

	bins := make([]int, nBins, nBins)
	t := 0
	for i := range bins {
		bins[i] = t
		t += binSize
	}

	return &randomState{
		binList:   bins,
		nBins:     nBins,
		cutoff:    nBins * smallBinSize,
		askedFor:  binCount,
		smallSize: smallBinSize,
	}
}

func (rs randomState) swizzle(n int) int {
	var bin, offset int
	if n < rs.cutoff {
		bin = n % rs.nBins
		offset = n / rs.nBins
	} else {
		bin = (n - rs.cutoff) % (rs.nBins - 1)
		offset = (n - rs.cutoff) / (rs.nBins - 1)
		offset += rs.smallSize
	}
	return rs.binList[bin] + offset
}

func (rs randomState) invSwizzle(n int) int {
	var bin, offset int
	bin = rs.nBins - 1
	for i := range rs.binList {
		if rs.binList[i] > n {
			bin = i - 1
			break
		}
	}
	offset = n - rs.binList[bin]
	if offset < rs.smallSize {
		return rs.nBins*offset + bin
	}
	return rs.cutoff + (rs.nBins-1)*(offset-rs.smallSize) + bin
}

// Given an integer n inside the range of the template,
// return the corresponding id string
func (t noidState) iton(n int) string {
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
	binCount   int
	template   string
	checkDigit bool
	pos        int
}

var (
	templateRegexp = regexp.MustCompile(`^(.*)\.([rsz])(\d*)([de]+)(k?)(\+\d+)?$`)
)

func parseTemplate(t string) (template, bool) {
	var result template

	matches := templateRegexp.FindStringSubmatch(t)

	if len(matches) == 0 {
		// error with match
		return result, false
	}

	result.slug = matches[1]
	result.generator = rune(matches[2][0])
	if len(matches[3]) > 0 {
		result.binCount, _ = strconv.Atoi(matches[3])
	}
	result.template = matches[4]
	result.checkDigit = matches[5] == "k"
	if len(matches[6]) > 1 {
		result.pos, _ = strconv.Atoi(matches[6])
	}

	if result.binCount > 0 && result.generator != 'r' {
		return result, false
	}

	return result, true
}

func (t template) String() string {
	s := fmt.Sprintf("%s.%c", t.slug, t.generator)
	if t.generator == 'r' && t.binCount > 0 && t.binCount != 293 {
		s += fmt.Sprintf("%d", t.binCount)
	}
	s += t.template
	if t.checkDigit {
		s += "k"
	}
	return s
}
