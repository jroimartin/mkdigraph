// Copyright (c) 2025 Roi Martin
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

// Mkdigraph generates random directed graphs. It uses a stream
// generator, so the full graph is not stored in memory, which makes
// it possible to generate graphs of arbitrary size.
//
// Usage:
//
//	mkdigraph [flags]
//
// The flags are:
//
//	-n n
//		Number of vertices (default 25).
//
//	-trials n
//		Number of edge creation trials per vertex (default 5).
//
//	-prob p
//		Success probability for each trial. p is a float value
//		between 0 and 1 (default 0.5).
//
//	-loops
//		Allow loops.
//
//	-multiedges
//		Allow multiple edges. If true, two or more edges with
//		the same tail vertex and the same head vertex are
//		allowed.
//
//	-words path
//		Choose vertex labels from a words file.
//
//	-dot
//		Emit DOT output.
//
//	-o output
//		Output file. The default is the standard output.
//
// Unless the -dot flag is specified, it prints the graph in the
// format:
//
//	tail [head]
//	...
//
// Each line specifies an edge, represented by two fields separated by
// a space character. The first field is the label of the tail vertex
// and the second field is the label of the head vertex. The head
// vertex is omitted for vertices with no outgoing edges.
//
// If -prob=0, the graph may have isolated vertices. Even if -prob=1,
// the resulting digraph might have vertices with no outgoing edges,
// depending on the -loops and -multiedges flags.
//
// If the -words flag is specified, vertex labels are selected from
// the provided words file. Each label is sanitized by removing any
// characters that match the regular expression '[^a-zA-Z]'. If the
// number of vertices exceeds the number of available labels, then
// duplicated labels are suffixed with the vertex number.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"maps"
	"os"
	"regexp"
	"slices"

	"github.com/jroimartin/randgraph"
)

func main() {
	var err error

	log.SetFlags(0)
	log.SetPrefix("mkdigraph: ")

	vertices := flag.Int("n", 25, "number of vertices")
	trials := flag.Int("trials", 5, "number of edge creation trials per vertex")
	prob := flag.Float64("prob", 0.5, "success probability for each trial")
	loops := flag.Bool("loops", false, "allow loops")
	multiedges := flag.Bool("multiedges", false, "allow multiple edges")
	wordsFile := flag.String("words", "", "choose vertex labels from a words file")
	emitDOT := flag.Bool("dot", false, "emit DOT output")
	outFile := flag.String("o", "", "output file")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 0 {
		usage()
		os.Exit(2)
	}

	var words []string
	if *wordsFile != "" {
		words, err = readWords(*wordsFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	opts := randgraph.BinomialOpts{
		Vertices:   *vertices,
		N:          *trials,
		P:          *prob,
		Loops:      *loops,
		Multiedges: *multiedges,
		Directed:   true,
		Labels:     words,
	}
	b, err := randgraph.NewBinomial(opts)
	if err != nil {
		log.Fatal(err)
	}
	r := randgraph.New(b)

	fout := os.Stdout
	if *outFile != "" {
		fout, err = os.Create(*outFile)
		if err != nil {
			log.Fatal(err)
		}
		defer fout.Close()
	}

	if *emitDOT {
		r.WriteDOT(fout)
	} else {
		writeSimple(fout, r.Graph())
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: mkdigraph [flags]")
	flag.PrintDefaults()
}

var invalidChars = regexp.MustCompile(`[^a-zA-Z]`)

func readWords(name string) ([]string, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	s := bufio.NewScanner(f)

	words := make(map[string]struct{})
	for s.Scan() {
		word := invalidChars.ReplaceAllString(s.Text(), "")
		if word != "" {
			words[word] = struct{}{}
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	return slices.Collect(maps.Keys(words)), nil
}

func writeSimple(w io.Writer, ch <-chan randgraph.Edge) {
	for edge := range ch {
		if edge.V1 != "" {
			fmt.Fprintf(w, "%v %v\n", edge.V0, edge.V1)
		} else {
			fmt.Fprintf(w, "%v\n", edge.V0)
		}
	}
}
