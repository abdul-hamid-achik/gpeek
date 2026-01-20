package search

import (
	"testing"
)

func TestMatcher_ExactMatch(t *testing.T) {
	tests := []struct {
		name          string
		pattern       string
		input         string
		caseSensitive bool
		wantMatched   bool
	}{
		{"simple match", "foo", "foobar", false, true},
		{"no match", "baz", "foobar", false, false},
		{"case insensitive", "FOO", "foobar", false, true},
		{"case sensitive no match", "FOO", "foobar", true, false},
		{"case sensitive match", "foo", "foobar", true, true},
		{"multiple matches", "a", "banana", false, true},
		{"empty pattern", "", "foobar", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMatcher(tt.pattern, ModeExact, tt.caseSensitive, false)
			result := m.Match(tt.input)
			if result.Matched != tt.wantMatched {
				t.Errorf("Match() = %v, want %v", result.Matched, tt.wantMatched)
			}
		})
	}
}

func TestMatcher_FuzzyMatch(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		input       string
		wantMatched bool
	}{
		{"sequential chars", "fb", "foobar", true},
		{"all chars", "foobar", "foobar", true},
		{"spaced chars", "far", "foobar", true},
		{"missing char", "xyz", "foobar", false},
		{"out of order", "bf", "foobar", false},
		{"empty pattern", "", "foobar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMatcher(tt.pattern, ModeFuzzy, false, false)
			result := m.Match(tt.input)
			if result.Matched != tt.wantMatched {
				t.Errorf("Match() = %v, want %v", result.Matched, tt.wantMatched)
			}
		})
	}
}

func TestMatcher_RegexMatch(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		input       string
		wantMatched bool
		wantError   bool
	}{
		{"simple regex", "foo.*", "foobar", true, false},
		{"no match", "^baz", "foobar", false, false},
		{"word boundary", "\\bfoo\\b", "foo bar", true, false},
		{"invalid regex", "[", "foobar", false, true},
		{"empty pattern", "", "foobar", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMatcher(tt.pattern, ModeRegex, false, false)
			if tt.wantError {
				if m.Error() == nil {
					t.Error("expected regex error, got nil")
				}
				return
			}
			result := m.Match(tt.input)
			if result.Matched != tt.wantMatched {
				t.Errorf("Match() = %v, want %v", result.Matched, tt.wantMatched)
			}
		})
	}
}

func TestMatcher_SmartCase(t *testing.T) {
	// Smart case: becomes case sensitive if pattern has uppercase
	tests := []struct {
		name        string
		pattern     string
		input       string
		wantMatched bool
	}{
		{"lowercase pattern matches any case", "foo", "FooBar", true},
		{"uppercase pattern is case sensitive", "Foo", "foobar", false},
		{"uppercase pattern matches", "Foo", "FooBar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMatcher(tt.pattern, ModeExact, false, true)
			result := m.Match(tt.input)
			if result.Matched != tt.wantMatched {
				t.Errorf("Match() = %v, want %v", result.Matched, tt.wantMatched)
			}
		})
	}
}

func TestMatcher_MatchPositions(t *testing.T) {
	m := NewMatcher("foo", ModeExact, false, false)
	result := m.Match("foo bar foo")

	if !result.Matched {
		t.Error("expected match")
	}
	if len(result.Matches) != 2 {
		t.Errorf("expected 2 match positions, got %d", len(result.Matches))
	}
	if result.Matches[0].Start != 0 || result.Matches[0].End != 3 {
		t.Errorf("unexpected first match position: %+v", result.Matches[0])
	}
}

func TestQuery_RegexError(t *testing.T) {
	// Test that RegexError returns error for invalid regex
	q := ParseQuery("/[/", QueryOptions{DefaultMode: ModeRegex})
	err := q.RegexError()
	if err == nil {
		t.Error("expected regex error for invalid pattern")
	}

	// Test that RegexError returns nil for valid regex
	q2 := ParseQuery("/foo.*/", QueryOptions{DefaultMode: ModeRegex})
	err2 := q2.RegexError()
	if err2 != nil {
		t.Errorf("unexpected error for valid regex: %v", err2)
	}

	// Test that non-regex queries return nil
	q3 := ParseQuery("foo", QueryOptions{DefaultMode: ModeFuzzy})
	err3 := q3.RegexError()
	if err3 != nil {
		t.Errorf("unexpected error for non-regex query: %v", err3)
	}
}

func TestFilter(t *testing.T) {
	items := []string{"apple", "banana", "cherry", "apricot"}
	query := ParseQuery("ap", DefaultQueryOptions())

	result := Filter(items, query, func(s string) string { return s })

	if len(result) != 2 {
		t.Errorf("expected 2 results, got %d", len(result))
	}
}

func TestFilter_Negation(t *testing.T) {
	items := []string{"apple", "banana", "cherry"}
	query := ParseQuery("!a", DefaultQueryOptions())

	result := Filter(items, query, func(s string) string { return s })

	// Should exclude "apple" and "banana" which contain 'a'
	if len(result) != 1 {
		t.Errorf("expected 1 result, got %d", len(result))
	}
	if result[0] != "cherry" {
		t.Errorf("expected 'cherry', got '%s'", result[0])
	}
}
