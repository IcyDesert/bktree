package bktree

import (
	"reflect"
	"testing"
)

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"a", "a", 0},
		{"a", "b", 1},
		{"abc", "abc", 0},
		{"kitten", "sitting", 3},
		{"saturday", "sunday", 3},
		{"book", "books", 1},
		{"boak", "book", 1},
		{"boak", "boo", 2},
	}
	for _, tt := range tests {
		got := Levenshtein(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
		// Symmetry
		gotRev := Levenshtein(tt.b, tt.a)
		if gotRev != tt.want {
			t.Errorf("Levenshtein(%q, %q) = %d, want %d (symmetry)", tt.b, tt.a, gotRev, tt.want)
		}
	}
}

func TestHamming(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "a", 0},
		{"a", "b", 1},
		{"karolin", "kathrin", 3},
		{"karolin", "kerstin", 3},
		{"1011101", "1001001", 2},
	}
	for _, tt := range tests {
		got := Hamming(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Hamming(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestHammingPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Hamming with unequal lengths should panic")
		}
	}()
	Hamming("abc", "ab")
}

func TestBKTreeQuery(t *testing.T) {
	tree := New(Levenshtein)
	words := []string{"book", "books", "cake", "boo", "cape", "cart"}
	for _, w := range words {
		tree.Add(w)
	}

	tests := []struct {
		word    string
		maxDist int
		want    []Result
	}{
		{
			word:    "boak",
			maxDist: 1,
			want:    []Result{{Word: "book", Distance: 1}},
		},
		{
			word:    "boak",
			maxDist: 2,
			want: []Result{
				{Word: "book", Distance: 1},
				{Word: "boo", Distance: 2},
				{Word: "books", Distance: 2},
			},
		},
		{
			word:    "xyz",
			maxDist: 3,
			want:    []Result{{Word: "boo", Distance: 3}},
		},
	}

	for _, tt := range tests {
		got := tree.Query(tt.word, tt.maxDist)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Query(%q, %d) = %v, want %v", tt.word, tt.maxDist, got, tt.want)
		}
	}
}

func TestBKTreeExists(t *testing.T) {
	tree := New(Levenshtein)
	tree.Add("hello")
	tree.Add("world")

	if !tree.Exists("hello", 0) {
		t.Error("Exists(hello, 0) should be true")
	}
	if !tree.Exists("hella", 1) {
		t.Errorf("Exists(hella, 1) should be true")
	}
	if tree.Exists("xyz", 2) {
		t.Error("Exists(xyz, 2) should be false")
	}
}

func TestBKTreeEmpty(t *testing.T) {
	tree := New(Levenshtein)
	if tree.Query("anything", 5) != nil {
		t.Error("Empty tree Query should return nil")
	}
	if tree.Exists("anything", 5) {
		t.Error("Empty tree Exists should return false")
	}
}

func TestBKTreeExactMatch(t *testing.T) {
	tree := New(Levenshtein)
	tree.Add("exact")

	got := tree.Query("exact", 0)
	want := []Result{{Word: "exact", Distance: 0}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Exact match Query = %v, want %v", got, want)
	}

	if !tree.Exists("exact", 0) {
		t.Error("Exact match Exists should be true")
	}
}

func TestBKTreeHamming(t *testing.T) {
	tree := New(Hamming)
	words := []string{"000", "001", "010", "111"}
	for _, w := range words {
		tree.Add(w)
	}

	got := tree.Query("001", 1)
	want := []Result{
		{Word: "001", Distance: 0},
		{Word: "000", Distance: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("BKTree Hamming Query = %v, want %v", got, want)
	}

	if !tree.Exists("001", 0) {
		t.Error("Exists(001, 0) should be true")
	}
	if tree.Exists("011", 0) {
		t.Error("Exists(011, 0) should be false")
	}
}

func TestForestHamming(t *testing.T) {
	forest := NewForest(Hamming)
	words := []string{"000", "001", "010", "111"}
	for _, w := range words {
		forest.Add(w)
	}

	// Hamming distance 1 from "001"
	got := forest.Query("001", 1)
	want := []Result{
		{Word: "001", Distance: 0},
		{Word: "000", Distance: 1},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("Forest Hamming Query = %v, want %v", got, want)
	}

	// Query a length-3 word not in the forest (safe, same length)
	if forest.Query("011", 0) != nil {
		t.Error("Query for non-existent word should return nil")
	}
}

func TestForestLevenshtein(t *testing.T) {
	forest := NewForest(Levenshtein)
	words := []string{"book", "books", "cake", "boo", "cape"}
	for _, w := range words {
		forest.Add(w)
	}

	got := forest.Query("boak", 2)

	// book(1), boo(2), books(2)
	if len(got) != 3 {
		t.Fatalf("Expected 3 results, got %d: %v", len(got), got)
	}

	// Check distances
	expected := map[string]int{
		"book":  1,
		"boo":   2,
		"books": 2,
	}
	for _, r := range got {
		wantDist, ok := expected[r.Word]
		if !ok {
			t.Errorf("Unexpected word %q in results", r.Word)
			continue
		}
		if r.Distance != wantDist {
			t.Errorf("Distance for %q = %d, want %d", r.Word, r.Distance, wantDist)
		}
	}

	// Verify sorted order
	for i := 1; i < len(got); i++ {
		if got[i].Distance < got[i-1].Distance {
			t.Error("Results not sorted by distance")
		}
	}
}

func TestForestExists(t *testing.T) {
	forest := NewForest(Levenshtein)
	forest.Add("hello")
	forest.Add("world")

	if !forest.Exists("hello", 0) {
		t.Error("Exists(hello, 0) should be true")
	}
	if !forest.Exists("hallo", 1) {
		t.Error("Exists(hallo, 1) should be true")
	}
	if forest.Exists("xyz", 2) {
		t.Error("Exists(xyz, 2) should be false")
	}
}

func TestForestEmpty(t *testing.T) {
	forest := NewForest(Levenshtein)
	if forest.Query("anything", 5) != nil {
		t.Error("Empty forest Query should return nil")
	}
	if forest.Exists("anything", 5) {
		t.Error("Empty forest Exists should return false")
	}
}

// TestBKTreeDuplicateInsertion tests that duplicate words are stored
// (BK-tree nodes with distance 0 can have children).
func TestBKTreeDuplicateInsertion(t *testing.T) {
	tree := New(Levenshtein)
	tree.Add("test")
	tree.Add("test")
	tree.Add("tent")

	got := tree.Query("test", 0)
	if len(got) != 2 {
		t.Errorf("Expected 2 exact matches for 'test', got %d", len(got))
	}
}

func TestLevenshteinUnicode(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		// Each CJK character is a single rune; byte length differs.
		{"你好", "你好", 0},
		{"你好", "您好", 1},
		{"中", "国", 1},
		{"hello", "你好", 5}, // completely different, 5 runes vs 2 runes
		{"café", "cafe", 1}, // é is one rune, e is one rune
		{"日本語", "日本語", 0},
		{"日本語", "日本话", 1},
	}
	for _, tt := range tests {
		got := Levenshtein(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("Levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestBKTreeNilDistance(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("New(nil) should panic")
		}
	}()
	New(nil)
}

func TestForestNilDistance(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewForest(nil) should panic")
		}
	}()
	NewForest(nil)
}

// TestForestSortTieBreaker verifies that when multiple results have the
// same distance, they are sorted lexicographically by word.
func TestForestSortTieBreaker(t *testing.T) {
	forest := NewForest(Levenshtein)
	// All of these are edit distance 1 from "apple" (last letter substituted).
	words := []string{"apply", "applx", "appla"}
	for _, w := range words {
		forest.Add(w)
	}

	got := forest.Query("apple", 1)
	if len(got) != 3 {
		t.Fatalf("Expected 3 results, got %d: %v", len(got), got)
	}

	want := []Result{
		{Word: "appla", Distance: 1},
		{Word: "applx", Distance: 1},
		{Word: "apply", Distance: 1},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Query = %v, want %v", got, want)
	}
}
