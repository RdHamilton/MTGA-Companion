package fuzzy

import (
	"sort"
	"strings"
)

// SearchResult represents a fuzzy search match with its score.
type SearchResult struct {
	Item  interface{}
	Score int
	Index int
}

// SearchOptions configures fuzzy search behavior.
type SearchOptions struct {
	// CaseSensitive enables case-sensitive matching
	CaseSensitive bool
	// MaxResults limits the number of results returned (0 = unlimited)
	MaxResults int
	// MinScore sets minimum score threshold (0-100)
	MinScore int
}

// DefaultSearchOptions returns sensible default search options.
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		CaseSensitive: false,
		MaxResults:    100,
		MinScore:      30, // 30% similarity threshold
	}
}

// Search performs fuzzy search on a list of strings.
// Returns results sorted by score (highest first).
func Search(query string, items []string, options SearchOptions) []SearchResult {
	if !options.CaseSensitive {
		query = strings.ToLower(query)
	}

	results := make([]SearchResult, 0, len(items))

	for i, item := range items {
		compareItem := item
		if !options.CaseSensitive {
			compareItem = strings.ToLower(item)
		}

		// Calculate similarity score
		score := calculateScore(query, compareItem)

		// Filter by minimum score
		if score >= options.MinScore {
			results = append(results, SearchResult{
				Item:  item,
				Score: score,
				Index: i,
			})
		}
	}

	// Sort by score descending, then by original index ascending
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Index < results[j].Index
	})

	// Limit results if specified
	if options.MaxResults > 0 && len(results) > options.MaxResults {
		results = results[:options.MaxResults]
	}

	return results
}

// calculateScore calculates a similarity score between query and target (0-100).
// Uses a combination of exact match, substring match, and Levenshtein distance.
func calculateScore(query, target string) int {
	if query == target {
		return 100 // Exact match
	}

	if len(query) == 0 || len(target) == 0 {
		return 0
	}

	// Check for substring match
	if strings.Contains(target, query) {
		// Score based on how much of the target is matched
		return 80 + (len(query) * 20 / len(target))
	}

	// Check for prefix match
	if strings.HasPrefix(target, query) {
		return 85
	}

	// Use Levenshtein distance for fuzzy matching
	distance := levenshteinDistance(query, target)
	maxLen := max(len(query), len(target))

	// Convert distance to similarity percentage
	similarity := 100 - (distance * 100 / maxLen)

	return similarity
}

// levenshteinDistance calculates the Levenshtein distance between two strings.
// This represents the minimum number of single-character edits required to
// change one string into the other.
func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create distance matrix
	d := make([][]int, len(s1)+1)
	for i := range d {
		d[i] = make([]int, len(s2)+1)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		d[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		d[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			d[i][j] = min(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}

	return d[len(s1)][len(s2)]
}

// min returns the minimum of three integers.
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

// max returns the maximum of two integers.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
