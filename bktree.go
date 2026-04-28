package bktree

import (
	"encoding/gob"
	"io"
	"sort"
)

// DistanceFunc defines a metric distance function between two strings.
type DistanceFunc func(a, b string) int

// Result represents a single match from a Query.
type Result struct {
	Word     string
	Distance int
}

// BKTree is a Burkhard-Keller tree for fast approximate string matching.
type BKTree struct {
	root *Node
	dist DistanceFunc
}

// Node is a node in the BK-tree.
type Node struct {
	Word     string
	Children map[int]*Node
}

// New creates a new empty BKTree with the given distance function.
// Panics if dist is nil.
func New(dist DistanceFunc) *BKTree {
	if dist == nil {
		panic("bktree: nil distance function")
	}
	return &BKTree{
		dist: dist,
	}
}

// Add inserts a word into the tree.
func (t *BKTree) Add(word string) {
	if t.root == nil {
		t.root = &Node{
			Word:     word,
			Children: make(map[int]*Node),
		}
		return
	}
	t.root.add(word, t.dist)
}

func (n *Node) add(word string, dist DistanceFunc) {
	d := dist(word, n.Word)
	if child, ok := n.Children[d]; ok {
		child.add(word, dist)
	} else {
		n.Children[d] = &Node{
			Word:     word,
			Children: make(map[int]*Node),
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

func (n *Node) query(word string, maxDist int, dist DistanceFunc, results *[]Result) {
	d := dist(word, n.Word)
	if d <= maxDist {
		*results = append(*results, Result{Word: n.Word, Distance: d})
	}
	for childDist, child := range n.Children {
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

func (n *Node) exists(word string, maxDist int, dist DistanceFunc) bool {
	d := dist(word, n.Word)
	if d <= maxDist {
		return true
	}
	for childDist, child := range n.Children {
		if childDist >= d-maxDist && childDist <= d+maxDist {
			if child.exists(word, maxDist, dist) {
				return true
			}
		}
	}
	return false
}

// Save writes the tree topology to w using gob encoding.
// The distance function is not persisted.
func (t *BKTree) Save(w io.Writer) error {
	enc := gob.NewEncoder(w)
	if t.root == nil {
		return enc.Encode(false) // nil marker
	}
	if err := enc.Encode(true); err != nil {
		return err
	}
	return enc.Encode(t.root)
}

// Load reads a tree topology from r using gob encoding.
func Load(r io.Reader, dist DistanceFunc) (*BKTree, error) {
	t := New(dist)
	dec := gob.NewDecoder(r)
	var hasRoot bool
	if err := dec.Decode(&hasRoot); err != nil {
		return nil, err
	}
	if !hasRoot {
		return t, nil
	}
	if err := dec.Decode(&t.root); err != nil {
		return nil, err
	}
	return t, nil
}
