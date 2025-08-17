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
	if !slices.Equal(got, want) {
		t.Errorf("unexpected word list: got: %v want: %v", got, want)
	}
}

var validSimpleOutput = regexp.MustCompile(`(?m)^(V: \d+ .+\n)+(E: \d+ \d+\n)+$`)

func TestWriteSimple(t *testing.T) {
	b, err := randgraph.NewBinomial(5, 2, 0.5)
	if err != nil {
		t.Fatal(err)
	}
	b.Loops = true
	b.Multiedges = true
	b.Directed = true
	r := randgraph.New(b)

	buf := &bytes.Buffer{}
	writeSimple(buf, r)
	out := buf.String()
	if !validSimpleOutput.MatchString(out) {
		t.Errorf("malformed output:\n%v", out)
	}
}

func TestLabel(t *testing.T) {
	tests := []struct {
		labels []string
		id     int
		want   string
	}{
		{
			labels: []string{"A", "B"},
			id:     0,
			want:   "A",
		},
		{
			labels: []string{"A", "B"},
			id:     1,
			want:   "B",
		},
		{
			labels: []string{"A", "B"},
			id:     2,
			want:   "A2",
		},
		{
			labels: []string{"A", "B"},
			id:     5,
			want:   "B5",
		},
		{
			labels: nil,
			id:     2,
			want:   "2",
		},
		{
			labels: []string{},
			id:     5,
			want:   "5",
		},
	}
	for _, tt := range tests {
		got := label(tt.labels, tt.id)
		if got != tt.want {
			t.Errorf("unexpected label: got: %q, want: %q", got, tt.want)
		}
	}
}
