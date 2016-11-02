package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

func main() {
	matches := parseFile("/home/marcots/Downloads/matches.tsv", "pool")
	s := newStats(matches)
	g := newGraph(s)
	g.shortCircuit()
	g.removeDirectCycles()
	g.print()
	// r := g.newRank()
	// fmt.Println(r)
}

type match struct {
	winner string
	loser  string
}

func parseFile(path, game string) []match {
	f, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	c := csv.NewReader(f)
	c.Comma = '\t'
	c.FieldsPerRecord = 5

	var matches []match
	for {
		record, err := c.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		// Only the game we want.
		if record[1] != game ||
			// Only valid matches.
			record[4] != "True" ||
			// No matches with more than one person
			strings.Contains(record[2], ",") || strings.Contains(record[3], ",") {
			continue
		}
		matches = append(matches, match{record[2], record[3]})
	}
	return matches
}

type stats map[match]int

func newStats(matches []match) stats {
	s := make(stats)
	for _, m := range matches {
		value := 1
		if m.loser < m.winner {
			value = -1
			m.winner, m.loser = m.loser, m.winner
		}
		s[m] += value
	}
	return s
}

type graph map[string]map[string]bool

func newGraph(s stats) graph {
	g := make(graph)
	for m, v := range s {
		if v < 0 {
			g.add(m.loser, m.winner)
		}
		if v > 0 {
			g.add(m.winner, m.loser)
		}
	}
	return g
}

func (g graph) add(w, l string) {
	if g[w] == nil {
		g[w] = make(map[string]bool)
	}
	g[w][l] = true
}

func (g graph) print() {
	fmt.Println("digraph stats {")
	for w, ls := range g {
		for l, b := range ls {
			if !b {
				continue
			}
			fmt.Printf("  %v -> %v\n", w, l)
		}
	}
	fmt.Println("}")
}

func (g graph) shortCircuit() {
	nodes := make(map[string]bool)
	for w, ls := range g {
		nodes[w] = true
		for l := range ls {
			nodes[l] = true
		}
	}
	// Floyd Warshall.
	for middle := range nodes {
		for from := range nodes {
			for to := range nodes {
				if g[from][middle] && g[middle][to] {
					g[from][to] = true
				}
			}
		}
	}
}

func (g graph) removeDirectCycles() {
	for w, ls := range g {
		for l, b := range ls {
			if !b {
				continue
			}
			if g[l][w] {
				g[w][l] = false
				g[l][w] = false
			}
		}
	}
}

type rank map[int][]string

func (g graph) newRank() rank {
	levels := make(map[string]int)
	for w := range g {
		g.newRankVisit(levels, w, 1)
	}
	r := make(rank)
	for p, l := range levels {
		r[l] = append(r[l], p)
	}
	for _, ps := range r {
		sort.Strings(ps)
	}
	return r
}

func (g graph) newRankVisit(levels map[string]int, w string, level int) {
	if levels[w] < level {
		levels[w] = level
	}
	for l, b := range g[w] {
		if !b {
			continue
		}
		g.newRankVisit(levels, l, level+1)
	}
}
