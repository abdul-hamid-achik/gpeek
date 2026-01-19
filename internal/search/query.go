package search

import (
	"strings"
)

// Query represents a parsed search query with options
type Query struct {
	Text          string
	Mode          MatchMode
	CaseSensitive bool
	SmartCase     bool
	Negated       bool // Exclude matches
}

// QueryOptions configures query parsing behavior
type QueryOptions struct {
	DefaultMode      MatchMode
	DefaultCaseSens  bool
	DefaultSmartCase bool
}

// DefaultQueryOptions returns sensible defaults
func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		DefaultMode:      ModeFuzzy,
		DefaultCaseSens:  false,
		DefaultSmartCase: true,
	}
}

// ParseQuery parses a search string into a Query
// Supports prefixes:
//   - ! prefix for negation (exclude matches)
//   - /pattern/ for regex mode
//   - "exact" for exact mode (quoted)
func ParseQuery(input string, opts QueryOptions) Query {
	q := Query{
		Mode:          opts.DefaultMode,
		CaseSensitive: opts.DefaultCaseSens,
		SmartCase:     opts.DefaultSmartCase,
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return q
	}

	// Check for negation prefix
	if strings.HasPrefix(input, "!") {
		q.Negated = true
		input = strings.TrimPrefix(input, "!")
		input = strings.TrimSpace(input)
	}

	// Check for regex mode: /pattern/
	if len(input) >= 2 && input[0] == '/' && input[len(input)-1] == '/' {
		q.Mode = ModeRegex
		q.Text = input[1 : len(input)-1]
		return q
	}

	// Check for exact mode: "pattern"
	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
		q.Mode = ModeExact
		q.Text = input[1 : len(input)-1]
		return q
	}

	q.Text = input
	return q
}

// CreateMatcher creates a Matcher from a Query
func (q Query) CreateMatcher() *Matcher {
	return NewMatcher(q.Text, q.Mode, q.CaseSensitive, q.SmartCase)
}

// Filter filters a slice of strings based on the query
func Filter[T any](items []T, query Query, getText func(T) string) []T {
	if query.Text == "" {
		return items
	}

	matcher := query.CreateMatcher()
	var results []T

	for _, item := range items {
		text := getText(item)
		result := matcher.Match(text)

		// Handle negation
		if query.Negated {
			if !result.Matched {
				results = append(results, item)
			}
		} else {
			if result.Matched {
				results = append(results, item)
			}
		}
	}

	return results
}

// FilterWithScore filters and returns items with their match scores for sorting
type ScoredItem[T any] struct {
	Item  T
	Score int
}

func FilterWithScore[T any](items []T, query Query, getText func(T) string) []ScoredItem[T] {
	if query.Text == "" {
		results := make([]ScoredItem[T], len(items))
		for i, item := range items {
			results[i] = ScoredItem[T]{Item: item, Score: 0}
		}
		return results
	}

	matcher := query.CreateMatcher()
	var results []ScoredItem[T]

	for _, item := range items {
		text := getText(item)
		result := matcher.Match(text)

		if query.Negated {
			if !result.Matched {
				results = append(results, ScoredItem[T]{Item: item, Score: 0})
			}
		} else {
			if result.Matched {
				results = append(results, ScoredItem[T]{Item: item, Score: result.Score})
			}
		}
	}

	return results
}

// MultiFieldQuery matches against multiple fields
type MultiFieldMatcher struct {
	matchers []*Matcher
	query    Query
}

// NewMultiFieldMatcher creates a matcher that can match against multiple fields
func NewMultiFieldMatcher(query Query) *MultiFieldMatcher {
	return &MultiFieldMatcher{
		matchers: []*Matcher{query.CreateMatcher()},
		query:    query,
	}
}

// MatchAny returns true if any field matches
func (m *MultiFieldMatcher) MatchAny(fields ...string) Result {
	var bestResult Result

	for _, field := range fields {
		for _, matcher := range m.matchers {
			result := matcher.Match(field)
			if result.Matched {
				if !bestResult.Matched || result.Score > bestResult.Score {
					bestResult = result
				}
			}
		}
	}

	// Handle negation at the multi-field level
	if m.query.Negated {
		bestResult.Matched = !bestResult.Matched
	}

	return bestResult
}
