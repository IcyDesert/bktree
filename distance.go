package bktree

// Levenshtein computes the edit distance between two strings
// using the Wagner-Fischer algorithm with O(min(len(a),len(b))) space.
// The distance is calculated over Unicode code points (runes), not bytes.
func Levenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)

	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}

	// Ensure rb is the shorter rune slice to minimize space.
	if len(rb) > len(ra) {
		ra, rb = rb, ra
	}

	prev := make([]int, len(rb)+1)
	curr := make([]int, len(rb)+1)

	for j := 0; j <= len(rb); j++ {
		prev[j] = j
	}

	for i := 1; i <= len(ra); i++ {
		curr[0] = i
		rai := ra[i-1]
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if rai == rb[j-1] {
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

	return prev[len(rb)]
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
