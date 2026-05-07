package bktree

import (
	"bytes"
	"fmt"
	"os"
	"reflect"
	"strings"
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
		{"hello", "你好", 5},  // completely different, 5 runes vs 2 runes
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

func TestBKTreeRoundTrip(t *testing.T) {
	original := New(Levenshtein)
	words := []string{"book", "books", "cake", "boo", "cape", "cart"}
	for _, w := range words {
		original.Add(w)
	}

	var buf bytes.Buffer
	if err := original.Save(&buf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(&buf, Levenshtein)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	queries := []struct {
		word    string
		maxDist int
	}{
		{"boak", 1},
		{"boak", 2},
		{"xyz", 3},
	}

	for _, q := range queries {
		want := original.Query(q.word, q.maxDist)
		got := loaded.Query(q.word, q.maxDist)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Query(%q, %d): got %v, want %v", q.word, q.maxDist, got, want)
		}
	}
}

func TestBKTreeEmptyRoundTrip(t *testing.T) {
	original := New(Levenshtein)

	var buf bytes.Buffer
	if err := original.Save(&buf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(&buf, Levenshtein)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Query("anything", 5) != nil {
		t.Error("Empty tree Query should return nil after round-trip")
	}
	if loaded.Exists("anything", 5) {
		t.Error("Empty tree Exists should return false after round-trip")
	}
}

func TestForestRoundTrip(t *testing.T) {
	original := NewForest(Levenshtein)
	words := []string{"book", "books", "cake", "boo", "cape"}
	for _, w := range words {
		original.Add(w)
	}

	var buf bytes.Buffer
	if err := original.Save(&buf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadForest(&buf, Levenshtein)
	if err != nil {
		t.Fatalf("LoadForest failed: %v", err)
	}

	queries := []struct {
		word    string
		maxDist int
	}{
		{"boak", 2},
		{"hello", 1},
		{"cake", 0},
	}

	for _, q := range queries {
		want := original.Query(q.word, q.maxDist)
		got := loaded.Query(q.word, q.maxDist)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Query(%q, %d): got %v, want %v", q.word, q.maxDist, got, want)
		}
	}
}

func TestForestEmptyRoundTrip(t *testing.T) {
	original := NewForest(Levenshtein)

	var buf bytes.Buffer
	if err := original.Save(&buf); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadForest(&buf, Levenshtein)
	if err != nil {
		t.Fatalf("LoadForest failed: %v", err)
	}

	if loaded.Query("anything", 5) != nil {
		t.Error("Empty forest Query should return nil after round-trip")
	}
	if loaded.Exists("anything", 5) {
		t.Error("Empty forest Exists should return false after round-trip")
	}
}

func TestMetricName(t *testing.T) {
	if got := metricName(Levenshtein); got != "Levenshtein" {
		t.Errorf("metricName(Levenshtein) = %q, want %q", got, "Levenshtein")
	}
	if got := metricName(Hamming); got != "Hamming" {
		t.Errorf("metricName(Hamming) = %q, want %q", got, "Hamming")
	}
	if got := metricName(nil); got != "nil" {
		t.Errorf("metricName(nil) = %q, want %q", got, "nil")
	}
}

func TestDefaultFilename(t *testing.T) {
	got := DefaultFilename("data.gob", Levenshtein)
	want := "data_levenshtein.gob"
	if got != want {
		t.Errorf("DefaultFilename(data.gob, Levenshtein) = %q, want %q", got, want)
	}

	got = DefaultFilename("/path/to/index", Hamming)
	want = "/path/to/index_hamming"
	if got != want {
		t.Errorf("DefaultFilename(/path/to/index, Hamming) = %q, want %q", got, want)
	}
}

func TestBKTreeFileRoundTrip(t *testing.T) {
	original := New(Levenshtein)
	words := []string{"book", "books", "cake", "boo", "cape", "cart"}
	for _, w := range words {
		original.Add(w)
	}

	f, err := os.CreateTemp("", "bktree_*.gob")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	defer os.Remove(f.Name())

	if err := original.Save(f); err != nil {
		f.Close()
		t.Fatalf("Save failed: %v", err)
	}
	f.Close()

	f, err = os.Open(f.Name())
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	loaded, err := Load(f, Levenshtein)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	queries := []struct {
		word    string
		maxDist int
	}{
		{"boak", 1},
		{"boak", 2},
		{"xyz", 3},
	}

	for _, q := range queries {
		want := original.Query(q.word, q.maxDist)
		got := loaded.Query(q.word, q.maxDist)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Query(%q, %d): got %v, want %v", q.word, q.maxDist, got, want)
		}
	}
}

func TestForestFileRoundTrip(t *testing.T) {
	original := NewForest(Levenshtein)
	words := []string{"book", "books", "cake", "boo", "cape"}
	for _, w := range words {
		original.Add(w)
	}

	f, err := os.CreateTemp("", "forest_*.gob")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	defer os.Remove(f.Name())

	if err := original.Save(f); err != nil {
		f.Close()
		t.Fatalf("Save failed: %v", err)
	}
	f.Close()

	f, err = os.Open(f.Name())
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	defer f.Close()

	loaded, err := LoadForest(f, Levenshtein)
	if err != nil {
		t.Fatalf("LoadForest failed: %v", err)
	}

	queries := []struct {
		word    string
		maxDist int
	}{
		{"boak", 2},
		{"hello", 1},
		{"cake", 0},
	}

	for _, q := range queries {
		want := original.Query(q.word, q.maxDist)
		got := loaded.Query(q.word, q.maxDist)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("Query(%q, %d): got %v, want %v", q.word, q.maxDist, got, want)
		}
	}
}

func ExampleForest_unsuitableDistance() {
	// A distance function where matches may have different lengths.
	// Forest partitions by length, so it can skip results that span buckets.
	prefixDist := func(a, b string) int {
		if strings.HasPrefix(a, b) || strings.HasPrefix(b, a) {
			return 0
		}
		return 1
	}

	forest := NewForest(prefixDist)
	forest.Add("golang") // length 6

	// "go" is a prefix of "golang" (distance 0), but Forest only
	// checks the length-2 bucket, which is empty.
	results := forest.Query("go", 0)
	fmt.Printf("Forest found %d results\n", len(results))

	// BKTree does not partition by length, so it finds the match.
	tree := New(prefixDist)
	tree.Add("golang")
	results = tree.Query("go", 0)
	fmt.Printf("BKTree found %d results\n", len(results))
	// Output:
	// Forest found 0 results
	// BKTree found 1 results
}

func ExampleForest_unicode() {
	// Levenshtein operates on runes, but Forest groups by byte length.
	// For ASCII text the two are the same; for Unicode they are not.
	forest := NewForest(Levenshtein)
	forest.Add("你好") // 2 runes, 6 bytes → bucket 6

	// "你" (1 rune, 3 bytes) and "你好" (2 runes, 6 bytes) have
	// Levenshtein distance 1. Forest queries bucket [2, 4] and
	// skips bucket 6, missing the result.
	results := forest.Query("你", 1)
	fmt.Printf("Forest found %d results\n", len(results))

	// BKTree does not partition by length, so it finds the match.
	tree := New(Levenshtein)
	tree.Add("你好")
	results = tree.Query("你", 1)
	fmt.Printf("BKTree found %d results\n", len(results))
	// Output:
	// Forest found 0 results
	// BKTree found 1 results
}
