# bktree

A Go implementation of the [Burkhard-Keller Tree](https://en.wikipedia.org/wiki/BK-tree) (BK-Tree) for fast approximate string matching in metric spaces.

## Features

- **Generic distance function** — plug in any metric (Levenshtein, Hamming, or your own)
- **Single tree + Forest** — `BKTree` for general use; `Forest` partitions by string length for natural Hamming support and Levenshtein length-based pruning
- **Exists shortcut** — `Exists(word, maxDist)` stops at the first match, faster than `Query` when you only need a boolean
- **Sorted results** — `Query` returns results ordered by distance (then word)
- **Zero dependencies** — standard library only

## Install

```bash
go get github.com/IcyDesert/bktree
```

## Quick Start

### Single Tree with Levenshtein

```go
package main

import (
    "fmt"
    "github.com/IcyDesert/bktree"
)

func main() {
    tree := bktree.New(bktree.Levenshtein)

    words := []string{"book", "books", "cake", "boo", "cape"}
    for _, w := range words {
        tree.Add(w)
    }

    // Find all words within edit distance 2 of "boak"
    results := tree.Query("boak", 2)
    for _, r := range results {
        fmt.Printf("%s (distance: %d)\n", r.Word, r.Distance)
    }
    // Output:
    // book (distance: 1)
    // boo (distance: 2)
    // books (distance: 2)

    // Just check if anything is close enough
    if tree.Exists("hallo", 1) {
        fmt.Println("found a near-match")
    }
}
```

### Forest with Hamming

`Forest` groups words by length into separate trees. This makes Hamming distance trivial — only the same-length tree is ever queried.

```go
forest := bktree.NewForest(bktree.Hamming)

words := []string{"000", "001", "010", "111"}
for _, w := range words {
    forest.Add(w)
}

results := forest.Query("001", 1)
for _, r := range results {
    fmt.Printf("%s (hamming: %d)\n", r.Word, r.Distance)
}
// Output:
// 001 (hamming: 0)
// 000 (hamming: 1)
```

### Custom Distance Function

```go
// Jaccard distance on character sets (example)
func jaccard(a, b string) int {
    // ... compute distance ...
    return distance
}

tree := bktree.New(jaccard)
tree.Add("foo")
tree.Add("bar")

results := tree.Query("baz", 2)
```

## How Forest Pruning Works

For Levenshtein and other reasonable metrics, if `dist(a, b) <= d` then `|len(a) - len(b)| <= d`. `Forest.Query` exploits this to skip entire length-buckets, reducing the search space.

For Hamming, the same logic collapses to checking only the exact-length bucket.
