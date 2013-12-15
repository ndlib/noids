// noid-tool
//
// a command line utility to decode and validate noids.
//
// usage:
//
// noid-tool -i <template list>
// noid-tool <template> [<noid list>]
//
// Options:
//  -i      print information about the given templates
//  -h      print this help text and exit
//
// when the -i option is given, noid-tool will display information about  the
// given templates to stdout. Otherwise noid-tool will output the sequence
// number of each noid with respect to the given template. If the noid is
// invalid, it will have a sequence number of -1. If no noids are given on
// the command line, noid-tool will take them from stdin, with each noid on its
// own line.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/dbrower/noids/noid"
)

func usage() {
    fmt.Println(`
noid-tool

a command line utility to decode and validate noids.

usage:

noid-tool -i <template list>
noid-tool <template> [<noid list>]

Options:
 -i      print information about the given templates
 -h      print this help text and exit

when the -i option is given, noid-tool will display information about  the
given templates to stdout. Otherwise noid-tool will output the sequence
number of each noid with respect to the given template. If the noid is
invalid, it will have a sequence number of -1. If no noids are given on
the command line, noid-tool will take them from stdin, with each noid on its
own line.`)
}

type Liner interface {
	Line() (string, error) // returns the next line, or an error
}

type arrayLiner struct {
	lines []string
	next  int
}

var LinesDone error = errors.New("EOF")

func NewArrayLiner(a []string) Liner {
	return &arrayLiner{lines: a}
}

func (a *arrayLiner) Line() (string, error) {
	if a.next >= len(a.lines) {
		return "", LinesDone
	}
	a.next++
	return a.lines[a.next - 1], nil
}

type readLiner struct {
	*bufio.Scanner
}

func NewReadLiner(r io.Reader) Liner {
	return &readLiner{bufio.NewScanner(r)}
}

func (r *readLiner) Line() (string, error) {
	if !r.Scan() {
		return "", LinesDone
	}
	return r.Text(), nil
}

func printTemplateInfo(t string) {
	n := getNoid(t)
	if n == nil {
		return
	}
	fmt.Printf("%s\tValid\n", t)
}

func getNoid(template string) *noid.Noid {
	var n noid.Noid
	n, err := noid.NewNoid(template)
	if err != nil {
		fmt.Printf("%s\tInvalid Template: %s\n", template, err.Error())
		return nil
	}
	return &n
}

func validateIds(n *noid.Noid, z Liner) {
	for {
		line, err := z.Line()
		if err != nil {
			break
		}
		var idx int = (*n).Index(line)
		fmt.Printf("%d\t%s\n", idx, line)
	}
}

func main() {
	var informational bool
        flag.Usage = usage

	flag.BoolVar(&informational, "i", false, "Display template information")
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("Usage:")
		flag.PrintDefaults()
		return
	}

	if informational {
		for _, t := range args {
			printTemplateInfo(t)
		}
	} else {
		var n *noid.Noid = getNoid(args[0])
		if n == nil {
			return
		}
		var z Liner
		if len(args) == 1 {
			z = NewReadLiner(os.Stdin)
		} else {
			z = NewArrayLiner(args[1:])
		}
		validateIds(n, z)
	}
}
