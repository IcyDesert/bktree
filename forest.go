package bktree

import "sort"

// Forest partitions words by length into multiple BK-Trees.
// This enables natural Hamming support and Levenshtein length-based pruning.
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
