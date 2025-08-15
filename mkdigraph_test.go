// Copyright (c) 2025 Roi Martin
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package main

import (
	"bytes"
	"regexp"
	"slices"
	"testing"

	"github.com/jroimartin/randgraph"
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

var validSimpleOutput = regexp.MustCompile(`(?m)^(\d+( \d+)?\n)+$`)

func TestWriteSimple(t *testing.T) {
	opts := randgraph.BinomialOpts{
		Vertices:   5,
		N:          2,
		P:          0.5,
		Loops:      true,
		Multiedges: true,
		Directed:   true,
	}
	b, err := randgraph.NewBinomial(opts)
	if err != nil {
		t.Fatal(err)
	}
	r := randgraph.New(b)

	buf := &bytes.Buffer{}
	writeSimple(buf, r.Graph())
	out := buf.String()
	if !validSimpleOutput.MatchString(out) {
		t.Errorf("malformed output:\n%v", out)
	}
}
