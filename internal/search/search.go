package search

import (
	"regexp"
	"strings"
	"unicode"
)

// MatchMode defines the type of matching algorithm to use
type MatchMode int

const (
	ModeExact MatchMode = iota
	ModeFuzzy
	ModeRegex
)

// Match represents a single match result with position information
type Match struct {
	Start int // Start index of match in the string
	End   int // End index (exclusive)
}

// Result represents the result of matching against a string
type Result struct {
	Matched  bool
	Score    int     // Higher is better (for fuzzy matching)
	Matches  []Match // All match positions for highlighting
}

// Matcher performs string matching with different algorithms
type Matcher struct {
	pattern       string
	lowerPattern  string
	mode          MatchMode
	caseSensitive bool
	smartCase     bool
	regex         *regexp.Regexp
	regexErr      error
}

// NewMatcher creates a new matcher with the given options
func NewMatcher(pattern string, mode MatchMode, caseSensitive, smartCase bool) *Matcher {
	m := &Matcher{
		pattern:       pattern,
		lowerPattern:  strings.ToLower(pattern),
		mode:          mode,
		caseSensitive: caseSensitive,
		smartCase:     smartCase,
	}

	// Smart case: if pattern contains uppercase, become case sensitive
	if smartCase && !caseSensitive {
		for _, r := range pattern {
			if unicode.IsUpper(r) {
				m.caseSensitive = true
				break
			}
		}
	}

	// Compile regex if in regex mode
	if mode == ModeRegex {
		flags := ""
		if !m.caseSensitive {
			flags = "(?i)"
		}
		m.regex, m.regexErr = regexp.Compile(flags + pattern)
	}

	return m
}

// Error returns any regex compilation error
func (m *Matcher) Error() error {
	return m.regexErr
}

// Match tests if the input string matches the pattern
func (m *Matcher) Match(input string) Result {
	if m.pattern == "" {
		return Result{Matched: true, Score: 0}
	}

	switch m.mode {
	case ModeExact:
		return m.matchExact(input)
	case ModeFuzzy:
		return m.matchFuzzy(input)
	case ModeRegex:
		return m.matchRegex(input)
	default:
		return Result{Matched: false}
	}
}

func (m *Matcher) matchExact(input string) Result {
	var searchIn, searchFor string

	if m.caseSensitive {
		searchIn = input
		searchFor = m.pattern
	} else {
		searchIn = strings.ToLower(input)
		searchFor = m.lowerPattern
	}

	idx := strings.Index(searchIn, searchFor)
	if idx == -1 {
		return Result{Matched: false}
	}

	// Find all matches
	var matches []Match
	offset := 0
	for {
		idx := strings.Index(searchIn[offset:], searchFor)
		if idx == -1 {
			break
		}
		matches = append(matches, Match{
			Start: offset + idx,
			End:   offset + idx + len(searchFor),
		})
		offset += idx + 1
	}

	return Result{
		Matched: true,
		Score:   100, // Exact match gets high score
		Matches: matches,
	}
}

func (m *Matcher) matchFuzzy(input string) Result {
	var searchIn string
	var searchFor string

	if m.caseSensitive {
		searchIn = input
		searchFor = m.pattern
	} else {
		searchIn = strings.ToLower(input)
		searchFor = m.lowerPattern
	}

	if len(searchFor) == 0 {
		return Result{Matched: true, Score: 0}
	}

	if len(searchFor) > len(searchIn) {
		return Result{Matched: false}
	}

	// Fuzzy matching: all pattern characters must appear in order
	var matches []Match
	patternIdx := 0
	score := 0
	prevMatchIdx := -1
	consecutive := 0

	for i := 0; i < len(searchIn) && patternIdx < len(searchFor); i++ {
		if searchIn[i] == searchFor[patternIdx] {
			matches = append(matches, Match{Start: i, End: i + 1})

			// Bonus for consecutive matches
			if prevMatchIdx == i-1 {
				consecutive++
				score += consecutive * 10
			} else {
				consecutive = 0
			}

			// Bonus for matching at start
			if i == 0 {
				score += 25
			}

			// Bonus for matching after separator (/, -, _, space)
			if i > 0 {
				prev := searchIn[i-1]
				if prev == '/' || prev == '-' || prev == '_' || prev == ' ' {
					score += 20
				}
			}

			score += 10 // Base score per match
			prevMatchIdx = i
			patternIdx++
		}
	}

	if patternIdx < len(searchFor) {
		return Result{Matched: false}
	}

	// Penalty for length difference
	lengthPenalty := (len(searchIn) - len(searchFor)) * 2
	score -= lengthPenalty

	return Result{
		Matched: true,
		Score:   score,
		Matches: matches,
	}
}

func (m *Matcher) matchRegex(input string) Result {
	if m.regex == nil {
		return Result{Matched: false}
	}

	locs := m.regex.FindAllStringIndex(input, -1)
	if locs == nil {
		return Result{Matched: false}
	}

	var matches []Match
	for _, loc := range locs {
		matches = append(matches, Match{Start: loc[0], End: loc[1]})
	}

	return Result{
		Matched: true,
		Score:   100,
		Matches: matches,
	}
}

// MatchString is a convenience function for simple matching
func MatchString(pattern, input string, mode MatchMode) bool {
	m := NewMatcher(pattern, mode, false, true)
	return m.Match(input).Matched
}

// FuzzyMatch is a convenience function for fuzzy matching
func FuzzyMatch(pattern, input string) Result {
	m := NewMatcher(pattern, ModeFuzzy, false, true)
	return m.Match(input)
}

// ExactMatch is a convenience function for exact matching
func ExactMatch(pattern, input string, caseSensitive bool) Result {
	m := NewMatcher(pattern, ModeExact, caseSensitive, false)
	return m.Match(input)
}
