package bktree

import "sort"

// DistanceFunc defines a metric distance function between two strings.
type DistanceFunc func(a, b string) int

// Result represents a single match from a Query.
type Result struct {
	Word     string
	Distance int
}

// BKTree is a Burkhard-Keller tree for fast approximate string matching.
type BKTree struct {
	root *node
	dist DistanceFunc
}

type node struct {
	word     string
	children map[int]*node
}

// New creates a new empty BKTree with the given distance function.
func New(dist DistanceFunc) *BKTree {
	return &BKTree{
		dist: dist,
	}
}

// Add inserts a word into the tree.
func (t *BKTree) Add(word string) {
	if t.root == nil {
		t.root = &node{
			word:     word,
			children: make(map[int]*node),
		}
		return
	}
	t.root.add(word, t.dist)
}

func (n *node) add(word string, dist DistanceFunc) {
	d := dist(word, n.word)
	if child, ok := n.children[d]; ok {
		child.add(word, dist)
	} else {
		n.children[d] = &node{
			word:     word,
			children: make(map[int]*node),
		}
	}
}

// Query returns all words in the tree within maxDist of the query word,
// sorted by distance ascending.
func (t *BKTree) Query(word string, maxDist int) []Result {
	if t.root == nil {
		return nil
	}
	var results []Result
	t.root.query(word, maxDist, t.dist, &results)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Distance != results[j].Distance {
			return results[i].Distance < results[j].Distance
		}
		return results[i].Word < results[j].Word
	})
	return results
}

func (n *node) query(word string, maxDist int, dist DistanceFunc, results *[]Result) {
	d := dist(word, n.word)
	if d <= maxDist {
		*results = append(*results, Result{Word: n.word, Distance: d})
	}
	for childDist, child := range n.children {
		if childDist >= d-maxDist && childDist <= d+maxDist {
			child.query(word, maxDist, dist, results)
		}
	}
}

// Exists returns true if there exists at least one word in the tree
// within maxDist of the query word. It short-circuits on the first match.
func (t *BKTree) Exists(word string, maxDist int) bool {
	if t.root == nil {
		return false
	}
	return t.root.exists(word, maxDist, t.dist)
}

func (n *node) exists(word string, maxDist int, dist DistanceFunc) bool {
	d := dist(word, n.word)
	if d <= maxDist {
		return true
	}
	for childDist, child := range n.children {
		if childDist >= d-maxDist && childDist <= d+maxDist {
			if child.exists(word, maxDist, dist) {
				return true
			}
		}
	}
	return false
}
