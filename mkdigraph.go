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
//	-edges n
//		Maximum number of outgoing edges per vertex
//		(default 5).
//
//	-prob p
//		Probability of creating an edge. p is a float value
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
// vertex is omitted for vertices with no outgoing edges. If the
// -multiedge flag is provided, the output may contain duplicated
// two-field entries. However, single-field entries are always unique.
//
// If -prob=0, the graph may have isolated vertices. If -prob=1, the
// resulting digraph might have vertices with no outgoing edges,
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
	"iter"
	"log"
	"maps"
	"math/rand/v2"
	"os"
	"regexp"
	"slices"
	"strconv"
)

func main() {
	var err error

	log.SetFlags(0)
	log.SetPrefix("mkdigraph: ")

	numVertices := flag.Int("n", 25, "number of vertices")
	maxEdges := flag.Int("edges", 5, "maximum number of outgoing edges per vertex")
	edgeProb := flag.Float64("prob", 0.5, "probability of creating an edge")
	allowLoops := flag.Bool("loops", false, "allow loops")
	allowMultiEdges := flag.Bool("multiedges", false, "allow multiple edges")
	wordsFile := flag.String("words", "", "choose vertex labels from a words file")
	emitDOT := flag.Bool("dot", false, "emit DOT output")
	outFile := flag.String("o", "", "output file")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 0 {
		usage()
		os.Exit(2)
	}

	if *numVertices < 0 {
		log.Fatalf("invalid number of vertices: %v", *numVertices)
	}
	if *maxEdges < 0 {
		log.Fatalf("invalid maximum number of outgoing edges: %v", *maxEdges)
	}
	if *edgeProb < 0 || *edgeProb > 1 {
		log.Fatalf("invalid edge probability: %v", *edgeProb)
	}

	var words []string
	if *wordsFile != "" {
		words, err = readWords(*wordsFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	opts := digraphOpts{
		NumVertices:     *numVertices,
		MaxEdges:        *maxEdges,
		EdgeProb:        *edgeProb,
		AllowLoops:      *allowLoops,
		AllowMultiEdges: *allowMultiEdges,
		Labels:          words,
	}

	fout := os.Stdout
	if *outFile != "" {
		fout, err = os.Create(*outFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	if *emitDOT {
		printDOT(fout, opts)
	} else {
		printText(fout, opts)
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

func printText(w io.Writer, opts digraphOpts) {
	for tail, heads := range digraph(opts) {
		if len(heads) == 0 {
			fmt.Fprintf(w, "%v\n", tail)
		} else {
			for _, head := range heads {
				fmt.Fprintf(w, "%v %v\n", tail, head)
			}
		}
	}
}

func printDOT(w io.Writer, opts digraphOpts) {
	fmt.Fprintln(w, "digraph {")
	for tail, heads := range digraph(opts) {
		if len(heads) == 0 {
			fmt.Fprintf(w, "  %v\n", tail)
		} else {
			for _, head := range heads {
				fmt.Fprintf(w, "  %v -> %v\n", tail, head)
			}
		}
	}
	fmt.Fprintln(w, "}")
}

type digraphOpts struct {
	NumVertices     int
	MaxEdges        int
	EdgeProb        float64
	AllowLoops      bool
	AllowMultiEdges bool
	Labels          []string
}

// randIntN is set by tests to produce predictable results.
var randIntN = rand.IntN

func digraph(opts digraphOpts) iter.Seq2[string, []string] {
	if opts.NumVertices < 0 {
		panic("invalid number of vertices")
	}
	if opts.MaxEdges < 0 {
		panic("invalid maximum number of outgoing edges")
	}
	if opts.EdgeProb < 0 || opts.EdgeProb > 1 {
		log.Fatalf("invalid edge probability: %v", opts.EdgeProb)
	}

	return func(yield func(string, []string) bool) {
		for itail := range opts.NumVertices {
			tail := label(opts.Labels, itail)

			var start int
			if opts.AllowLoops {
				start = 0
			} else {
				if itail == opts.NumVertices-1 {
					// No possible heads.
					yield(tail, []string{})
					return
				}
				start = itail + 1
			}

			heads := make([]string, 0)
			selHeads := make(map[int]struct{})
			for range opts.MaxEdges {
				if rand.Float64() < opts.EdgeProb {
					ihead := start + randIntN(opts.NumVertices-start)
					if !opts.AllowMultiEdges {
						if _, found := selHeads[ihead]; found {
							continue
						}
						selHeads[ihead] = struct{}{}
					}
					heads = append(heads, label(opts.Labels, ihead))
				}
			}

			if !yield(tail, heads) {
				return
			}
		}
	}
}

func label(labels []string, n int) string {
	if len(labels) == 0 {
		return strconv.Itoa(n)
	}
	i := n % len(labels)
	if n < len(labels) {
		return labels[i]
	}
	return labels[i] + strconv.Itoa(n)
}
