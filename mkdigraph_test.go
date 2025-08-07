// Copyright (c) 2025 Roi Martin
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"maps"
	"math/rand/v2"
	"regexp"
	"slices"
	"testing"
)

func TestReadWords(t *testing.T) {
	want := []string{"FirstWord", "SecondWord"}
	got, err := readWords("testdata/words")
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	slices.Sort(want)
	if slices.Compare(got, want) != 0 {
		t.Errorf("unexpected word list: got: %v want: %v", got, want)
	}
}

var validText = regexp.MustCompile(`(?m)^(\d+( \d+)?\n)+$`)

func TestPrintText(t *testing.T) {
	oldRandIntN := randIntN
	randIntN = testRand().IntN
	t.Cleanup(func() { randIntN = oldRandIntN })

	buf := &bytes.Buffer{}
	opts := digraphOpts{
		NumVertices:     5,
		MaxEdges:        2,
		EdgeProb:        0.5,
		AllowLoops:      true,
		AllowMultiEdges: true,
	}
	printText(buf, opts)
	s := buf.String()
	if !validText.MatchString(s) {
		t.Errorf("malformed output:\n%v", s)
	}
}

var validDOT = regexp.MustCompile(`(?m)^digraph {\n(  \d+( -> \d+)?\n)+}$`)

func TestPrintDOT(t *testing.T) {
	oldRandIntN := randIntN
	randIntN = testRand().IntN
	t.Cleanup(func() { randIntN = oldRandIntN })

	buf := &bytes.Buffer{}
	opts := digraphOpts{
		NumVertices:     5,
		MaxEdges:        2,
		EdgeProb:        0.5,
		AllowLoops:      true,
		AllowMultiEdges: true,
	}
	printDOT(buf, opts)
	s := buf.String()
	if !validDOT.MatchString(s) {
		t.Errorf("malformed output:\n%v", s)
	}
}

func TestDigraph(t *testing.T) {
	tests := []struct {
		name string
		opts digraphOpts
		want map[string][]string
	}{
		{
			name: "1 vertex with loops and multiedges",
			opts: digraphOpts{
				NumVertices:     1,
				MaxEdges:        2,
				EdgeProb:        1,
				AllowLoops:      true,
				AllowMultiEdges: true,
			},
			want: map[string][]string{
				"0": {"0", "0"},
			},
		},
		{
			name: "1 vertex with multiedges",
			opts: digraphOpts{
				NumVertices:     1,
				MaxEdges:        2,
				EdgeProb:        1,
				AllowLoops:      false,
				AllowMultiEdges: true,
			},
			want: map[string][]string{
				"0": {},
			},
		},
		{
			name: "1 vertex",
			opts: digraphOpts{
				NumVertices:     1,
				MaxEdges:        2,
				EdgeProb:        1,
				AllowLoops:      false,
				AllowMultiEdges: false,
			},
			want: map[string][]string{
				"0": {},
			},
		},
		{
			name: "2 vertices with multiedges",
			opts: digraphOpts{
				NumVertices:     2,
				MaxEdges:        2,
				EdgeProb:        1,
				AllowLoops:      false,
				AllowMultiEdges: true,
			},
			want: map[string][]string{
				"0": {"1", "1"},
				"1": {},
			},
		},
		{
			name: "2 vertices",
			opts: digraphOpts{
				NumVertices:     2,
				MaxEdges:        2,
				EdgeProb:        1,
				AllowLoops:      false,
				AllowMultiEdges: false,
			},
			want: map[string][]string{
				"0": {"1"},
				"1": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldRandIntN := randIntN
			randIntN = testRand().IntN
			t.Cleanup(func() { randIntN = oldRandIntN })

			g := maps.Collect(digraph(tt.opts))

			if len(g) != len(tt.want) {
				t.Errorf("unexpected number of vertices: got: %v want: %v", len(g), len(tt.want))
			}

			for gotTail, gotHeads := range g {
				wantHeads, ok := tt.want[gotTail]
				if !ok {
					t.Errorf("could not find vertex %v", gotTail)
					continue
				}
				slices.Sort(gotHeads)
				slices.Sort(wantHeads)
				if slices.Compare(gotHeads, wantHeads) != 0 {
					t.Errorf("unexpected outgoing edges in vertex %v: got: %v want: %v", gotTail, gotHeads, wantHeads)
				}
			}
		})
	}
}

func TestDigraph_edgeless(t *testing.T) {
	oldRandIntN := randIntN
	randIntN = testRand().IntN
	t.Cleanup(func() { randIntN = oldRandIntN })

	const numVertices = 5

	opts := digraphOpts{
		NumVertices: numVertices,
		MaxEdges:    0,
		EdgeProb:    0,
	}
	g := maps.Collect(digraph(opts))

	if len(g) != numVertices {
		t.Errorf("unexpected number of vertices: got: %v want: %v", len(g), numVertices)
	}

	for tail, heads := range g {
		if len(heads) != 0 {
			t.Errorf("vertex %v has %v outgoing edges", tail, len(heads))
		}
	}
}

func TestDigraph_order_zero(t *testing.T) {
	oldRandIntN := randIntN
	randIntN = testRand().IntN
	t.Cleanup(func() { randIntN = oldRandIntN })

	opts := digraphOpts{
		NumVertices: 0,
		MaxEdges:    1,
		EdgeProb:    1,
	}
	g := maps.Collect(digraph(opts))

	if len(g) != 0 {
		t.Errorf("expected an order-zero graph: got %v vertices", len(g))
	}
}

func TestLabel(t *testing.T) {
	tests := []struct {
		labels []string
		n      int
		want   string
	}{
		{
			labels: []string{"A", "B"},
			n:      0,
			want:   "A",
		},
		{
			labels: []string{"A", "B"},
			n:      1,
			want:   "B",
		},
		{
			labels: []string{"A", "B"},
			n:      2,
			want:   "A2",
		},
		{
			labels: []string{"A", "B"},
			n:      5,
			want:   "B5",
		},
		{
			labels: nil,
			n:      2,
			want:   "2",
		},
		{
			labels: []string{},
			n:      5,
			want:   "5",
		},
	}
	for _, tt := range tests {
		got := label(tt.labels, tt.n)
		if got != tt.want {
			t.Errorf("unexpected label: got: %q want: %q", got, tt.want)
		}
	}
}

func testRand() *rand.Rand {
	return rand.New(rand.NewPCG(1, 2))
}

func BenchmarkDigraph(b *testing.B) {
	benchmarks := []struct {
		name string
		opts digraphOpts
	}{
		{
			name: "1000 100 0.5",
			opts: digraphOpts{
				NumVertices:     1000,
				MaxEdges:        100,
				EdgeProb:        0.5,
				AllowLoops:      false,
				AllowMultiEdges: false,
			},
		},
		{
			name: "1000 100 0.5 with loops",
			opts: digraphOpts{
				NumVertices:     1000,
				MaxEdges:        100,
				EdgeProb:        0.5,
				AllowLoops:      true,
				AllowMultiEdges: false,
			},
		},
		{
			name: "1000 100 0.5 with multiedges",
			opts: digraphOpts{
				NumVertices:     1000,
				MaxEdges:        100,
				EdgeProb:        0.5,
				AllowLoops:      false,
				AllowMultiEdges: true,
			},
		},
		{
			name: "1000 100 0.5 with loops and multiedges",
			opts: digraphOpts{
				NumVertices:     1000,
				MaxEdges:        100,
				EdgeProb:        0.5,
				AllowLoops:      true,
				AllowMultiEdges: true,
			},
		},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			for b.Loop() {
				for range digraph(bb.opts) {
				}
			}
		})
	}
}
