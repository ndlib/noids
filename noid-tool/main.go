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
// When the -i option is given, noid-tool will display information about the
// given templates to stdout. Otherwise noid-tool will output the sequence
// number of each noid with respect to the given template. Invalid noids
// will have a sequence number of -1. If no noids are given on the command
// line, noid-tool will take them from stdin, with each noid on its
// own line.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/ndlib/noids/noid"
)

func usage() {
	fmt.Println(`noid-tool

a command line utility to decode and validate noids.

usage:

noid-tool info [<template list>]
noid-tool valid <template> [<noid list>]
noid-tool generate <template> [<number list>]

Options:
 -h      print this help text and exit

'info'
    Display information about the given templates to stdout.

'valid'
    Output the sequence number of each id with respect to the given template.
    Invalid ids will have a sequence number of -1.

'generate'
    Output the ids associated to each given sequence number.


If no ids are given on the command line, they will be taken from stdin,
with each id on its own line.
`)
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

func printTemplateInfo(t string) {
	n := getNoid(t)
	if n == nil {
		return
	}
	pos, max := (*n).Count()
	var used float64 = 0
	if max != -1 {
		used = (float64(pos) / float64(max)) * 100
	}
	fmt.Printf("%s\tValid, Pos = %d, Max = %d, %0.2f%% used\n", t, pos, max, used)
}

func validateId(n *noid.Noid, id string) {
	var idx int = (*n).Index(id)
	fmt.Printf("%d\t%s\n", idx, id)
}

func generate(n *noid.Noid, index string) {
	var i int
	var result string
	var err error
	if i, err = strconv.Atoi(index); err == nil {
		_, max := (*n).Count()
		if max == -1 || i < max {
			(*n).AdvanceTo(i)
			result = (*n).Mint()
		}
	}
	fmt.Printf("%s\t%s\n", index, result)
}

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if len(args) == 0 {
		flag.Usage()
		return
	}

	// all of the subcommands are essentially line oriented.
	// Reduce each one to its appropriate processing function, f
	var f func(string)
	var rest []string
	switch args[0] {
	case "info":
		f = printTemplateInfo
		rest = args[1:]
	case "valid", "generate":
		if len(args) == 1 {
			flag.Usage()
			return
		}
		var n *noid.Noid
		if n = getNoid(args[1]); n == nil {
			return
		}
		if args[0] == "valid" {
			f = func(s string) { validateId(n, s) }
		} else {
			f = func(s string) { generate(n, s) }
		}
		rest = args[2:]
	default:
		flag.Usage()
		return
	}

	// take input from either stdin or the command line
	if len(rest) == 0 {
		b := bufio.NewScanner(os.Stdin)
		for b.Scan() {
			f(b.Text())
		}
	} else {
		for _, id := range rest {
			f(id)
		}
	}
}
