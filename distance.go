package bktree

// Levenshtein computes the edit distance between two strings
// using the Wagner-Fischer algorithm with O(min(len(a),len(b))) space.
func Levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Ensure b is the shorter string to minimize space.
	if len(b) > len(a) {
		a, b = b, a
	}

	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)

	for j := 0; j <= len(b); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr[0] = i
		ai := a[i-1]
		for j := 1; j <= len(b); j++ {
			cost := 1
			if ai == b[j-1] {
				cost = 0
			}
			curr[j] = min(
				curr[j-1]+1,    // insertion
				prev[j]+1,      // deletion
				prev[j-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}

	return prev[len(b)]
}

// Hamming computes the Hamming distance between two equal-length strings.
// Panics if the strings have different lengths.
func Hamming(a, b string) int {
	if len(a) != len(b) {
		panic("bktree: Hamming distance requires equal-length strings")
	}
	dist := 0
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			dist++
		}
	}
	return dist
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
