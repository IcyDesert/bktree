package bktree

import (
	"encoding/gob"
	"io"
	"sort"
)

// Forest groups words by byte length into separate BK-trees.
// Query skips trees whose byte length differs from the query by more than
// maxDist. This works well for Hamming (which compares bytes) and for
// Levenshtein on ASCII text. For Levenshtein on non-ASCII text, byte
// length and rune length may differ; use BKTree if that matters.
//
// For distance functions not based on string length, use BKTree.
type Forest struct {
	trees map[int]*BKTree
	dist  DistanceFunc
}

// NewForest creates a new empty Forest with the given distance function.
// Panics if dist is nil.
func NewForest(dist DistanceFunc) *Forest {
	if dist == nil {
		panic("bktree: nil distance function")
	}
	return &Forest{
		trees: make(map[int]*BKTree),
		dist:  dist,
	}
}

// Add inserts a word into the forest, routing it to the tree
// keyed by its length.
func (f *Forest) Add(word string) {
	length := len(word)
	tree, ok := f.trees[length]
	if !ok {
		tree = New(f.dist)
		f.trees[length] = tree
	}
	tree.Add(word)
}

// Query returns all words in the forest within maxDist of the query word.
// It only queries trees whose length is in [len(word)-maxDist, len(word)+maxDist].
// Results are sorted by distance ascending.
func (f *Forest) Query(word string, maxDist int) []Result {
	queryLen := len(word)
	minLen := max(queryLen-maxDist, 0)
	maxLen := queryLen + maxDist

	var results []Result
	for length, tree := range f.trees {
		if length >= minLen && length <= maxLen {
			results = append(results, tree.Query(word, maxDist)...)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Distance != results[j].Distance {
			return results[i].Distance < results[j].Distance
		}
		return results[i].Word < results[j].Word
	})
	return results
}

// Exists returns true if there exists at least one word in the forest
// within maxDist of the query word. Short-circuits on first match.
func (f *Forest) Exists(word string, maxDist int) bool {
	queryLen := len(word)
	minLen := max(queryLen-maxDist, 0)
	maxLen := queryLen + maxDist

	for length, tree := range f.trees {
		if length >= minLen && length <= maxLen {
			if tree.Exists(word, maxDist) {
				return true
			}
		}
	}
	return false
}

// Save writes the forest topology to w using gob encoding.
// The distance function is not persisted.
func (f *Forest) Save(w io.Writer) error {
	// Encode as map[int]*Node to avoid gob encoding BKTree.dist (a function).
	nodes := make(map[int]*Node, len(f.trees))
	for length, tree := range f.trees {
		nodes[length] = tree.root
	}
	return gob.NewEncoder(w).Encode(nodes)
}

// LoadForest reads a forest topology from r using gob encoding.
func LoadForest(r io.Reader, dist DistanceFunc) (*Forest, error) {
	f := NewForest(dist)
	var nodes map[int]*Node
	if err := gob.NewDecoder(r).Decode(&nodes); err != nil {
		return nil, err
	}
	f.trees = make(map[int]*BKTree, len(nodes))
	for length, root := range nodes {
		tree := New(dist)
		tree.root = root
		f.trees[length] = tree
	}
	return f, nil
}
